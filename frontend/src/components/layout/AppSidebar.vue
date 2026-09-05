<template>
  <aside
    id="admin-sidebar"
    class="sidebar"
    :class="[
      sidebarCollapsed ? 'sidebar--collapsed' : 'components-layout-app-sidebar__aside',
      { 'components-layout-app-sidebar__aside-2': !mobileOpen }
    ]"
  >
    <!-- Logo/Brand -->
    <div
      v-if="!adminMenuOnly"
      class="sidebar-header"
      :class="{ 'sidebar-header-collapsed': sidebarCollapsed }"
    >
      <!-- Custom Logo or Default Logo -->
      <router-link
        :to="homePath"
        class="components-layout-app-sidebar__router-link sidebar-logo"
        @click="handleMenuItemClick(homePath)"
      >
        <img v-if="settingsLoaded" :src="siteLogo || '/logo.svg'" alt="Logo" class="components-layout-app-sidebar__image" />
      </router-link>
      <div class="sidebar-brand" :class="{ 'sidebar-brand-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
        <router-link
          :to="homePath"
          class="components-layout-app-sidebar__router-link-2 sidebar-brand-title"
          @click="handleMenuItemClick(homePath)"
        >
          {{ siteName }}
        </router-link>
        <!-- Version Badge -->
        <VersionBadge :version="siteVersion" />
      </div>
    </div>

    <!-- Navigation -->
    <nav ref="sidebarNavRef" class="sidebar-nav scrollbar-hide" @pointerdown="primeSidebarActiveIndicator">
      <span
        ref="sidebarActiveIndicatorRef"
        class="sidebar-active-indicator"
        aria-hidden="true"
      ></span>
      <!-- Admin View: Admin menu first, then personal menu -->
      <template v-if="isAdmin">
        <!-- Admin sections grouped by responsibility -->
        <div v-for="section in adminNavSections" :key="section.key" class="sidebar-section">
          <div
            class="sidebar-section-title"
            :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }"
            :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
          >
            <span
              class="sidebar-section-title-text"
              :class="{ 'sidebar-section-title-text-collapsed': sidebarCollapsed }"
            >
              {{ section.label }}
            </span>
          </div>
          <template v-for="item in section.items" :key="item.path">
            <!-- Collapsible group (has children) -->
            <template v-if="item.children?.length">
              <button
                type="button"
                class="components-layout-app-sidebar__action sidebar-link"
                :class="{
                  'sidebar-link-active':
                    isCollapsedFlyoutOpen(item)
                    || (isGroupActive(item) && !isGroupExpanded(item)),
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :id="item.path === '/admin/accounts' ? 'sidebar-channel-manage' : undefined"
                :aria-expanded="sidebarCollapsed ? isCollapsedFlyoutOpen(item) : isGroupExpanded(item)"
                :aria-haspopup="sidebarCollapsed ? 'menu' : undefined"
                @click="handleGroupClick(item, $event)"
                @mouseenter="onCollapsedGroupEnter(item, $event)"
                @mouseleave="scheduleCloseCollapsedFlyout"
              >
                <component :is="item.icon" class="components-layout-app-sidebar__component" />
                <span
                  class="sidebar-label sidebar-label-flex"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                >
                  <span class="components-layout-app-sidebar__text">{{ item.label }}</span>
                  <ChevronDownIcon
                    class="components-layout-app-sidebar__chevron-down-icon"
                    :class="isGroupExpanded(item) ? 'components-layout-app-sidebar__chevron-down-icon-2' : ''"
                  />
                </span>
              </button>
              <!-- Children -->
              <div v-if="!sidebarCollapsed && isGroupExpanded(item)" class="components-layout-app-sidebar__panel">
                <router-link
                  v-for="child in item.children"
                  :key="child.path"
                  :to="{ path: child.path, query: child.query }"
                  class="components-layout-app-sidebar__router-link-3 sidebar-link"
                  :class="{ 'sidebar-link-active': route.path === child.path }"
                  @click="handleMenuItemClick(child.path)"
                >
                  <component :is="child.icon" class="components-layout-app-sidebar__component-2" />
                  <span>{{ child.label }}</span>
                  <span v-if="child.badge" class="sidebar-badge">{{ child.badge > 99 ? '99+' : child.badge }}</span>
                </router-link>
              </div>
            </template>
            <!-- Normal item (no children) -->
            <router-link
              v-else
              :to="{ path: item.path, query: item.query }"
              class="components-layout-app-sidebar__router-link-4 sidebar-link"
              :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :id="
                item.path === '/admin/accounts'
                  ? 'sidebar-channel-manage'
                  : item.path === '/admin/groups'
                    ? 'sidebar-group-manage'
                    : item.path === '/admin/redeem'
                      ? 'sidebar-wallet'
                      : undefined
              "
              @click="handleMenuItemClick(item.path)"
            >
              <span v-if="item.iconSvg" class="components-layout-app-sidebar__component sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
              <component v-else :is="item.icon" class="components-layout-app-sidebar__component" />
              <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
              <span v-if="item.badge" class="sidebar-badge">{{ item.badge > 99 ? '99+' : item.badge }}</span>
            </router-link>
          </template>
        </div>

        <!-- Personal Section for Admin (hidden in simple mode) -->
        <div v-if="!adminMenuOnly && !authStore.isSimpleMode" class="sidebar-section">
          <div class="sidebar-section-title" :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
            <span class="sidebar-section-title-text" :class="{ 'sidebar-section-title-text-collapsed': sidebarCollapsed }">
              {{ t('nav.myAccount') }}
            </span>
          </div>

          <router-link
            v-for="item in personalNavItems"
            :key="item.path"
            :to="item.path"
            class="components-layout-app-sidebar__router-link-4 sidebar-link"
            :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
            :title="sidebarCollapsed ? item.label : undefined"
            :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
            @click="handleMenuItemClick(item.path)"
          >
            <span v-if="item.iconSvg" class="components-layout-app-sidebar__component sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
            <component v-else :is="item.icon" class="components-layout-app-sidebar__component" />
            <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
            <span v-if="item.badge" class="sidebar-badge">{{ item.badge > 99 ? '99+' : item.badge }}</span>
          </router-link>
        </div>
      </template>

      <!-- Regular User View -->
      <template v-else-if="!appStore.backendModeEnabled">
        <div class="sidebar-section">
          <router-link
            v-for="item in userNavItems"
            :key="item.path"
            :to="item.path"
            class="components-layout-app-sidebar__router-link-4 sidebar-link"
            :class="{ 'sidebar-link-active': isActive(item.path), 'sidebar-link-collapsed': sidebarCollapsed }"
            :title="sidebarCollapsed ? item.label : undefined"
            :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
            @click="handleMenuItemClick(item.path)"
          >
            <span v-if="item.iconSvg" class="components-layout-app-sidebar__component sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
            <component v-else :is="item.icon" class="components-layout-app-sidebar__component" />
            <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ item.label }}</span>
            <span v-if="item.badge" class="sidebar-badge">{{ item.badge > 99 ? '99+' : item.badge }}</span>
          </router-link>
        </div>
      </template>
    </nav>

    <!-- Bottom Section -->
    <div
      class="components-layout-app-sidebar__panel-2"
      :class="adminMenuOnly && 'components-layout-app-sidebar__panel-4'"
    >
      <!-- Theme Toggle -->
      <button
        v-if="!adminMenuOnly"
        @click="toggleTheme"
        class="components-layout-app-sidebar__action-2 sidebar-link"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? (isDark ? t('nav.lightMode') : t('nav.darkMode')) : undefined"
      >
        <SunIcon v-if="isDark" class="components-layout-app-sidebar__sun-icon" />
        <MoonIcon v-else class="components-layout-app-sidebar__component" />
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{
          isDark ? t('nav.lightMode') : t('nav.darkMode')
        }}</span>
      </button>

      <!-- Collapse Button -->
      <button
        @click="toggleSidebar"
        class="components-layout-app-sidebar__action-3 sidebar-link"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? t('nav.expand') : t('nav.collapse')"
      >
        <ChevronDoubleLeftIcon v-if="!sidebarCollapsed" class="components-layout-app-sidebar__component" />
        <ChevronDoubleRightIcon v-else class="components-layout-app-sidebar__component" />
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{
          sidebarCollapsed ? t('nav.expand') : t('nav.collapse')
        }}</span>
      </button>
    </div>
  </aside>

  <Teleport to="body">
    <div
      v-if="collapsedFlyoutItem?.children?.length"
      ref="collapsedFlyoutPanelRef"
      class="dropdown sidebar-collapsed-flyout"
      :style="collapsedFlyoutStyle"
      role="menu"
      :aria-label="collapsedFlyoutItem.label"
      @mouseenter="clearCollapsedFlyoutLeaveTimer"
      @mouseleave="scheduleCloseCollapsedFlyout"
      @click.stop
    >
      <p class="sidebar-collapsed-flyout__title">{{ collapsedFlyoutItem.label }}</p>
      <router-link
        v-for="child in collapsedFlyoutItem.children"
        :key="child.path"
        :to="{ path: child.path, query: child.query }"
        class="dropdown-item"
        :class="{ 'sidebar-collapsed-flyout__item--active': route.path === child.path }"
        role="menuitem"
        @click="onCollapsedFlyoutChildClick(child.path)"
      >
        <component :is="child.icon" class="components-layout-app-sidebar__component-2" />
        <span>{{ child.label }}</span>
        <span v-if="child.badge" class="sidebar-badge">{{ child.badge > 99 ? '99+' : child.badge }}</span>
      </router-link>
    </div>
  </Teleport>

  <!-- Mobile Overlay -->
  <transition name="fade">
    <div
      v-if="mobileOpen"
      class="components-layout-app-sidebar__panel-3"
      @click="closeMobile"
    ></div>
  </transition>
