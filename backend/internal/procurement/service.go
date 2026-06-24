package procurement

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type PO struct {
	ID            uuid.UUID  `json:"id"`
	SupplierID    uuid.UUID  `json:"supplier_id"`
	SupplierName  string     `json:"supplier_name"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	WarehouseName string     `json:"warehouse_name"`
	PONumber      string     `json:"po_number"`
	Status        string     `json:"status"`
	TotalAmount   float64    `json:"total_amount"`
	OrderedAt     time.Time  `json:"ordered_at"`
	ExpectedAt    *time.Time `json:"expected_at,omitempty"`
	ItemCount     int        `json:"item_count"`
}

type POItem struct {
	ID           uuid.UUID `json:"id"`
	MaterialID   uuid.UUID `json:"material_id"`
	MaterialName string    `json:"material_name"`
	MaterialSKU  string    `json:"material_sku"`
	QtyOrdered   float64   `json:"qty_ordered"`
	QtyReceived  float64   `json:"qty_received"`
	UnitPrice    float64   `json:"unit_price"`
}

type PODetail struct {
	PO
	Items []POItem `json:"items"`
}

type CreateInput struct {
	SupplierID  uuid.UUID   `json:"supplier_id"`
	WarehouseID uuid.UUID   `json:"warehouse_id"`
	PONumber    string      `json:"po_number"`
	ExpectedAt  *time.Time  `json:"expected_at"`
	Items       []ItemInput `json:"items"`
}

type ItemInput struct {
	MaterialID uuid.UUID `json:"material_id"`
	QtyOrdered float64   `json:"qty_ordered"`
	UnitPrice  float64   `json:"unit_price"`
}

type ReceiveInput struct {
	Note  string             `json:"note"`
	Items []ReceiveItemInput `json:"items"`
}

type ReceiveItemInput struct {
	POItemID    uuid.UUID  `json:"po_item_id"`
	QtyReceived float64    `json:"qty_received"`
	BatchNo     string     `json:"batch_no"`
	ExpiryAt    *time.Time `json:"expiry_at"`
}

func List(ctx context.Context, workspaceID uuid.UUID) ([]PO, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT po.id, po.supplier_id, s.name, po.warehouse_id, w.name,
		       po.po_number, po.status::text, po.total_amount,
		       po.ordered_at, po.expected_at,
		       (SELECT COUNT(*) FROM po_items pi WHERE pi.po_id = po.id)
		FROM purchase_orders po
		JOIN suppliers s  ON s.id  = po.supplier_id
		JOIN warehouses w ON w.id  = po.warehouse_id
		WHERE po.workspace_id = $1
		ORDER BY po.created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PO
	for rows.Next() {
		var p PO
		if err := rows.Scan(
			&p.ID, &p.SupplierID, &p.SupplierName, &p.WarehouseID, &p.WarehouseName,
			&p.PONumber, &p.Status, &p.TotalAmount,
			&p.OrderedAt, &p.ExpectedAt, &p.ItemCount,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if list == nil {
		list = []PO{}
	}
	return list, nil
}

// getPO fetches a single PO summary row (joined supplier/warehouse + item count).
// Avoids re-listing every PO just to return one after a status change.
func getPO(ctx context.Context, workspaceID, id uuid.UUID) (*PO, error) {
	var p PO
	err := db.Pool.QueryRow(ctx, `
		SELECT po.id, po.supplier_id, s.name, po.warehouse_id, w.name,
		       po.po_number, po.status::text, po.total_amount,
		       po.ordered_at, po.expected_at,
		       (SELECT COUNT(*) FROM po_items pi WHERE pi.po_id = po.id)
		FROM purchase_orders po
		JOIN suppliers s  ON s.id = po.supplier_id
		JOIN warehouses w ON w.id = po.warehouse_id
		WHERE po.workspace_id = $1 AND po.id = $2`,
		workspaceID, id,
	).Scan(
		&p.ID, &p.SupplierID, &p.SupplierName, &p.WarehouseID, &p.WarehouseName,
		&p.PONumber, &p.Status, &p.TotalAmount,
		&p.OrderedAt, &p.ExpectedAt, &p.ItemCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get po: %w", err)
	}
	return &p, nil
}

func GetDetail(ctx context.Context, workspaceID, id uuid.UUID) (*PODetail, error) {
	var d PODetail
	err := db.Pool.QueryRow(ctx, `
		SELECT po.id, po.supplier_id, s.name, po.warehouse_id, w.name,
		       po.po_number, po.status::text, po.total_amount,
		       po.ordered_at, po.expected_at, 0
		FROM purchase_orders po
		JOIN suppliers s  ON s.id = po.supplier_id
		JOIN warehouses w ON w.id = po.warehouse_id
		WHERE po.workspace_id=$1 AND po.id=$2`,
		workspaceID, id,
	).Scan(
		&d.ID, &d.SupplierID, &d.SupplierName, &d.WarehouseID, &d.WarehouseName,
		&d.PONumber, &d.Status, &d.TotalAmount,
		&d.OrderedAt, &d.ExpectedAt, &d.ItemCount,
	)
	if err != nil {
		return nil, fmt.Errorf("po not found")
	}

	rows, err := db.Pool.Query(ctx, `
		SELECT pi.id, pi.material_id, m.name, m.sku,
		       pi.qty_ordered, pi.qty_received, pi.unit_price
		FROM po_items pi
		JOIN materials m ON m.id = pi.material_id
		WHERE pi.po_id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it POItem
		if err := rows.Scan(&it.ID, &it.MaterialID, &it.MaterialName, &it.MaterialSKU,
			&it.QtyOrdered, &it.QtyReceived, &it.UnitPrice); err != nil {
			return nil, err
		}
		d.Items = append(d.Items, it)
	}
	d.ItemCount = len(d.Items)
	return &d, nil
}

