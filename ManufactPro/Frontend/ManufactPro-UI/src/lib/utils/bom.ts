// Pure mapping helpers for the BOM view. UI-free so they can be unit-tested
// and benchmarked without a running backend.
import type { Material, ProductVariant, BomItem } from '$lib/types'

export interface BomRecipe {
  variantId: string
  variant: string
  product: string
  items: { id: string; materialId: string; material: string; qty: number; unit: string; optional: boolean }[]
  totalMaterials: number
}

// Build the per-variant BOM recipes from already-fetched data.
// `itemsByVariant[i]` must correspond to `variants[i]`.
export function buildBomRecipes(
  variants: ProductVariant[],
  materials: Material[],
  itemsByVariant: BomItem[][],
): BomRecipe[] {
  // Index materials by id once — avoids an O(n) find() per BOM item (O(v·i·m) → O(v·i)).
  const materialName = new Map(materials.map(m => [m.id, m.name]))

  return variants.map((variant, idx) => {
    const items = itemsByVariant[idx].map(item => ({
      id: item.id,
      materialId: item.material_id,
      material: materialName.get(item.material_id) ?? 'Unknown',
      qty: item.qty,
      unit: item.unit,
      optional: item.is_optional,
    }))
    return {
      variantId: variant.id,
      variant: variant.sku,
      product: variant.sku,
      items,
      totalMaterials: items.length,
    }
  })
}