</template>

<script setup lang="ts">
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import gsap from 'gsap'
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import VersionBadge from '@/components/common/VersionBadge.vue'
import { sanitizeSvg } from '@/utils/sanitize'
import { sanitizeUrl } from '@/utils/url'
import {
  getNavigationIndicatorGeometry,
  setNavigationIndicatorGeometry
} from '@/utils/navigationMotion'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import { adminSupportTicketsAPI, supportTicketsAPI } from '@/api/supportTickets'

interface NavItem {
  path: string
  query?: Record<string, string>
  label: string
  icon: unknown
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: NavItem[]
  /**
   * When true, the parent item only toggles the expand/collapse state and
   * does NOT navigate to its `path`. The `path` is purely a stable key.
   */
  expandOnly?: boolean
  /**
   * 可选的功能开关 getter。返回 false 时菜单项被隐藏；返回 undefined/true 时显示。
   * 宽容策略（undefined → 显示）避免 public settings 未加载完成时菜单闪烁消失。
   * Getter 里访问的 reactive 来源（store / composable）会被 computed 自动追踪，
   * 开关切换时菜单自动更新。
   */
  featureFlag?: () => boolean | undefined
  badge?: number
}

interface NavSection {
  key: string
  label: string
  items: NavItem[]
}

const { adminMenuOnly = false } = defineProps<{
  adminMenuOnly?: boolean
}>()

