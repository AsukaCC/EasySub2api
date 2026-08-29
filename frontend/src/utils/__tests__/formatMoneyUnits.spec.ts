import { describe, expect, it } from 'vitest'
import { formatCNY, formatPointAmount, formatPoints, formatUSD } from '../format'

describe('platform point formatting', () => {
  it('keeps at least two and at most eight fraction digits', () => {
    expect(formatPointAmount(1, 'en')).toBe('1.00')
    expect(formatPointAmount(1.2, 'en')).toBe('1.20')
    expect(formatPointAmount(1.23456789, 'en')).toBe('1.23456789')
    expect(formatPointAmount(1.234567891, 'en')).toBe('1.23456789')
  })

  it('trims optional trailing zeroes and localizes the unit', () => {
    expect(formatPoints(12.340000, 'zh')).toBe('12.34 积分')
    expect(formatPoints(12.345600, 'en')).toBe('12.3456 points')
  })

  it('normalizes missing and non-finite values to zero', () => {
    expect(formatPointAmount(undefined, 'en')).toBe('0.00')
    expect(formatPointAmount(Number.NaN, 'en')).toBe('0.00')
  })
})

describe('explicit money formatting', () => {
  it('keeps payment CNY separate from internal USD', () => {
    expect(formatCNY(10, 'zh')).toContain('10.00')
    expect(formatCNY(10, 'zh')).toContain('¥')
    expect(formatUSD(10, 'en')).toBe('$10.00')
  })
})
