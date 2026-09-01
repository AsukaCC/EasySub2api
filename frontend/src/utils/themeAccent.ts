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
const LIGHT_TEXT_BACKGROUND = '#f5f5f7'
const DARK_TEXT_BACKGROUND = '#191921'
const MIN_TEXT_CONTRAST = 4.6

type Rgb = readonly [number, number, number]

function hexToRgb(value: string): Rgb {
  return [
    Number.parseInt(value.slice(1, 3), 16),
    Number.parseInt(value.slice(3, 5), 16),
    Number.parseInt(value.slice(5, 7), 16),
  ]
}

function rgbToHex([red, green, blue]: Rgb): string {
  const channel = (value: number) => Math.round(value).toString(16).padStart(2, '0')
  return `#${channel(red)}${channel(green)}${channel(blue)}`
}

function mixColor(color: string, target: string, amount: number): string {
  const sourceRgb = hexToRgb(color)
  const targetRgb = hexToRgb(target)
  return rgbToHex([
    sourceRgb[0] + (targetRgb[0] - sourceRgb[0]) * amount,
    sourceRgb[1] + (targetRgb[1] - sourceRgb[1]) * amount,
    sourceRgb[2] + (targetRgb[2] - sourceRgb[2]) * amount,
  ])
}

function relativeLuminance(value: string): number {
  const channels = hexToRgb(value).map((channel) => {
    const normalized = channel / 255
    return normalized <= 0.04045
      ? normalized / 12.92
      : ((normalized + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * channels[0]! + 0.7152 * channels[1]! + 0.0722 * channels[2]!
}

function contrastRatio(foreground: string, background: string): number {
  const foregroundLuminance = relativeLuminance(foreground)
  const backgroundLuminance = relativeLuminance(background)
  const lighter = Math.max(foregroundLuminance, backgroundLuminance)
  const darker = Math.min(foregroundLuminance, backgroundLuminance)
  return (lighter + 0.05) / (darker + 0.05)
}

function ensureContrast(
  color: string,
  background: string,
  target: '#000000' | '#ffffff',
): string {
  if (contrastRatio(color, background) >= MIN_TEXT_CONTRAST) return color

  for (let step = 1; step <= 100; step += 1) {
    const candidate = mixColor(color, target, step / 100)
    if (contrastRatio(candidate, background) >= MIN_TEXT_CONTRAST) return candidate
  }
  return target
}

function deriveAccessibleAccentColors(accent: string) {
  const lightLink = ensureContrast(accent, LIGHT_TEXT_BACKGROUND, '#000000')
  const darkLink = ensureContrast(accent, DARK_TEXT_BACKGROUND, '#ffffff')
  const lightHover = ensureContrast(mixColor(lightLink, '#000000', 0.14), LIGHT_TEXT_BACKGROUND, '#000000')
  const darkHover = ensureContrast(mixColor(darkLink, '#ffffff', 0.18), DARK_TEXT_BACKGROUND, '#ffffff')
  const onAccent = contrastRatio('#09090b', accent) >= contrastRatio('#ffffff', accent)
    ? '#09090b'
    : '#ffffff'

  return { lightLink, lightHover, darkLink, darkHover, onAccent }
}

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
    const rootStyle = document.documentElement.style
    const derived = deriveAccessibleAccentColors(accent)
    rootStyle.setProperty('--theme-accent', accent)
    rootStyle.setProperty('--theme-accent-link-light', derived.lightLink)
    rootStyle.setProperty('--theme-accent-link-light-hover', derived.lightHover)
    rootStyle.setProperty('--theme-accent-link-dark', derived.darkLink)
    rootStyle.setProperty('--theme-accent-link-dark-hover', derived.darkHover)
    rootStyle.setProperty('--theme-accent-on', derived.onAccent)
  }
  return accent
}

export function applyResolvedThemeAccent(siteDefault?: string | null): string {
  return applyThemeAccent(resolveThemeAccent(siteDefault))
}