func Create(ctx context.Context, workspaceID, userID uuid.UUID, in CreateInput) (*PODetail, error) {
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("at least one item required")
	}
	poNum := in.PONumber
	if poNum == "" {
		poNum = fmt.Sprintf("PO-%d", time.Now().UnixMilli())
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var poID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO purchase_orders
			(workspace_id, supplier_id, warehouse_id, po_number, expected_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id`,
		workspaceID, in.SupplierID, in.WarehouseID, poNum, in.ExpectedAt,
	).Scan(&poID)
	if err != nil {
		return nil, fmt.Errorf("create po: %w", err)
	}

	total := 0.0
	for _, it := range in.Items {
		if it.QtyOrdered <= 0 {
			return nil, fmt.Errorf("qty_ordered must be > 0")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO po_items (po_id, material_id, qty_ordered, unit_price)
			VALUES ($1,$2,$3,$4)`,
			poID, it.MaterialID, it.QtyOrdered, it.UnitPrice,
		); err != nil {
			return nil, fmt.Errorf("insert po_item: %w", err)
		}
		total += it.QtyOrdered * it.UnitPrice
	}

	if _, err = tx.Exec(ctx,
		`UPDATE purchase_orders SET total_amount=$1 WHERE id=$2`, total, poID,
	); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return GetDetail(ctx, workspaceID, poID)
}

func Send(ctx context.Context, workspaceID, id uuid.UUID) (*PO, error) {
	tag, err := db.Pool.Exec(ctx,
		`UPDATE purchase_orders SET status='sent'
		WHERE id=$1 AND workspace_id=$2 AND status='draft'`,
		id, workspaceID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("po not found or not in draft status")
	}
	return getPO(ctx, workspaceID, id)
}

func Cancel(ctx context.Context, workspaceID, id uuid.UUID) (*PO, error) {
	tag, err := db.Pool.Exec(ctx,
		`UPDATE purchase_orders SET status='cancelled'
		WHERE id=$1 AND workspace_id=$2 AND status IN ('draft','sent')`,
		id, workspaceID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("po not found or cannot be cancelled")
	}
	return getPO(ctx, workspaceID, id)
}

