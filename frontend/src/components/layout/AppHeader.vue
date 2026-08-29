<template>
  <header class="components-layout-app-header__header glass">
    <div class="components-layout-app-header__panel">
      <!-- Left: Brand + Page Title -->
      <div class="components-layout-app-header__panel-2">
        <router-link
          :to="homePath"
          class="components-layout-app-header__router-link"
          :aria-label="siteName"
        >
          <span class="components-layout-app-header__text">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="components-layout-app-header__image" />
          </span>
          <span class="components-layout-app-header__text-2">
            {{ siteName }}
          </span>
        </router-link>

        <div class="components-layout-app-header__panel-3">
          <h1 class="components-layout-app-header__heading">
            {{ pageTitle }}
          </h1>
          <p v-if="pageDescription" class="components-layout-app-header__description">
            {{ pageDescription }}
          </p>
        </div>
      </div>

      <!-- Right: Announcements + Docs + Language + Subscriptions + Balance + User Dropdown -->
      <div class="components-layout-app-header__panel-4">
        <!-- Announcement Bell -->
        <AnnouncementBell v-if="user" />

        <!-- Docs Link -->
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="components-layout-app-header__link"
        >
          <Icon name="book" size="sm" />
          <span class="components-layout-app-header__text-3">{{ t('nav.docs') }}</span>
        </a>

        <!-- Theme Toggle -->
        <button
          type="button"
          class="btn-ghost btn-icon"
          :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
          :aria-label="isDark ? t('nav.lightMode') : t('nav.darkMode')"
          @click="toggleTheme"
        >
          <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
        </button>

        <!-- Language Switcher -->
        <LocaleSwitcher />

        <!-- Subscription Progress (for users with active subscriptions) -->
        <SubscriptionProgressMini v-if="user" />

        <!-- Balance Display -->
        <div
          v-if="user"
          class="components-layout-app-header__panel-5"
        >
          <svg
            class="components-layout-app-header__icon"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
            />
          </svg>
          <span class="components-layout-app-header__text-4">
            {{ formatPoints(availableBalance) }}
          </span>
        </div>

        <!-- User Dropdown -->
        <div v-if="user" class="components-layout-app-header__panel-10" ref="dropdownRef">
          <button
            ref="triggerRef"
            @click="toggleDropdown"
            class="components-layout-app-header__action"
            :aria-label="t('common.userMenu')"
          >
            <div class="components-layout-app-header__panel-11">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                class="components-layout-app-header__image-2"
              >
              <span v-else>{{ userInitials }}</span>
            </div>
            <div class="components-layout-app-header__panel-12">
              <div class="components-layout-app-header__panel-13">
                {{ displayName }}
              </div>
              <div class="components-layout-app-header__description">
                {{ t('admin.users.roles.' + user.role) }}
              </div>
            </div>
            <Icon name="chevronDown" size="sm" class="components-layout-app-header__icon-2" />
          </button>

          <!-- Dropdown Menu -->
          <Teleport to="body">
            <transition name="dropdown">
              <div
                v-if="dropdownOpen"
                ref="panelRef"
                class="components-layout-app-header__panel-14 dropdown"
                :style="panelStyle"
                @click.stop
              >
              <!-- User Info -->
              <div class="components-layout-app-header__panel-15">
                <div class="components-layout-app-header__panel-13">
                  {{ displayName }}
                </div>
                <div class="components-layout-app-header__description">{{ user.email }}</div>
              </div>

              <!-- Balance (mobile only) -->
              <div class="components-layout-app-header__panel-16">
                <div class="components-layout-app-header__description">
                  {{ t('common.availableBalance') }}
                </div>
                <div class="components-layout-app-header__panel-17">
                  {{ formatPoints(availableBalance) }}
                </div>
              </div>

              <div class="components-layout-app-header__panel-19">
                <router-link to="/profile" @click="closeDropdown" class="dropdown-item">
                  <Icon name="user" size="sm" />
                  {{ t('nav.profile') }}
                </router-link>

                <router-link to="/keys" @click="closeDropdown" class="dropdown-item">
                  <Icon name="key" size="sm" />
                  {{ t('nav.apiKeys') }}
                </router-link>

                <a
                  v-if="authStore.isAdmin"
                  href="https://github.com/AsukaCC/EasySub2api"
                  target="_blank"
                  rel="noopener noreferrer"
                  @click="closeDropdown"
                  class="dropdown-item"
                >
                  <svg class="components-layout-app-header__icon-3" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      fill-rule="evenodd"
                      clip-rule="evenodd"
                      d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                    />
                  </svg>
                  {{ t('nav.github') }}
                </a>

              </div>

              <!-- Contact Support (only show if configured) -->
              <div
                v-if="contactInfo"
                class="components-layout-app-header__panel-20"
              >
                <div class="components-layout-app-header__panel-21">
                  <svg
                    class="components-layout-app-header__icon-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155"
                    />
                  </svg>
                  <span>{{ t('common.contactSupport') }}:</span>
                  <span class="components-layout-app-header__text-10">{{
                    contactInfo
                  }}</span>
                </div>
              </div>

              <div v-if="showOnboardingButton" class="components-layout-app-header__panel-22">
                <button @click="handleReplayGuide" class="components-layout-app-header__action-2 dropdown-item">
                  <svg class="components-layout-app-header__icon-3" fill="currentColor" viewBox="0 0 24 24">
                    <path
                      d="M12 2a10 10 0 100 20 10 10 0 000-20zm0 14a1 1 0 110 2 1 1 0 010-2zm1.07-7.75c0-.6-.49-1.25-1.32-1.25-.7 0-1.22.4-1.43 1.02a1 1 0 11-1.9-.62A3.41 3.41 0 0111.8 5c2.02 0 3.25 1.4 3.25 2.9 0 2-1.83 2.55-2.43 3.12-.43.4-.47.75-.47 1.23a1 1 0 01-2 0c0-1 .16-1.82 1.1-2.7.69-.64 1.82-1.05 1.82-2.06z"
                    />
                  </svg>
                  {{ $t('onboarding.restartTour') }}
                </button>
              </div>

              <div class="components-layout-app-header__panel-22">
                <button
                  @click="handleLogout"
                  class="components-layout-app-header__action-3 dropdown-item"
                >
                  <svg
                    class="components-layout-app-header__icon-3"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75"
                    />
                  </svg>
                  {{ t('nav.logout') }}
                </button>
              </div>
              </div>
            </transition>
          </Teleport>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { formatPoints } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const onboardingStore = useOnboardingStore()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, dropdownOpen, {
  maxWidth: 224,
  maxHeightRatio: 0.85,
  minComfortableHeight: 220
})
const contactInfo = computed(() => appStore.contactInfo)
const docUrl = computed(() => sanitizeUrl(appStore.docUrl))
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const homePath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const isDark = ref(document.documentElement.classList.contains('dark'))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const availableBalance = computed(() => Number(user.value?.available_balance ?? user.value?.balance ?? 0))

