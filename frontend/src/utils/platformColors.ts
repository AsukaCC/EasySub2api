/**
 * Centralized platform color definitions.
 *
 * All components that need platform-specific styling should import from here
 * instead of defining their own color mappings.
 */

export type Platform =
  | 'anthropic'
  | 'openai'
  | 'gemini'
  | 'antigravity'
  | 'grok'
  | 'kimi'
  | 'zhipu'
  | 'deepseek'
  | 'composite'

// ── Badge (bg + text + border, for inline badges with border) ───────
const BADGE: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state',
  openai: 'utils-platform-colors__state-2',
  gemini: 'utils-platform-colors__state-2',
  antigravity: 'utils-platform-colors__state-7',
  grok: 'utils-platform-colors__state-5',
  kimi: 'utils-platform-colors__state-6',
  zhipu: 'utils-platform-colors__state-7',
  deepseek: 'utils-platform-colors__state-8',
  composite: 'utils-platform-colors__state-9',
}
const BADGE_DEFAULT = 'utils-platform-colors__state-10'

// ── Light badge (softer bg, no border) ──────────────────────────────
const BADGE_LIGHT: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-11',
  openai: 'utils-platform-colors__state-12',
  gemini: 'utils-platform-colors__state-12',
  antigravity: 'utils-platform-colors__state-17',
  grok: 'utils-platform-colors__state-15',
  kimi: 'utils-platform-colors__state-16',
  zhipu: 'utils-platform-colors__state-17',
  deepseek: 'utils-platform-colors__state-18',
  composite: 'utils-platform-colors__state-19',
}

// ── Border ──────────────────────────────────────────────────────────
const BORDER: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-20',
  openai: 'utils-platform-colors__state-21',
  gemini: 'utils-platform-colors__state-21',
  antigravity: 'utils-platform-colors__state-26',
  grok: 'utils-platform-colors__state-24',
  kimi: 'utils-platform-colors__state-25',
  zhipu: 'utils-platform-colors__state-26',
  deepseek: 'utils-platform-colors__state-27',
  composite: 'utils-platform-colors__state-28',
}
const BORDER_DEFAULT = 'utils-platform-colors__state-29'

// ── Border strong (higher-contrast platform tint, e.g. plaza group cards) ──
const BORDER_STRONG: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-30',
  openai: 'utils-platform-colors__state-31',
  gemini: 'utils-platform-colors__state-31',
  antigravity: 'utils-platform-colors__state-36',
  grok: 'utils-platform-colors__state-34',
  kimi: 'utils-platform-colors__state-35',
  zhipu: 'utils-platform-colors__state-36',
  deepseek: 'utils-platform-colors__state-37',
  composite: 'utils-platform-colors__state-38',
}
const BORDER_STRONG_DEFAULT = 'utils-platform-colors__state-39'

// ── Accent (single raw color per platform; consumers derive washes/tints
//    from it via CSS color-mix, e.g. plaza paid-price zone) ──
const ACCENT: Record<Platform, string> = {
  anthropic: '#f97316', // orange-500
  openai: '#22c55e', // green-500
  gemini: '#3b82f6', // blue-500
  antigravity: '#8b5cf6', // violet-500
  grok: '#71717a', // zinc-500
  kimi: '#ec4899', // pink-500
  zhipu: '#6366f1', // indigo-500
  deepseek: '#14b8a6', // teal-500
  composite: '#06b6d4', // cyan-500
}
const ACCENT_DEFAULT = '#0a84ff' // primary-500 (apple blue)

// ── Accent bar (gradient) ───────────────────────────────────────────
const ACCENT_BAR: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-40',
  openai: 'utils-platform-colors__state-41',
  gemini: 'utils-platform-colors__state-41',
  antigravity: 'utils-platform-colors__state-46',
  grok: 'utils-platform-colors__state-44',
  kimi: 'utils-platform-colors__state-45',
  zhipu: 'utils-platform-colors__state-46',
  deepseek: 'utils-platform-colors__state-47',
  composite: 'utils-platform-colors__state-48',
}
const ACCENT_BAR_DEFAULT = 'utils-platform-colors__state-49'

// ── Text (price, icon) ─────────────────────────────────────────────
const TEXT: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-50',
  openai: 'utils-platform-colors__state-51',
  gemini: 'utils-platform-colors__state-51',
  antigravity: 'utils-platform-colors__state-56',
  grok: 'utils-platform-colors__state-54',
  kimi: 'utils-platform-colors__state-55',
  zhipu: 'utils-platform-colors__state-56',
  deepseek: 'utils-platform-colors__state-57',
  composite: 'utils-platform-colors__state-58',
}
const TEXT_DEFAULT = 'utils-platform-colors__state-59'

// ── Icon (check mark etc.) ──────────────────────────────────────────
const ICON: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-60',
  openai: 'utils-platform-colors__state-61',
  gemini: 'utils-platform-colors__state-61',
  antigravity: 'utils-platform-colors__state-65',
  grok: 'utils-platform-colors__state-54',
  kimi: 'utils-platform-colors__state-64',
  zhipu: 'utils-platform-colors__state-65',
  deepseek: 'utils-platform-colors__state-66',
  composite: 'utils-platform-colors__state-67',
}
const ICON_DEFAULT = 'utils-platform-colors__state-68'

