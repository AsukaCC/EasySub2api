<template>
  <div
    v-if="user"
    ref="dropdownRef"
    class="components-layout-app-nav-user-module__panel"
    @keydown.esc.prevent.stop="closeDropdown"
  >
    <button
      ref="triggerRef"
      type="button"
      class="app-nav-user-trigger"
      :class="dropdownOpen && 'app-nav-user-trigger-open'"
      :aria-label="t('common.userMenu')"
      :aria-expanded="dropdownOpen"
      aria-controls="app-nav-user-menu"
      aria-haspopup="menu"
      @click="toggleDropdown"
      @keydown.escape="closeDropdown"
    >
      <span class="app-nav-user-avatar">
        <img
          v-if="avatarUrl"
          :src="avatarUrl"
          :alt="displayName"
          class="components-layout-app-nav-user-module__image"
        >
        <span v-else>{{ userInitials }}</span>
      </span>

      <span class="components-layout-app-nav-user-module__text">
        <span class="components-layout-app-nav-user-module__text-2">
          {{ displayName }}
        </span>
        <span class="components-layout-app-nav-user-module__text-3">
          {{ formatPoints(availableBalance) }}
        </span>
      </span>

      <Icon
        name="chevronDown"
        size="xs"
        class="components-layout-app-nav-user-module__icon"
        :class="dropdownOpen && 'components-layout-app-nav-user-module__icon-3'"
      />
    </button>

    <Teleport to="body">
      <transition name="nav-user-dropdown">
        <div
          v-if="dropdownOpen"
          ref="panelRef"
          id="app-nav-user-menu"
          class="components-layout-app-nav-user-module__panel-2 dropdown app-nav-user-dropdown"
          :style="panelStyle"
          role="menu"
          @click.stop
        >
        <div class="components-layout-app-nav-user-module__panel-3">
          <span class="components-layout-app-nav-user-module__text-4 app-nav-user-avatar">
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="components-layout-app-nav-user-module__image"
            >
            <span v-else>{{ userInitials }}</span>
          </span>
          <span class="components-layout-app-nav-user-module__text-5">
            <span class="components-layout-app-nav-user-module__text-6">
              {{ displayName }}
            </span>
            <span class="components-layout-app-nav-user-module__text-7">
              {{ user.email }}
            </span>
          </span>
          <span class="components-layout-app-nav-user-module__text-8">
            {{ t('admin.users.roles.' + user.role) }}
          </span>
        </div>

        <div class="components-layout-app-nav-user-module__panel-4 app-nav-user-balance-section">
          <div class="components-layout-app-nav-user-module__panel-5">
            <span class="components-layout-app-nav-user-module__text-9">{{ balanceAvailableText }}</span>
            <span class="components-layout-app-nav-user-module__text-10">
              {{ formatPoints(availableBalance) }}
            </span>
          </div>
          <SubscriptionProgressMini class="app-nav-user-balance-subscription" />
        </div>

        <div class="components-layout-app-nav-user-module__panel-8">
          <AnnouncementBell />
          <button
            type="button"
            class="app-nav-user-tool"
            :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
            :aria-label="isDark ? t('nav.lightMode') : t('nav.darkMode')"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
          <button
            type="button"
            class="app-nav-user-tool"
            :title="t('nav.customBackground')"
            :aria-label="t('nav.customBackground')"
            @click="triggerBackgroundUpload"
          >
            <Icon name="photo" size="sm" />
          </button>
          <button
            v-if="customBgActive"
            type="button"
            class="app-nav-user-tool"
            :title="t('nav.resetBackground')"
            :aria-label="t('nav.resetBackground')"
            @click="handleResetBackground"
          >
            <Icon name="refresh" size="sm" />
          </button>
          <input
            ref="backgroundFileInput"
            type="file"
            accept="image/*"
            class="components-layout-app-nav-user-module__field"
            @change="handleBackgroundFileChange"
          />
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="app-nav-user-tool"
            :title="t('nav.docs')"
            :aria-label="t('nav.docs')"
          >
            <Icon name="book" size="sm" />
          </a>
        </div>

        <router-link to="/profile" class="dropdown-item" role="menuitem" @click="closeDropdown">
          <Icon name="user" size="sm" />
          {{ t('nav.profile') }}
        </router-link>

        <a
          v-if="authStore.isAdmin"
          href="https://github.com/AsukaCC/EasySub2api"
          target="_blank"
          rel="noopener noreferrer"
          class="dropdown-item"
          role="menuitem"
          @click="closeDropdown"
        >
          <Icon name="externalLink" size="sm" />
          {{ t('nav.github') }}
        </a>

        <div
          v-if="contactInfo"
          class="components-layout-app-nav-user-module__panel-9"
        >
          <div class="components-layout-app-nav-user-module__panel-10">
            <Icon name="chat" size="xs" class="components-layout-app-nav-user-module__icon-2" />
            <span class="components-layout-app-nav-user-module__text-13">
              {{ t('common.contactSupport') }}: {{ contactInfo }}
            </span>
          </div>
        </div>

        <button
          v-if="showOnboardingButton"
          type="button"
          class="components-layout-app-nav-user-module__action dropdown-item"
          role="menuitem"
          @click="handleReplayGuide"
        >
          <Icon name="questionCircle" size="sm" />
          {{ t('onboarding.restartTour') }}
        </button>

        <button
          type="button"
          class="components-layout-app-nav-user-module__action-2 dropdown-item"
          role="menuitem"
          @click="handleLogout"
        >
          <Icon name="login" size="sm" class="components-layout-app-nav-user-module__icon-3" />
          {{ t('nav.logout') }}
        </button>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { formatPoints } from '@/utils/format'