// 只在标准模式的管理员下显示新手引导按钮
const showOnboardingButton = computed(() => {
  return !authStore.isSimpleMode && user.value?.role === 'admin'
})

const userInitials = computed(() => {
  if (!user.value) return ''
  // Prefer username, fallback to email
  if (user.value.username) {
    return user.value.username.substring(0, 2).toUpperCase()
  }
  if (user.value.email) {
    // Get the part before @ and take first 2 chars
    const localPart = user.value.email.split('@')[0]
    return localPart.substring(0, 2).toUpperCase()
  }
  return ''
})

const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

const pageTitle = computed(() => {
  // For custom pages, use the menu item's label instead of generic "自定义页面"
  if (route.name === 'CustomPage') {
    const id = route.params.id as string
    const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
    const menuItem = publicItems.find((item) => item.id === id)
      ?? (authStore.isAdmin ? adminSettingsStore.customMenuItems.find((item) => item.id === id) : undefined)
    if (menuItem?.label) return menuItem.label
  }
  const titleKey = route.meta.titleKey as string
  if (titleKey) {
    return t(titleKey)
  }
  return (route.meta.title as string) || ''
})

const pageDescription = computed(() => {
  const descKey = route.meta.descriptionKey as string
  if (descKey) {
    return t(descKey)
  }
  return (route.meta.description as string) || ''
})

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    // Ignore logout errors - still redirect to login
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide() {
  closeDropdown()
  onboardingStore.replay()
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  if (dropdownRef.value?.contains(target) || panelRef.value?.contains(target)) return
  closeDropdown()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
