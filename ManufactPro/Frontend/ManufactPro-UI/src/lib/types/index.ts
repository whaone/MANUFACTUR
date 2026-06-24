export type BadgeVariant = 'success' | 'warning' | 'critical' | 'info' | 'neutral' | 'primary'

export interface Workspace {
  id: string
  name: string
  costing_method: 'fifo' | 'weighted_average'
  created_at: string
}

export interface Branch {
  id: string
  workspace_id: string
  name: string
  address: string
}

export interface Warehouse {
  id: string
  branch_id: string
  name: string
  code: string
  is_default: boolean
}

export interface Material {
  id: string
  workspace_id: string
  sku: string
  name: string
  unit: 'meter' | 'pcs' | 'gram' | 'liter' | 'kg' | 'lusin'
  category: string
  min_stock: number
  barcode: string
  image_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface MaterialBatch {
  id: string
  material_id: string
  warehouse_id: string
  batch_no: string
  qty_received: number
  qty_remaining: number
  unit_cost: number
  supplier_id: string
  received_at: string
  expiry_at?: string
}

export interface Product {
  id: string
  workspace_id: string
  name: string
  category: string
  description?: string
  image_url?: string
  created_at: string
  updated_at: string
  variants?: ProductVariant[]
}

export interface ProductVariant {
  id: string
  product_id: string
  sku: string
  barcode: string
  attributes: Record<string, string>
  sell_price: number
  is_active: boolean
  created_at: string
}

export interface BomItem {
  id: string
  product_variant_id: string
  material_id: string
  qty: number
  unit: string
  is_optional: boolean
  note?: string
}

export type MovementType =
  | 'IN_PURCHASE'
  | 'IN_PRODUCTION'
  | 'OUT_PRODUCTION'
  | 'OUT_SALES'
  | 'TRANSFER_OUT'
  | 'TRANSFER_IN'
  | 'ADJUSTMENT'
  | 'WASTE'
  | 'REJECT'

export interface StockMovement {
  id: string
  workspace_id: string
  warehouse_id: string
  item_type: 'material' | 'product_variant'
  item_id: string
  qty: number
  movement_type: MovementType
  reference_type?: string
  reference_id?: string
  unit_cost: number
  reason?: string
  created_by: string
  created_at: string
}

export type ProductionStatus = 'draft' | 'in_progress' | 'completed' | 'cancelled'

export interface ProductionOrder {
  id: string
  workspace_id: string
  warehouse_id: string
  product_variant_id: string
  qty_planned: number
  status: ProductionStatus
  planned_at: string
  started_at?: string
  completed_at?: string
  total_cost: number
  created_by: string
}

export interface ProductionOutput {
  id: string
  production_order_id: string
  qty_good: number
  qty_reject: number
  qty_waste: number
  reject_reason?: string
  waste_reason?: string
  recorded_at: string
}

export interface Supplier {
  id: string
  workspace_id: string
  name: string
  contact?: string
  email?: string
  phone?: string
  payment_term?: string
}

export type POStatus = 'draft' | 'sent' | 'partial_received' | 'received' | 'cancelled'

export interface PurchaseOrder {
  id: string
  workspace_id: string
  supplier_id: string
  warehouse_id: string
  po_number: string
  status: POStatus
  total_amount: number
  ordered_at: string
  expected_at?: string
}

export interface PoItem {
  id: string
  po_id: string
  material_id: string
  qty_ordered: number
  qty_received: number
  unit_price: number
}

export interface GoodsReceipt {
  id: string
  po_id: string
  received_at: string
  received_by: string
  note?: string
}

export type UserRole = 'owner' | 'admin' | 'production' | 'warehouse' | 'viewer'

export interface User {
  id: string
  workspace_id: string
  name: string
  email: string
  role: UserRole
}

export interface SyncQueueItem {
  id: string
  operation: 'create' | 'update' | 'delete'
  entity: string
  payload: unknown
  client_timestamp: number
  status: 'pending' | 'syncing' | 'synced' | 'failed'
  retries: number
  error?: string
}
