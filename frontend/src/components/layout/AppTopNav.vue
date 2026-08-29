<template>
  <div class="app-top-nav-wrap">
    <nav ref="navRootRef" class="app-top-nav-shell" aria-label="Primary navigation">
      <router-link
        :to="homePath"
        class="app-top-nav-brand"
        :title="siteName"
        :aria-label="siteName"
      >
        <img
          :src="siteLogo || '/logo.svg'"
          alt=""
          class="app-top-nav__logo"
        />
      </router-link>

      <button
        v-if="isAdmin"
        type="button"
        class="app-top-nav-menu-toggle"
        :title="t('common.toggleMenu')"
        :aria-label="t('common.toggleMenu')"
        :aria-expanded="appStore.mobileOpen"
        aria-controls="admin-sidebar"
        @click="appStore.toggleMobileSidebar()"
      >
        <Icon name="menu" size="md" />
      </button>

      <div class="app-top-nav-scroll">
        <div class="app-top-nav-list">
          <router-link
            v-for="item in visibleNavItems"
            :key="item.path"
            :to="{ path: item.path, query: item.query }"
            class="app-top-nav-item"
            :class="isItemActive(item) && 'app-top-nav-item-active'"
            :aria-current="isItemActive(item) ? 'page' : undefined"
            :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
            @click="handleMenuItemClick(item.path)"
          >
            <span
              v-if="item.iconSvg"
              class="app-top-nav-custom-icon"
              v-html="sanitizeSvg(item.iconSvg)"
            ></span>
            <Icon v-else :name="item.icon" size="sm" class="app-top-nav-icon" />
            <span class="app-top-nav-label">{{ item.label }}</span>
          </router-link>
        </div>
      </div>

      <div v-if="authStore.user" class="app-top-nav-user">
        <AppNavUserModule />
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import { sanitizeSvg } from '@/utils/sanitize'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'
import AppNavUserModule from './AppNavUserModule.vue'

interface NavItem {
  path: string
  label: string
  icon: any
  query?: Record<string, string>
  iconSvg?: string
  hideInSimpleMode?: boolean
  exact?: boolean
  featureFlag?: () => boolean | undefined
}

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()

const navRootRef = ref<HTMLElement | null>(null)
const isAdmin = computed(() => authStore.isAdmin)
const homePath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)

const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)

const customMenuItemsForUser = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
})

function applyFeatureFlags(items: NavItem[]): NavItem[] {
  return items.filter((item) => item.featureFlag?.() !== false)
}

function finalizeNav(items: NavItem[]): NavItem[] {
  const visible = applyFeatureFlags(items)
  if (!authStore.isSimpleMode) return visible
  return visible.filter((item) => !item.hideInSimpleMode)
}

// 菜单顺序:高频核心(仪表盘/密钥/用量)→ 账务闭环(购买/订阅/订单/兑换)
// → 渠道信息(可用渠道/渠道状态)→ 其他(推广/自定义),同类相邻。
function buildSelfNavItems(withDashboard: boolean): NavItem[] {
  const items: NavItem[] = []
  if (withDashboard) {
    items.push({ path: '/dashboard', label: t('nav.dashboard'), icon: 'home' })
  }
  items.push(
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    { path: '/usage', label: t('nav.usage'), icon: 'chart', hideInSimpleMode: true },
    {
      path: '/purchase',
      label: t('nav.buySubscription'),
      icon: 'dollar',
      hideInSimpleMode: true,
      featureFlag: flagPayment
    },
    { path: '/subscriptions', label: t('nav.mySubscriptions'), icon: 'creditCard', hideInSimpleMode: true },
    {
      path: '/orders',
      label: t('nav.myOrders'),
      icon: 'document',
      hideInSimpleMode: true,
      featureFlag: flagPayment
    },
    { path: '/redeem', label: t('nav.redeem'), icon: 'gift', hideInSimpleMode: true },
    {
      path: '/available-channels',
      label: t('nav.availableChannels'),
      icon: 'server',
      hideInSimpleMode: true,
      featureFlag: flagAvailableChannels
    },
    { path: '/monitor', label: t('nav.channelStatus'), icon: 'chartBar', featureFlag: flagChannelMonitor },
    {
      path: '/affiliate',
      label: t('nav.affiliate'),
      icon: 'users',
      hideInSimpleMode: true,
      featureFlag: flagAffiliate
    },
    { path: '/profile', label: t('nav.profile'), icon: 'user' },
    ...customMenuItemsForUser.value.map((item): NavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      icon: 'grid',
      iconSvg: item.icon_svg
    }))
  )
  return items
}

const userNavItems = computed((): NavItem[] =>
  finalizeNav(buildSelfNavItems(true)).filter((item) => item.path !== '/profile')
)

