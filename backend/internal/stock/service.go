package stock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type CurrentStockItem struct {
	WarehouseID    uuid.UUID `json:"warehouse_id"`
	WarehouseName  string    `json:"warehouse_name"`
	ItemType       string    `json:"item_type"`
	ItemID         uuid.UUID `json:"item_id"`
	ItemName       string    `json:"item_name"`
	ItemSKU        string    `json:"item_sku"`
	QtyOnHand      float64   `json:"qty_on_hand"`
	UnitCost       float64   `json:"unit_cost"`
	Value          float64   `json:"value"`
	LastMovementAt time.Time `json:"last_movement_at"`
}

type Movement struct {
	ID            uuid.UUID  `json:"id"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	ItemType      string     `json:"item_type"`
	ItemID        uuid.UUID  `json:"item_id"`
	Qty           float64    `json:"qty"`
	MovementType  string     `json:"movement_type"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	UnitCost      float64    `json:"unit_cost"`
	Reason        *string    `json:"reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type TransferInput struct {
	FromWarehouseID uuid.UUID `json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID `json:"to_warehouse_id"`
	ItemType        string    `json:"item_type"`
	ItemID          uuid.UUID `json:"item_id"`
	Qty             float64   `json:"qty"`
	Reason          string    `json:"reason"`
}

type AdjustmentInput struct {
	WarehouseID uuid.UUID `json:"warehouse_id"`
	ItemType    string    `json:"item_type"`
	ItemID      uuid.UUID `json:"item_id"`
	Qty         float64   `json:"qty"`
	Reason      string    `json:"reason"`
}

func CurrentStock(ctx context.Context, workspaceID uuid.UUID) ([]CurrentStockItem, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT
			v.warehouse_id,
			w.name AS warehouse_name,
			v.item_type::text,
			v.item_id,
			COALESCE(
				CASE WHEN v.item_type = 'material'        THEN m.name
				     WHEN v.item_type = 'product_variant' THEN pr.name || ' — ' || pv.sku
				END, ''
			) AS item_name,
			COALESCE(
				CASE WHEN v.item_type = 'material'        THEN m.sku
				     WHEN v.item_type = 'product_variant' THEN pv.sku
				END, ''
			) AS item_sku,
			v.qty_on_hand,
			COALESCE(mc.avg_cost, 0)                  AS unit_cost,
			v.qty_on_hand * COALESCE(mc.avg_cost, 0)  AS value,
			v.last_movement_at
		FROM v_current_stock v
		JOIN warehouses w ON w.id = v.warehouse_id
		LEFT JOIN materials m        ON m.id = v.item_id  AND v.item_type = 'material'
		LEFT JOIN product_variants pv ON pv.id = v.item_id AND v.item_type = 'product_variant'
		LEFT JOIN products pr         ON pr.id = pv.product_id
		LEFT JOIN (
			SELECT material_id,
			       SUM(qty_remaining * unit_cost) / NULLIF(SUM(qty_remaining), 0) AS avg_cost
			FROM material_batches
			WHERE qty_remaining > 0
			GROUP BY material_id
		) mc ON mc.material_id = v.item_id AND v.item_type = 'material'
		WHERE v.workspace_id = $1 AND v.qty_on_hand > 0
		ORDER BY w.name, v.item_type, item_name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CurrentStockItem
	for rows.Next() {
		var it CurrentStockItem
		if err := rows.Scan(
			&it.WarehouseID, &it.WarehouseName,
			&it.ItemType, &it.ItemID,
			&it.ItemName, &it.ItemSKU,
			&it.QtyOnHand, &it.UnitCost, &it.Value, &it.LastMovementAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []CurrentStockItem{}
	}
	return items, nil
}

func ListMovements(ctx context.Context, workspaceID uuid.UUID) ([]Movement, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, warehouse_id, item_type::text, item_id, qty,
		       movement_type::text, reference_type, reference_id,
		       unit_cost, reason, created_at
		FROM stock_movements
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT 500`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var moves []Movement
	for rows.Next() {
		var m Movement
		if err := rows.Scan(
			&m.ID, &m.WarehouseID, &m.ItemType, &m.ItemID, &m.Qty,
			&m.MovementType, &m.ReferenceType, &m.ReferenceID,
			&m.UnitCost, &m.Reason, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		moves = append(moves, m)
	}
	if moves == nil {
		moves = []Movement{}
	}
	return moves, nil
}

func Transfer(ctx context.Context, workspaceID, userID uuid.UUID, in TransferInput) error {
	if in.Qty <= 0 {
		return fmt.Errorf("qty must be > 0")
	}
	if in.FromWarehouseID == in.ToWarehouseID {
		return fmt.Errorf("source and destination warehouse must differ")
	}

	// Check available stock
	var available float64
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(qty_on_hand, 0) FROM v_current_stock
		WHERE workspace_id=$1 AND warehouse_id=$2 AND item_id=$3 AND item_type=$4::item_type`,
		workspaceID, in.FromWarehouseID, in.ItemID, in.ItemType,
	).Scan(&available)
	if err != nil {
		return fmt.Errorf("stock check failed: %w", err)
	}
	if available < in.Qty {
		return fmt.Errorf("insufficient stock: available %.2f, requested %.2f", available, in.Qty)
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ins := `INSERT INTO stock_movements
		(workspace_id, warehouse_id, item_type, item_id, qty, movement_type, reason, created_by)
		VALUES ($1,$2,$3::item_type,$4,$5,$6::movement_type,$7,$8)`

	reason := &in.Reason
	if _, err = tx.Exec(ctx, ins, workspaceID, in.FromWarehouseID, in.ItemType, in.ItemID, -in.Qty, "TRANSFER_OUT", reason, userID); err != nil {
		return fmt.Errorf("insert TRANSFER_OUT: %w", err)
	}
	if _, err = tx.Exec(ctx, ins, workspaceID, in.ToWarehouseID, in.ItemType, in.ItemID, in.Qty, "TRANSFER_IN", reason, userID); err != nil {
		return fmt.Errorf("insert TRANSFER_IN: %w", err)
	}
	return tx.Commit(ctx)
}

func Adjustment(ctx context.Context, workspaceID, userID uuid.UUID, in AdjustmentInput) (*Movement, error) {
	if in.Qty == 0 {
		return nil, fmt.Errorf("qty must not be zero")
	}

	var m Movement
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO stock_movements
		(workspace_id, warehouse_id, item_type, item_id, qty, movement_type, reason, created_by)
		VALUES ($1,$2,$3::item_type,$4,$5,'ADJUSTMENT'::movement_type,$6,$7)
		RETURNING id, warehouse_id, item_type::text, item_id, qty,
		          movement_type::text, reference_type, reference_id, unit_cost, reason, created_at`,
		workspaceID, in.WarehouseID, in.ItemType, in.ItemID, in.Qty, in.Reason, userID,
	).Scan(
		&m.ID, &m.WarehouseID, &m.ItemType, &m.ItemID, &m.Qty,
		&m.MovementType, &m.ReferenceType, &m.ReferenceID, &m.UnitCost, &m.Reason, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("adjustment: %w", err)
	}
	return &m, nil
}