import {
  clearCustomBackground,
  hasCustomBackground,
  setCustomBackground
} from '@/utils/customBackground'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, dropdownOpen, {
  maxWidth: 288,
  maxHeightRatio: 0.85,
  minComfortableHeight: 260
})
const isDark = ref(document.documentElement.classList.contains('dark'))
const contactInfo = computed(() => appStore.contactInfo)
const docUrl = computed(() => sanitizeUrl(appStore.docUrl))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const availableBalance = computed(() => Number(user.value?.available_balance ?? user.value?.balance ?? 0))
const balanceAvailableText = computed(() =>
  t('common.availableBalance') === 'common.availableBalance'
    ? '可用余额'
    : t('common.availableBalance')
)
const showOnboardingButton = computed(
  () => !authStore.isSimpleMode && user.value?.role === 'admin'
)

const userInitials = computed(() => {
  if (!user.value) return ''
  const source = user.value.username || user.value.email?.split('@')[0] || ''
  return source.substring(0, 2).toUpperCase()
})

const displayName = computed(
  () => user.value?.username || user.value?.email?.split('@')[0] || ''
)
function toggleTheme(): void {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

const backgroundFileInput = ref<HTMLInputElement | null>(null)
const customBgActive = ref(hasCustomBackground())

function triggerBackgroundUpload(): void {
  backgroundFileInput.value?.click()
}

async function handleBackgroundFileChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    await setCustomBackground(file)
    customBgActive.value = true
  } catch {
    appStore.showToast('error', t('nav.backgroundSaveFailed'))
  }
}

function handleResetBackground(): void {
  clearCustomBackground()
  customBgActive.value = false
}

function toggleDropdown(): void {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown(): void {
  dropdownOpen.value = false
}

async function handleLogout(): Promise<void> {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide(): void {
  closeDropdown()
  onboardingStore.replay()
}

function handleClickOutside(event: MouseEvent): void {
  const target = event.target as Node
  if (dropdownRef.value?.contains(target) || panelRef.value?.contains(target)) return
  closeDropdown()
}

function handleDocumentKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') closeDropdown()
}

watch(() => route.fullPath, closeDropdown)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleDocumentKeydown)
})

</script>

<style scoped>
.app-nav-user-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0.625rem 0.25rem 0.25rem;
  border: 1px solid var(--glass-border);
  border-radius: 999px;
  background-color: var(--glass-layer-inset-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background-color 160ms ease, border-color 160ms ease;
}

.app-nav-user-trigger:hover {
  background-color: var(--glass-layer-inset-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.app-nav-user-trigger:focus-visible {
  border-color: rgb(10 132 255 / 0.5);
  outline: none;
  box-shadow: 0 0 0 3px rgb(10 132 255 / 0.18);
}

.app-nav-user-trigger-open {
  border-color: rgb(10 132 255 / 0.35);
  background-color: rgb(10 132 255 / 0.1);
}

.app-nav-user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  overflow: hidden;
  border-radius: 999px;
  background: linear-gradient(180deg, #0a84ff, #007aff);
  color: #fff;
  font-size: var(--font-size-xs);
  font-weight: 600;
}

.app-nav-user-dropdown {
  max-width: calc(100vw - 1.5rem);
}

.app-nav-user-balance-section {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
}

.app-nav-user-balance-subscription {
  align-self: stretch;
}

.app-nav-user-balance-subscription :deep(.components-common-subscription-progress-mini__action) {
  min-width: 2.75rem;
  height: 100%;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.app-nav-user-tool {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: color 150ms ease, border-color 150ms ease, background-color 150ms ease;
}

.app-nav-user-tool:hover {
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  color: var(--color-text-primary);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.nav-user-dropdown-enter-active,
.nav-user-dropdown-leave-active {
  transition:
    opacity 150ms ease,
    transform 150ms ease;
}

.nav-user-dropdown-enter-from,
.nav-user-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
</style>
