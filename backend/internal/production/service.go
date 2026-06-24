package production

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type Order struct {
	ID               uuid.UUID  `json:"id"`
	WarehouseID      uuid.UUID  `json:"warehouse_id"`
	WarehouseName    string     `json:"warehouse_name"`
	ProductVariantID uuid.UUID  `json:"product_variant_id"`
	VariantSKU       string     `json:"variant_sku"`
	ProductName      string     `json:"product_name"`
	QtyPlanned       float64    `json:"qty_planned"`
	Status           string     `json:"status"`
	PlannedAt        *time.Time `json:"planned_at,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	TotalCost        float64    `json:"total_cost"`
	CreatedAt        time.Time  `json:"created_at"`
}

type CreateInput struct {
	WarehouseID      uuid.UUID  `json:"warehouse_id"`
	ProductVariantID uuid.UUID  `json:"product_variant_id"`
	QtyPlanned       float64    `json:"qty_planned"`
	PlannedAt        *time.Time `json:"planned_at"`
}

type OutputInput struct {
	QtyGood      float64 `json:"qty_good"`
	QtyReject    float64 `json:"qty_reject"`
	QtyWaste     float64 `json:"qty_waste"`
	RejectReason string  `json:"reject_reason"`
	WasteReason  string  `json:"waste_reason"`
}

const orderCols = `
	po.id, po.warehouse_id, w.name,
	po.product_variant_id, pv.sku, p.name,
	po.qty_planned, po.status::text,
	po.planned_at, po.started_at, po.completed_at,
	po.total_cost, po.created_at`

func scanOrder(row interface {
	Scan(...any) error
}) (*Order, error) {
	var o Order
	err := row.Scan(
		&o.ID, &o.WarehouseID, &o.WarehouseName,
		&o.ProductVariantID, &o.VariantSKU, &o.ProductName,
		&o.QtyPlanned, &o.Status,
		&o.PlannedAt, &o.StartedAt, &o.CompletedAt,
		&o.TotalCost, &o.CreatedAt,
	)
	return &o, err
}

const orderJoins = `
	FROM production_orders po
	JOIN warehouses w        ON w.id  = po.warehouse_id
	JOIN product_variants pv ON pv.id = po.product_variant_id
	JOIN products p          ON p.id  = pv.product_id`

func List(ctx context.Context, workspaceID uuid.UUID) ([]Order, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+orderCols+orderJoins+`
		WHERE po.workspace_id = $1
		ORDER BY po.created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	if orders == nil {
		orders = []Order{}
	}
	return orders, nil
}

