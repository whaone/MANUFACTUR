<script lang="ts">
  import { link } from 'svelte-routing'
  import Card from '$lib/components/ui/Card.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import {
    Package,
    Factory,
    TrendingUp,
    AlertTriangle,
    PackagePlus,
    ClipboardList,
    Truck,
    ArrowRight,
    Activity,
  } from '@lucide/svelte'
  import { api } from '$lib/api'
  import { onMount, type Component } from 'svelte'
  import { formatNumber } from '$lib/utils/format'
  import type { Material, Product, StockMovement } from '$lib/types'
  import type { StockView, DashboardReport } from '$lib/types/views'

  type StatusKind = 'success' | 'warning' | 'critical' | 'info'

  let materials = $state<Material[]>([])
  let products = $state<Product[]>([])
  let rawStock = $state<StockView[]>([])
  let movements = $state<StockMovement[]>([])
  let dashboard = $state<DashboardReport>({ stock_value: 0, completed_orders: 0, total_qty_produced: 0, received_pos: 0, material_cost_month: 0 })
  let loading = $state(true)

  onMount(async () => {
    try {
      const [mats, prods, stk, moves, dash] = await Promise.all([
        api.materials.list(),
        api.products.list(),
        api.stock.list(),
        api.stock.movements(),
        api.reports.dashboard(),
      ])
      materials = mats
      products = prods
      rawStock = stk
      movements = moves
      dashboard = dash
    } catch (e) {
      console.error('Dashboard load failed:', e)
    } finally {
      loading = false
    }
  })

  // ── Low stock (merge stock + material min_stock) ──
  let lowStockItems = $derived(
    rawStock
      .filter((s) => s.item_type === 'material')
      .map((s) => {
        const m = materials.find((x) => x.id === s.item_id)
        return {
          name: s.item_name,
          sku: s.item_sku,
          current: s.qty_on_hand,
          min: m?.min_stock ?? 0,
          unit: m?.unit ?? 'pcs',
          warehouse: s.warehouse_name,
        }
      })
      .filter((s) => s.min > 0 && s.current < s.min)
  )

  let lowCount = $derived(lowStockItems.length)

  interface Kpi { label: string; value: string; note: string; critical?: boolean; icon: Component; color: string }

  let kpis = $derived<Kpi[]>([
    { label: 'Total Materials', value: formatNumber(materials.length), note: 'tercatat', icon: Package, color: 'text-primary bg-primary/10' },
    { label: 'Active Products', value: formatNumber(products.length), note: 'aktif', icon: Factory, color: 'text-status-info bg-status-info/10' },
    { label: 'Nilai Persediaan', value: `Rp ${(dashboard.stock_value / 1e6).toFixed(1)}M`, note: 'total stok', icon: TrendingUp, color: 'text-status-success bg-status-success/10' },
    { label: 'Low Stock Items', value: formatNumber(lowCount), note: 'perlu reorder', critical: true, icon: AlertTriangle, color: 'text-status-critical bg-status-critical/10' },
  ])

  // ── Recent activity from movements ──
  const MOVE_META: Record<string, { label: string; status: StatusKind }> = {
    IN_PURCHASE:    { label: 'Goods Receipt', status: 'info' },
    OUT_PRODUCTION: { label: 'Produksi', status: 'warning' },
    IN_PRODUCTION:  { label: 'Output Produksi', status: 'success' },
    TRANSFER_OUT:   { label: 'Transfer Keluar', status: 'info' },
    TRANSFER_IN:    { label: 'Transfer Masuk', status: 'info' },
    ADJUSTMENT:     { label: 'Penyesuaian', status: 'warning' },
  }

  function nameForItem(id: string): string {
    const m = materials.find((x) => x.id === id)
    return m?.name ?? id.slice(0, 8)
  }

  function relTime(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime()
    const min = Math.floor(diff / 60000)
    if (min < 1) return 'baru saja'
    if (min < 60) return `${min} mnt lalu`
    const hr = Math.floor(min / 60)
    if (hr < 24) return `${hr} jam lalu`
    const d = Math.floor(hr / 24)
    return `${d} hari lalu`
  }

  let activities = $derived(
    movements.slice(0, 6).map((mv) => {
      const meta = MOVE_META[mv.movement_type] ?? { label: mv.movement_type, status: 'info' as StatusKind }
      const sign = mv.qty > 0 ? '+' : ''
      return {
        action: meta.label,
        status: meta.status,
        entity: nameForItem(mv.item_id),
        detail: `${sign}${mv.qty}${mv.reason ? ' · ' + mv.reason : ''}`,
        time: relTime(mv.created_at),
      }
    })
  )
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-[32px] font-bold text-on-surface tracking-tight leading-tight">Dashboard</h1>
    <p class="text-on-surface-variant mt-1">Overview of your manufacturing operations</p>
  </div>

  <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 md:gap-5">
    {#each kpis as kpi}
      <Card variant="glass" class="flex flex-col gap-3">
        <div class="flex items-start justify-between">
          <span class="text-xs font-semibold text-on-surface-variant uppercase tracking-wider">{kpi.label}</span>
          <div class="p-2 rounded-lg {kpi.color}">
            <kpi.icon class="w-4 h-4" />
          </div>
        </div>
        <div>
          <p class="text-2xl font-bold text-on-surface">{kpi.value}</p>
          <p class="text-xs mt-0.5 {kpi.critical ? 'text-status-critical' : 'text-on-surface-variant'}">{kpi.note}</p>
        </div>
      </Card>
    {/each}
  </div>

  <div class="grid grid-cols-1 xl:grid-cols-3 gap-5">
    <div class="xl:col-span-2 space-y-4">
      <Card variant="glass">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Quick Actions</h2>
        </div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <a href="/production" use:link class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl bg-surface-container-low hover:bg-surface-container-high transition-colors border border-outline-variant/30">
            <ClipboardList class="w-6 h-6 text-primary" />
            <span class="text-xs font-medium text-on-surface">New Production</span>
          </a>
          <a href="/procurement" use:link class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl bg-surface-container-low hover:bg-surface-container-high transition-colors border border-outline-variant/30">
            <PackagePlus class="w-6 h-6 text-primary" />
            <span class="text-xs font-medium text-on-surface">New PO</span>
          </a>
          <a href="/inventory" use:link class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl bg-surface-container-low hover:bg-surface-container-high transition-colors border border-outline-variant/30">
            <Truck class="w-6 h-6 text-primary" />
            <span class="text-xs font-medium text-on-surface">Transfer</span>
          </a>
          <a href="/materials" use:link class="flex flex-col items-center justify-center gap-2 p-4 rounded-xl bg-surface-container-low hover:bg-surface-container-high transition-colors border border-outline-variant/30">
            <Package class="w-6 h-6 text-primary" />
            <span class="text-xs font-medium text-on-surface">Add Material</span>
          </a>
        </div>
      </Card>

      <Card variant="glass">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Recent Activity</h2>
          <a href="/reports" use:link class="text-xs font-medium text-primary hover:underline flex items-center gap-1">
            View All <ArrowRight class="w-3 h-3" />
          </a>
        </div>
        <div class="space-y-2">
          {#if activities.length === 0}
            <p class="text-sm text-outline py-6 text-center">{loading ? 'Memuat…' : 'Belum ada aktivitas'}</p>
          {/if}
          {#each activities as act}
            <div class="flex items-start gap-3 p-3 rounded-lg hover:bg-surface-container-low transition-colors">
              <div class="mt-0.5">
                {#if act.status === 'success'}
                  <Activity class="w-4 h-4 text-status-success" />
                {:else if act.status === 'critical'}
                  <AlertTriangle class="w-4 h-4 text-status-critical" />
                {:else}
                  <Activity class="w-4 h-4 text-status-info" />
                {/if}
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <Badge variant={act.status ?? 'info'}>{act.action}</Badge>
                  <span class="text-xs font-medium text-on-surface">{act.entity}</span>
                </div>
                <p class="text-xs text-on-surface-variant mt-0.5">{act.detail}</p>
              </div>
              <span class="text-xs text-outline shrink-0">{act.time}</span>
            </div>
          {/each}
        </div>
      </Card>
    </div>

    <div class="xl:col-span-1">
      <Card variant="glass">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Low Stock Alert</h2>
          <Badge variant="critical">{lowCount} items</Badge>
        </div>
        <div class="space-y-3">
          {#if lowStockItems.length === 0}
            <p class="text-sm text-outline py-6 text-center">{loading ? 'Memuat…' : 'Semua stok aman'}</p>
          {/if}
          {#each lowStockItems as item}
            <div class="flex items-center gap-3 p-3 rounded-lg bg-surface-container-low/50">
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-on-surface truncate">{item.name}</p>
                <p class="text-xs text-on-surface-variant">{item.sku} · {item.warehouse}</p>
              </div>
              <div class="text-right shrink-0">
                <p class="text-sm font-semibold text-status-critical">{item.current}/{item.min}</p>
                <p class="text-xs text-outline">{item.unit}</p>
              </div>
            </div>
          {/each}
        </div>
        <div class="mt-4 pt-4 border-t border-outline-variant/30">
          <a href="/inventory" use:link class="text-xs font-medium text-primary hover:underline flex items-center gap-1">
            Manage Inventory <ArrowRight class="w-3 h-3" />
          </a>
        </div>
      </Card>
    </div>
  </div>
</div>