// applyFeatureFlags 递归过滤掉 featureFlag() === false 的节点（含子节点）。
// 使用 `!== false` 宽容语义：undefined（设置未加载）或 true 都视为显示。
function applyFeatureFlags(items: NavItem[]): NavItem[] {
  const out: NavItem[] = []
  for (const item of items) {
    if (item.featureFlag && item.featureFlag() === false) continue
    if (item.children) {
      out.push({ ...item, children: applyFeatureFlags(item.children) })
    } else {
      out.push(item)
    }
  }
  return out
}

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const adminSettingsStore = useAdminSettingsStore()
const supportUserSummary = ref({ total: 0, unread: 0, featureEnabled: false, loaded: false })
const supportAdminSummary = ref({ total: 0, unread: 0, featureEnabled: false, loaded: false })

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed && !appStore.mobileOpen)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)
const sidebarNavRef = ref<HTMLElement | null>(null)
const sidebarActiveIndicatorRef = ref<HTMLElement | null>(null)
let sidebarActiveIndicatorReady = false
let sidebarResizeObserver: ResizeObserver | null = null
const collapsedFlyoutPath = ref<string | null>(null)
const collapsedFlyoutPinned = ref(false)
const collapsedFlyoutTriggerRef = ref<HTMLElement | null>(null)
const collapsedFlyoutOpen = computed(() => sidebarCollapsed.value && collapsedFlyoutPath.value !== null)
const { panelRef: collapsedFlyoutPanelRef, style: collapsedFlyoutStyle, update: updateCollapsedFlyout } = useFloatingPanel(
  collapsedFlyoutTriggerRef,
  collapsedFlyoutOpen,
  {
    placement: 'right',
    align: 'start',
    gap: 8,
    maxWidth: 220,
    maxHeightRatio: 0.7,
    minComfortableHeight: 96,
    zIndex: 50
  }
)
let collapsedFlyoutLeaveTimer: number | null = null
const isDark = ref(document.documentElement.classList.contains('dark'))

const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

// Track which parent nav groups are expanded
const expandedGroups = ref<Set<string>>(new Set())

// Site settings from appStore (cached, no flicker)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteVersion = computed(() => appStore.siteVersion)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

// SVG Icon Components
const DashboardIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z'
        })
      ]
    )
}

const KeyIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z'
        })
      ]
    )
}

const ChartIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z'
        })
      ]
    )
}

const GiftIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M21 11.25v8.25a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5v-8.25M12 4.875A2.625 2.625 0 109.375 7.5H12m0-2.625V7.5m0-2.625A2.625 2.625 0 1114.625 7.5H12m0 0V21m-8.625-9.75h18c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125h-18c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z'
        })
      ]
    )
}

const UserIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z'
        })
      ]
    )
}

const UsersIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z'
        })
      ]
    )
}

const FolderIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z'
        })
      ]
    )
}

const ChannelIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0l4.179 2.25L12 17.25 2.25 12m15.321-2.25l4.179 2.25L12 17.25l-9.75-5.25'
        })
      ]
    )
}

const CreditCardIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z'
        })
      ]
    )
}

const RechargeSubscriptionIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'currentColor', viewBox: '0 0 1024 1024' },
      [
        h('path', {
          d: 'M512 992C247.3 992 32 776.7 32 512S247.3 32 512 32s480 215.3 480 480c0 84.4-22.2 167.4-64.2 240-8.9 15.3-28.4 20.6-43.7 11.7-15.3-8.8-20.5-28.4-11.7-43.7 36.4-62.9 55.6-134.8 55.6-208 0-229.4-186.6-416-416-416S96 282.6 96 512s186.6 416 416 416c17.7 0 32 14.3 32 32s-14.3 32-32 32z'
        }),
        h('path', {
          d: 'M640 512H384c-17.7 0-32-14.3-32-32s14.3-32 32-32h256c17.7 0 32 14.3 32 32s-14.3 32-32 32zM640 640H384c-17.7 0-32-14.3-32-32s14.3-32 32-32h256c17.7 0 32 14.3 32 32s-14.3 32-32 32z'
        }),
        h('path', {
          d: 'M512 480c-8.2 0-16.4-3.1-22.6-9.4l-128-128c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l128 128c12.5 12.5 12.5 32.8 0 45.3-6.3 6.3-14.5 9.4-22.7 9.4z'
        }),
        h('path', {
          d: 'M512 480c-8.2 0-16.4-3.1-22.6-9.4-12.5-12.5-12.5-32.8 0-45.3l128-128c12.5-12.5 32.8-12.5 45.3 0s12.5 32.8 0 45.3l-128 128c-6.3 6.3-14.5 9.4-22.7 9.4z'
        }),
        h('path', {
          d: 'M512 736c-17.7 0-32-14.3-32-32V448c0-17.7 14.3-32 32-32s32 14.3 32 32v256c0 17.7-14.3 32-32 32zM896 992H512c-17.7 0-32-14.3-32-32s14.3-32 32-32h306.8l-73.4-73.4c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l128 128c9.2 9.2 11.9 22.9 6.9 34.9S908.9 992 896 992z'
        })
      ]
    )
}

const GlobeIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418'
        })
      ]
    )
}

const ServerIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z'
        })
      ]
    )
}

const DownloadIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M7.5 10.5L12 15m0 0l4.5-4.5M12 15V3'
        })
      ]
    )
}

const BellIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9a6 6 0 10-12 0v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0'
        })
      ]
    )
}

const TicketIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M16.5 6v.75m0 3v.75m0 3v.75m0 3V18m-9-5.25h5.25M7.5 15h3M3.375 5.25c-.621 0-1.125.504-1.125 1.125v3.026a2.999 2.999 0 010 5.198v3.026c0 .621.504 1.125 1.125 1.125h17.25c.621 0 1.125-.504 1.125-1.125v-3.026a2.999 2.999 0 010-5.198V6.375c0-.621-.504-1.125-1.125-1.125H3.375z'
        })
      ]
    )
}

const CogIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z'
        }),
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z'
        })
      ]
    )
}

const SunIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z'
        })
      ]
    )
}

const MoonIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z'
        })
      ]
    )
}

const ChevronDoubleLeftIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm18.75 4.5-7.5 7.5 7.5 7.5m-6-15L5.25 12l7.5 7.5'
        })
      ]
    )
}

const OrderIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15a2.25 2.25 0 012.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z'
        })
      ]
    )
}

const OrderListIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z'
        })
      ]
    )
}

const ChevronDoubleRightIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm5.25 4.5 7.5 7.5-7.5 7.5m6-15 7.5 7.5-7.5 7.5'
        })
      ]
    )
}

const SignalIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.348 14.651a3.75 3.75 0 010-5.303m5.304 0a3.75 3.75 0 010 5.303m-7.425 2.122a6.75 6.75 0 010-9.546m9.546 0a6.75 6.75 0 010 9.546M5.106 18.894c-3.808-3.807-3.808-9.98 0-13.788m13.788 0c3.808 3.807 3.808 9.98 0 13.788M12 12h.008v.008H12V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z'
        })
      ]
    )
}

const ShieldIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z'
        })
      ]
    )
}

const PriceTagIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z'
        }),
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M6 6h.008v.008H6V6z'
        })
      ]
    )
}

const ChevronDownIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'm19.5 8.25-7.5 7.5-7.5-7.5'
        })
      ]
    )
}

// Public-settings flags go through the registry in utils/featureFlags.ts,
// which handles the opt-in vs opt-out fallback when settings haven't loaded
// yet. Admin-only flags (not in public settings) stay inline below.
const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled
const flagAdminChannelMonitor = () => adminSettingsStore.channelMonitorEnabled
const flagAdminModelPlaza = () => adminSettingsStore.modelPlazaEnabled
const flagAdminRiskControl = () => adminSettingsStore.riskControlEnabled
const flagAdminAffiliate = () => adminSettingsStore.affiliateEnabled
// Ticket navigation is fail-closed while its summary is loading. Public settings are
// injected before the shell renders, so an enabled feature still appears immediately;
// disabled users only see the entry after the summary confirms historical tickets.
const flagSupportTickets = () =>
  appStore.cachedPublicSettings?.support_tickets_enabled === true ||
  (supportUserSummary.value.loaded && (
    supportUserSummary.value.featureEnabled || supportUserSummary.value.total > 0
  ))

// buildSelfNavItems 构造用户自己的导航项（用户端主菜单和管理员的"我的账户"子菜单共享这组声明）。
// withDashboard=true 时包含仪表盘（用户端），false 时不含（管理员的个人区已经有独立仪表盘入口）。
//
// 条目顺序与顶部导航(AppTopNav)保持一致：高频核心（密钥/用量）→
// 账务闭环（购买/订阅/订单/兑换）→ 渠道信息（可用渠道/渠道状态）→ 其他（推广/资料）。
function buildSelfNavItems(withDashboard: boolean): NavItem[] {
  const items: NavItem[] = []
  if (withDashboard) {
    items.push({ path: '/dashboard', label: t('nav.dashboard'), icon: DashboardIcon })
  }
  items.push(
    { path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon },
    { path: '/usage', label: t('nav.usage'), icon: ChartIcon, hideInSimpleMode: true },
    { path: '/purchase', label: t('nav.buySubscription'), icon: RechargeSubscriptionIcon, hideInSimpleMode: true, featureFlag: flagPayment },
    { path: '/subscriptions', label: t('nav.mySubscriptions'), icon: CreditCardIcon, hideInSimpleMode: true },
    { path: '/orders', label: t('nav.myOrders'), icon: OrderListIcon, hideInSimpleMode: true, featureFlag: flagPayment },
    { path: '/tickets', label: t('nav.supportTickets'), icon: TicketIcon, hideInSimpleMode: true, featureFlag: flagSupportTickets, badge: supportUserSummary.value.unread },
    { path: '/redeem', label: t('nav.redeem'), icon: GiftIcon, hideInSimpleMode: true },
    { path: '/available-channels', label: t('nav.availableChannels'), icon: ChannelIcon, hideInSimpleMode: true, featureFlag: flagAvailableChannels },
    { path: '/monitor', label: t('nav.channelStatus'), icon: SignalIcon, featureFlag: flagChannelMonitor },
    { path: '/affiliate', label: t('nav.affiliate'), icon: UsersIcon, hideInSimpleMode: true, featureFlag: flagAffiliate },
    { path: '/profile', label: t('nav.profile'), icon: UserIcon },
    ...customMenuItemsForUser.value.map((item): NavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      icon: null,
      iconSvg: item.icon_svg,
    })),
  )
  return items
}

// finalizeNav 合并三重过滤：featureFlag 过滤 + simple 模式过滤。
function finalizeNav(items: NavItem[]): NavItem[] {
  const visible = applyFeatureFlags(items)
  return authStore.isSimpleMode ? visible.filter(item => !item.hideInSimpleMode) : visible
}

// User navigation items (for regular users)
const userNavItems = computed((): NavItem[] => finalizeNav(buildSelfNavItems(true)))

// Personal navigation items (for admin's "My Account" section, without Dashboard).
// Admins access 可用渠道 from this section just like regular users — there is no
// separate admin entry, since the page is purely a user-facing view.
const personalNavItems = computed((): NavItem[] => finalizeNav(buildSelfNavItems(false)))

// Custom menu items filtered by visibility
const customMenuItemsForUser = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
})

const customMenuItemsForAdmin = computed(() => {
  return adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
})