// ── Button (solid bg) ───────────────────────────────────────────────
const BUTTON: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-69',
  openai: 'utils-platform-colors__state-70',
  gemini: 'utils-platform-colors__state-70',
  antigravity: 'utils-platform-colors__state-75',
  grok: 'utils-platform-colors__state-73',
  kimi: 'utils-platform-colors__state-74',
  zhipu: 'utils-platform-colors__state-75',
  deepseek: 'utils-platform-colors__state-76',
  composite: 'utils-platform-colors__state-77',
}
const BUTTON_DEFAULT = 'utils-platform-colors__state-78'

// ── Discount badge ──────────────────────────────────────────────────
const DISCOUNT: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-79',
  openai: 'utils-platform-colors__state-80',
  gemini: 'utils-platform-colors__state-80',
  antigravity: 'utils-platform-colors__state-85',
  grok: 'utils-platform-colors__state-83',
  kimi: 'utils-platform-colors__state-84',
  zhipu: 'utils-platform-colors__state-85',
  deepseek: 'utils-platform-colors__state-86',
  composite: 'utils-platform-colors__state-87',
}
const DISCOUNT_DEFAULT = 'utils-platform-colors__state-88'

// ── Header gradient (subscription confirm) ─────────────────────────
const GRADIENT: Record<Platform, string> = {
  anthropic: 'utils-platform-colors__state-89',
  openai: 'utils-platform-colors__state-90',
  gemini: 'utils-platform-colors__state-90',
  antigravity: 'utils-platform-colors__state-95',
  grok: 'utils-platform-colors__state-93',
  kimi: 'utils-platform-colors__state-94',
  zhipu: 'utils-platform-colors__state-95',
  deepseek: 'utils-platform-colors__state-96',
  composite: 'utils-platform-colors__state-97',
}
const GRADIENT_DEFAULT = 'utils-platform-colors__state-98'

// ── Header text (light text on gradient bg) ────────────────────────
const GRADIENT_TEXT: Record<Platform, string> = {
  anthropic: 'status-text--warning',
  openai: 'status-text--success',
  gemini: 'status-text--info',
  antigravity: 'status-text--accent',
  grok: 'status-text--neutral',
  kimi: 'status-text--accent',
  zhipu: 'status-text--info',
  deepseek: 'status-text--success',
  composite: 'status-text--info',
}
const GRADIENT_TEXT_DEFAULT = 'status-text--neutral'

const GRADIENT_SUBTEXT: Record<Platform, string> = {
  anthropic: 'status-text--warning',
  openai: 'status-text--success',
  gemini: 'status-text--info',
  antigravity: 'status-text--accent',
  grok: 'status-text--neutral',
  kimi: 'status-text--accent',
  zhipu: 'status-text--info',
  deepseek: 'status-text--success',
  composite: 'status-text--info',
}
const GRADIENT_SUBTEXT_DEFAULT = 'status-text--neutral'

// ── Public API ──────────────────────────────────────────────────────

function isPlatform(p: string): p is Platform {
  return (
    p === 'anthropic' ||
    p === 'openai' ||
    p === 'gemini' ||
    p === 'antigravity' ||
    p === 'grok' ||
    p === 'kimi' ||
    p === 'zhipu' ||
    p === 'deepseek' ||
    p === 'composite'
  )
}

export function platformBadgeClass(p: string): string {
  return isPlatform(p) ? BADGE[p] : BADGE_DEFAULT
}

export function platformBadgeLightClass(p: string): string {
  return isPlatform(p) ? BADGE_LIGHT[p] : BADGE_DEFAULT
}

export function platformBorderClass(p: string): string {
  return isPlatform(p) ? BORDER[p] : BORDER_DEFAULT
}

export function platformBorderStrongClass(p: string): string {
  return isPlatform(p) ? BORDER_STRONG[p] : BORDER_STRONG_DEFAULT
}

export function platformAccentColor(p: string): string {
  return isPlatform(p) ? ACCENT[p] : ACCENT_DEFAULT
}

export function platformAccentBarClass(p: string): string {
  return isPlatform(p) ? ACCENT_BAR[p] : ACCENT_BAR_DEFAULT
}

export function platformTextClass(p: string): string {
  return isPlatform(p) ? TEXT[p] : TEXT_DEFAULT
}

export function platformIconClass(p: string): string {
  return isPlatform(p) ? ICON[p] : ICON_DEFAULT
}

export function platformButtonClass(p: string): string {
  return isPlatform(p) ? BUTTON[p] : BUTTON_DEFAULT
}

export function platformDiscountClass(p: string): string {
  return isPlatform(p) ? DISCOUNT[p] : DISCOUNT_DEFAULT
}

export function platformGradientClass(p: string): string {
  return isPlatform(p) ? GRADIENT[p] : GRADIENT_DEFAULT
}

export function platformGradientTextClass(p: string): string {
  return isPlatform(p) ? GRADIENT_TEXT[p] : GRADIENT_TEXT_DEFAULT
}

export function platformGradientSubtextClass(p: string): string {
  return isPlatform(p) ? GRADIENT_SUBTEXT[p] : GRADIENT_SUBTEXT_DEFAULT
}

export function platformLabel(p: string): string {
  switch (p) {
    case 'anthropic': return 'Anthropic'
    case 'openai': return 'OpenAI'
    case 'gemini': return 'Gemini'
    case 'antigravity': return 'Antigravity'
    case 'grok': return 'Grok'
    case 'kimi': return 'Kimi'
    case 'zhipu': return 'Zhipu GLM'
    case 'deepseek': return 'DeepSeek'
    case 'composite': return 'Composite'
    default: return p || 'API'
  }
}
