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