// Admin navigation grouped by responsibility. Routes and permission rules stay unchanged.
const adminNavSections = computed((): NavSection[] => {
  const sections: NavSection[] = [
    {
      key: 'overview',
      label: t('nav.adminOverview'),
      items: [{ path: '/admin/dashboard', label: t('nav.dashboard'), icon: DashboardIcon }],
    },
    {
      key: 'resources',
      label: t('nav.resourceManagement'),
      items: [
        {
          path: '/admin/accounts',
          label: t('nav.accounts'),
          icon: GlobeIcon,
          children: [
            { path: '/admin/accounts', label: t('nav.accountList'), icon: GlobeIcon },
          ],
        },
        { path: '/admin/groups', label: t('nav.groups'), icon: FolderIcon, hideInSimpleMode: true },
        {
          path: '/admin/channels',
          label: t('nav.channelManagement'),
          icon: ChannelIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            { path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: PriceTagIcon },
            {
              path: '/admin/channels/monitor',
              label: t('nav.channelMonitor'),
              icon: SignalIcon,
              featureFlag: flagAdminChannelMonitor,
            },
            { path: '/admin/channels/monitor/settings', label: t('nav.settingsChannelMonitor'), icon: CogIcon },
          ],
        },
        { path: '/admin/proxies', label: t('nav.proxies'), icon: ServerIcon },
        {
          path: '/model-plaza',
          label: t('nav.modelPlaza'),
          icon: PriceTagIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            {
              path: '/model-plaza',
              query: { embedded: '1' },
              label: t('nav.modelPlaza'),
              icon: PriceTagIcon,
              featureFlag: flagAdminModelPlaza,
            },
            { path: '/admin/model-plaza/settings', label: t('nav.settingsModelPlaza'), icon: CogIcon },
          ],
        },
      ],
    },
    {
      key: 'users-business',
      label: t('nav.userBusinessManagement'),
      items: [
        {
          path: '/admin/users',
          label: t('nav.users'),
          icon: UsersIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            { path: '/admin/users', label: t('nav.userList'), icon: UsersIcon },
            { path: '/admin/users/levels', label: t('nav.userLevels'), icon: ChartIcon },
          ],
        },
        { path: '/admin/subscriptions', label: t('nav.subscriptions'), icon: CreditCardIcon, hideInSimpleMode: true },
        { path: '/admin/tickets', label: t('nav.ticketManagement'), icon: TicketIcon, badge: supportAdminSummary.value.unread },
        {
          path: '/admin/orders',
          label: t('nav.orderManagement'),
          icon: OrderIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            {
              path: '/admin/orders/dashboard',
              label: t('nav.paymentDashboard'),
              icon: ChartIcon,
              featureFlag: flagAdminPayment,
            },
            { path: '/admin/orders', label: t('nav.orderManagement'), icon: OrderIcon, featureFlag: flagAdminPayment },
            {
              path: '/admin/orders/plans',
              label: t('nav.paymentPlans'),
              icon: CreditCardIcon,
              featureFlag: flagAdminPayment,
            },
            { path: '/admin/orders/settings', label: t('nav.settingsPayment'), icon: CogIcon },
          ],
        },
        { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: TicketIcon, hideInSimpleMode: true },
        { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: GiftIcon, hideInSimpleMode: true },
        {
          path: '/admin/affiliates',
          label: t('nav.affiliateManagement'),
          icon: UsersIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            {
              path: '/admin/affiliates/invites',
              label: t('nav.affiliateInviteRecords'),
              icon: UsersIcon,
              featureFlag: flagAdminAffiliate,
            },
            {
              path: '/admin/affiliates/rebates',
              label: t('nav.affiliateRebateRecords'),
              icon: OrderIcon,
              featureFlag: flagAdminAffiliate,
            },
            {
              path: '/admin/affiliates/transfers',
              label: t('nav.affiliateTransferRecords'),
              icon: CreditCardIcon,
              featureFlag: flagAdminAffiliate,
            },
            { path: '/admin/affiliates/settings', label: t('nav.settingsAffiliate'), icon: CogIcon },
          ],
        },
        { path: '/admin/announcements', label: t('nav.announcements'), icon: BellIcon },
      ],
    },
    {
      key: 'operations-security',
      label: t('nav.operationsSecurity'),
      items: [
        { path: '/admin/usage', label: t('nav.usage'), icon: ChartIcon },
        {
          path: '/admin/ops',
          label: t('nav.ops'),
          icon: ChartIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            { path: '/admin/ops', label: t('nav.opsDashboard'), icon: ChartIcon, featureFlag: flagOpsMonitoring },
            { path: '/admin/ops/settings', label: t('nav.settingsOpsMonitoring'), icon: CogIcon },
          ],
        },
        {
          path: '/admin/security-audit',
          label: t('nav.securityAudit'),
          icon: ShieldIcon,
          expandOnly: true,
          children: [
            { path: '/admin/risk-control', label: t('nav.contentModeration'), icon: ShieldIcon },
            {
              path: '/admin/prompt-audit',
              label: t('nav.promptAudit'),
              icon: ShieldIcon,
              featureFlag: flagAdminRiskControl,
            },
          ],
        },
        { path: '/admin/audit-logs', label: t('nav.auditLogs'), icon: ShieldIcon, hideInSimpleMode: true },
      ],
    },
    {
      key: 'system',
      label: t('nav.systemManagement'),
      items: [
        {
          path: '/admin/settings',
          label: t('nav.settings'),
          icon: CogIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            { path: '/admin/settings/platform', label: t('nav.settingsPlatform'), icon: DashboardIcon },
            { path: '/admin/settings/features', label: t('nav.settingsFeatures'), icon: CogIcon },
            { path: '/admin/settings/access', label: t('nav.settingsAccess'), icon: UsersIcon },
            { path: '/admin/settings/gateway', label: t('nav.settingsGateway'), icon: ServerIcon },
            { path: '/admin/settings/compliance', label: t('nav.settingsCompliance'), icon: ShieldIcon },
            { path: '/admin/settings/operations', label: t('nav.settingsOperations'), icon: CogIcon },
            { path: '/admin/settings/updates', label: t('nav.settingsUpdates'), icon: DownloadIcon },
          ],
        },
      ],
    },
  ]

  const visibleSections = sections.map((section) => ({
    ...section,
    items: applyFeatureFlags(section.items),
  }))
  const customItems: NavItem[] = customMenuItemsForAdmin.value.map((item) => ({
    path: `/custom/${item.id}`,
    label: item.label,
    icon: null,
    iconSvg: item.icon_svg,
  }))

  if (authStore.isSimpleMode) {
    const simpleSections = visibleSections.map((section) => ({
      ...section,
      items: section.items.filter((item) => !item.hideInSimpleMode),
    }))

    simpleSections.find((section) => section.key === 'resources')?.items.push({
      path: '/keys',
      label: t('nav.apiKeys'),
      icon: KeyIcon,
    })
    simpleSections.find((section) => section.key === 'system')?.items.push({
      path: '/admin/settings/features',
      label: t('nav.settings'),
      icon: CogIcon,
    })

    if (customItems.length > 0) {
      simpleSections.push({ key: 'custom', label: t('nav.customPages'), items: customItems })
    }
    return simpleSections.filter((section) => section.items.length > 0)
  }

  if (customItems.length > 0) {
    visibleSections.push({ key: 'custom', label: t('nav.customPages'), items: customItems })
  }
  return visibleSections.filter((section) => section.items.length > 0)
})

