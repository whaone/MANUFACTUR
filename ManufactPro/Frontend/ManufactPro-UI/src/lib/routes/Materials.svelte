<script lang="ts">
  import { onMount } from 'svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { compressImage } from '$lib/utils/image'
  import type { Material } from '$lib/types'
  import {
    Search,
    Plus,
    Package,
    Edit2,
    Trash2,
    QrCode,
  } from '@lucide/svelte'

  let materials = $state<Material[]>([])
  let isLoading = $state(true)
  let search = $state('')
  let filterCategory = $state('')
  let showForm = $state(false)
  let editingMaterial = $state<Material | null>(null)

  let formData = $state({
    name: '',
    sku: '',
    unit: 'pcs' as 'meter' | 'pcs' | 'gram' | 'liter' | 'kg' | 'lusin',
    category: '',
    min_stock: 0,
    barcode: '',
    is_active: true,
    image_url: '',
  })

  const categories = ['Tekstil', 'Aksesoris', 'Bahan Baku', 'Packaging']
  const units = [
    { value: 'meter', label: 'Meter' },
    { value: 'pcs', label: 'Pcs' },
    { value: 'gram', label: 'Gram' },
    { value: 'liter', label: 'Liter' },
    { value: 'kg', label: 'Kg' },
    { value: 'lusin', label: 'Lusin' },
  ]

  let filteredMaterials = $derived(
    materials.filter((m) => {
      const matchSearch = !search || m.name.toLowerCase().includes(search.toLowerCase()) || m.sku.toLowerCase().includes(search.toLowerCase())
      const matchCategory = !filterCategory || m.category === filterCategory
      return matchSearch && matchCategory
    })
  )

  onMount(async () => {
    await loadMaterials()
  })

  async function loadMaterials() {
    isLoading = true
    materials = await api.materials.list()
    isLoading = false
  }

  function openAddForm() {
    editingMaterial = null
    formData = { name: '', sku: '', unit: 'pcs', category: '', min_stock: 0, barcode: '', is_active: true, image_url: '' }
    showForm = true
  }

  function openEditForm(mat: Material) {
    editingMaterial = mat
    formData = {
      name: mat.name,
      sku: mat.sku,
      unit: mat.unit,
      category: mat.category,
      min_stock: mat.min_stock,
      barcode: mat.barcode,
      is_active: mat.is_active,
      image_url: mat.image_url || '',
    }
    showForm = true
  }

  async function handleImageUpload(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) return
    try {
      formData.image_url = await compressImage(file)
    } catch (err) {
      console.error('Failed to compress image', err)
    }
  }

  async function handleSave() {
    if (!formData.name || !formData.sku) return

    if (editingMaterial) {
      await api.materials.update(editingMaterial.id, formData)
    } else {
      await api.materials.create(formData)
    }
    showForm = false
    await loadMaterials()
  }

  async function handleDelete(id: string) {
    if (!confirm('Hapus material ini?')) return
    try {
      await api.materials.delete(id)
      await loadMaterials()
      toast.success('Material dihapus')
    } catch (e) {
      toast.error('Gagal menghapus material', e instanceof Error ? e.message : 'error')
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Materials</h1>
      <p class="text-on-surface-variant mt-1">Manage your raw materials and supplies</p>
    </div>
    <Button onclick={openAddForm}>
      <Plus class="w-4 h-4" />
      Add Material
    </Button>
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="flex flex-col md:flex-row gap-3">
        <div class="flex-1 relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
          <input
            type="text"
            bind:value={search}
            placeholder="Search by name or SKU..."
            class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
          />
        </div>
        <div class="flex gap-2">
          <select
            bind:value={filterCategory}
            class="px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 appearance-none cursor-pointer"
          >
            <option value="">All Categories</option>
            {#each categories as cat}
              <option value={cat}>{cat}</option>
            {/each}
          </select>
        </div>
      </div>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-outline-variant bg-surface-container-low">
            <th class="px-2 sm:px-4 py-3 text-left text-xs font-semibold text-on-surface-variant tracking-wider uppercase"></th>
            <th class="px-2 sm:px-4 py-3 text-left text-xs font-semibold text-on-surface-variant tracking-wider uppercase">SKU</th>
            <th class="px-2 sm:px-4 py-3 text-left text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Material Name</th>
            <th class="hidden md:table-cell px-4 py-3 text-left text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Category</th>
            <th class="hidden sm:table-cell px-4 py-3 text-center text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Unit</th>
            <th class="hidden md:table-cell px-4 py-3 text-right text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Min Stock</th>
            <th class="px-2 sm:px-4 py-3 text-center text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Status</th>
            <th class="px-2 sm:px-4 py-3 text-right text-xs font-semibold text-on-surface-variant tracking-wider uppercase">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr>
              <td colspan="8" class="px-4 py-12 text-center text-outline">Loading...</td>
            </tr>
          {:else if filteredMaterials.length === 0}
            <tr>
              <td colspan="8" class="px-4 py-12 text-center text-outline">
                <div class="flex flex-col items-center gap-2">
                  <Package class="w-8 h-8" />
                  <span>No materials found</span>
                </div>
              </td>
            </tr>
          {:else}
            {#each filteredMaterials as mat}
              <tr class="border-b border-outline-variant/30 last:border-0 hover:bg-surface-container-low/50 transition-colors">
                <td class="px-2 sm:px-4 py-3">
                  {#if mat.image_url}
                    <img src={mat.image_url} alt={mat.name} class="w-8 h-8 rounded object-cover border border-outline-variant/30" />
                  {:else}
                    <div class="w-8 h-8 rounded bg-surface-container-high flex items-center justify-center text-on-surface-variant">
                      <Package class="w-4 h-4" />
                    </div>
                  {/if}
                </td>
                <td class="px-2 sm:px-4 py-3 font-mono text-xs text-on-surface-variant">{mat.sku}</td>
                <td class="px-2 sm:px-4 py-3 font-medium text-on-surface">
                  <span class="line-clamp-2">{mat.name}</span>
                </td>
                <td class="hidden md:table-cell px-4 py-3">
                  <Badge variant="primary">{mat.category}</Badge>
                </td>
                <td class="hidden sm:table-cell px-4 py-3 text-center text-on-surface-variant">{mat.unit}</td>
                <td class="hidden md:table-cell px-4 py-3 text-right font-medium text-on-surface">{mat.min_stock}</td>
                <td class="px-2 sm:px-4 py-3 text-center">
                  {#if mat.is_active}
                    <Badge variant="success">Active</Badge>
                  {:else}
                    <Badge variant="neutral">Inactive</Badge>
                  {/if}
                </td>
                <td class="px-2 sm:px-4 py-3 text-right">
                  <div class="flex items-center justify-end gap-0.5 sm:gap-1">
                    <button onclick={() => openEditForm(mat)} class="p-1.5 rounded hover:bg-surface-container-high transition-colors cursor-pointer" title="Edit">
                      <Edit2 class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-on-surface-variant" />
                    </button>
                    <button class="p-1.5 rounded hover:bg-surface-container-high transition-colors cursor-pointer" title="Barcode">
                      <QrCode class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-on-surface-variant" />
                    </button>
                    <button onclick={() => handleDelete(mat.id)} class="p-1.5 rounded hover:bg-status-critical/10 transition-colors cursor-pointer" title="Delete">
                      <Trash2 class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-status-critical" />
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    {#if !isLoading}
      <div class="p-4 border-t border-outline-variant/30 flex items-center justify-between text-xs text-on-surface-variant">
        <span>{filteredMaterials.length} of {materials.length} materials</span>
      </div>
    {/if}
  </Card>
</div>

<Modal bind:open={showForm} title={editingMaterial ? 'Edit Material' : 'Add New Material'}>
  <div class="space-y-4">
    <div class="flex items-center gap-4">
      <div class="w-16 h-16 shrink-0 rounded-lg border border-outline-variant/50 overflow-hidden bg-surface-container flex items-center justify-center">
        {#if formData.image_url}
          <img src={formData.image_url} alt="Thumbnail" class="w-full h-full object-cover" />
        {:else}
          <Package class="w-8 h-8 text-on-surface-variant/50" />
        {/if}
      </div>
      <div class="flex-1">
        <label for="material-image" class="block text-sm font-medium text-on-surface mb-1">Material Image</label>
        <input 
          id="material-image"
          type="file" 
          accept="image/*" 
          onchange={handleImageUpload}
          class="block w-full text-sm text-on-surface-variant file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20 transition-colors"
        />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <Input label="Material Name" placeholder="e.g. Kain Cotton Combed 30s" bind:value={formData.name} required />
      <Input label="SKU" placeholder="e.g. FAB-001" bind:value={formData.sku} required />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <Select label="Category" bind:value={formData.category} options={categories.map(c => ({ value: c, label: c }))} placeholder="Select category" required />
      <Select label="Unit" bind:value={formData.unit} options={units} required />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <Input label="Minimum Stock" type="number" placeholder="0" bind:value={formData.min_stock} />
      <Input label="Barcode" placeholder="Auto-generated if empty" bind:value={formData.barcode} />
    </div>
  </div>

  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showForm = false }}>Cancel</Button>
      <Button onclick={handleSave}>
        {editingMaterial ? 'Update' : 'Create'} Material
      </Button>
    </div>
  {/snippet}
</Modal>
