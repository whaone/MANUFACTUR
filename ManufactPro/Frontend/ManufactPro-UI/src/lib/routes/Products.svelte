<script lang="ts">
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Modal from '$lib/components/ui/Modal.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { onMount } from 'svelte'
  import { api } from '$lib/api'
  import { toast } from '$lib/stores/toast'
  import {
    Search,
    Plus,
    Boxes,
    ChevronDown,
    ChevronRight,
    Edit2,
    Trash2,
    Wand2
  } from '@lucide/svelte'

  interface Variant {
    id: string
    sku: string
    barcode: string
    attributes: Record<string, string>
    sell_price: number
    is_active: boolean
  }

  interface ProductItem {
    id: string
    name: string
    category: string
    description: string
    variants: Variant[]
    expanded: boolean
  }

  let products = $state<ProductItem[]>([])
  let loading = $state(true)

  onMount(async () => {
    try {
      const data = await api.products.list()
      products = data.map((p: any) => ({
        id: p.id,
        name: p.name,
        category: p.category ?? '',
        description: p.description ?? '',
        variants: (p.variants ?? []).map((v: any) => ({
          id: v.id,
          sku: v.sku,
          barcode: v.barcode ?? '',
          attributes: v.attributes ?? {},
          sell_price: v.sell_price,
          is_active: v.is_active,
        })),
        expanded: false,
      }))
    } catch (e) {
      console.error('Failed to load products:', e)
    } finally {
      loading = false
    }
  })

  let search = $state('')
  let showForm = $state(false)
  let formData = $state({ name: '', category: '', description: '' })

  let filtered = $derived(
    products.filter((p) => !search || p.name.toLowerCase().includes(search.toLowerCase()))
  )

  type AttrRow = { key: string; value: string }

  const CATEGORY_PRESETS: Record<string, AttrRow[]> = {
    pakaian:   [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
    tekstil:   [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
    clothing:  [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
    makanan:   [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
    food:      [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
    'f&b':     [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
    minuman:   [{ key: 'Rasa', value: '' }, { key: 'Volume (ml)', value: '' }],
    beverage:  [{ key: 'Rasa', value: '' }, { key: 'Volume (ml)', value: '' }],
    elektronik:[{ key: 'Warna', value: '' }, { key: 'Kapasitas', value: '' }],
  }

  function getCategoryPreset(cat: string): AttrRow[] {
    const lower = cat.toLowerCase()
    for (const [key, rows] of Object.entries(CATEGORY_PRESETS)) {
      if (lower.includes(key)) return rows.map(r => ({ ...r }))
    }
    return [{ key: '', value: '' }]
  }

  let showVariantForm = $state(false)
  let variantFormData = $state({ sku: '', barcode: '', price: 0, status: 'Active' })
  let attrRows = $state<AttrRow[]>([])
  let activeProductId = $state('')
  let editingVariantId = $state('')

  function addAttrRow() { attrRows = [...attrRows, { key: '', value: '' }] }
  function removeAttrRow(i: number) { attrRows = attrRows.filter((_, idx) => idx !== i) }

  // --- Matrix Generator ---
  type Dimension = { key: string; values: string[]; inputVal: string }

  function generateCombos(dims: Dimension[]): Record<string, string>[] {
    const valid = dims.filter(d => d.key.trim() && d.values.length > 0)
    if (valid.length === 0) return []
    const [first, ...rest] = valid
    const restCombos = rest.length > 0 ? generateCombos(rest) : [{}]
    return first.values.flatMap(v =>
      restCombos.map(combo => ({ [first.key.trim()]: v, ...combo }))
    )
  }

  function buildSku(productName: string, attrs: Record<string, string>): string {
    const prefix = productName.replace(/\s+/g, '').slice(0, 3).toUpperCase()
    const parts = Object.values(attrs).map(v => v.replace(/\s+/g, '').slice(0, 3).toUpperCase())
    return [prefix, ...parts].join('-')
  }

  let showMatrixForm = $state(false)
  let dimensions = $state<Dimension[]>([])
  let matrixBasePrice = $state(0)
  let matrixGenerating = $state(false)

  let matrixCombos = $derived(generateCombos(dimensions))
  let matrixSelected = $state<boolean[]>([])

  $effect(() => {
    matrixSelected = matrixCombos.map(() => true)
  })

  function openMatrixForm(productId: string) {
    activeProductId = productId
    const cat = products.find(p => p.id === productId)?.category ?? ''
    const preset = getCategoryPreset(cat)
    dimensions = preset.map(r => ({ key: r.key, values: [], inputVal: '' }))
    matrixBasePrice = 0
    showMatrixForm = true
  }

  function addDimensionValue(i: number) {
    const val = dimensions[i].inputVal.trim()
    if (!val || dimensions[i].values.includes(val)) return
    dimensions[i].values = [...dimensions[i].values, val]
    dimensions[i].inputVal = ''
  }

  function removeDimensionValue(i: number, v: string) {
    dimensions[i].values = dimensions[i].values.filter(x => x !== v)
  }

  function addDimension() {
    dimensions = [...dimensions, { key: '', values: [], inputVal: '' }]
  }

  function removeDimension(i: number) {
    dimensions = dimensions.filter((_, idx) => idx !== i)
  }

  async function handleBulkCreate() {
    if (!activeProductId || matrixGenerating) return
    matrixGenerating = true
    const toCreate = matrixCombos.filter((_, i) => matrixSelected[i])
    const product = products.find(p => p.id === activeProductId)
    const created: Variant[] = []
    for (const attrs of toCreate) {
      const sku = buildSku(product?.name ?? 'PRD', attrs)
      try {
        const v = await api.productVariants.create(activeProductId, {
          sku,
          barcode: '',
          sell_price: matrixBasePrice,
          is_active: true,
          attributes: attrs,
        })
        created.push(v)
      } catch {
        // skip on duplicate SKU — user edit manual after generate
      }
    }
    products = products.map(p =>
      p.id === activeProductId ? { ...p, variants: [...p.variants, ...created] } : p
    )
    matrixGenerating = false
    showMatrixForm = false
  }

  function toggleExpand(id: string) {
    products = products.map((p) => p.id === id ? { ...p, expanded: !p.expanded } : p)
  }

  async function handleSave() {
    if (!formData.name) return
    try {
      const p = await api.products.create({
        name: formData.name,
        category: formData.category,
        description: formData.description,
        image_url: '',
      })
      products = [...products, { id: p.id, name: p.name, category: p.category ?? '', description: p.description ?? '', variants: [], expanded: false }]
    } catch (e) {
      console.error('Failed to create product:', e)
    }
    showForm = false
    formData = { name: '', category: '', description: '' }
  }

  function openVariantForm(productId: string, variant?: Variant) {
    activeProductId = productId
    if (variant) {
      editingVariantId = variant.id
      variantFormData = { sku: variant.sku, barcode: variant.barcode, price: variant.sell_price, status: variant.is_active ? 'Active' : 'Inactive' }
      attrRows = Object.entries(variant.attributes).map(([key, value]) => ({ key, value }))
      if (attrRows.length === 0) attrRows = [{ key: '', value: '' }]
    } else {
      editingVariantId = ''
      variantFormData = { sku: '', barcode: '', price: 0, status: 'Active' }
      const cat = products.find(p => p.id === productId)?.category ?? ''
      attrRows = getCategoryPreset(cat)
    }
    showVariantForm = true
  }

  async function handleSaveVariant() {
    if (!variantFormData.sku || !activeProductId) return

    const attrs: Record<string, string> = {}
    attrRows.filter(r => r.key.trim() && r.value.trim())
      .forEach(r => { attrs[r.key.trim()] = r.value.trim() })

    const is_active = variantFormData.status === 'Active'
    
    if (editingVariantId) {
      const updatedData = {
        sku: variantFormData.sku,
        barcode: variantFormData.barcode,
        sell_price: variantFormData.price,
        is_active,
        attributes: attrs
      }
      await api.productVariants.update(editingVariantId, updatedData)
      
      products = products.map(p => {
        if (p.id === activeProductId) {
          return {
            ...p,
            variants: p.variants.map(v => v.id === editingVariantId ? { ...v, ...updatedData } : v)
          }
        }
        return p
      })
    } else {
      const newData = {
        sku: variantFormData.sku,
        barcode: variantFormData.barcode,
        sell_price: variantFormData.price,
        is_active,
        attributes: attrs
      }
      const newVariant = await api.productVariants.create(activeProductId, newData)
      
      products = products.map(p => {
        if (p.id === activeProductId) {
          return {
            ...p,
            variants: [...p.variants, newVariant]
          }
        }
        return p
      })
    }
    showVariantForm = false
  }

  async function handleDeleteVariant(productId: string, variantId: string) {
    if (!confirm('Are you sure you want to delete this variant?')) return
    await api.productVariants.delete(variantId)
    products = products.map(p => {
      if (p.id === productId) {
        return {
          ...p,
          variants: p.variants.filter(v => v.id !== variantId)
        }
      }
      return p
    })
  }

  async function handleDeleteProduct(productId: string, name: string) {
    if (!confirm(`Hapus produk "${name}" beserta semua varian & resep BOM-nya?`)) return
    try {
      await api.products.delete(productId)
      products = products.filter(p => p.id !== productId)
      toast.success('Produk dihapus', name)
    } catch (e) {
      toast.error('Gagal menghapus produk', e instanceof Error ? e.message : 'error')
    }
  }

  const categories = ['Apparel', 'F&B', 'Kerajinan', 'Lainnya']
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Products</h1>
      <p class="text-on-surface-variant mt-1">Manage products and variants</p>
    </div>
    <Button onclick={() => { formData = { name: '', category: '', description: '' }; showForm = true }}>
      <Plus class="w-4 h-4" />
      Add Product
    </Button>
  </div>

  <Card variant="glass" padding="none">
    <div class="p-4 border-b border-outline-variant/30">
      <div class="relative w-full">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-outline" />
        <input
          type="text"
          bind:value={search}
          placeholder="Search products..."
          class="w-full pl-10 pr-4 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
        />
      </div>
    </div>

    <div class="divide-y divide-outline-variant/30">
      {#if filtered.length === 0}
        <div class="px-4 py-12 text-center text-outline">
          <Boxes class="w-8 h-8 mx-auto mb-2" />
          <span>No products found</span>
        </div>
      {:else}
        {#each filtered as product}
          <div>
            <div class="w-full flex items-center gap-4 px-4 py-3 hover:bg-surface-container-low/50 transition-colors">
              <button
                onclick={() => toggleExpand(product.id)}
                class="flex items-center gap-4 flex-1 min-w-0 cursor-pointer text-left"
              >
                <div class="p-2 rounded-lg bg-primary/10">
                  <Boxes class="w-5 h-5 text-primary" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-on-surface">{product.name}</p>
                  <p class="text-xs text-on-surface-variant">{product.category} · {product.variants.length} variant(s)</p>
                </div>
                <Badge variant="info">{product.variants.length} variants</Badge>
                {#if product.expanded}
                  <ChevronDown class="w-5 h-5 text-on-surface-variant" />
                {:else}
                  <ChevronRight class="w-5 h-5 text-on-surface-variant" />
                {/if}
              </button>
              <button
                onclick={() => handleDeleteProduct(product.id, product.name)}
                class="p-2 rounded-lg hover:bg-status-critical/10 transition-colors cursor-pointer shrink-0"
                title="Hapus produk"
              >
                <Trash2 class="w-4 h-4 text-status-critical" />
              </button>
            </div>

            {#if product.expanded}
              <div class="px-4 pb-3">
                <div class="flex justify-between items-center mb-2 ml-12">
                  <span class="text-xs text-on-surface-variant font-medium">Variant List</span>
                  <div class="flex gap-2">
                    <Button size="sm" variant="secondary" onclick={() => openMatrixForm(product.id)}>
                      <Wand2 class="w-3 h-3" />
                      Generate
                    </Button>
                    <Button size="sm" onclick={() => openVariantForm(product.id)}>
                      <Plus class="w-3 h-3" />
                      Add Variant
                    </Button>
                  </div>
                </div>
                {#if product.variants.length > 0}
                <div class="ml-12 rounded-lg border border-outline-variant/30 overflow-hidden">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="bg-surface-container-low">
                        <th class="px-3 py-2 text-left text-xs font-semibold text-on-surface-variant uppercase">SKU</th>
                        <th class="px-3 py-2 text-left text-xs font-semibold text-on-surface-variant uppercase">Attributes</th>
                        <th class="px-3 py-2 text-right text-xs font-semibold text-on-surface-variant uppercase">Price</th>
                        <th class="px-3 py-2 text-center text-xs font-semibold text-on-surface-variant uppercase">Status</th>
                        <th class="px-3 py-2 text-center text-xs font-semibold text-on-surface-variant uppercase">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each product.variants as variant}
                        <tr class="border-t border-outline-variant/20">
                          <td class="px-3 py-2 font-mono text-xs text-on-surface-variant">{variant.sku}</td>
                          <td class="px-3 py-2">
                            <div class="flex flex-wrap gap-1">
                              {#each Object.entries(variant.attributes) as [key, val]}
                                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-surface-container-high text-xs">
                                  <span class="text-outline">{key}:</span>
                                  <span class="font-medium text-on-surface">{val}</span>
                                </span>
                              {/each}
                            </div>
                          </td>
                          <td class="px-3 py-2 text-right font-medium text-on-surface">
                            Rp {variant.sell_price.toLocaleString('id-ID')}
                          </td>
                          <td class="px-3 py-2 text-center">
                            {#if variant.is_active}
                              <Badge variant="success">Active</Badge>
                            {:else}
                              <Badge variant="neutral">Inactive</Badge>
                            {/if}
                          </td>
                          <td class="px-3 py-2 text-center">
                            <div class="flex items-center justify-center gap-1">
                              <button
                                onclick={() => openVariantForm(product.id, variant)}
                                class="p-1 rounded hover:bg-surface-container-high text-on-surface-variant hover:text-primary transition-colors"
                                title="Edit variant"
                              >
                                <Edit2 class="w-3.5 h-3.5" />
                              </button>
                              <button
                                onclick={() => handleDeleteVariant(product.id, variant.id)}
                                class="p-1 rounded hover:bg-surface-container-high text-on-surface-variant hover:text-red-500 transition-colors"
                                title="Delete variant"
                              >
                                <Trash2 class="w-3.5 h-3.5" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
                {:else}
                  <div class="ml-12 text-center py-4 text-xs text-outline">
                    No variants yet. Add your first variant.
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      {/if}
    </div>
  </Card>
</div>

<Modal bind:open={showForm} title="Add New Product">
  <div class="space-y-4">
    <Input label="Product Name" placeholder="e.g. Kaos Polos" bind:value={formData.name} required />
    <Select label="Category" bind:value={formData.category} options={categories.map(c => ({ value: c, label: c }))} placeholder="Select category" />
    <Input label="Description" placeholder="Product description..." bind:value={formData.description} />
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showForm = false }}>Cancel</Button>
      <Button onclick={handleSave}>Create Product</Button>
    </div>
  {/snippet}
</Modal>

<Modal bind:open={showMatrixForm} title="Generate Variants">
  <div class="space-y-4">
    {#each dimensions as dim, i}
      <div class="space-y-1.5">
        <div class="flex items-center gap-2">
          <input
            bind:value={dim.key}
            placeholder="Nama dimensi (e.g. Ukuran)"
            class="flex-1 px-3 py-1.5 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
          />
          <button type="button" onclick={() => removeDimension(i)} class="p-1.5 text-outline hover:text-red-500 transition-colors">×</button>
        </div>
        <div class="flex flex-wrap items-center gap-1.5 pl-1">
          {#each dim.values as val}
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-primary/10 text-primary text-xs font-medium">
              {val}
              <button type="button" onclick={() => removeDimensionValue(i, val)} class="hover:text-red-500 leading-none">×</button>
            </span>
          {/each}
          <input
            bind:value={dim.inputVal}
            placeholder="Tambah nilai + Enter"
            onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addDimensionValue(i) } }}
            onblur={() => addDimensionValue(i)}
            class="px-2 py-0.5 rounded text-sm bg-transparent border-b border-outline-variant focus:outline-none focus:border-primary w-36"
          />
        </div>
      </div>
    {/each}

    <button type="button" onclick={addDimension} class="text-xs text-primary hover:text-primary/80 font-medium">+ Dimensi Baru</button>

    <div class="pt-1 border-t border-outline-variant/30">
      <span class="block text-xs font-semibold text-on-surface-variant uppercase tracking-wide mb-1.5">Harga Dasar (Rp)</span>
      <input
        type="number"
        bind:value={matrixBasePrice}
        placeholder="0"
        class="w-full px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
      />
    </div>

    {#if matrixCombos.length > 0}
      <div class="border-t border-outline-variant/30 pt-2">
        <p class="text-xs font-semibold text-on-surface-variant uppercase tracking-wide mb-2">
          Preview {matrixCombos.length} kombinasi
        </p>
        <div class="max-h-48 overflow-y-auto space-y-1 pr-1">
          {#each matrixCombos as combo, i}
            <label class="flex items-center gap-2 cursor-pointer hover:bg-surface-container-low rounded px-2 py-1">
              <input type="checkbox" bind:checked={matrixSelected[i]} class="accent-primary" />
              <span class="text-sm text-on-surface">
                {Object.values(combo).join(' · ')}
              </span>
              <span class="ml-auto font-mono text-xs text-outline">
                {buildSku(products.find(p => p.id === activeProductId)?.name ?? '', combo)}
              </span>
            </label>
          {/each}
        </div>
      </div>
    {:else}
      <p class="text-xs text-outline text-center py-2">Tambahkan dimensi dan nilai untuk melihat preview kombinasi</p>
    {/if}
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-between gap-3">
      <span class="text-xs text-outline">
        {matrixSelected.filter(Boolean).length} dari {matrixCombos.length} variant akan dibuat
      </span>
      <div class="flex gap-2">
        <Button variant="secondary" onclick={() => { showMatrixForm = false }}>Batal</Button>
        <Button
          onclick={handleBulkCreate}
          disabled={matrixGenerating || matrixSelected.filter(Boolean).length === 0}
        >
          {matrixGenerating ? 'Membuat...' : `Buat ${matrixSelected.filter(Boolean).length} Variant`}
        </Button>
      </div>
    </div>
  {/snippet}
</Modal>

<Modal bind:open={showVariantForm} title={editingVariantId ? 'Edit Variant' : 'Add New Variant'}>
  <div class="space-y-4">
    <Input label="SKU" placeholder="e.g. KP-M-MERAH" bind:value={variantFormData.sku} required />
    <Input label="Barcode" placeholder="e.g. 8991001000001" bind:value={variantFormData.barcode} />
    <div class="space-y-2">
      <span class="block text-xs font-semibold text-on-surface-variant uppercase tracking-wide">Atribut</span>
      {#each attrRows as row, i}
        <div class="flex items-center gap-2">
          <input
            bind:value={row.key}
            placeholder="Ukuran"
            class="w-28 flex-shrink-0 px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
          />
          <span class="text-on-surface-variant text-sm">:</span>
          <input
            bind:value={row.value}
            placeholder="M"
            class="flex-1 px-3 py-2 rounded-lg text-sm bg-surface-container-low border border-outline-variant focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
          />
          <button
            type="button"
            onclick={() => removeAttrRow(i)}
            class="p-1.5 rounded-lg text-outline hover:text-red-500 hover:bg-red-50 transition-colors"
            title="Hapus atribut"
          >×</button>
        </div>
      {/each}
      <button
        type="button"
        onclick={addAttrRow}
        class="text-xs text-primary hover:text-primary/80 font-medium transition-colors"
      >+ Tambah Atribut</button>
    </div>
    <Input label="Price (Rp)" type="number" placeholder="e.g. 65000" bind:value={variantFormData.price} />
    <Select label="Status" bind:value={variantFormData.status} options={[{ value: 'Active', label: 'Active' }, { value: 'Inactive', label: 'Inactive' }]} />
  </div>
  {#snippet footer()}
    <div class="flex items-center justify-end gap-3">
      <Button variant="secondary" onclick={() => { showVariantForm = false }}>Cancel</Button>
      <Button onclick={handleSaveVariant}>{editingVariantId ? 'Update' : 'Create'} Variant</Button>
    </div>
  {/snippet}
</Modal>