function toggleSidebar() {
  appStore.toggleSidebar()
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function closeMobile() {
  appStore.setMobileOpen(false)
}

function revealAdminMenu() {
  if (window.innerWidth < 1024) {
    appStore.setMobileOpen(true)
    return
  }
  appStore.setSidebarCollapsed(false)
}

function handleMenuItemClick(itemPath: string) {
  if (mobileOpen.value) {
    setTimeout(() => {
      appStore.setMobileOpen(false)
    }, 150)
  }

  // Map paths to tour selectors
  const pathToSelector: Record<string, string> = {
    '/admin/groups': '#sidebar-group-manage',
    '/admin/accounts': '#sidebar-channel-manage',
    '/keys': '[data-tour="sidebar-my-keys"]'
  }

  const selector = pathToSelector[itemPath]
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}

function isGroupActive(item: NavItem): boolean {
  if (!item.children) return false
  return item.children.some(child => route.path === child.path)
}

function isGroupExpanded(item: NavItem): boolean {
  return expandedGroups.value.has(item.path) || isGroupActive(item)
}

function getSidebarNavigationGeometry(target: HTMLElement) {
  const nav = sidebarNavRef.value
  if (!nav) return null
  const navRect = nav.getBoundingClientRect()
  const targetRect = target.getBoundingClientRect()
  return {
    x: targetRect.left - navRect.left + nav.scrollLeft,
    y: targetRect.top - navRect.top + nav.scrollTop,
    width: targetRect.width,
    height: targetRect.height
  }
}

function primeSidebarActiveIndicator(event: PointerEvent): void {
  if (event.button !== 0) return
  const target = (event.target as HTMLElement | null)?.closest<HTMLElement>('.sidebar-link')
  const indicator = sidebarActiveIndicatorRef.value
  const geometry = target ? getSidebarNavigationGeometry(target) : null
  if (!indicator || !geometry) return

  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  gsap.to(indicator, {
    ...geometry,
    opacity: 1,
    duration: reduceMotion ? 0 : 0.42,
    ease: 'power3.out',
    overwrite: 'auto'
  })
  sidebarActiveIndicatorReady = true
}

function persistCurrentSidebarIndicatorGeometry(): void {
  const indicator = sidebarActiveIndicatorRef.value
  if (!indicator || !sidebarActiveIndicatorReady) return
  const geometry = getSidebarNavigationGeometry(indicator)
  if (geometry) setNavigationIndicatorGeometry('admin-sidebar', geometry)
}

function updateSidebarActiveIndicator(animate = true): void {
  void nextTick(() => {
    window.requestAnimationFrame(() => {
      const nav = sidebarNavRef.value
      const indicator = sidebarActiveIndicatorRef.value
      const activeItem = nav?.querySelector<HTMLElement>('.sidebar-link-active')
      if (!nav || !indicator || !activeItem) {
        if (indicator) gsap.set(indicator, { opacity: 0 })
        return
      }

      const navRect = nav.getBoundingClientRect()
      const itemRect = activeItem.getBoundingClientRect()
      const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
      const nextGeometry = {
        x: itemRect.left - navRect.left + nav.scrollLeft,
        y: itemRect.top - navRect.top + nav.scrollTop,
        width: itemRect.width,
        height: itemRect.height
      }

      if (!sidebarActiveIndicatorReady && animate && !reduceMotion) {
        const previousGeometry = getNavigationIndicatorGeometry('admin-sidebar')
        if (previousGeometry) {
          gsap.set(indicator, { ...previousGeometry, opacity: 1 })
          sidebarActiveIndicatorReady = true
        }
      }
      const duration = animate && sidebarActiveIndicatorReady && !reduceMotion ? 0.42 : 0

      gsap.to(indicator, {
        ...nextGeometry,
        opacity: 1,
        duration,
        ease: 'power3.out',
        overwrite: 'auto'
      })
      setNavigationIndicatorGeometry('admin-sidebar', nextGeometry)
      sidebarActiveIndicatorReady = true
    })
  })
}

function updateSidebarActiveIndicatorOnResize(): void {
  updateSidebarActiveIndicator(false)
}

function toggleGroup(item: NavItem) {
  if (expandedGroups.value.has(item.path)) {
    expandedGroups.value.delete(item.path)
  } else {
    expandedGroups.value.add(item.path)
  }
}

function findAdminNavItem(path: string): NavItem | null {
  for (const section of adminNavSections.value) {
    const found = section.items.find((item) => item.path === path)
    if (found) return found
  }
  return null
}

const collapsedFlyoutItem = computed(() => (
  collapsedFlyoutPath.value ? findAdminNavItem(collapsedFlyoutPath.value) : null
))

function isCollapsedFlyoutOpen(item: NavItem): boolean {
  return collapsedFlyoutOpen.value && collapsedFlyoutPath.value === item.path
}

function clearCollapsedFlyoutLeaveTimer() {
  if (collapsedFlyoutLeaveTimer !== null) {
    window.clearTimeout(collapsedFlyoutLeaveTimer)
    collapsedFlyoutLeaveTimer = null
  }
}

function closeCollapsedFlyout() {
  clearCollapsedFlyoutLeaveTimer()
  collapsedFlyoutPath.value = null
  collapsedFlyoutPinned.value = false
  collapsedFlyoutTriggerRef.value = null
}

function openCollapsedFlyout(item: NavItem, trigger: EventTarget | null) {
  if (!sidebarCollapsed.value) return
  clearCollapsedFlyoutLeaveTimer()
  collapsedFlyoutPath.value = item.path
  collapsedFlyoutTriggerRef.value = trigger instanceof HTMLElement ? trigger : null
  void nextTick(updateCollapsedFlyout)
}

function onCollapsedGroupEnter(item: NavItem, event: MouseEvent) {
  openCollapsedFlyout(item, event.currentTarget)
}

function scheduleCloseCollapsedFlyout() {
  if (collapsedFlyoutPinned.value) return
  clearCollapsedFlyoutLeaveTimer()
  collapsedFlyoutLeaveTimer = window.setTimeout(() => {
    closeCollapsedFlyout()
  }, 160)
}

function onCollapsedFlyoutChildClick(path: string) {
  closeCollapsedFlyout()
  handleMenuItemClick(path)
}

function onCollapsedFlyoutPointerDown(event: MouseEvent) {
  const target = event.target as Node | null
  if (!target) return
  if (collapsedFlyoutTriggerRef.value?.contains(target)) return
  if (collapsedFlyoutPanelRef.value?.contains(target)) return
  closeCollapsedFlyout()
}

function onCollapsedFlyoutKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closeCollapsedFlyout()
}

