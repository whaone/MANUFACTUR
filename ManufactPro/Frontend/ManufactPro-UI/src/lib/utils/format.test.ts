import { describe, it, expect } from 'vitest'
import { formatNumber, formatRupiah } from './format'

describe('formatNumber', () => {
  it('groups thousands with the id-ID locale', () => {
    expect(formatNumber(1234567)).toBe('1.234.567')
  })

  it('leaves small numbers unchanged', () => {
    expect(formatNumber(0)).toBe('0')
    expect(formatNumber(999)).toBe('999')
  })
})

describe('formatRupiah', () => {
  it('prefixes "Rp " and groups thousands', () => {
    expect(formatRupiah(65000)).toBe('Rp 65.000')
  })

  it('handles zero', () => {
    expect(formatRupiah(0)).toBe('Rp 0')
  })
})