const visibleNavItems = computed(() => {
  if (appStore.backendModeEnabled && !isAdmin.value) return []
  return userNavItems.value
})

function isItemActive(item: NavItem): boolean {
  if (item.exact) return route.path === item.path
  return route.path === item.path || route.path.startsWith(`${item.path}/`)
}

function handleMenuItemClick(path: string): void {
  const selector = path === '/keys' ? '[data-tour="sidebar-my-keys"]' : undefined
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function scrollActiveItemIntoView(): void {
  window.requestAnimationFrame(() => {
    navRootRef.value
      ?.querySelector<HTMLElement>('.app-top-nav-item-active')
      ?.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
  })
}

watch(
  () => route.fullPath,
  scrollActiveItemIntoView
)

onMounted(() => {
  scrollActiveItemIntoView()
})
</script>

<style scoped>
.app-top-nav__logo {
  width: 2.25rem;
  height: 2.25rem;
  object-fit: contain;
}

.app-top-nav-wrap {
  position: sticky;
  top: 0;
  z-index: 50;
  padding: 0.75rem clamp(1rem, 2vw, 2rem) 0;
}

/* 玻璃材质由 styles/glass.scss 中的 .app-top-nav-shell 提供 */
.app-top-nav-shell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: var(--app-shell-height);
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-xl);
  box-shadow:
    var(--glass-shadow),
    0 1px 0 var(--glass-highlight) inset;
}

.app-top-nav-brand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--radius-lg);
}

.app-top-nav-menu-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--radius-lg);
  border: 1px solid transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background-color 150ms ease, border-color 150ms ease;

  &:hover {
    color: var(--color-text-primary);
    border-color: var(--glass-border);
    background-color: var(--glass-bg-subtle-hover);
    -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
    backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  }
}

@media (min-width: 1024px) {
  .app-top-nav-menu-toggle {
    display: none;
  }
}

.app-top-nav-scroll {
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.app-top-nav-scroll::-webkit-scrollbar {
  display: none;
}

.app-top-nav-list {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  width: max-content;
}

.app-top-nav-user {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  margin-left: auto;
}

.app-top-nav-item {
  position: relative;
  isolation: isolate;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  white-space: nowrap;
  transition: color 180ms ease, border-color 180ms ease, box-shadow 180ms ease;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: -1;
    border-radius: inherit;
    opacity: 0;
    background-color: var(--glass-bg-subtle-hover);
    -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
    backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
    transition: opacity 180ms ease, background-color 180ms ease, backdrop-filter 180ms ease;
  }

  &:hover {
    color: var(--color-text-primary);
    border-color: var(--glass-border);
    box-shadow: 0 2px 8px rgba(15, 23, 42, 0.04), 0 1px 0 var(--glass-highlight) inset;
  }

  &:hover::before,
  &:focus-visible::before {
    opacity: 1;
  }

  &:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.25), 0 1px 0 var(--glass-highlight) inset;
  }
}

.app-top-nav-item-active {
  border-color: var(--glass-border-active);
  background-color: var(--glass-tint-brand);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  color: var(--color-text-brand);
  box-shadow:
    0 4px 16px rgba(10, 132, 255, 0.12),
    0 1px 0 var(--glass-highlight-hover) inset;

  &::before,
  &:hover::before {
    opacity: 0;
  }

  &::after {
    position: absolute;
    right: 0.75rem;
    bottom: 0.25rem;
    left: 0.75rem;
    height: 2px;
    border-radius: 9999px;
    content: '';
    background: linear-gradient(90deg, #0a84ff, #5e5ce6);
    box-shadow: 0 0 8px rgba(10, 132, 255, 0.5);
  }

  .app-top-nav-icon {
    color: var(--color-primary);
  }

  &:hover {
    color: var(--color-text-brand);
    border-color: var(--glass-border-hover);
  }

  .dark & {
    border-color: var(--glass-border-active);
    background-color: var(--glass-tint-brand);
    color: var(--color-text-primary);
    box-shadow:
      0 6px 20px rgba(0, 0, 0, 0.35),
      0 1px 0 var(--glass-highlight-hover) inset;

    &:hover {
      color: var(--color-text-primary);
    }
  }
}

.app-top-nav-label {
  line-height: var(--line-height-tight);
}

.app-top-nav-icon {
  flex-shrink: 0;
  color: var(--color-text-tertiary);
  transition: color 160ms ease;
}

.app-top-nav-custom-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 1rem;
  height: 1rem;
}

.app-top-nav-custom-icon :deep(svg) {
  display: block;
  width: 1rem;
  height: 1rem;
  max-width: 100%;
  max-height: 100%;
}
</style>