/**
 * Click handler for collapsible parent items.
 * - When sidebar is collapsed: open a flyout with the secondary menu.
 * - When `expandOnly` is true: only toggle expand state.
 * - Otherwise (default, e.g. /admin/orders): navigate to the parent path
 *   (router-link semantics) and ensure the group is expanded.
 */
function handleGroupClick(item: NavItem, event?: MouseEvent) {
  if (sidebarCollapsed.value) {
    if (isCollapsedFlyoutOpen(item) && collapsedFlyoutPinned.value) {
      closeCollapsedFlyout()
      return
    }
    collapsedFlyoutPinned.value = true
    openCollapsedFlyout(item, event?.currentTarget ?? null)
    return
  }
  if (item.expandOnly) {
    toggleGroup(item)
    return
  }
  // Push to path and ensure expanded
  if (route.path !== item.path) {
    router.push(item.path)
  }
  if (!expandedGroups.value.has(item.path)) {
    expandedGroups.value.add(item.path)
  }
  handleMenuItemClick(item.path)
}

// Initialize theme
const savedTheme = localStorage.getItem('theme')
if (
  savedTheme === 'dark' ||
  (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
) {
  isDark.value = true
  document.documentElement.classList.add('dark')
}

// Fetch admin settings (for feature-gated nav items like Ops).
watch(
  isAdmin,
  (v) => {
    if (v) {
      adminSettingsStore.fetch()
    }
  },
  { immediate: true }
)

async function fetchSupportSummary() {
  try {
    const [userResponse, adminResponse] = await Promise.all([
      supportTicketsAPI.summary(),
      isAdmin.value ? adminSupportTicketsAPI.summary() : Promise.resolve(null),
    ])
    supportUserSummary.value = {
      total: userResponse.data.total || 0,
      unread: userResponse.data.unread || 0,
      featureEnabled: userResponse.data.feature_enabled || false,
      loaded: true,
    }
    if (adminResponse) {
      supportAdminSummary.value = {
        total: adminResponse.data.total || 0,
        unread: adminResponse.data.unread || 0,
        featureEnabled: adminResponse.data.feature_enabled || false,
        loaded: true,
      }
    }
  } catch {
    supportUserSummary.value.loaded = true
    supportAdminSummary.value.loaded = true
  }
}

watch(collapsedFlyoutOpen, (open) => {
  if (open) {
    document.addEventListener('mousedown', onCollapsedFlyoutPointerDown)
    document.addEventListener('keydown', onCollapsedFlyoutKeydown)
  } else {
    document.removeEventListener('mousedown', onCollapsedFlyoutPointerDown)
    document.removeEventListener('keydown', onCollapsedFlyoutKeydown)
  }
})

watch(sidebarCollapsed, () => {
  closeCollapsedFlyout()
  updateSidebarActiveIndicator()
})

watch(() => route.fullPath, () => {
  closeCollapsedFlyout()
  updateSidebarActiveIndicator()
}, { flush: 'post' })

watch(
  () => Array.from(expandedGroups.value).sort().join('|'),
  () => updateSidebarActiveIndicator(),
  { flush: 'post' }
)

watch(
  () => adminNavSections.value.flatMap((section) => section.items.map((item) => item.path)).join('|'),
  () => updateSidebarActiveIndicator(false),
  { flush: 'post' }
)

watch(collapsedFlyoutPath, () => {
  if (collapsedFlyoutOpen.value) void nextTick(updateCollapsedFlyout)
})

onMounted(() => {
  window.addEventListener('app:open-admin-menu', revealAdminMenu)
  window.addEventListener('support-tickets:updated', fetchSupportSummary)
  void fetchSupportSummary()
  if (isAdmin.value) {
    adminSettingsStore.fetch()
  }
  updateSidebarActiveIndicator()
  if (typeof ResizeObserver !== 'undefined') {
    sidebarResizeObserver = new ResizeObserver(() => updateSidebarActiveIndicator())
    if (sidebarNavRef.value) sidebarResizeObserver.observe(sidebarNavRef.value)
  }
  window.addEventListener('resize', updateSidebarActiveIndicatorOnResize)
  // Restore sidebar scroll position after route change re-mounts the component
  if (appStore.sidebarScrollTop > 0 && sidebarNavRef.value) {
    void nextTick(() => {
      if (sidebarNavRef.value) {
        sidebarNavRef.value.scrollTop = appStore.sidebarScrollTop
      }
    })
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('app:open-admin-menu', revealAdminMenu)
  window.removeEventListener('support-tickets:updated', fetchSupportSummary)
  window.removeEventListener('resize', updateSidebarActiveIndicatorOnResize)
  document.removeEventListener('mousedown', onCollapsedFlyoutPointerDown)
  document.removeEventListener('keydown', onCollapsedFlyoutKeydown)
  sidebarResizeObserver?.disconnect()
  persistCurrentSidebarIndicatorGeometry()
  if (sidebarActiveIndicatorRef.value) gsap.killTweensOf(sidebarActiveIndicatorRef.value)
  closeCollapsedFlyout()
  if (sidebarNavRef.value) {
    appStore.sidebarScrollTop = sidebarNavRef.value.scrollTop
  }
})
</script>

<style scoped>
.sidebar-logo {
  flex: 0 0 2.25rem;
  min-width: 2.25rem;
}

.sidebar-active-indicator {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 0;
  display: block;
  width: 0;
  height: 0;
  pointer-events: none;
  opacity: 0;
  border: 1px solid var(--glass-border-active);
  border-radius: var(--radius-lg);
  background-color: var(--glass-tint-brand);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow:
    0 4px 16px color-mix(in srgb, var(--theme-accent) 12%, transparent),
    0 1px 0 var(--glass-highlight-hover) inset;
  will-change: transform, width, height;
}

.sidebar-badge {
  margin-left: auto;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  display: inline-grid;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--color-danger) 35%, transparent);
  border-radius: 999px;
  background: var(--glass-tint-danger);
  color: var(--color-text-danger);
  font-size: var(--font-size-2xs);
  line-height: 1;
}

