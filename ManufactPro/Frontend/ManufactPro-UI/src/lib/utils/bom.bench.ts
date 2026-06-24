import { bench, describe } from 'vitest'
import { buildBomRecipes } from './bom'
import type { Material, ProductVariant, BomItem } from '$lib/types'

// --- Fixtures -------------------------------------------------------------
const VARIANT_COUNT = 50
const ITEMS_PER_VARIANT = 8
const MATERIAL_COUNT = 200
const API_LATENCY_MS = 5 // simulated per-request round-trip

const materials = Array.from({ length: MATERIAL_COUNT }, (_, i) => ({
  id: `mat-${i}`, name: `Material ${i}`,
}) as unknown as Material)

const variants = Array.from({ length: VARIANT_COUNT }, (_, i) => ({
  id: `var-${i}`, sku: `SKU-${i}`,
}) as unknown as ProductVariant)

const itemsByVariant: BomItem[][] = variants.map((_, vi) =>
  Array.from({ length: ITEMS_PER_VARIANT }, (_, ii) => ({
    id: `bom-${vi}-${ii}`,
    material_id: `mat-${(vi * ITEMS_PER_VARIANT + ii) % MATERIAL_COUNT}`,
    qty: 1, unit: 'pcs', is_optional: false,
  }) as unknown as BomItem),
)

const listByVariant = (variantId: string): Promise<BomItem[]> =>
  new Promise(resolve => {
    const idx = Number(variantId.split('-')[1])
    setTimeout(() => resolve(itemsByVariant[idx]), API_LATENCY_MS)
  })

// --- Fetch strategy: N+1 sequential vs parallel ---------------------------
describe('BOM fetch strategy', () => {
  bench('sequential (N+1 waterfall — before)', async () => {
    const out: BomItem[][] = []
    for (const v of variants) out.push(await listByVariant(v.id))
  })

  bench('parallel Promise.all (after)', async () => {
    await Promise.all(variants.map(v => listByVariant(v.id)))
  })
})

// --- Mapping: find() per item vs Map index --------------------------------
describe('BOM mapping', () => {
  bench('find() per item (before)', () => {
    variants.map((variant, idx) => {
      const items = itemsByVariant[idx].map(item => ({
        material: materials.find(m => m.id === item.material_id)?.name || 'Unknown',
      }))
      return items
    })
  })

  bench('Map index (after — buildBomRecipes)', () => {
    buildBomRecipes(variants, materials, itemsByVariant)
  })
})
