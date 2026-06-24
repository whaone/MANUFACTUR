package reports

import (
	"context"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type DashboardData struct {
	CompletedOrders  int     `json:"completed_orders"`
	TotalQtyProduced float64 `json:"total_qty_produced"`
	ReceivedPOs      int     `json:"received_pos"`
	StockValue       float64 `json:"stock_value"`
	MaterialCostMonth float64 `json:"material_cost_month"`
}

type HppMarginItem struct {
	VariantID   uuid.UUID `json:"variant_id"`
	VariantSKU  string    `json:"variant_sku"`
	ProductName string    `json:"product_name"`
	QtyProduced float64   `json:"qty_produced"`
	TotalCost   float64   `json:"total_cost"`
	HppPerUnit  float64   `json:"hpp_per_unit"`
	SellPrice   float64   `json:"sell_price"`
	Margin      float64   `json:"margin"`
	MarginPct   float64   `json:"margin_pct"`
}

type TrendItem struct {
	Month       string  `json:"month"`
	QtyProduced float64 `json:"qty_produced"`
	OrderCount  int     `json:"order_count"`
}

func Dashboard(ctx context.Context, workspaceID uuid.UUID) (*DashboardData, error) {
	var d DashboardData

	// Completed orders + qty produced this month
	err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT po.id), COALESCE(SUM(out.qty_good), 0)
		FROM production_orders po
		LEFT JOIN production_outputs out ON out.production_order_id = po.id
		WHERE po.workspace_id = $1 AND po.status = 'completed'
		  AND DATE_TRUNC('month', po.completed_at) = DATE_TRUNC('month', NOW())`,
		workspaceID,
	).Scan(&d.CompletedOrders, &d.TotalQtyProduced)
	if err != nil {
		return nil, err
	}

	// Received POs this month
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM purchase_orders
		WHERE workspace_id = $1 AND status = 'received'
		  AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())`,
		workspaceID,
	).Scan(&d.ReceivedPOs)
	if err != nil {
		return nil, err
	}

	// Total stock value (material batches with remaining qty)
	err = db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(mb.qty_remaining * mb.unit_cost), 0)
		FROM material_batches mb
		JOIN materials m ON m.id = mb.material_id
		WHERE m.workspace_id = $1 AND mb.qty_remaining > 0`,
		workspaceID,
	).Scan(&d.StockValue)
	if err != nil {
		return nil, err
	}

	// Material cost consumed this month (OUT_PRODUCTION)
	err = db.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(ABS(qty) * unit_cost), 0)
		FROM stock_movements
		WHERE workspace_id = $1 AND movement_type = 'OUT_PRODUCTION'
		  AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())`,
		workspaceID,
	).Scan(&d.MaterialCostMonth)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func HppMargin(ctx context.Context, workspaceID uuid.UUID) ([]HppMarginItem, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT
			po.product_variant_id,
			pv.sku,
			p.name,
			COALESCE(SUM(out.qty_good), 0)                   AS qty_produced,
			COALESCE(SUM(po.total_cost), 0)                  AS total_cost,
			CASE WHEN COALESCE(SUM(out.qty_good), 0) > 0
			     THEN SUM(po.total_cost) / SUM(out.qty_good)
			     ELSE 0 END                                   AS hpp_per_unit,
			pv.sell_price
		FROM production_orders po
		JOIN production_outputs out ON out.production_order_id = po.id
		JOIN product_variants pv   ON pv.id = po.product_variant_id
		JOIN products p            ON p.id  = pv.product_id
		WHERE po.workspace_id = $1 AND po.status = 'completed'
		GROUP BY po.product_variant_id, pv.sku, p.name, pv.sell_price
		ORDER BY qty_produced DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HppMarginItem
	for rows.Next() {
		var it HppMarginItem
		if err := rows.Scan(
			&it.VariantID, &it.VariantSKU, &it.ProductName,
			&it.QtyProduced, &it.TotalCost, &it.HppPerUnit, &it.SellPrice,
		); err != nil {
			return nil, err
		}
		it.Margin = it.SellPrice - it.HppPerUnit
		if it.SellPrice > 0 {
			it.MarginPct = it.Margin / it.SellPrice * 100
		}
		items = append(items, it)
	}
	if items == nil {
		items = []HppMarginItem{}
	}
	return items, nil
}

func ProductionTrend(ctx context.Context, workspaceID uuid.UUID) ([]TrendItem, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', po.completed_at), 'YYYY-MM') AS month,
			COALESCE(SUM(out.qty_good), 0)                           AS qty_produced,
			COUNT(DISTINCT po.id)                                     AS order_count
		FROM production_orders po
		LEFT JOIN production_outputs out ON out.production_order_id = po.id
		WHERE po.workspace_id = $1 AND po.status = 'completed'
		  AND po.completed_at >= NOW() - INTERVAL '6 months'
		GROUP BY DATE_TRUNC('month', po.completed_at)
		ORDER BY DATE_TRUNC('month', po.completed_at) ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TrendItem
	for rows.Next() {
		var it TrendItem
		if err := rows.Scan(&it.Month, &it.QtyProduced, &it.OrderCount); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []TrendItem{}
	}
	return items, nil
}