.sidebar-header-collapsed {
  gap: 0;
  padding-left: 1.125rem;
  padding-right: 1.125rem;
}

.sidebar-brand {
  min-width: 0;
  flex: 1 1 auto;
  white-space: nowrap;
  transition:
    max-width 0.22s ease,
    opacity 0.14s ease,
    transform 0.14s ease;
  max-width: 12rem;
}

.sidebar-brand-collapsed {
  max-width: 0;
  overflow: hidden;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

.sidebar-brand-title {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-link-collapsed {
  justify-content: center;
  gap: 0;
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.sidebar-link-collapsed .sidebar-badge {
  display: none;
}

.sidebar-section-title {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 1.25rem;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar-section-title-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}

.sidebar-section-title::after {
  content: '';
  position: absolute;
  left: 0.75rem;
  right: 0.75rem;
  top: 50%;
  height: 1px;
  background: var(--color-border);
  opacity: 0;
  transform: translateY(-50%);
  transition: opacity 0.18s ease;
}

.dark .sidebar-section-title::after {
  background: rgb(46 46 51);
}

.sidebar-section-title-text-collapsed {
  opacity: 0;
  transform: translateX(-4px);
}

.sidebar-section-title-collapsed::after {
  opacity: 1;
  transition-delay: 0.08s;
}

.sidebar-label {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    max-width 0.2s ease,
    opacity 0.12s ease,
    transform 0.12s ease;
  max-width: 12rem;
}

.sidebar-label-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.sidebar-label-collapsed {
  max-width: 0;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

/* Custom SVG icon in sidebar: constrain size without overriding uploaded SVG colors */
.sidebar-svg-icon {
  color: currentColor;
}

.sidebar-svg-icon :deep(svg) {
  display: block;
  width: 1.25rem;
  height: 1.25rem;
}

.sidebar-collapsed-flyout {
  overflow-y: auto;
  min-width: 11.5rem;
}

.sidebar-collapsed-flyout__title {
  margin: 0;
  padding: 0.375rem 0.75rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.sidebar-collapsed-flyout .dropdown-item {
  justify-content: flex-start;
}

.sidebar-collapsed-flyout__item--active {
  color: var(--color-text-brand);
  background-color: var(--glass-tint-brand);
}

.sidebar-collapsed-flyout :deep(svg) {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

@media (max-width: 1023px) {
  .components-layout-app-sidebar__aside-2 {
    transform: translateX(-100%);
  }

  .components-layout-app-sidebar__panel-3 {
    inset: calc(var(--app-shell-sticky-offset) + env(safe-area-inset-top, 0px)) 0 0;
    z-index: 40;
  }
}
</style>
