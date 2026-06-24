import { describe, it, expect } from 'vitest'
import { getCategoryPreset, generateCombos, buildSku, type Dimension } from './variants'

describe('getCategoryPreset', () => {
  it('matches a known category by substring', () => {
    expect(getCategoryPreset('Pakaian Pria')).toEqual([
      { key: 'Ukuran', value: '' },
      { key: 'Warna', value: '' },
    ])
  })

  it('is case-insensitive', () => {
    expect(getCategoryPreset('MINUMAN')).toEqual([
      { key: 'Rasa', value: '' },
      { key: 'Volume (ml)', value: '' },
    ])
  })

  it('falls back to a single blank row for unknown category', () => {
    expect(getCategoryPreset('xyz')).toEqual([{ key: '', value: '' }])
  })

  it('returns fresh row copies, not shared references', () => {
    const a = getCategoryPreset('food')
    const b = getCategoryPreset('food')
    a[0].value = 'mutated'
    expect(b[0].value).toBe('')
  })
})

describe('generateCombos', () => {
  const dim = (key: string, values: string[]): Dimension => ({ key, values, inputVal: '' })

  it('returns empty when no valid dimension', () => {
    expect(generateCombos([])).toEqual([])
    expect(generateCombos([dim('Ukuran', [])])).toEqual([])
    expect(generateCombos([dim('  ', ['M'])])).toEqual([])
  })

  it('produces single-dimension combos', () => {
    expect(generateCombos([dim('Ukuran', ['S', 'M'])])).toEqual([
      { Ukuran: 'S' },
      { Ukuran: 'M' },
    ])
  })

  it('produces the cartesian product of two dimensions', () => {
    const result = generateCombos([dim('Ukuran', ['S', 'M']), dim('Warna', ['Merah', 'Biru'])])
    expect(result).toEqual([
      { Ukuran: 'S', Warna: 'Merah' },
      { Ukuran: 'S', Warna: 'Biru' },
      { Ukuran: 'M', Warna: 'Merah' },
      { Ukuran: 'M', Warna: 'Biru' },
    ])
  })

  it('trims dimension keys', () => {
    expect(generateCombos([dim('  Ukuran  ', ['M'])])).toEqual([{ Ukuran: 'M' }])
  })

  it('skips invalid dimensions mixed with valid ones', () => {
    const result = generateCombos([dim('Ukuran', ['M']), dim('', ['x']), dim('Warna', ['Merah'])])
    expect(result).toEqual([{ Ukuran: 'M', Warna: 'Merah' }])
  })
})

describe('buildSku', () => {
  it('joins uppercase 3-char prefix and attr values', () => {
    expect(buildSku('Kaos Polos', { Ukuran: 'Medium', Warna: 'Merah' })).toBe('KAO-MED-MER')
  })

  it('strips whitespace before slicing', () => {
    expect(buildSku('  ab cd  ', { x: 'a b' })).toBe('ABC-AB')
  })

  it('handles a product name with no attributes', () => {
    expect(buildSku('Tas', {})).toBe('TAS')
  })
})
