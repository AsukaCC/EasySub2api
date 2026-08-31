import { afterEach, describe, expect, it } from 'vitest'
import {
  DEFAULT_THEME_ACCENT,
  applyThemeAccent,
  clearStoredThemeAccent,
  normalizeThemeAccent,
  resolveThemeAccent,
  setStoredThemeAccent,
} from '@/utils/themeAccent'

describe('themeAccent', () => {
  afterEach(() => {
    clearStoredThemeAccent()
    document.documentElement.style.removeProperty('--theme-accent')
  })

  it('normalizes 3-digit hex and rejects invalid values', () => {
    expect(normalizeThemeAccent('#abc')).toBe('#aabbcc')
    expect(normalizeThemeAccent('#0A84FF')).toBe('#0a84ff')
    expect(normalizeThemeAccent('blue')).toBe(DEFAULT_THEME_ACCENT)
    expect(normalizeThemeAccent('')).toBe(DEFAULT_THEME_ACCENT)
  })

  it('prefers the stored override over the site default', () => {
    setStoredThemeAccent('#7c3aed')
    expect(resolveThemeAccent('#0a84ff')).toBe('#7c3aed')
  })

  it('writes the accent onto the document element', () => {
    applyThemeAccent('#059669')
    expect(document.documentElement.style.getPropertyValue('--theme-accent')).toBe('#059669')
  })
})
