import { mount } from '@vue/test-utils'
import { defineComponent, type ComputedRef } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  useThemeColors,
  type ThemeColors
} from '@/composables/useThemeColors'

const root = document.documentElement
const originalClass = root.getAttribute('class')
const originalStyle = root.getAttribute('style')

afterEach(() => {
  vi.restoreAllMocks()

  if (originalClass == null) root.removeAttribute('class')
  else root.setAttribute('class', originalClass)

  if (originalStyle == null) root.removeAttribute('style')
  else root.setAttribute('style', originalStyle)
})

function mountThemeProbe(): {
  colors: ComputedRef<ThemeColors>
  unmount: () => void
} {
  let colors!: ComputedRef<ThemeColors>
  const wrapper = mount(defineComponent({
    setup() {
      colors = useThemeColors()
      return () => null
    }
  }))

  return {
    colors,
    unmount: () => wrapper.unmount()
  }
}

describe('useThemeColors', () => {
  it('reads semantic colors from the root element', () => {
    root.style.setProperty('--color-text-primary', '#101010')
    root.style.setProperty('--color-text-secondary', '#202020')
    root.style.setProperty('--color-text-tertiary', '#303030')
    root.style.setProperty('--color-border', 'rgba(40, 40, 40, 0.5)')
    root.style.setProperty('--color-surface-elevated', '#505050')

    const probe = mountThemeProbe()

    expect(probe.colors.value).toEqual({
      textPrimary: '#101010',
      textSecondary: '#202020',
      textTertiary: '#303030',
      grid: 'rgba(40, 40, 40, 0.5)',
      elevatedSurface: '#505050'
    })

    probe.unmount()
  })

  it('refreshes after root class and style changes', async () => {
    const values = {
      light: '#62626b',
      dark: '#a1a1aa',
      custom: '#abcdef'
    }
    const getComputedStyle = vi.spyOn(window, 'getComputedStyle')
    getComputedStyle.mockImplementation(() => ({
      getPropertyValue(token: string) {
        if (token === '--color-text-tertiary') {
          return root.style.getPropertyValue(token) || (root.classList.contains('dark') ? values.dark : values.light)
        }
        return ''
      }
    }) as CSSStyleDeclaration)

    root.classList.remove('dark')
    const probe = mountThemeProbe()
    expect(probe.colors.value.textTertiary).toBe(values.light)

    root.classList.add('dark')
    await vi.waitFor(() => {
      expect(probe.colors.value.textTertiary).toBe(values.dark)
    })

    root.style.setProperty('--color-text-tertiary', values.custom)
    await vi.waitFor(() => {
      expect(probe.colors.value.textTertiary).toBe(values.custom)
    })

    probe.unmount()
  })
})