// Receive creates a goods receipt, inserts material_batches + stock_movements IN_PURCHASE,
// updates po_items qty_received, and advances PO status.
func Receive(ctx context.Context, workspaceID, userID, poID uuid.UUID, in ReceiveInput) (*PODetail, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lock PO, check receivable status
	var warehouseID uuid.UUID
	var status string
	err = tx.QueryRow(ctx,
		`SELECT warehouse_id, status::text FROM purchase_orders
		WHERE id=$1 AND workspace_id=$2 AND status IN ('sent','partial_received')
		FOR UPDATE`,
		poID, workspaceID,
	).Scan(&warehouseID, &status)
	if err != nil {
		return nil, fmt.Errorf("po not found or not in receivable status")
	}

	// Insert goods_receipt header
	var grID uuid.UUID
	err = tx.QueryRow(ctx,
		`INSERT INTO goods_receipts (po_id, received_by, note) VALUES ($1,$2,$3) RETURNING id`,
		poID, userID, in.Note,
	).Scan(&grID)
	if err != nil {
		return nil, fmt.Errorf("insert goods_receipt: %w", err)
	}

	for _, item := range in.Items {
		if item.QtyReceived <= 0 {
			continue
		}

		// Get material_id + unit_price + qty tracking from po_item
		var materialID uuid.UUID
		var unitPrice, qtyOrdered, qtyReceived float64
		err = tx.QueryRow(ctx,
			`SELECT material_id, unit_price, qty_ordered, qty_received
			 FROM po_items WHERE id=$1 AND po_id=$2 FOR UPDATE`,
			item.POItemID, poID,
		).Scan(&materialID, &unitPrice, &qtyOrdered, &qtyReceived)
		if err != nil {
			return nil, fmt.Errorf("po_item not found: %v", item.POItemID)
		}

		// Guard: cannot receive more than ordered.
		if qtyReceived+item.QtyReceived > qtyOrdered+1e-9 {
			return nil, fmt.Errorf("over-receive on po_item %v: ordered %.4f, already %.4f, attempted %.4f",
				item.POItemID, qtyOrdered, qtyReceived, item.QtyReceived)
		}

		batchNo := item.BatchNo
		if batchNo == "" {
			batchNo = fmt.Sprintf("BATCH-%d", time.Now().UnixMilli())
		}

		// Insert material_batch
		if _, err = tx.Exec(ctx, `
			INSERT INTO material_batches
				(material_id, warehouse_id, batch_no, qty_received, qty_remaining, unit_cost, supplier_id, expiry_at)
			SELECT $1,$2,$3,$4,$4,$5,po.supplier_id,$6
			FROM purchase_orders po WHERE po.id=$7`,
			materialID, warehouseID, batchNo, item.QtyReceived, unitPrice, item.ExpiryAt, poID,
		); err != nil {
			return nil, fmt.Errorf("insert batch: %w", err)
		}

		// Insert stock_movement IN_PURCHASE
		if _, err = tx.Exec(ctx, `
			INSERT INTO stock_movements
				(workspace_id, warehouse_id, item_type, item_id, qty, movement_type, reference_type, reference_id, unit_cost, created_by)
			VALUES ($1,$2,'material',$3,$4,'IN_PURCHASE'::movement_type,'goods_receipt',$5,$6,$7)`,
			workspaceID, warehouseID, materialID, item.QtyReceived, grID, unitPrice, userID,
		); err != nil {
			return nil, fmt.Errorf("insert IN_PURCHASE movement: %w", err)
		}

		// Update po_item qty_received
		if _, err = tx.Exec(ctx,
			`UPDATE po_items SET qty_received = qty_received + $1 WHERE id=$2`,
			item.QtyReceived, item.POItemID,
		); err != nil {
			return nil, err
		}
	}

	// Determine new PO status: received if all items fully received
	var pendingCount int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM po_items WHERE po_id=$1 AND qty_received < qty_ordered`,
		poID,
	).Scan(&pendingCount)
	if err != nil {
		return nil, err
	}
	newStatus := "partial_received"
	if pendingCount == 0 {
		newStatus = "received"
	}
	if _, err = tx.Exec(ctx,
		`UPDATE purchase_orders SET status=$1::po_status WHERE id=$2`, newStatus, poID,
	); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return GetDetail(ctx, workspaceID, poID)
}
