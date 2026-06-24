// Pure helpers for product-variant attribute presets and the matrix generator.
// Kept UI-free so they can be unit-tested in isolation.

export type AttrRow = { key: string; value: string }
export type Dimension = { key: string; values: string[]; inputVal: string }

export const CATEGORY_PRESETS: Record<string, AttrRow[]> = {
  pakaian:    [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
  tekstil:    [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
  clothing:   [{ key: 'Ukuran', value: '' }, { key: 'Warna', value: '' }],
  makanan:    [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
  food:       [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
  'f&b':      [{ key: 'Rasa', value: '' }, { key: 'Ukuran Porsi', value: '' }],
  minuman:    [{ key: 'Rasa', value: '' }, { key: 'Volume (ml)', value: '' }],
  beverage:   [{ key: 'Rasa', value: '' }, { key: 'Volume (ml)', value: '' }],
  elektronik: [{ key: 'Warna', value: '' }, { key: 'Kapasitas', value: '' }],
}

// Resolve a category string to its attribute preset. Substring match against the
// preset keys; falls back to a single blank row when nothing matches.
export function getCategoryPreset(cat: string): AttrRow[] {
  const lower = cat.toLowerCase()
  for (const [key, rows] of Object.entries(CATEGORY_PRESETS)) {
    if (lower.includes(key)) return rows.map(r => ({ ...r }))
  }
  return [{ key: '', value: '' }]
}

// Cartesian product of all valid dimensions (non-empty key + at least one value).
// Returns one attribute map per combination; empty when no valid dimension exists.
export function generateCombos(dims: Dimension[]): Record<string, string>[] {
  const valid = dims.filter(d => d.key.trim() && d.values.length > 0)
  if (valid.length === 0) return []
  const [first, ...rest] = valid
  const restCombos = rest.length > 0 ? generateCombos(rest) : [{}]
  return first.values.flatMap(v =>
    restCombos.map(combo => ({ [first.key.trim()]: v, ...combo }))
  )
}

// Build an uppercase, dash-joined SKU from the product name prefix and attr values.
export function buildSku(productName: string, attrs: Record<string, string>): string {
  const prefix = productName.replace(/\s+/g, '').slice(0, 3).toUpperCase()
  const parts = Object.values(attrs).map(v => v.replace(/\s+/g, '').slice(0, 3).toUpperCase())
  return [prefix, ...parts].join('-')
}
