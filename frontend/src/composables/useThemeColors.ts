import { computed, onBeforeUnmount, onMounted, ref, type ComputedRef } from 'vue'

export interface ThemeColors {
  textPrimary: string
  textSecondary: string
  textTertiary: string
  grid: string
  elevatedSurface: string
}

const FALLBACK_COLORS: ThemeColors = {
  textPrimary: '#09090b',
  textSecondary: '#3f3f46',
  textTertiary: '#62626b',
  grid: 'rgba(228, 228, 231, 0.80)',
  elevatedSurface: 'rgba(255, 255, 255, 0.88)'
}

const themeRevision = ref(0)
let themeObserver: MutationObserver | null = null
let subscriberCount = 0

function refreshThemeColors(): void {
  themeRevision.value += 1
}

function subscribeToThemeChanges(): void {
  subscriberCount += 1

  if (
    themeObserver ||
    typeof document === 'undefined' ||
    typeof MutationObserver === 'undefined'
  ) {
    refreshThemeColors()
    return
  }

  themeObserver = new MutationObserver(refreshThemeColors)
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class', 'style']
  })
  refreshThemeColors()
}

function unsubscribeFromThemeChanges(): void {
  subscriberCount = Math.max(0, subscriberCount - 1)
  if (subscriberCount > 0) return

  themeObserver?.disconnect()
  themeObserver = null
}

function readThemeColors(): ThemeColors {
  if (typeof document === 'undefined' || typeof window === 'undefined') {
    return FALLBACK_COLORS
  }

  const styles = window.getComputedStyle(document.documentElement)
  const readToken = (token: string, fallback: string): string =>
    styles.getPropertyValue(token).trim() || fallback

  return {
    textPrimary: readToken('--color-text-primary', FALLBACK_COLORS.textPrimary),
    textSecondary: readToken('--color-text-secondary', FALLBACK_COLORS.textSecondary),
    textTertiary: readToken('--color-text-tertiary', FALLBACK_COLORS.textTertiary),
    grid: readToken('--color-border', FALLBACK_COLORS.grid),
    elevatedSurface: readToken('--color-surface-elevated', FALLBACK_COLORS.elevatedSurface)
  }
}

/**
 * Resolves CSS theme tokens for canvas-based charts and refreshes them when the
 * root theme class or inline theme variables change.
 */
export function useThemeColors(): ComputedRef<ThemeColors> {
  let subscribed = false

  onMounted(() => {
    subscribed = true
    subscribeToThemeChanges()
  })

  onBeforeUnmount(() => {
    if (!subscribed) return
    subscribed = false
    unsubscribeFromThemeChanges()
  })

  return computed(() => {
    themeRevision.value
    return readThemeColors()
  })
}
