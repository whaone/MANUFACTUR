import { describe, it, expect } from 'vitest'
import { formatDate, colorForBar, barWidth } from './reports'

describe('formatDate', () => {
  it('formats an ISO date as "D Mon YYYY"', () => {
    expect(formatDate('2026-06-24')).toBe('24 Jun 2026')
  })

  it('strips a leading zero from the day', () => {
    expect(formatDate('2026-01-05')).toBe('5 Jan 2026')
  })

  it('returns empty string for empty input', () => {
    expect(formatDate('')).toBe('')
  })
})

describe('colorForBar', () => {
  it('maps margin percentage to the right tier class', () => {
    expect(colorForBar(50)).toBe('bg-status-success')
    expect(colorForBar(45)).toBe('bg-status-success')
    expect(colorForBar(40)).toBe('bg-primary')
    expect(colorForBar(35)).toBe('bg-primary')
    expect(colorForBar(25)).toBe('bg-status-warning')
    expect(colorForBar(20)).toBe('bg-status-warning')
    expect(colorForBar(10)).toBe('bg-status-critical')
    expect(colorForBar(0)).toBe('bg-status-critical')
  })
})

describe('barWidth', () => {
  it('clamps to a minimum of 15', () => {
    expect(barWidth(0)).toBe(15)
    expect(barWidth(2)).toBe(15)
  })

  it('clamps to a maximum of 100', () => {
    expect(barWidth(50)).toBe(100)
    expect(barWidth(200)).toBe(100)
  })

  it('scales by 5 within range', () => {
    expect(barWidth(10)).toBe(50)
    expect(barWidth(18)).toBe(90)
  })
})
