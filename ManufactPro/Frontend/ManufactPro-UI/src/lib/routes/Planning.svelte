<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$lib/components/ui/Card.svelte'
  import Button from '$lib/components/ui/Button.svelte'
  import Badge from '$lib/components/ui/Badge.svelte'
  import Input from '$lib/components/ui/Input.svelte'
  import Select from '$lib/components/ui/Select.svelte'
  import { api } from '$lib/api'
  import type { ProductVariant, Material, BomItem } from '$lib/types'
  import {
    Calculator,
    Package,
    AlertTriangle,
    CheckCircle,
    XCircle,
  } from '@lucide/svelte'

  interface MaterialRequirement {
    materialId: string
    materialName: string
    materialSku: string
    qtyNeeded: number
    unit: string
    stockAvailable: number
    shortfall: number
    status: 'ok' | 'low' | 'out'
  }

  let variants = $state<ProductVariant[]>([])
  let materials = $state<Material[]>([])
  let selectedVariantId = $state('')
  let targetQty = $state(0)
  let requirements = $state<MaterialRequirement[]>([])
  let showResults = $state(false)

  let variantOptions = $derived(variants.map(v => ({ value: v.id, label: `${v.sku} - Rp ${v.sell_price.toLocaleString('id-ID')}` })))

  onMount(async () => {
    materials = await api.materials.list()
    variants = await api.productVariants.list()
  })

  async function calculateRequirements() {
    if (!selectedVariantId || targetQty <= 0) return

    const bomItems = await api.bomItems.listByVariant(selectedVariantId)
    
    requirements = bomItems.map(item => {
      const material = materials.find(m => m.id === item.material_id)
      const qtyNeeded = item.qty * targetQty
      const stockAvailable = 1000
      const shortfall = Math.max(0, qtyNeeded - stockAvailable)
      
      let status: 'ok' | 'low' | 'out' = 'ok'
      if (stockAvailable <= 0) status = 'out'
      else if (shortfall > 0) status = 'low'
      
      return {
        materialId: item.material_id,
        materialName: material?.name || 'Unknown',
        materialSku: material?.sku || '',
        qtyNeeded,
        unit: item.unit,
        stockAvailable,
        shortfall,
        status
      }
    })
    
    showResults = true
  }

  function resetForm() {
    selectedVariantId = ''
    targetQty = 0
    requirements = []
    showResults = false
  }
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-[32px] font-bold text-on-surface tracking-tight">Material Planning</h1>
    <p class="text-on-surface-variant mt-1">Calculate material requirements before production</p>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="lg:col-span-1">
      <Card variant="glass">
        <div class="flex items-center gap-2 mb-4">
          <Calculator class="w-5 h-5 text-primary" />
          <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Calculator</h2>
        </div>
        
        <div class="space-y-4">
          <Select 
            label="Product Variant" 
            bind:value={selectedVariantId} 
            options={variantOptions} 
            placeholder="Select variant..."
            required 
          />
          
          <Input 
            label="Target Quantity" 
            type="number" 
            placeholder="e.g. 100" 
            bind:value={targetQty}
            required 
          />
          
          <div class="flex gap-2 pt-2">
            <Button onclick={calculateRequirements} class="flex-1" disabled={!selectedVariantId || targetQty <= 0}>
              Calculate
            </Button>
            {#if showResults}
              <Button variant="secondary" onclick={resetForm}>
                Reset
              </Button>
            {/if}
          </div>
        </div>
      </Card>
    </div>

    <div class="lg:col-span-2">
      {#if showResults && requirements.length > 0}
        <Card variant="glass">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-sm font-semibold text-on-surface uppercase tracking-wider">Material Requirements</h2>
            <Badge variant={requirements.some(r => r.status !== 'ok') ? 'critical' : 'success'}>
              {requirements.filter(r => r.status === 'ok').length}/{requirements.length} Available
            </Badge>
          </div>

          <div class="space-y-3">
            {#each requirements as req}
              <div class="flex items-center gap-4 p-3 rounded-lg bg-surface-container-low/50 border border-outline-variant/30">
                <div class="p-2 rounded-lg {req.status === 'ok' ? 'bg-status-success/10' : req.status === 'low' ? 'bg-status-warning/10' : 'bg-status-critical/10'}">
                  <Package class="w-5 h-5 {req.status === 'ok' ? 'text-status-success' : req.status === 'low' ? 'text-status-warning' : 'text-status-critical'}" />
                </div>
                
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-on-surface">{req.materialName}</p>
                  <p class="text-xs text-on-surface-variant font-mono">{req.materialSku}</p>
                </div>
                
                <div class="text-right">
                  <p class="text-sm font-semibold text-on-surface">
                    Need: {req.qtyNeeded} {req.unit}
                  </p>
                  <p class="text-xs text-on-surface-variant">
                    Stock: {req.stockAvailable} {req.unit}
                  </p>
                  {#if req.shortfall > 0}
                    <p class="text-xs text-status-critical font-medium">
                      Short: {req.shortfall} {req.unit}
                    </p>
                  {/if}
                </div>
                
                <div class="shrink-0">
                  {#if req.status === 'ok'}
                    <CheckCircle class="w-5 h-5 text-status-success" />
                  {:else if req.status === 'low'}
                    <AlertTriangle class="w-5 h-5 text-status-warning" />
                  {:else}
                    <XCircle class="w-5 h-5 text-status-critical" />
                  {/if}
                </div>
              </div>
            {/each}
          </div>

          {#if requirements.some(r => r.status !== 'ok')}
            <div class="mt-4 p-3 rounded-lg bg-status-warning/10 border border-status-warning/30">
              <div class="flex items-start gap-2">
                <AlertTriangle class="w-5 h-5 text-status-warning shrink-0 mt-0.5" />
                <div class="flex-1">
                  <p class="text-sm font-medium text-status-warning">Material Shortfall Detected</p>
                  <p class="text-xs text-on-surface-variant mt-1">
                    Some materials have insufficient stock. Consider creating purchase orders or reducing production quantity.
                  </p>
                </div>
              </div>
            </div>
          {/if}
        </Card>
      {:else if showResults}
        <Card variant="glass">
          <div class="text-center py-12 text-outline">
            <Calculator class="w-8 h-8 mx-auto mb-2" />
            <span>No BOM defined for selected variant</span>
          </div>
        </Card>
      {:else}
        <Card variant="glass">
          <div class="text-center py-12 text-outline">
            <Calculator class="w-8 h-8 mx-auto mb-2" />
            <span>Select variant and quantity to calculate requirements</span>
          </div>
        </Card>
      {/if}
    </div>
  </div>
</div>
