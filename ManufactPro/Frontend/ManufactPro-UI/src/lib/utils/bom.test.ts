import { describe, it, expect } from 'vitest'
import { buildBomRecipes } from './bom'
import type { Material, ProductVariant, BomItem } from '$lib/types'

const mat = (id: string, name: string) => ({ id, name }) as unknown as Material
const variant = (id: string, sku: string) => ({ id, sku }) as unknown as ProductVariant
const item = (id: string, material_id: string, extra: Partial<BomItem> = {}) =>
  ({ id, material_id, qty: 2, unit: 'pcs', is_optional: false, ...extra }) as unknown as BomItem

describe('buildBomRecipes', () => {
  it('maps variants to recipes with resolved material names', () => {
    const recipes = buildBomRecipes(
      [variant('v1', 'SKU-1')],
      [mat('m1', 'Cotton'), mat('m2', 'Thread')],
      [[item('b1', 'm1'), item('b2', 'm2', { unit: 'meter', is_optional: true })]],
    )
    expect(recipes).toEqual([
      {
        variantId: 'v1',
        variant: 'SKU-1',
        product: 'SKU-1',
        totalMaterials: 2,
        items: [
          { id: 'b1', materialId: 'm1', material: 'Cotton', qty: 2, unit: 'pcs', optional: false },
          { id: 'b2', materialId: 'm2', material: 'Thread', qty: 2, unit: 'meter', optional: true },
        ],
      },
    ])
  })

  it('falls back to "Unknown" for missing materials', () => {
    const recipes = buildBomRecipes([variant('v1', 'S')], [], [[item('b1', 'gone')]])
    expect(recipes[0].items[0].material).toBe('Unknown')
  })

  it('pairs itemsByVariant positionally with variants', () => {
    const recipes = buildBomRecipes(
      [variant('v1', 'A'), variant('v2', 'B')],
      [mat('m1', 'X')],
      [[item('b1', 'm1')], []],
    )
    expect(recipes[0].totalMaterials).toBe(1)
    expect(recipes[1].totalMaterials).toBe(0)
    expect(recipes[1].items).toEqual([])
  })

  it('returns empty array for no variants', () => {
    expect(buildBomRecipes([], [], [])).toEqual([])
  })
})
