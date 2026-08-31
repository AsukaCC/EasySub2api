export const DEFAULT_THEME_ACCENT = '#0a84ff'
export const THEME_ACCENT_STORAGE_KEY = 'theme-accent'

export const THEME_ACCENT_PRESETS = [
  { id: 'blue', value: '#0a84ff' },
  { id: 'purple', value: '#7c3aed' },
  { id: 'green', value: '#059669' },
  { id: 'orange', value: '#ea580c' },
  { id: 'pink', value: '#db2777' },
  { id: 'teal', value: '#0891b2' },
] as const

const HEX_RE = /^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/

export function normalizeThemeAccent(value: string | null | undefined): string {
  const raw = (value ?? '').trim()
  if (!HEX_RE.test(raw)) return DEFAULT_THEME_ACCENT
  if (raw.length === 4) {
    return `#${raw[1]}${raw[1]}${raw[2]}${raw[2]}${raw[3]}${raw[3]}`.toLowerCase()
  }
  return raw.toLowerCase()
}

export function getStoredThemeAccent(): string | null {
  if (typeof window === 'undefined') return null
  const stored = window.localStorage.getItem(THEME_ACCENT_STORAGE_KEY)
  return stored ? normalizeThemeAccent(stored) : null
}

export function setStoredThemeAccent(value: string): string {
  const accent = normalizeThemeAccent(value)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(THEME_ACCENT_STORAGE_KEY, accent)
  }
  return accent
}

export function clearStoredThemeAccent(): void {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(THEME_ACCENT_STORAGE_KEY)
}

export function resolveThemeAccent(siteDefault?: string | null): string {
  return getStoredThemeAccent() ?? normalizeThemeAccent(siteDefault)
}

export function applyThemeAccent(value?: string | null): string {
  const accent = normalizeThemeAccent(value)
  if (typeof document !== 'undefined') {
    document.documentElement.style.setProperty('--theme-accent', accent)
  }
  return accent
}

export function applyResolvedThemeAccent(siteDefault?: string | null): string {
  return applyThemeAccent(resolveThemeAccent(siteDefault))
}
