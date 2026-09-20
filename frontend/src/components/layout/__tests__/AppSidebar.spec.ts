import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../styles/_layout.scss')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar mobile visibility', () => {
  it('moves the closed sidebar outside the mobile viewport', () => {
    expect(componentSource).toMatch(
      /@media \(max-width: 1023px\)[\s\S]*?\.components-layout-app-sidebar__aside-2\s*\{[\s\S]*?transform: translateX\(-100%\);/,
    )
  })

  it('keeps the drawer below the mobile top navigation', () => {
    expect(styleSource).toContain(
      'inset: calc(var(--app-shell-sticky-offset) + env(safe-area-inset-top, 0px)) auto 0 0;',
    )
    expect(styleSource).toContain('z-index: var(--z-sidebar);')
    expect(componentSource).toContain(
      'inset: calc(var(--app-shell-sticky-offset) + env(safe-area-inset-top, 0px)) 0 0;',
    )
    expect(componentSource).toContain('z-index: calc(var(--z-sidebar) - 1);')
  })

  it('keeps the mobile overlay behind the drawer so menu items remain clickable', () => {
    expect(styleSource).toContain('z-index: var(--z-sidebar);')
    expect(componentSource).toContain('z-index: calc(var(--z-sidebar) - 1);')
  })
})

describe('AppSidebar mobile group navigation', () => {
  it('expands parent items without navigating or closing the mobile drawer', () => {
    const handler = componentSource.match(
      /function handleGroupClick\(item: NavItem, event\?: MouseEvent\) \{[\s\S]*?\n\}/,
    )?.[0]

    expect(handler).toBeDefined()
    expect(handler).toContain(`if (mobileOpen.value) {
    toggleGroup(item)
    return
  }`)
    expect(handler?.indexOf('if (mobileOpen.value)')).toBeLessThan(
      handler?.indexOf('if (sidebarCollapsed.value)') ?? -1,
    )
    expect(handler?.indexOf('if (mobileOpen.value)')).toBeLessThan(
      handler?.indexOf('router.push(item.path)') ?? -1,
    )
  })
})

describe('AppSidebar site billing navigation', () => {
  it('uses the unified mode for the purchase label and subscription visibility', () => {
    expect(componentSource).toContain('resolveSiteBillingMode(appStore.cachedPublicSettings)')
    expect(componentSource).toContain('const purchaseNavLabel = computed')
    expect(componentSource).toContain('featureFlag: flagSubscription')
  })
})

describe('AppSidebar layout placement', () => {
  it('keeps the desktop sidebar in document flow', () => {
    const sidebarBlock = styleSource.match(/^\.sidebar \{[\s\S]*?\n\}/m)
    expect(sidebarBlock?.[0]).not.toContain('position: fixed')
    expect(sidebarBlock?.[0]).toContain('flex: 0 0 16rem')
  })

  it('uses a fixed drawer only below the desktop breakpoint', () => {
    expect(styleSource).toMatch(
      /@media \(max-width: 1023px\)\s*\{\s*\.sidebar \{\s*position: fixed;/,
    )
  })

  it('collapses to an icon rail instead of hiding the sidebar', () => {
    expect(componentSource).not.toContain('v-show="!sidebarCollapsed"')
    expect(componentSource).not.toContain('admin-menu-trigger')
    expect(styleSource).toContain('.sidebar--collapsed')
    expect(styleSource).toContain('flex: 0 0 4.5rem')
  })

  it('opens a teleported secondary menu flyout when the icon rail is collapsed', () => {
    expect(componentSource).toContain('sidebar-collapsed-flyout')
    expect(componentSource).toContain("placement: 'right'")
    expect(componentSource).toContain('onCollapsedGroupEnter')
    expect(componentSource).toContain('<Teleport to="body">')
  })
})
