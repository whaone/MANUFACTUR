<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import { toISO } from '$lib/utils/date'
  import { formatRupiah } from '$lib/utils/format'
  import { onMount } from 'svelte'
  import {
    Factory,
    Plus,
    Search,
    Clock,
    Play,
    CheckCircle,
    XCircle,
  } from '@lucide/svelte'
  import type { ProductionStatus as Status, BadgeVariant, BomItem } from '$lib/types'
  import type { ProductionOrderView as ProdOrder } from '$lib/types/views'

  // BOM list rows carry joined material display fields beyond BomItem.
  type BomCheckRow = BomItem & { material_name?: string; material_sku?: string }

  let orders = $state<ProdOrder[]>([])
  let loading = $state(true)
  let variantOptions = $state<{ value: string; label: string }[]>([])
  let warehouseOptions = $state<{ value: string; label: string }[]>([])

  onMount(async () => {
    try {
      const [ordersData, variantsData, warehousesData] = await Promise.all([
        api.production.list(),
        api.productVariants.list(),
        api.warehouses.list(),
      ])
      orders = ordersData
      variantOptions = variantsData.map((v) => ({
        value: v.id,
        label: `${v.product_name ?? ''} — ${v.sku}`,
      }))
      warehouseOptions = warehousesData.map((w) => ({
        value: w.id,
        label: w.name,
      }))
    } catch (e) {
      console.error('Failed to load production data:', e)
    } finally {
      loading = false
    }
  })

  let search = $state('')
  let filterStatus = $state('')
  let showCreate = $state(false)
  let newOrder = $state({ variantId: '', warehouseId: '', qty: 0, plannedAt: '' })
  let creating = $state(false)

  let showOutput = $state(false)
  let outputData = $state({ production_order_id: '', qty_good: 0, qty_reject: 0, qty_waste: 0, reject_reason: '', waste_reason: '' })
  let savingOutput = $state(false)

  let starting = $state<string | null>(null)

  let filtered = $derived(
    orders.filter((o) => {
      const q = search.toLowerCase()
      const matchSearch = !search || o.product_name.toLowerCase().includes(q) || o.variant_sku.toLowerCase().includes(q)
      const matchStatus = !filterStatus || o.status === filterStatus
      return matchSearch && matchStatus
    })
  )

  function statusVariant(s: Status): BadgeVariant {
    const map: Record<Status, BadgeVariant> = {
      completed: 'success', in_progress: 'info', draft: 'warning', cancelled: 'critical',
    }
    return map[s] ?? 'neutral'
  }

  function statusLabel(s: Status) {
    return { completed: 'Selesai', in_progress: 'Berjalan', draft: 'Draft', cancelled: 'Dibatalkan' }[s] ?? s
  }

  async function handleCreate() {
    if (!newOrder.variantId || !newOrder.warehouseId || newOrder.qty <= 0) return
    creating = true
    try {
      const o = await api.production.create({
        warehouse_id: newOrder.warehouseId,
        product_variant_id: newOrder.variantId,
        qty_planned: Number(newOrder.qty),
        planned_at: toISO(newOrder.plannedAt),
      })
      orders = [o, ...orders]
      showCreate = false
      newOrder = { variantId: '', warehouseId: '', qty: 0, plannedAt: '' }
      toast.success('Production order dibuat', `${o.product_name} — ${o.variant_sku}`)
    } catch (e) {
      toast.error('Gagal membuat order', e instanceof Error ? e.message : 'error')
    } finally {
      creating = false
    }
  }

  // Pre-flight: cek recipe (BOM) terisi + stok material cukup sebelum mulai produksi.
  // Return daftar peringatan; kosong = aman.
  async function preflightStart(order: ProdOrder): Promise<string[]> {
    const bom = await api.bomItems.listByVariant(order.product_variant_id)
    if (!bom || bom.length === 0) {
      return ['Harap tambah material & isi resep (BOM) varian ini dulu sebelum produksi. Buka menu BOM → Add.']
    }
    const stock = await api.stock.list()
    const warnings: string[] = []
    for (const b of bom as BomCheckRow[]) {
      const onHand = stock
        .filter((s) => s.item_type === 'material' && s.item_id === b.material_id && s.warehouse_id === order.warehouse_id)
        .reduce((a, s) => a + s.qty_on_hand, 0)
      const need = b.qty * order.qty_planned
      if (onHand < need) {
        const name = b.material_name ?? b.material_sku ?? 'Material'
        warnings.push(`${name}: butuh ${need} ${b.unit ?? ''}, tersedia ${onHand}`)
      }
    }
    if (warnings.length > 0) {
      warnings.push('Tambah stok lewat Procurement → Terima Barang.')
    }
    return warnings
  }

  async function handleStart(id: string) {
    const order = orders.find((x) => x.id === id)
    if (!order) return
    starting = id
    try {
      const warnings = await preflightStart(order)
      if (warnings.length > 0) {
        toast.warning('Belum bisa mulai produksi', warnings)
        return
      }
      const o = await api.production.start(id)
      orders = orders.map(x => x.id === id ? o : x)
      toast.success('Produksi dimulai', `${order.product_name} — ${order.variant_sku}`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'error'
      if (msg.includes('insufficient stock')) {
        toast.warning('Stok material tidak cukup', 'Lengkapi stok atau periksa resep (BOM) sebelum mulai produksi.')
      } else {
        toast.error('Gagal memulai produksi', msg)
      }
    } finally {
      starting = null
    }
  }

  function openOutput(order: ProdOrder) {
    outputData = { production_order_id: order.id, qty_good: 0, qty_reject: 0, qty_waste: 0, reject_reason: '', waste_reason: '' }
    showOutput = true
  }

  async function handleRecordOutput() {
    if (savingOutput) return
    savingOutput = true
    try {
      const o = await api.production.recordOutput({
        ...outputData,
        qty_good: Number(outputData.qty_good),
        qty_reject: Number(outputData.qty_reject),
        qty_waste: Number(outputData.qty_waste),
      })
      orders = orders.map(x => x.id === outputData.production_order_id ? o : x)
      showOutput = false
      toast.success('Output dicatat', `Qty good: ${outputData.qty_good}`)
    } catch (e) {
      toast.error('Gagal catat output', e instanceof Error ? e.message : 'error')
    } finally {
      savingOutput = false
    }
  }

  async function handleCancel(id: string) {
    if (!confirm('Batalkan production order ini?')) return
    try {
      const o = await api.production.cancel(id)
      orders = orders.map(x => x.id === id ? o : x)
      toast.info('Order dibatalkan')
    } catch (e) {
      toast.error('Gagal membatalkan', e instanceof Error ? e.message : 'error')
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Production Orders</h1>
      <p class="text-on-surface-variant mt-1">Manage production runs and outputs</p>
    </div>
    <Button onclick={() => { showCreate = true }}>
      <Plus class="w-4 h-4" />
      New Production
    </Button>
  </div>

  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    {#each (['draft', 'in_progress', 'completed', 'cancelled'] as Status[]) as s}
      <button
        onclick={() => { filterStatus = filterStatus === s ? '' : s }}
        class={cn(
          'p-4 rounded-xl border transition-all text-left cursor-pointer',
          filterStatus === s
            ? 'border-primary bg-primary/5 shadow-sm'
            : 'border-outline-variant/30 bg-surface-container-lowest hover:bg-surface-container-low',
        )}
      >
        <p class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant mb-1">{statusLabel(s)}</p>
        <p class="text-xl font-bold text-on-surface">{orders.filter(o => o.status === s).length}</p>
      </button>
    {/each}
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="relative w-full">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
        <input
          type="text"
          bind:value={search}
          placeholder="Cari produk atau variant..."
          class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
        />
      </div>
    </div>

    <div class="divide-y divide-outline-variant/30">
      {#if loading}
        <div class="px-4 py-12 text-center text-outline text-sm">Memuat...</div>
      {:else if filtered.length === 0}
        <div class="px-4 py-12 text-center text-outline">
          <Factory class="w-8 h-8 mx-auto mb-2" />
          <span>Tidak ada production order</span>
        </div>
      {:else}
        {#each filtered as order}
          <div class="flex items-center gap-4 px-4 py-4 hover:bg-surface-container-low/50 transition-colors">
            <div class="hidden md:flex p-2 rounded-lg bg-primary/10 flex-shrink-0">
              <Factory class="w-5 h-5 text-primary" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <Badge variant={statusVariant(order.status)}>{statusLabel(order.status)}</Badge>
              </div>
              <p class="font-medium text-on-surface">{order.product_name} — <span class="font-mono text-sm">{order.variant_sku}</span></p>
              <p class="text-xs text-on-surface-variant mt-0.5">
                Qty: {order.qty_planned} · Gudang: {order.warehouse_name}
                {#if order.planned_at}· Rencana: {order.planned_at.slice(0,10)}{/if}
              </p>
            </div>
            <div class="flex flex-col items-end gap-2 flex-shrink-0">
              {#if order.status === 'draft'}
                <div class="flex gap-1.5">
                  <Button size="sm" onclick={() => handleStart(order.id)} disabled={starting === order.id}>
                    <Play class="w-3.5 h-3.5" />
                    {starting === order.id ? 'Memulai...' : 'Mulai Produksi'}
                  </Button>
                  <button
                    onclick={() => handleCancel(order.id)}
                    class="p-1.5 rounded text-outline hover:text-red-500 hover:bg-red-50 transition-colors"
                    title="Batalkan"
                  >
                    <XCircle class="w-4 h-4" />
                  </button>
                </div>
              {:else if order.status === 'in_progress'}
                <div class="flex flex-col items-end gap-1.5">
                  <div class="flex items-center gap-1 text-status-info">
                    <Clock class="w-4 h-4" />
                    <span class="text-xs font-medium">Berjalan</span>
                  </div>
                  {#if order.total_cost > 0}
                    <span class="text-xs text-on-surface-variant">HPP: {formatRupiah(order.total_cost)}</span>
                  {/if}
                  <Button size="sm" onclick={() => openOutput(order)}>
                    <CheckCircle class="w-3.5 h-3.5" />
                    Catat Output
                  </Button>
                </div>
              {:else if order.status === 'completed'}
                <div class="text-right">
                  <p class="text-sm font-semibold text-on-surface">HPP: {formatRupiah(order.total_cost)}</p>
                  {#if order.completed_at}
                    <p class="text-xs text-on-surface-variant">{order.completed_at.slice(0,10)}</p>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </Card>
</div>

<Modal bind:open={showCreate} title="Production Order Baru">
  <div class="space-y-4">
    <Select
      label="Produk / Variant"
      bind:value={newOrder.variantId}
      options={variantOptions}
      placeholder="Pilih variant produk"
    />
    <Select
      label="Gudang"
      bind:value={newOrder.warehouseId}
      options={warehouseOptions}
      placeholder="Pilih gudang"
    />
    <Input label="Qty Rencana" type="number" placeholder="100" bind:value={newOrder.qty} required />
    <Input label="Tanggal Rencana" type="date" bind:value={newOrder.plannedAt} />
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showCreate = false }}>Batal</Button>
      <Button onclick={handleCreate} disabled={creating}>
        {creating ? 'Membuat...' : 'Buat Order'}
      </Button>
    </div>
  {/snippet}
</Modal>

<Modal bind:open={showOutput} title="Catat Output Produksi">
  <div class="space-y-4">
    <Input label="Qty Baik" type="number" bind:value={outputData.qty_good} required />
    <Input label="Qty Reject" type="number" bind:value={outputData.qty_reject} />
    <Input label="Qty Waste" type="number" bind:value={outputData.qty_waste} />
    {#if outputData.qty_reject > 0}
      <Input label="Alasan Reject" placeholder="Cacat jahitan..." bind:value={outputData.reject_reason} />
    {/if}
    {#if outputData.qty_waste > 0}
      <Input label="Alasan Waste" placeholder="Bahan sisa..." bind:value={outputData.waste_reason} />
    {/if}
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showOutput = false }}>Batal</Button>
      <Button onclick={handleRecordOutput} disabled={savingOutput}>
        {savingOutput ? 'Menyimpan...' : 'Simpan Output'}
      </Button>
    </div>
  {/snippet}
</Modal>
