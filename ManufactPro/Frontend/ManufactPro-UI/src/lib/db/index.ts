import Dexie, { type Table } from 'dexie'
import type {
  Material,
  MaterialBatch,
  Product,
  ProductVariant,
  BomItem,
  StockMovement,
  ProductionOrder,
  ProductionOutput,
  Supplier,
  PurchaseOrder,
  PoItem,
  GoodsReceipt,
  SyncQueueItem,
} from '$lib/types'

export class ManufactProDB extends Dexie {
  materials!: Table<Material, string>
  materialBatches!: Table<MaterialBatch, string>
  products!: Table<Product, string>
  productVariants!: Table<ProductVariant, string>
  bomItems!: Table<BomItem, string>
  stockMovements!: Table<StockMovement, string>
  productionOrders!: Table<ProductionOrder, string>
  productionOutputs!: Table<ProductionOutput, string>
  suppliers!: Table<Supplier, string>
  purchaseOrders!: Table<PurchaseOrder, string>
  poItems!: Table<PoItem, string>
  goodsReceipts!: Table<GoodsReceipt, string>
  syncQueue!: Table<SyncQueueItem, string>

  constructor() {
    super('ManufactPro')
    this.version(1).stores({
      materials: 'id, workspace_id, sku, name, category, barcode, is_active',
      materialBatches: 'id, material_id, warehouse_id, batch_no',
      products: 'id, workspace_id, name, category',
      productVariants: 'id, product_id, sku, barcode, is_active',
      bomItems: 'id, product_variant_id, material_id',
      stockMovements: 'id, workspace_id, warehouse_id, item_id, item_type, movement_type, created_at',
      productionOrders: 'id, workspace_id, warehouse_id, product_variant_id, status, planned_at',
      productionOutputs: 'id, production_order_id',
      suppliers: 'id, workspace_id, name',
      purchaseOrders: 'id, workspace_id, supplier_id, status',
      poItems: 'id, po_id, material_id',
      goodsReceipts: 'id, po_id',
      syncQueue: 'id, status, client_timestamp',
    })
  }
}

export const db = new ManufactProDB()