func Create(ctx context.Context, workspaceID, userID uuid.UUID, in CreateInput) (*Order, error) {
	if in.QtyPlanned <= 0 {
		return nil, fmt.Errorf("qty_planned must be > 0")
	}
	var id uuid.UUID
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO production_orders
			(workspace_id, warehouse_id, product_variant_id, qty_planned, planned_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id`,
		workspaceID, in.WarehouseID, in.ProductVariantID, in.QtyPlanned, in.PlannedAt, userID,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return GetByID(ctx, workspaceID, id)
}

func GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*Order, error) {
	row := db.Pool.QueryRow(ctx,
		`SELECT `+orderCols+orderJoins+`
		WHERE po.workspace_id = $1 AND po.id = $2`,
		workspaceID, id,
	)
	o, err := scanOrder(row)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

func Cancel(ctx context.Context, workspaceID, id uuid.UUID) (*Order, error) {
	tag, err := db.Pool.Exec(ctx,
		`UPDATE production_orders SET status = 'cancelled'
		WHERE id=$1 AND workspace_id=$2 AND status IN ('draft','in_progress')`,
		id, workspaceID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("order not found or already completed/cancelled")
	}
	return GetByID(ctx, workspaceID, id)
}

// Start transitions draft→in_progress and executes FIFO material deduction.
func Start(ctx context.Context, workspaceID, userID, orderID uuid.UUID) (*Order, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock order, verify status
	var variantID uuid.UUID
	var warehouseID uuid.UUID
	var qtyPlanned float64
	err = tx.QueryRow(ctx,
		`SELECT product_variant_id, warehouse_id, qty_planned
		FROM production_orders
		WHERE id=$1 AND workspace_id=$2 AND status='draft'
		FOR UPDATE`,
		orderID, workspaceID,
	).Scan(&variantID, &warehouseID, &qtyPlanned)
	if err != nil {
		return nil, fmt.Errorf("order not found or not in draft status")
	}

	// Get BOM items
	type bomItem struct {
		materialID uuid.UUID
		qty        float64
	}
	bomRows, err := tx.Query(ctx,
		`SELECT b.material_id, b.qty
		FROM bom_items b
		JOIN product_variants pv ON pv.id = b.product_variant_id
		JOIN products p ON p.id = pv.product_id
		WHERE b.product_variant_id=$1 AND p.workspace_id=$2`,
		variantID, workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer bomRows.Close()

	var bom []bomItem
	for bomRows.Next() {
		var item bomItem
		if err := bomRows.Scan(&item.materialID, &item.qty); err != nil {
			return nil, err
		}
		bom = append(bom, item)
	}
	bomRows.Close()

	totalCost := 0.0

	// FIFO deduction per BOM item
	for _, b := range bom {
		needed := b.qty * qtyPlanned

		batchRows, err := tx.Query(ctx,
			`SELECT id, qty_remaining, unit_cost FROM material_batches
			WHERE material_id=$1 AND warehouse_id=$2 AND qty_remaining > 0
			ORDER BY received_at ASC
			FOR UPDATE SKIP LOCKED`,
			b.materialID, warehouseID,
		)
		if err != nil {
			return nil, err
		}

		var batches []fifoBatch
		for batchRows.Next() {
			var bt fifoBatch
			if err := batchRows.Scan(&bt.id, &bt.qtyRemaining, &bt.unitCost); err != nil {
				batchRows.Close()
				return nil, err
			}
			batches = append(batches, bt)
		}
		batchRows.Close()

		deductions, cost, shortfall := allocateFIFO(batches, needed)
		// Not enough stock across all FIFO batches → abort whole order (rollback via defer).
		if shortfall > 0 {
			return nil, fmt.Errorf("insufficient stock for material %s: short %.4f", b.materialID, shortfall)
		}
		totalCost += cost

		for _, d := range deductions {
			if _, err = tx.Exec(ctx,
				`UPDATE material_batches SET qty_remaining = qty_remaining - $1 WHERE id=$2`,
				d.deduct, d.batchID,
			); err != nil {
				return nil, fmt.Errorf("deduct batch: %w", err)
			}
			if _, err = tx.Exec(ctx,
				`INSERT INTO stock_movements
					(workspace_id, warehouse_id, item_type, item_id, qty, movement_type, reference_type, reference_id, unit_cost, created_by)
				VALUES ($1,$2,'material',$3,$4,'OUT_PRODUCTION'::movement_type,'production_order',$5,$6,$7)`,
				workspaceID, warehouseID, b.materialID, -d.deduct, orderID, d.unitCost, userID,
			); err != nil {
				return nil, fmt.Errorf("insert OUT_PRODUCTION: %w", err)
			}
		}
	}

	if _, err = tx.Exec(ctx,
		`UPDATE production_orders
		SET status='in_progress', started_at=NOW(), total_cost=$1
		WHERE id=$2`,
		totalCost, orderID,
	); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return GetByID(ctx, workspaceID, orderID)
}

// RecordOutput records production output and transitions in_progress→completed.
func RecordOutput(ctx context.Context, workspaceID, userID, orderID uuid.UUID, in OutputInput) (*Order, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var variantID, warehouseID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT product_variant_id, warehouse_id FROM production_orders
		WHERE id=$1 AND workspace_id=$2 AND status='in_progress'
		FOR UPDATE`,
		orderID, workspaceID,
	).Scan(&variantID, &warehouseID)
	if err != nil {
		return nil, fmt.Errorf("order not found or not in_progress")
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO production_outputs
			(production_order_id, qty_good, qty_reject, qty_waste, reject_reason, waste_reason)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		orderID, in.QtyGood, in.QtyReject, in.QtyWaste, in.RejectReason, in.WasteReason,
	); err != nil {
		return nil, fmt.Errorf("insert output: %w", err)
	}

	if in.QtyGood > 0 {
		if _, err = tx.Exec(ctx,
			`INSERT INTO stock_movements
				(workspace_id, warehouse_id, item_type, item_id, qty, movement_type, reference_type, reference_id, created_by)
			VALUES ($1,$2,'product_variant',$3,$4,'IN_PRODUCTION'::movement_type,'production_order',$5,$6)`,
			workspaceID, warehouseID, variantID, in.QtyGood, orderID, userID,
		); err != nil {
			return nil, fmt.Errorf("insert IN_PRODUCTION: %w", err)
		}
	}

	if _, err = tx.Exec(ctx,
		`UPDATE production_orders SET status='completed', completed_at=NOW() WHERE id=$1`,
		orderID,
	); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return GetByID(ctx, workspaceID, orderID)
}
