<script lang="ts">
  import { cn } from '$lib/utils/cn'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { Warehouse, Search, ArrowRight } from '@lucide/svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import { onMount } from 'svelte'
  import { formatRupiah } from '$lib/utils/format'
  import type { Material } from '$lib/types'
  import type { StockView } from '$lib/types/views'

  interface StockRow {
    key: string
    name: string
    sku: string
    category: string
    current: number
    min_stock: number
    unit: string
    warehouse: string
    warehouse_id: string
    item_type: string
    item_id: string
    value: number
  }

  let rawStock = $state<StockView[]>([])
  let materials = $state<Material[]>([])
  let warehouseList = $state<{ id: string; name: string }[]>([])
  let loading = $state(true)
  let errMsg = $state('')

  const WH_COLORS = ['bg-primary/10 text-primary', 'bg-status-info/10 text-status-info', 'bg-status-success/10 text-status-success', 'bg-status-warning/10 text-status-warning']

  async function loadStock() {
    const [stk, mats, whs] = await Promise.all([
      api.stock.list(),
      api.materials.list(),
      api.warehouses.list(),
    ])
    rawStock = stk
    materials = mats
    warehouseList = whs.map((w) => ({ id: w.id, name: w.name }))
  }

  onMount(async () => {
    try {
      await loadStock()
    } catch (e) {
      errMsg = e instanceof Error ? e.message : 'Gagal memuat stok'
      console.error(e)
    } finally {
      loading = false
    }
  })

  // Merge stock + material (min_stock, unit, category) by item_id
  let stocks = $derived<StockRow[]>(
    rawStock.map((s) => {
      const m = s.item_type === 'material' ? materials.find((x) => x.id === s.item_id) : null
      return {
        key: `${s.warehouse_id}:${s.item_type}:${s.item_id}`,
        name: s.item_name,
        sku: s.item_sku,
        category: m?.category || (s.item_type === 'product_variant' ? 'Produk Jadi' : '—'),
        current: s.qty_on_hand,
        min_stock: m?.min_stock ?? 0,
        unit: m?.unit || 'pcs',
        warehouse: s.warehouse_name,
        warehouse_id: s.warehouse_id,
        item_type: s.item_type,
        item_id: s.item_id,
        value: s.value,
      }
    })
  )

  let search = $state('')
  let filterWh = $state('')
  let filterLow = $state(false)
  let filterCategory = $state('')

  let warehouseColors = $derived(
    Object.fromEntries(warehouseList.map((w, i) => [w.name, WH_COLORS[i % WH_COLORS.length]]))
  )

  let filtered = $derived(
    stocks.filter((s) => {
      const m = !search || s.name.toLowerCase().includes(search.toLowerCase()) || s.sku.toLowerCase().includes(search.toLowerCase())
      const w = !filterWh || s.warehouse === filterWh
      const l = !filterLow || (s.min_stock > 0 && s.current < s.min_stock)
      const c = !filterCategory || s.category === filterCategory
      return m && w && l && c
    })
  )

  let lowCount = $derived(stocks.filter((s) => s.min_stock > 0 && s.current < s.min_stock).length)
  let totalValue = $derived(filtered.reduce((a, s) => a + s.value, 0))
  let categories = $derived([...new Set(stocks.map((s) => s.category))])

  // ── Transfer ──
  let showTransfer = $state(false)
  let transfer = $state({ from: '', to: '', itemKey: '', qty: 0, reason: '' })
  let transferErr = $state('')
  let transferring = $state(false)

  // items available in the selected "from" warehouse
  let transferItems = $derived(
    rawStock
      .filter((s) => !transfer.from || s.warehouse_id === transfer.from)
      .map((s) => ({ value: `${s.item_type}:${s.item_id}`, label: `${s.item_name} (${s.qty_on_hand})` }))
  )

  function openTransfer() {
    transfer = { from: warehouseList[0]?.id ?? '', to: warehouseList[1]?.id ?? '', itemKey: '', qty: 0, reason: '' }
    transferErr = ''
    showTransfer = true
  }

  async function handleTransfer() {
    transferErr = ''
    if (!transfer.from || !transfer.to || !transfer.itemKey || transfer.qty <= 0) {
      transferErr = 'Lengkapi gudang asal, tujuan, item, dan qty'
      return
    }
    if (transfer.from === transfer.to) {
      transferErr = 'Gudang asal dan tujuan harus berbeda'
      return
    }
    const [item_type, item_id] = transfer.itemKey.split(':')

    // Pre-flight: stok di gudang asal cukup?
    const avail = rawStock
      .filter((s) => s.warehouse_id === transfer.from && s.item_type === item_type && s.item_id === item_id)
      .reduce((a, s) => a + s.qty_on_hand, 0)
    if (Number(transfer.qty) > avail + 1e-9) {
      transferErr = `Stok tidak cukup: tersedia ${avail}, diminta ${transfer.qty}`
      toast.warning('Stok gudang asal kurang', transferErr)
      return
    }

    transferring = true
    try {
      await api.stock.transfer({
        from_warehouse_id: transfer.from,
        to_warehouse_id: transfer.to,
        item_type,
        item_id,
        qty: Number(transfer.qty),
        reason: transfer.reason,
      })
      await loadStock()
      showTransfer = false
      toast.success('Transfer berhasil', `${transfer.qty} dipindah antar gudang`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Transfer gagal'
      transferErr = msg
      if (msg.toLowerCase().includes('insufficient') || msg.toLowerCase().includes('stok')) {
        toast.warning('Stok tidak cukup untuk transfer', msg)
      } else {
        toast.error('Transfer gagal', msg)
      }
    } finally {
      transferring = false
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Inventory</h1>
      <p class="text-on-surface-variant mt-1">Multi-warehouse stock management</p>
    </div>
    <Button onclick={openTransfer}>
      <ArrowRight class="w-4 h-4" />
      Transfer
    </Button>
  </div>

  {#if errMsg}
    <div class="px-4 py-3 rounded-lg bg-status-critical/10 text-status-critical text-sm">{errMsg}</div>
  {/if}

  <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
    <button
      onclick={() => { filterLow = false }}
      class={cn(
        'p-4 rounded-xl border transition-all text-left cursor-pointer',
        !filterLow
          ? 'border-primary bg-primary/5 shadow-sm'
          : 'border-outline-variant/30 bg-surface-container-lowest hover:bg-surface-container-low',
      )}
    >
      <p class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Total Items</p>
      <p class="text-xl font-bold text-on-surface mt-1">{stocks.length} SKUs</p>
    </button>
    <button
      onclick={() => { filterLow = !filterLow }}
      class={cn(
        'p-4 rounded-xl border transition-all text-left cursor-pointer',
        filterLow
          ? 'border-primary bg-status-critical/5 shadow-sm'
          : 'border-outline-variant/30 bg-surface-container-lowest hover:bg-surface-container-low',
      )}
    >
      <p class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Low Stock Items</p>
      <p class="text-xl font-bold text-on-surface mt-1">{lowCount} need reorder</p>
    </button>
    <Card variant="glass">
      <p class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">Total Value</p>
      <p class="text-xl font-bold text-on-surface mt-1">{formatRupiah(totalValue)}</p>
    </Card>
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
          <input type="text" bind:value={search} placeholder="Search inventory..." class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary" />
        </div>
        <select bind:value={filterWh} class="px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant appearance-none cursor-pointer">
          <option value="">All Warehouses</option>
          {#each warehouseList as wh}<option value={wh.name}>{wh.name}</option>{/each}
        </select>
        <select bind:value={filterCategory} class="px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant appearance-none cursor-pointer">
          <option value="">All Categories</option>
          {#each categories as cat}<option value={cat}>{cat}</option>{/each}
        </select>
        <label class="flex items-center gap-2 text-sm text-on-surface-variant cursor-pointer px-3 py-2 rounded-lg hover:bg-surface-container-low transition-colors">
          <input type="checkbox" bind:checked={filterLow} class="rounded border-outline-variant text-primary focus:ring-primary" />
          Low stock only
        </label>
      </div>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-outline-variant bg-surface-container-low">
            <th class="px-4 py-3 text-left text-xs font-semibold text-on-surface-variant uppercase">Item</th>
            <th class="px-4 py-3 text-center text-xs font-semibold text-on-surface-variant uppercase">Warehouse</th>
            <th class="px-4 py-3 text-right text-xs font-semibold text-on-surface-variant uppercase">Stock</th>
            <th class="px-4 py-3 text-right text-xs font-semibold text-on-surface-variant uppercase">Value</th>
            <th class="px-4 py-3 text-center text-xs font-semibold text-on-surface-variant uppercase">Status</th>
          </tr>
        </thead>
        <tbody>
          {#if loading}
            <tr><td colspan="5" class="px-4 py-12 text-center text-outline">Memuat…</td></tr>
          {:else if filtered.length === 0}
            <tr><td colspan="5" class="px-4 py-12 text-center text-outline"><Warehouse class="w-8 h-8 mx-auto mb-2" /><span>No inventory data</span></td></tr>
          {:else}
            {#each filtered as item (item.key)}
              <tr class="border-b border-outline-variant/20 hover:bg-surface-container-low/50 transition-colors">
                <td class="px-4 py-3">
                  <p class="font-medium text-on-surface">{item.name}</p>
                  <p class="font-mono text-xs text-on-surface-variant">{item.sku} · {item.category}</p>
                </td>
                <td class="px-4 py-3 text-center">
                  <span class="inline-flex px-2 py-0.5 rounded-full text-xs font-medium {warehouseColors[item.warehouse] ?? 'bg-surface-container-high'}">{item.warehouse}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="font-medium {item.min_stock > 0 && item.current < item.min_stock ? 'text-status-critical' : 'text-on-surface'}">{item.current}</span>
                  <span class="text-on-surface-variant text-xs ml-1">{item.unit}</span>
                </td>
                <td class="px-4 py-3 text-right font-medium text-on-surface">
                  {formatRupiah(item.value)}
                </td>
                <td class="px-4 py-3 text-center">
                  {#if item.min_stock > 0 && item.current < item.min_stock}
                    <Badge variant="critical">Low ({item.min_stock})</Badge>
                  {:else}
                    <Badge variant="success">OK</Badge>
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </Card>
</div>

<Modal bind:open={showTransfer} title="Transfer Between Warehouses">
  <div class="space-y-4">
    {#if transferErr}
      <div class="px-3 py-2 rounded-lg bg-status-critical/10 text-status-critical text-sm">{transferErr}</div>
    {/if}
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <Select label="From" bind:value={transfer.from} options={warehouseList.map((w) => ({ value: w.id, label: w.name }))} />
      <Select label="To" bind:value={transfer.to} options={warehouseList.map((w) => ({ value: w.id, label: w.name }))} />
    </div>
    <Select label="Item" bind:value={transfer.itemKey} options={transferItems} placeholder="Pilih item dari gudang asal" />
    <Input label="Quantity" type="number" placeholder="0" bind:value={transfer.qty} required />
    <Input label="Reason (opsional)" placeholder="Alasan transfer" bind:value={transfer.reason} />
  </div>
  {#snippet footer()}
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-end gap-3">
      <Button variant="secondary" onclick={() => { showTransfer = false }}>Cancel</Button>
      <Button onclick={handleTransfer} disabled={transferring}>
        <ArrowRight class="w-4 h-4" />
        {transferring ? 'Memproses…' : 'Execute Transfer'}
      </Button>
    </div>
  {/snippet}
</Modal>
