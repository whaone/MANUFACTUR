<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import { Warehouse, Plus, Building2 } from '@lucide/svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'

  interface Branch { id: string; name: string; address: string }
  interface WarehouseItem { id: string; branch_id: string; name: string; code: string; is_default: boolean }

  let branches = $state<Branch[]>([])
  let warehouses = $state<WarehouseItem[]>([])
  let loading = $state(true)
  let showBranchForm = $state(false)
  let showWarehouseForm = $state(false)
  let branchForm = $state({ name: '', address: '' })
  let warehouseForm = $state({ branch_id: '', name: '', code: '', is_default: false })
  let saving = $state(false)

  onMount(async () => {
    try {
      [branches, warehouses] = await Promise.all([api.warehouses.listBranches(), api.warehouses.list()])
    } catch (e) { toast.error('Gagal memuat data gudang', e instanceof Error ? e.message : 'error') }
    finally { loading = false }
  })

  async function saveBranch() {
    if (!branchForm.name) return
    saving = true
    try {
      const b = await api.warehouses.createBranch(branchForm)
      branches = [...branches, b]
      showBranchForm = false
      branchForm = { name: '', address: '' }
      toast.success('Cabang ditambahkan', b.name)
    } catch (e) { toast.error('Gagal menyimpan cabang', e instanceof Error ? e.message : 'error') }
    finally { saving = false }
  }

  async function saveWarehouse() {
    if (!warehouseForm.name || !warehouseForm.code || !warehouseForm.branch_id) return
    saving = true
    try {
      const w = await api.warehouses.create(warehouseForm)
      warehouses = [...warehouses, w]
      showWarehouseForm = false
      warehouseForm = { branch_id: '', name: '', code: '', is_default: false }
      toast.success('Gudang ditambahkan', w.name)
    } catch (e) { toast.error('Gagal menyimpan gudang', e instanceof Error ? e.message : 'error') }
    finally { saving = false }
  }

  function getBranchName(id: string) {
    return branches.find(b => b.id === id)?.name ?? '—'
  }
</script>

<div class="p-6 space-y-6">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <Warehouse class="w-6 h-6 text-primary" />
      <div>
        <h1 class="text-xl font-semibold text-on-surface">Gudang & Cabang</h1>
        <p class="text-sm text-on-surface-variant">{branches.length} cabang, {warehouses.length} gudang</p>
      </div>
    </div>
    <div class="flex gap-2">
      <Button variant="secondary" onclick={() => showBranchForm = true}>
        <Building2 class="w-4 h-4 mr-1" /> Cabang Baru
      </Button>
      <Button onclick={() => showWarehouseForm = true} disabled={branches.length === 0}>
        <Plus class="w-4 h-4 mr-1" /> Gudang Baru
      </Button>
    </div>
  </div>

  {#if loading}
    <div class="p-8 text-center text-on-surface-variant">Memuat...</div>
  {:else}
    <!-- Branches -->
    <Card>
      <div class="p-4 border-b border-outline-variant/30">
        <h2 class="font-medium text-on-surface">Daftar Cabang</h2>
      </div>
      {#if branches.length === 0}
        <div class="p-8 text-center text-on-surface-variant">Belum ada cabang. Buat cabang dulu sebelum tambah gudang.</div>
      {:else}
        <div class="divide-y divide-outline-variant/10">
          {#each branches as b}
            <div class="flex items-center gap-4 p-4">
              <Building2 class="w-5 h-5 text-primary shrink-0" />
              <div class="flex-1">
                <div class="font-medium text-on-surface">{b.name}</div>
                {#if b.address}<div class="text-sm text-on-surface-variant">{b.address}</div>{/if}
              </div>
              <span class="text-xs text-on-surface-variant">
                {warehouses.filter(w => w.branch_id === b.id).length} gudang
              </span>
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <!-- Warehouses -->
    <Card>
      <div class="p-4 border-b border-outline-variant/30">
        <h2 class="font-medium text-on-surface">Daftar Gudang</h2>
      </div>
      {#if warehouses.length === 0}
        <div class="p-8 text-center text-on-surface-variant">Belum ada gudang</div>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-outline-variant/20 text-left">
                <th class="p-3 font-medium text-on-surface-variant">Nama</th>
                <th class="p-3 font-medium text-on-surface-variant">Kode</th>
                <th class="p-3 font-medium text-on-surface-variant">Cabang</th>
                <th class="p-3 font-medium text-on-surface-variant">Default</th>
              </tr>
            </thead>
            <tbody>
              {#each warehouses as w}
                <tr class="border-b border-outline-variant/10 hover:bg-surface-container-low/40">
                  <td class="p-3 font-medium text-on-surface">{w.name}</td>
                  <td class="p-3 text-on-surface-variant font-mono">{w.code}</td>
                  <td class="p-3 text-on-surface-variant">{getBranchName(w.branch_id)}</td>
                  <td class="p-3">
                    {#if w.is_default}
                      <span class="px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-full">Default</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </Card>
  {/if}
</div>

<Modal open={showBranchForm} title="Tambah Cabang" onclose={() => showBranchForm = false}>
  <div class="space-y-4 p-1">
    <Input label="Nama Cabang" bind:value={branchForm.name} placeholder="Cabang Utama" />
    <Input label="Alamat" bind:value={branchForm.address} placeholder="Jl. ..." />
    <div class="flex gap-3 pt-2">
      <Button variant="secondary" onclick={() => showBranchForm = false} class="flex-1">Batal</Button>
      <Button onclick={saveBranch} disabled={saving} class="flex-1">Simpan</Button>
    </div>
  </div>
</Modal>

<Modal open={showWarehouseForm} title="Tambah Gudang" onclose={() => showWarehouseForm = false}>
  <div class="space-y-4 p-1">
    <div class="space-y-1.5">
      <label class="text-sm font-medium text-on-surface">Cabang</label>
      <select bind:value={warehouseForm.branch_id}
        class="w-full rounded-xl border border-outline-variant bg-surface-container-lowest px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/40">
        <option value="">Pilih cabang...</option>
        {#each branches as b}
          <option value={b.id}>{b.name}</option>
        {/each}
      </select>
    </div>
    <Input label="Nama Gudang" bind:value={warehouseForm.name} placeholder="Gudang Utama" />
    <Input label="Kode" bind:value={warehouseForm.code} placeholder="GDG-01" />
    <label class="flex items-center gap-2 text-sm text-on-surface">
      <input type="checkbox" bind:checked={warehouseForm.is_default} class="rounded" />
      Set sebagai gudang default
    </label>
    <div class="flex gap-3 pt-2">
      <Button variant="secondary" onclick={() => showWarehouseForm = false} class="flex-1">Batal</Button>
      <Button onclick={saveWarehouse} disabled={saving} class="flex-1">Simpan</Button>
    </div>
  </div>
</Modal>
