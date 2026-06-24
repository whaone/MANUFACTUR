<script lang="ts">
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import type { Material, ProductVariant, BomItem } from '$lib/types'
  import { buildBomRecipes, type BomRecipe } from '$lib/utils/bom'
  import {
    ClipboardList,
    ChevronRight,
    Plus,
    Search,
    Package,
    Edit,
    Trash2,
    X,
  } from '@lucide/svelte'
  import { onMount } from 'svelte'

  interface MaterialRow {
    id: string
    materialId: string
    qty: number
    unit: string
    optional: boolean
  }

  let boms = $state<BomRecipe[]>([])
  let materials = $state<Material[]>([])
  let variants = $state<ProductVariant[]>([])
  let search = $state('')
  let expandedIdx = $state(-1)
  let showForm = $state(false)
  let showEditModal = $state(false)
  let showAddItemModal = $state(false)
  let activeVariantId = $state('')
  let editingItem: { bomVariantId: string; item: MaterialRow } | null = $state(null)
  let newBom = $state({ variantId: '', materialRows: [] as MaterialRow[] })
  let editItem = $state({ materialId: '', qty: 0, unit: 'pcs', optional: false })
  let addItem = $state({ materialId: '', qty: 0, unit: 'pcs', optional: false })

  let filtered = $derived(
    boms.filter((b) => !search || b.product.toLowerCase().includes(search.toLowerCase()) || b.variant.toLowerCase().includes(search.toLowerCase()))
  )

  let materialOptions = $derived(materials.map(m => ({ value: m.id, label: `${m.name} (${m.sku})` })))
  let variantOptions = $derived(variants.map(v => ({ value: v.id, label: v.sku })))
  let unitOptions = [
    { value: 'meter', label: 'Meter' },
    { value: 'pcs', label: 'Pcs' },
    { value: 'gram', label: 'Gram' },
    { value: 'liter', label: 'Liter' }
  ]

  onMount(async () => {
    await loadData()
  })

  async function loadData() {
    const [materialsData, variantsData] = await Promise.all([
      api.materials.list(),
      api.productVariants.list(),
    ])
    materials = materialsData
    variants = variantsData

    // Fetch every variant's BOM items in parallel — avoids the N+1 request
    // waterfall of awaiting each variant sequentially.
    const itemsByVariant = await Promise.all(
      variants.map(variant => api.bomItems.listByVariant(variant.id))
    )

    boms = buildBomRecipes(variants, materials, itemsByVariant)
  }

  function addMaterialRow() {
    newBom.materialRows = [...newBom.materialRows, { id: crypto.randomUUID(), materialId: '', qty: 0, unit: 'pcs', optional: false }]
  }

  function removeMaterialRow(id: string) {
    newBom.materialRows = newBom.materialRows.filter(r => r.id !== id)
  }

  async function handleCreateBom() {
    if (!newBom.variantId || newBom.materialRows.length === 0) return
    
    for (const row of newBom.materialRows) {
      if (!row.materialId || row.qty <= 0) continue
      await api.bomItems.create({
        product_variant_id: newBom.variantId,
        material_id: row.materialId,
        qty: row.qty,
        unit: row.unit,
        is_optional: row.optional
      })
    }
    
    await loadData()
    showForm = false
    newBom = { variantId: '', materialRows: [] }
  }

  function openEditModal(variantId: string, item: MaterialRow) {
    editingItem = { bomVariantId: variantId, item }
    editItem = { materialId: item.materialId, qty: item.qty, unit: item.unit, optional: item.optional }
    showEditModal = true
  }

  async function handleUpdateItem() {
    if (!editingItem) return
    await api.bomItems.update(editingItem.item.id, {
      material_id: editItem.materialId,
      qty: editItem.qty,
      unit: editItem.unit,
      is_optional: editItem.optional
    })
    await loadData()
    showEditModal = false
    editingItem = null
  }

  async function handleDeleteItem(itemId: string) {
    if (!confirm('Hapus item BOM ini?')) return
    try {
      await api.bomItems.delete(itemId)
      // Hapus dari state lokal langsung — tidak tergantung reload penuh (yang bisa lambat/gagal).
      boms = boms.map(b => {
        const items = b.items.filter(it => it.id !== itemId)
        return { ...b, items, totalMaterials: items.length }
      })
      toast.success('Item BOM dihapus')
    } catch (e) {
      toast.error('Gagal menghapus item BOM', e instanceof Error ? e.message : 'error')
    }
  }

  function openAddItemModal(variantId: string) {
    activeVariantId = variantId
    addItem = { materialId: '', qty: 0, unit: 'pcs', optional: false }
    showAddItemModal = true
  }

  async function handleAddItemToBom() {
    if (!addItem.materialId || addItem.qty <= 0) return
    await api.bomItems.create({
      product_variant_id: activeVariantId,
      material_id: addItem.materialId,
      qty: addItem.qty,
      unit: addItem.unit,
      is_optional: addItem.optional
    })
    await loadData()
    showAddItemModal = false
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Bill of Materials</h1>
      <p class="text-on-surface-variant mt-1">Production recipes per product variant</p>
    </div>
    <Button onclick={() => { showForm = true }}>
      <Plus class="w-4 h-4" />
      New BOM
    </Button>
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="relative w-full">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
        <input
          type="text"
          bind:value={search}
          placeholder="Search BOM by product or variant..."
          class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
        />
      </div>
    </div>

    <div class="divide-y divide-outline-variant/30">
      {#each filtered as bom, i}
        <div>
          <button
            onclick={() => { expandedIdx = expandedIdx === i ? -1 : i }}
            class="w-full flex items-center gap-4 px-4 py-3 hover:bg-surface-container-low/50 transition-colors cursor-pointer text-left"
          >
            <div class="p-2 rounded-lg bg-primary/10">
              <ClipboardList class="w-5 h-5 text-primary" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-on-surface">{bom.product}</p>
              <p class="text-xs text-on-surface-variant">Variant: {bom.variant}</p>
            </div>
            <Badge variant="info">{bom.totalMaterials} materials</Badge>
            {#if expandedIdx === i}
              <ChevronRight class="w-5 h-5 text-on-surface-variant rotate-90" />
            {:else}
              <ChevronRight class="w-5 h-5 text-on-surface-variant" />
            {/if}
          </button>

          {#if expandedIdx === i}
            <div class="px-4 pb-4">
              <div class="ml-12 rounded-xl border border-outline-variant/30 bg-surface-container-lowest overflow-hidden">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="bg-surface-container-low">
                      <th class="px-4 py-2.5 text-left text-xs font-semibold text-on-surface-variant uppercase">Material</th>
                      <th class="px-4 py-2.5 text-right text-xs font-semibold text-on-surface-variant uppercase">Qty</th>
                      <th class="px-4 py-2.5 text-left text-xs font-semibold text-on-surface-variant uppercase">Unit</th>
                      <th class="px-4 py-2.5 text-center text-xs font-semibold text-on-surface-variant uppercase">Required</th>
                      <th class="px-4 py-2.5 text-right text-xs font-semibold text-on-surface-variant uppercase">
                        <Button variant="ghost" size="sm" class="h-6 px-2 py-0 text-xs gap-1" onclick={() => openAddItemModal(bom.variantId)}>
                          <Plus class="w-3 h-3" />
                          Add
                        </Button>
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each bom.items as item}
                      <tr class="border-t border-outline-variant/20 hover:bg-surface-container-low/30">
                        <td class="px-4 py-2.5 flex items-center gap-2">
                          <Package class="w-4 h-4 text-outline shrink-0" />
                          <span class="font-medium text-on-surface">{item.material}</span>
                        </td>
                        <td class="px-4 py-2.5 text-right font-mono font-medium text-on-surface">{item.qty}</td>
                        <td class="px-4 py-2.5 text-on-surface-variant">{item.unit}</td>
                        <td class="px-4 py-2.5 text-center">
                          {#if item.optional}
                            <Badge variant="warning">Optional</Badge>
                          {:else}
                            <Badge variant="success">Required</Badge>
                          {/if}
                        </td>
                        <td class="px-4 py-2.5 text-right">
                          <div class="flex items-center justify-end gap-2">
                            <button class="p-1.5 text-outline hover:text-primary hover:bg-primary/10 rounded-lg transition-colors cursor-pointer" onclick={() => openEditModal(bom.variantId, item)}>
                              <Edit class="w-4 h-4" />
                            </button>
                            <button class="p-1.5 text-outline hover:text-error hover:bg-error/10 rounded-lg transition-colors cursor-pointer" onclick={() => handleDeleteItem(item.id)}>
                              <Trash2 class="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </Card>
</div>

<Modal bind:open={showForm} title="New BOM" size="lg">
  <div class="space-y-6">
    <Select label="Product Variant" bind:value={newBom.variantId} options={variantOptions} required />
    
    <div>
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-sm font-semibold text-on-surface-variant uppercase tracking-wide">Materials</h3>
        <Button variant="secondary" size="sm" onclick={addMaterialRow}>
          <Plus class="w-4 h-4" />
          Add Material
        </Button>
      </div>

      {#if newBom.materialRows.length === 0}
        <div class="text-center py-8 text-on-surface-variant bg-surface-container-low/50 rounded-xl border border-dashed border-outline-variant">
          No materials added yet. Click "Add Material" to start building the recipe.
        </div>
      {:else}
        <div class="space-y-3">
          {#each newBom.materialRows as row (row.id)}
            <div class="flex items-end gap-3 p-3 rounded-xl bg-surface-container-low border border-outline-variant/30">
              <div class="flex-1 min-w-[200px]">
                <Select label="Material" bind:value={row.materialId} options={materialOptions} />
              </div>
              <div class="w-24">
                <Input label="Qty" type="number" bind:value={row.qty} />
              </div>
              <div class="w-32">
                <Select label="Unit" bind:value={row.unit} options={unitOptions} />
              </div>
              <div class="w-32 flex items-center h-[38px] px-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" bind:checked={row.optional} class="w-4 h-4 rounded text-primary focus:ring-primary" />
                  <span class="text-sm text-on-surface-variant">Optional</span>
                </label>
              </div>
              <button 
                class="p-2 text-outline hover:text-error hover:bg-error/10 rounded-lg transition-colors cursor-pointer shrink-0 h-[38px]"
                onclick={() => removeMaterialRow(row.id)}
              >
                <X class="w-5 h-5" />
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showForm = false }}>Cancel</Button>
      <Button onclick={handleCreateBom} disabled={!newBom.variantId || newBom.materialRows.length === 0}>Save BOM</Button>
    </div>
  {/snippet}
</Modal>

<Modal bind:open={showEditModal} title="Edit BOM Item">
  <div class="space-y-4">
    <Select label="Material" bind:value={editItem.materialId} options={materialOptions} required />
    <div class="grid grid-cols-2 gap-4">
      <Input label="Quantity" type="number" bind:value={editItem.qty} required />
      <Select label="Unit" bind:value={editItem.unit} options={unitOptions} required />
    </div>
    <label class="flex items-center gap-2 cursor-pointer mt-2">
      <input type="checkbox" bind:checked={editItem.optional} class="w-4 h-4 rounded text-primary focus:ring-primary" />
      <span class="text-sm text-on-surface">Mark as Optional</span>
    </label>
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showEditModal = false }}>Cancel</Button>
      <Button onclick={handleUpdateItem}>Save Changes</Button>
    </div>
  {/snippet}
</Modal>

<Modal bind:open={showAddItemModal} title="Add Material to BOM">
  <div class="space-y-4">
    <Select label="Material" bind:value={addItem.materialId} options={materialOptions} required />
    <div class="grid grid-cols-2 gap-4">
      <Input label="Quantity" type="number" bind:value={addItem.qty} required />
      <Select label="Unit" bind:value={addItem.unit} options={unitOptions} required />
    </div>
    <label class="flex items-center gap-2 cursor-pointer mt-2">
      <input type="checkbox" bind:checked={addItem.optional} class="w-4 h-4 rounded text-primary focus:ring-primary" />
      <span class="text-sm text-on-surface">Mark as Optional</span>
    </label>
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showAddItemModal = false }}>Cancel</Button>
      <Button onclick={handleAddItemToBom} disabled={!addItem.materialId || addItem.qty <= 0}>Add Material</Button>
    </div>
  {/snippet}
</Modal>
