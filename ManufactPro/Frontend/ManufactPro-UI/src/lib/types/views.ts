// Read-model / view types returned by list & report endpoints. These are the
// joined, denormalized shapes the backend serves for display — distinct from the
// normalized row types in ./index.ts. Centralized here so views and the api layer
// share one source of truth instead of redefining interfaces per route.
import type {
  ProductVariant,
  ProductionStatus,
  POStatus,
  UserRole,
} from './index'

// Product variant list rows carry the parent product name (joined).
export type ProductVariantView = ProductVariant & { product_name?: string }

// /api/stock — per-warehouse on-hand view with item display fields + valuation.
export interface StockView {
  warehouse_id: string
  warehouse_name: string
  item_type: string
  item_id: string
  item_name: string
  item_sku: string
  qty_on_hand: number
  unit_cost: number
  value: number
}

// /api/reports/dashboard
export interface DashboardReport {
  completed_orders: number
  total_qty_produced: number
  received_pos: number
  stock_value: number
  material_cost_month: number
}

// /api/reports/hpp-margin
export interface HppMarginItem {
  variant_sku: string
  product_name: string
  hpp_per_unit: number
  sell_price: number
  margin: number
  margin_pct: number
  qty_produced: number
  total_cost: number
}

// /api/reports/production-trend
export interface ProductionTrendItem {
  month: string
  qty_produced: number
  order_count: number
}

// /api/production/orders — order with joined warehouse + product/variant labels.
export interface ProductionOrderView {
  id: string
  warehouse_id: string
  warehouse_name: string
  product_variant_id: string
  variant_sku: string
  product_name: string
  qty_planned: number
  status: ProductionStatus
  planned_at: string | null
  started_at: string | null
  completed_at: string | null
  total_cost: number
  created_at: string
}

// /api/procurement/purchase-orders — list row with joined names + item count.
export interface PurchaseOrderView {
  id: string
  supplier_id: string
  supplier_name: string
  warehouse_id: string
  warehouse_name: string
  po_number: string
  status: POStatus
  total_amount: number
  ordered_at: string
  expected_at: string | null
  item_count: number
}

// /api/procurement/purchase-orders/:id — list row plus expanded line items.
export interface PurchaseOrderDetailView extends PurchaseOrderView {
  items: {
    id: string
    material_id: string
    material_name: string
    material_sku: string
    qty_ordered: number
    qty_received: number
    unit_price: number
  }[]
}

// /api/users
export interface UserView {
  id: string
  name: string
  email: string
  role: UserRole
  created_at: string
}
