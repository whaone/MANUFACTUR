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
  import { onMount } from 'svelte'
  import { Truck, Plus, Search, Send, Package, CheckCircle2, Trash2 } from '@lucide/svelte'
  import type { POStatus as PoStatus, BadgeVariant } from '$lib/types'

  interface PO {
    id: string
    supplier_id: string
    supplier_name: string
    warehouse_id: string
    warehouse_name: string
    po_number: string
    status: PoStatus
    total_amount: number
    ordered_at: string
    expected_at: string | null
    item_count: number
  }

  interface PODetail extends PO {
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

  let orders = $state<PO[]>([])
  let loading = $state(true)
  let supplierOptions = $state<{ value: string; label: string }[]>([])
  let warehouseOptions = $state<{ value: string; label: string }[]>([])
  let materialOptions = $state<{ value: string; label: string }[]>([])

  onMount(async () => {
    try {
      const [pos, suppliers, warehouses, materials] = await Promise.all([
        api.procurement.list(),
        api.suppliers.list(),
        api.warehouses.list(),
        api.materials.list(),
      ])
      orders = pos
      supplierOptions = suppliers.map((s: any) => ({ value: s.id, label: s.name }))
      warehouseOptions = warehouses.map((w: any) => ({ value: w.id, label: w.name }))
      materialOptions = materials.map((m: any) => ({ value: m.id, label: `${m.name} (${m.sku})` }))
    } catch (e) {
      console.error('Failed to load procurement data:', e)
    } finally {
      loading = false
    }
  })

  let search = $state('')
  let filterStatus = $state('')

  // Create PO form
  let showCreate = $state(false)
  let createData = $state({ supplierId: '', warehouseId: '', poNumber: '', expectedAt: '' })
  type LineItem = { materialId: string; qty: number; unitPrice: number }
  let lineItems = $state<LineItem[]>([{ materialId: '', qty: 0, unitPrice: 0 }])
  let creating = $state(false)

  // Receive form
  let showReceive = $state(false)
  let receiveDetail = $state<PODetail | null>(null)
  type ReceiveLine = { po_item_id: string; qty_received: number; batch_no: string }
  let receiveLines = $state<ReceiveLine[]>([])
  let receiveNote = $state('')
  let receiving = $state(false)

  let filtered = $derived(
    orders.filter(o => {
      const q = search.toLowerCase()
      const ms = !search || o.po_number.toLowerCase().includes(q) || o.supplier_name.toLowerCase().includes(q)
      const ss = !filterStatus || o.status === filterStatus
      return ms && ss
    })
  )

  function statusVariant(s: PoStatus): BadgeVariant {
    const map: Record<PoStatus, BadgeVariant> = {
      received: 'success', partial_received: 'warning', sent: 'info', draft: 'neutral', cancelled: 'critical',
    }
    return map[s] ?? 'neutral'
  }

  function statusLabel(s: PoStatus) {
    return { draft: 'Draft', sent: 'Terkirim', partial_received: 'Parsial', received: 'Diterima', cancelled: 'Dibatalkan' }[s] ?? s
  }

  function addLineItem() { lineItems = [...lineItems, { materialId: '', qty: 0, unitPrice: 0 }] }
  function removeLineItem(i: number) { lineItems = lineItems.filter((_, idx) => idx !== i) }

  async function handleCreate() {
    if (!createData.supplierId || !createData.warehouseId) return
    const validItems = lineItems.filter(it => it.materialId && it.qty > 0)
    if (validItems.length === 0) return
    creating = true
    try {
      const po = await api.procurement.create({
        supplier_id: createData.supplierId,
        warehouse_id: createData.warehouseId,
        po_number: createData.poNumber || undefined,
        expected_at: toISO(createData.expectedAt),
        items: validItems.map(it => ({ material_id: it.materialId, qty_ordered: it.qty, unit_price: it.unitPrice })),
      })
      orders = [{ ...po, item_count: po.items?.length ?? validItems.length }, ...orders]
      showCreate = false
      createData = { supplierId: '', warehouseId: '', poNumber: '', expectedAt: '' }
      lineItems = [{ materialId: '', qty: 0, unitPrice: 0 }]
      toast.success('PO dibuat', po.po_number)
    } catch (e) {
      toast.error('Gagal membuat PO', e instanceof Error ? e.message : 'error')
    } finally {
      creating = false
    }
  }

  async function handleSend(id: string) {
    try {
      const po = await api.procurement.send(id)
      orders = orders.map(o => o.id === id ? { ...o, status: po.status } : o)
      toast.success('PO dikirim ke supplier')
    } catch (e) {
      toast.error('Gagal mengirim PO', e instanceof Error ? e.message : 'error')
    }
  }

  async function openReceive(po: PO) {
    try {
      const detail = await api.procurement.getDetail(po.id)
      receiveDetail = detail
      receiveLines = detail.items.map((it: any) => ({
        po_item_id: it.id,
        qty_received: it.qty_ordered - it.qty_received,
        batch_no: '',
      }))
      receiveNote = ''
      showReceive = true
    } catch (e) {
      toast.error('Gagal memuat detail PO', e instanceof Error ? e.message : 'error')
    }
  }

  async function handleReceive() {
    if (!receiveDetail || receiving) return

    // Pre-flight: jangan terima lebih dari sisa yang dipesan
    const overReceive: string[] = []
    for (const l of receiveLines) {
      const item = receiveDetail.items.find((it: any) => it.id === l.po_item_id)
      if (!item) continue
      const remaining = item.qty_ordered - item.qty_received
      if (l.qty_received > remaining + 1e-9) {
        const name = item.material_name ?? item.material_sku ?? 'Material'
        overReceive.push(`${name}: sisa pesanan ${remaining}, diinput ${l.qty_received}`)
      }
    }
    if (overReceive.length > 0) {
      toast.warning('Jumlah terima melebihi pesanan', overReceive)
      return
    }

    receiving = true
    try {
      const validLines = receiveLines.filter(l => l.qty_received > 0)
      if (validLines.length === 0) {
        toast.warning('Tidak ada item untuk diterima', 'Isi qty terima minimal satu item.')
        return
      }
      const updated = await api.procurement.receive(receiveDetail.id, {
        note: receiveNote,
        items: validLines.map(l => ({ po_item_id: l.po_item_id, qty_received: l.qty_received, batch_no: l.batch_no || undefined })),
      })
      orders = orders.map(o => o.id === updated.id ? { ...updated, item_count: updated.items?.length ?? o.item_count } : o)
      showReceive = false
      toast.success('Barang diterima', `Status PO: ${updated.status}`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'error'
      if (msg.includes('over-receive')) {
        toast.warning('Jumlah terima melebihi pesanan', 'Periksa kembali qty terima per item.')
      } else {
        toast.error('Gagal menerima barang', msg)
      }
    } finally {
      receiving = false
    }
  }

  async function handleCancel(id: string) {
    if (!confirm('Batalkan PO ini?')) return
    try {
      await api.procurement.cancel(id)
      orders = orders.map(o => o.id === id ? { ...o, status: 'cancelled' as PoStatus } : o)
      toast.info('PO dibatalkan')
    } catch (e) {
      toast.error('Gagal membatalkan PO', e instanceof Error ? e.message : 'error')
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Procurement</h1>
      <p class="text-on-surface-variant mt-1">Purchase Orders & Goods Receipt</p>
    </div>
    <Button onclick={() => { showCreate = true }}>
      <Plus class="w-4 h-4" />
      New Purchase Order
    </Button>
  </div>

  <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
    {#each (['draft', 'sent', 'partial_received', 'received'] as PoStatus[]) as s}
      <button
        onclick={() => { filterStatus = filterStatus === s ? '' : s }}
        class={cn(
          'p-4 rounded-xl border transition-all text-left cursor-pointer',
          filterStatus === s ? 'border-primary bg-primary/5 shadow-sm' : 'border-outline-variant/30 bg-surface-container-lowest hover:bg-surface-container-low',
        )}
      >
        <p class="text-xs font-semibold uppercase tracking-wider text-on-surface-variant">{statusLabel(s)}</p>
        <p class="text-xl font-bold text-on-surface mt-1">{orders.filter(o => o.status === s).length}</p>
      </button>
    {/each}
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="relative w-full">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
        <input type="text" bind:value={search} placeholder="Cari nomor PO atau supplier..."
          class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary" />
      </div>
    </div>
    <div class="divide-y divide-outline-variant/30">
      {#if loading}
        <div class="px-4 py-12 text-center text-outline text-sm">Memuat...</div>
      {:else if filtered.length === 0}
        <div class="px-4 py-12 text-center text-outline"><Truck class="w-8 h-8 mx-auto mb-2" /><span>Tidak ada purchase order</span></div>
      {:else}
        {#each filtered as po}
          <div class="flex items-center gap-4 px-4 py-4 hover:bg-surface-container-low/50 transition-colors">
            <div class="hidden md:flex p-2 rounded-lg bg-primary/10 flex-shrink-0">
              <Truck class="w-5 h-5 text-primary" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-mono text-xs text-on-surface-variant">{po.po_number}</span>
                <Badge variant={statusVariant(po.status)}>{statusLabel(po.status)}</Badge>
              </div>
              <p class="font-medium text-on-surface">{po.supplier_name}</p>
              <p class="text-xs text-on-surface-variant">{po.warehouse_name} · {po.item_count} item · {po.ordered_at.slice(0,10)}</p>
            </div>
            <div class="flex flex-col items-end gap-1.5 flex-shrink-0">
              <p class="text-sm font-semibold text-on-surface">Rp {po.total_amount.toLocaleString('id-ID')}</p>
              <div class="flex gap-1">
                {#if po.status === 'draft'}
                  <Button size="sm" variant="secondary" onclick={() => handleSend(po.id)}>
                    <Send class="w-3 h-3" /> Kirim
                  </Button>
                  <button onclick={() => handleCancel(po.id)} class="p-1.5 rounded text-outline hover:text-red-500 hover:bg-red-50 transition-colors" title="Batalkan">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                {:else if po.status === 'sent' || po.status === 'partial_received'}
                  <Button size="sm" onclick={() => openReceive(po)}>
                    <Package class="w-3 h-3" /> Terima Barang
                  </Button>
                {:else if po.status === 'received'}
                  <span class="text-xs text-status-success flex items-center gap-1">
                    <CheckCircle2 class="w-3 h-3" /> Selesai
                  </span>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </Card>
</div>

<!-- Create PO Modal -->
<Modal bind:open={showCreate} title="Purchase Order Baru">
  <div class="space-y-4">
    <Select label="Supplier" bind:value={createData.supplierId} options={supplierOptions} placeholder="Pilih supplier" />
    <Select label="Gudang Tujuan" bind:value={createData.warehouseId} options={warehouseOptions} placeholder="Pilih gudang" />
    <div class="grid grid-cols-2 gap-3">
      <Input label="No. PO (opsional)" placeholder="PO-2024-001" bind:value={createData.poNumber} />
      <Input label="Tanggal Expected" type="date" bind:value={createData.expectedAt} />
    </div>

    <div class="space-y-2">
      <span class="block text-xs font-semibold text-on-surface-variant uppercase tracking-wide">Item Pembelian</span>
      {#each lineItems as item, i}
        <div class="flex gap-2 items-end">
          <div class="flex-1">
            <Select bind:value={item.materialId} options={materialOptions} placeholder="Material" />
          </div>
          <div class="w-20">
            <Input label="" placeholder="Qty" type="number" bind:value={item.qty} />
          </div>
          <div class="w-28">
            <Input label="" placeholder="Harga/unit" type="number" bind:value={item.unitPrice} />
          </div>
          {#if lineItems.length > 1}
            <button type="button" onclick={() => removeLineItem(i)} class="p-2 text-outline hover:text-red-500 mb-0.5">×</button>
          {/if}
        </div>
      {/each}
      <button type="button" onclick={addLineItem} class="text-xs text-primary hover:text-primary/80 font-medium">+ Tambah Item</button>
    </div>
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showCreate = false }}>Batal</Button>
      <Button onclick={handleCreate} disabled={creating}>{creating ? 'Membuat...' : 'Buat PO'}</Button>
    </div>
  {/snippet}
</Modal>

<!-- Goods Receipt Modal -->
<Modal bind:open={showReceive} title="Terima Barang">
  {#if receiveDetail}
    <div class="space-y-4">
      <div class="text-sm text-on-surface-variant">
        PO: <span class="font-mono font-medium text-on-surface">{receiveDetail.po_number}</span>
        · Supplier: {receiveDetail.supplier_name}
      </div>
      <div class="space-y-3">
        {#each receiveDetail.items as item, i}
          <div class="p-3 rounded-lg bg-surface-container-low space-y-2">
            <p class="text-sm font-medium text-on-surface">{item.material_name} <span class="font-mono text-xs text-outline">({item.material_sku})</span></p>
            <p class="text-xs text-on-surface-variant">Dipesan: {item.qty_ordered} · Sudah diterima: {item.qty_received}</p>
            <div class="flex gap-2">
              <Input label="Qty Terima" type="number" bind:value={receiveLines[i].qty_received} />
              <Input label="No. Batch" placeholder="Auto" bind:value={receiveLines[i].batch_no} />
            </div>
          </div>
        {/each}
      </div>
      <Input label="Catatan Penerimaan" placeholder="Opsional..." bind:value={receiveNote} />
    </div>
  {/if}
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showReceive = false }}>Batal</Button>
      <Button onclick={handleReceive} disabled={receiving}>{receiving ? 'Menyimpan...' : 'Konfirmasi Penerimaan'}</Button>
    </div>
  {/snippet}
</Modal>
