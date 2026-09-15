import { describe, expect, it } from 'vitest'
import {
  calculateImageSize,
  findPresetForSize,
  imageBillingTier,
  parseSize,
} from './size'

function maxEdge(size: string) {
  const parsed = parseSize(size)
  expect(parsed).not.toBeNull()
  return Math.max(parsed!.width, parsed!.height)
}

describe('image workbench size presets', () => {
  it('keeps 1K presets inside the backend 1K long-edge bucket', () => {
    for (const ratio of ['1:1', '3:2', '2:3', '4:3', '3:4'] as const) {
      const size = calculateImageSize('1K', ratio)
      expect(size).toBeTruthy()
      expect(maxEdge(size!)).toBeLessThanOrEqual(1024)
      expect(imageBillingTier(size!)).toBe('1K')
    }
  })

  it('does not treat official HD landscape as a 1K selection', () => {
    expect(calculateImageSize('1K', '3:2')).not.toBe('1536x1024')
    expect(findPresetForSize('1536x1024')).not.toEqual({ tier: '1K', ratio: '3:2' })
    expect(imageBillingTier('1536x1024')).toBe('2K')
  })

  it('promotes 1K 16:9 to a valid 2K-billed size instead of a fake 1K', () => {
    const size = calculateImageSize('1K', '16:9')
    expect(size).toBeTruthy()
    expect(imageBillingTier(size!)).toBe('2K')
    expect(maxEdge(size!)).toBeGreaterThan(1024)
  })

  it('uses official 2K landscape instead of a 4K-billed 2560x1440', () => {
    expect(calculateImageSize('2K', '16:9')).toBe('2048x1152')
    expect(imageBillingTier('2048x1152')).toBe('2K')
    expect(imageBillingTier('2560x1440')).toBe('4K')
    expect(findPresetForSize('2560x1440')).toBeNull()
  })

  it('keeps 2K presets inside the backend 2K long-edge bucket', () => {
    for (const ratio of ['1:1', '3:2', '2:3', '16:9', '9:16', '4:3', '3:4', '21:9'] as const) {
      const size = calculateImageSize('2K', ratio)
      expect(size).toBeTruthy()
      expect(maxEdge(size!)).toBeLessThanOrEqual(2048)
      expect(imageBillingTier(size!)).toBe('2K')
    }
  })

  it('classifies auto and unknown sizes as the backend default 2K tier', () => {
    expect(imageBillingTier('auto')).toBe('2K')
    expect(imageBillingTier('')).toBe('2K')
    expect(imageBillingTier('not-a-size')).toBe('2K')
  })
})
