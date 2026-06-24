<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import { Truck, Plus, Edit2, Trash2 } from '@lucide/svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import type { Supplier } from '$lib/types'

  let suppliers = $state<Supplier[]>([])
  let loading = $state(true)
  let showForm = $state(false)
  let editingId = $state<string | null>(null)
  let form = $state({ name: '', contact: '', email: '', phone: '', payment_term: '' })
  let saving = $state(false)

  onMount(async () => {
    try { suppliers = await api.suppliers.list() }
    catch (e) { toast.error('Gagal memuat supplier', e instanceof Error ? e.message : 'error') }
    finally { loading = false }
  })

  function openCreate() {
    editingId = null
    form = { name: '', contact: '', email: '', phone: '', payment_term: '' }
    showForm = true
  }

  function openEdit(s: Supplier) {
    editingId = s.id
    form = { name: s.name, contact: s.contact ?? '', email: s.email ?? '', phone: s.phone ?? '', payment_term: s.payment_term ?? '' }
    showForm = true
  }

  async function save() {
    if (!form.name) return
    saving = true
    try {
      if (editingId) {
        const updated = await api.suppliers.update(editingId, form)
        suppliers = suppliers.map(s => s.id === editingId ? updated : s)
        toast.success('Supplier diperbarui', updated.name)
      } else {
        const created = await api.suppliers.create(form)
        suppliers = [...suppliers, created]
        toast.success('Supplier ditambahkan', created.name)
      }
      showForm = false
    } catch (e) { toast.error('Gagal menyimpan supplier', e instanceof Error ? e.message : 'error') }
    finally { saving = false }
  }

  async function remove(id: string) {
    if (!confirm('Hapus supplier ini?')) return
    try {
      await api.suppliers.delete(id)
      suppliers = suppliers.filter(s => s.id !== id)
      toast.success('Supplier dihapus')
    } catch (e) { toast.error('Gagal menghapus supplier', e instanceof Error ? e.message : 'error') }
  }
</script>

<div class="p-6 space-y-6">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <Truck class="w-6 h-6 text-primary" />
      <div>
        <h1 class="text-xl font-semibold text-on-surface">Supplier</h1>
        <p class="text-sm text-on-surface-variant">{suppliers.length} supplier terdaftar</p>
      </div>
    </div>
    <Button onclick={openCreate}><Plus class="w-4 h-4 mr-1" /> Tambah Supplier</Button>
  </div>

  <Card>
    {#if loading}
      <div class="p-8 text-center text-on-surface-variant">Memuat...</div>
    {:else if suppliers.length === 0}
      <div class="p-8 text-center text-on-surface-variant">Belum ada supplier</div>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-outline-variant/20 text-left">
              <th class="p-3 font-medium text-on-surface-variant">Nama</th>
              <th class="p-3 font-medium text-on-surface-variant">Kontak</th>
              <th class="p-3 font-medium text-on-surface-variant">Email</th>
              <th class="p-3 font-medium text-on-surface-variant">Telepon</th>
              <th class="p-3 font-medium text-on-surface-variant">Termin</th>
              <th class="p-3"></th>
            </tr>
          </thead>
          <tbody>
            {#each suppliers as s}
              <tr class="border-b border-outline-variant/10 hover:bg-surface-container-low/40 transition-colors">
                <td class="p-3 font-medium text-on-surface">{s.name}</td>
                <td class="p-3 text-on-surface-variant">{s.contact ?? '—'}</td>
                <td class="p-3 text-on-surface-variant">{s.email ?? '—'}</td>
                <td class="p-3 text-on-surface-variant">{s.phone ?? '—'}</td>
                <td class="p-3">
                  {#if s.payment_term}
                    <span class="px-2 py-0.5 bg-surface-container-high text-on-surface-variant text-xs rounded-full">{s.payment_term}</span>
                  {:else}—{/if}
                </td>
                <td class="p-3">
                  <div class="flex gap-1 justify-end">
                    <button onclick={() => openEdit(s)} class="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors text-on-surface-variant">
                      <Edit2 class="w-4 h-4" />
                    </button>
                    <button onclick={() => remove(s.id)} class="p-1.5 rounded-lg hover:bg-status-critical/10 hover:text-status-critical transition-colors text-on-surface-variant">
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </Card>
</div>

<Modal open={showForm} title={editingId ? 'Edit Supplier' : 'Tambah Supplier'} onclose={() => showForm = false}>
  <div class="space-y-4 p-1">
    <Input label="Nama Supplier" bind:value={form.name} placeholder="PT Tekstil Jaya" />
    <Input label="Kontak" bind:value={form.contact} placeholder="Nama PIC" />
    <Input label="Email" type="email" bind:value={form.email} placeholder="email@supplier.com" />
    <Input label="Telepon" bind:value={form.phone} placeholder="021-xxx" />
    <Input label="Termin Pembayaran" bind:value={form.payment_term} placeholder="30 hari" />
    <div class="flex gap-3 pt-2">
      <Button variant="secondary" onclick={() => showForm = false} class="flex-1">Batal</Button>
      <Button onclick={save} disabled={saving} class="flex-1">{saving ? 'Menyimpan...' : 'Simpan'}</Button>
    </div>
  </div>
</Modal>
