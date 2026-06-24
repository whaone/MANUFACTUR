import type { Material, Supplier, BomItem, ProductVariant, Product, Warehouse, Branch, StockMovement } from '$lib/types'
import type {
  ProductVariantView,
  StockView,
  DashboardReport,
  HppMarginItem,
  ProductionTrendItem,
  ProductionOrderView,
  PurchaseOrderView,
  PurchaseOrderDetailView,
  UserView,
} from '$lib/types/views'
import { auth } from '$lib/stores/auth'
import { syncedWrite } from '$lib/sync/mutate'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

function getToken(): string | null {
  try {
    const s = localStorage.getItem('manufactpro_auth')
    return s ? JSON.parse(s).access_token : null
  } catch {
    return null
  }
}

function doFetch(path: string, token: string | null, options?: RequestInit): Promise<Response> {
  return fetch(`${BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  })
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  let res = await doFetch(path, getToken(), options)

  // Access token expired → try one silent refresh, then retry once.
  if (res.status === 401) {
    if (await auth.refresh()) {
      res = await doFetch(path, getToken(), options)
    } else {
      // Unrecoverable: session invalid/expired and no usable refresh token.
      // Kick to login instead of leaving every menu in a broken 401 state.
      auth.logout()
      throw new Error('Sesi berakhir, silakan login kembali')
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  materials: {
    list: () => apiFetch<Material[]>('/api/materials'),
    getById: (id: string) => apiFetch<Material>(`/api/materials/${id}`),
    create: (data: Omit<Material, 'id' | 'created_at' | 'updated_at' | 'workspace_id'>) =>
      syncedWrite<Material>('material', 'create', { ...data },
        () => apiFetch<Material>('/api/materials', { method: 'POST', body: JSON.stringify(data) }),
        () => ({ ...data, id: crypto.randomUUID() }) as Material),
    update: (id: string, data: Partial<Material>) =>
      syncedWrite<Material>('material', 'update', { id, ...data },
        () => apiFetch<Material>(`/api/materials/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
        () => ({ id, ...data }) as Material),
    delete: (id: string) =>
      syncedWrite<void>('material', 'delete', { id },
        () => apiFetch<void>(`/api/materials/${id}`, { method: 'DELETE' }),
        () => undefined as void),
  },

  suppliers: {
    list: () => apiFetch<Supplier[]>('/api/suppliers'),
    create: (data: Omit<Supplier, 'id' | 'workspace_id'>) =>
      syncedWrite<Supplier>('supplier', 'create', { ...data },
        () => apiFetch<Supplier>('/api/suppliers', { method: 'POST', body: JSON.stringify(data) }),
        () => ({ ...data, id: crypto.randomUUID() }) as Supplier),
    update: (id: string, data: Partial<Supplier>) =>
      syncedWrite<Supplier>('supplier', 'update', { id, ...data },
        () => apiFetch<Supplier>(`/api/suppliers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
        () => ({ id, ...data }) as Supplier),
    delete: (id: string) =>
      syncedWrite<void>('supplier', 'delete', { id },
        () => apiFetch<void>(`/api/suppliers/${id}`, { method: 'DELETE' }),
        () => undefined as void),
  },

  products: {
    list: () => apiFetch<Product[]>('/api/products'),
    getById: (id: string) => apiFetch<Product>(`/api/products/${id}`),
    create: (data: Omit<Product, 'id' | 'created_at' | 'updated_at' | 'workspace_id' | 'variants'>) =>
      syncedWrite<Product>('product', 'create', { ...data },
        () => apiFetch<Product>('/api/products', { method: 'POST', body: JSON.stringify(data) }),
        () => ({ ...data, id: crypto.randomUUID() }) as Product),
    update: (id: string, data: Partial<Product>) =>
      syncedWrite<Product>('product', 'update', { id, ...data },
        () => apiFetch<Product>(`/api/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
        () => ({ id, ...data }) as Product),
    delete: (id: string) =>
      syncedWrite<void>('product', 'delete', { id },
        () => apiFetch<void>(`/api/products/${id}`, { method: 'DELETE' }),
        () => undefined as void),
  },

  productVariants: {
    list: () => apiFetch<ProductVariantView[]>('/api/variants'),
    create: (productId: string, data: Omit<ProductVariant, 'id' | 'created_at' | 'product_id'>) =>
      apiFetch<ProductVariant>(`/api/products/${productId}/variants`, {
        method: 'POST', body: JSON.stringify(data),
      }),
    update: (id: string, data: Partial<ProductVariant>) =>
      apiFetch<ProductVariant>(`/api/variants/${id}`, {
        method: 'PUT', body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      apiFetch<void>(`/api/variants/${id}`, { method: 'DELETE' }),
  },

  bomItems: {
    listByVariant: (variantId: string) =>
      apiFetch<BomItem[]>(`/api/variants/${variantId}/bom`),
    create: (data: Omit<BomItem, 'id'>) =>
      apiFetch<BomItem>(`/api/variants/${data.product_variant_id}/bom`, {
        method: 'POST', body: JSON.stringify(data),
      }),
    update: (id: string, data: Partial<BomItem>) =>
      apiFetch<BomItem>(`/api/bom/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) =>
      apiFetch<void>(`/api/bom/${id}`, { method: 'DELETE' }),
  },

  users: {
    list: () => apiFetch<UserView[]>('/api/users'),
    invite: (data: { name: string; email: string; role: string; password: string }) =>
      apiFetch<UserView>('/api/users', { method: 'POST', body: JSON.stringify(data) }),
    updateRole: (id: string, role: string) =>
      apiFetch<UserView>(`/api/users/${id}/role`, { method: 'PATCH', body: JSON.stringify({ role }) }),
    delete: (id: string) =>
      apiFetch<void>(`/api/users/${id}`, { method: 'DELETE' }),
  },

  warehouses: {
    list: () => apiFetch<Warehouse[]>('/api/warehouses'),
    listBranches: () => apiFetch<Branch[]>('/api/branches'),
    create: (data: { branch_id: string; name: string; code: string; is_default: boolean }) =>
      apiFetch<Warehouse>('/api/warehouses', { method: 'POST', body: JSON.stringify(data) }),
    createBranch: (data: { name: string; address: string }) =>
      apiFetch<Branch>('/api/branches', { method: 'POST', body: JSON.stringify(data) }),
  },

  sync: {
    push: (items: {
      id: string
      operation: 'create' | 'update' | 'delete'
      entity: string
      payload: unknown
      client_timestamp: number
    }[]) =>
      apiFetch<{
        results: { id: string; status: 'synced' | 'failed'; server_id?: string; error?: string }[]
        synced: number
        failed: number
      }>('/api/sync', { method: 'POST', body: JSON.stringify({ items }) }),
  },

  stock: {
    list: () => apiFetch<StockView[]>('/api/stock'),
    movements: () => apiFetch<StockMovement[]>('/api/stock/movements'),
    transfer: (data: {
      from_warehouse_id: string
      to_warehouse_id: string
      item_type: string
      item_id: string
      qty: number
      reason?: string
    }) => apiFetch<StockMovement>('/api/stock/transfer', { method: 'POST', body: JSON.stringify(data) }),
    adjustment: (data: {
      warehouse_id: string
      item_type: string
      item_id: string
      qty: number
      reason?: string
    }) => apiFetch<StockMovement>('/api/stock/adjustment', { method: 'POST', body: JSON.stringify(data) }),
  },

  reports: {
    dashboard: () => apiFetch<DashboardReport>('/api/reports/dashboard'),
    hppMargin: () => apiFetch<HppMarginItem[]>('/api/reports/hpp-margin'),
    productionTrend: () => apiFetch<ProductionTrendItem[]>('/api/reports/production-trend'),
  },

  procurement: {
    list: () => apiFetch<PurchaseOrderView[]>('/api/procurement/purchase-orders'),
    getDetail: (id: string) => apiFetch<PurchaseOrderDetailView>(`/api/procurement/purchase-orders/${id}`),
    create: (data: {
      supplier_id: string
      warehouse_id: string
      po_number?: string
      expected_at?: string | null
      items: { material_id: string; qty_ordered: number; unit_price: number }[]
    }) => apiFetch<PurchaseOrderView>('/api/procurement/purchase-orders', { method: 'POST', body: JSON.stringify(data) }),
    send: (id: string) =>
      apiFetch<PurchaseOrderView>(`/api/procurement/purchase-orders/${id}/send`, { method: 'PATCH' }),
    receive: (id: string, data: {
      note?: string
      items: { po_item_id: string; qty_received: number; batch_no?: string; expiry_at?: string | null }[]
    }) => apiFetch<PurchaseOrderView>(`/api/procurement/purchase-orders/${id}/receive`, { method: 'POST', body: JSON.stringify(data) }),
    cancel: (id: string) =>
      apiFetch<PurchaseOrderView>(`/api/procurement/purchase-orders/${id}/cancel`, { method: 'PATCH' }),
  },

  production: {
    list: () => apiFetch<ProductionOrderView[]>('/api/production/orders'),
    create: (data: {
      warehouse_id: string
      product_variant_id: string
      qty_planned: number
      planned_at?: string | null
    }) => apiFetch<ProductionOrderView>('/api/production/orders', { method: 'POST', body: JSON.stringify(data) }),
    start: (id: string) =>
      apiFetch<ProductionOrderView>(`/api/production/orders/${id}/start`, { method: 'POST' }),
    cancel: (id: string) =>
      apiFetch<ProductionOrderView>(`/api/production/orders/${id}/cancel`, { method: 'PATCH' }),
    recordOutput: (data: {
      production_order_id: string
      qty_good: number
      qty_reject: number
      qty_waste: number
      reject_reason?: string
      waste_reason?: string
    }) =>
      apiFetch<ProductionOrderView>(`/api/production/orders/${data.production_order_id}/output`, {
        method: 'POST', body: JSON.stringify(data),
      }),
  },
}
