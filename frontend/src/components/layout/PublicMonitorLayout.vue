<template>
  <div class="public-monitor-layout">
    <header class="public-monitor-layout__header">
      <div class="public-monitor-layout__bar">
        <router-link to="/home" class="public-monitor-layout__brand">
          <img :src="siteLogo || '/logo.svg'" alt="" class="public-monitor-layout__logo" />
          <span>{{ siteName }}</span>
        </router-link>

        <div class="public-monitor-layout__actions">
          <LocaleSwitcher />
          <button
            type="button"
            class="public-monitor-layout__icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
          </button>
          <router-link to="/login" class="public-monitor-layout__login">
            {{ t('home.login') }}
          </router-link>
        </div>
      </div>
    </header>

    <main class="public-monitor-layout__main">
      <slot />
    </main>

    <footer class="public-monitor-layout__footer">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const appStore = useAppStore()
const isDark = ref(document.documentElement.classList.contains('dark'))

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'EasySub2api')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const currentYear = new Date().getFullYear()

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>

<style scoped>
.public-monitor-layout {
  display: grid;
  min-height: 100vh;
  grid-template-rows: auto 1fr auto;
  background: var(--color-page);
  color: var(--color-text-primary);
}

.public-monitor-layout__header {
  padding: 1rem 1.25rem 0;
}

.public-monitor-layout__bar {
  display: flex;
  width: min(1120px, 100%);
  min-height: 3.5rem;
  margin: 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  background: var(--glass-bg);
  box-shadow: var(--shadow-sm);
}

.public-monitor-layout__brand {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 0.65rem;
  color: var(--color-text-primary);
  font-weight: 700;
  text-decoration: none;
}

.public-monitor-layout__brand span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.public-monitor-layout__logo {
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  object-fit: contain;
}

.public-monitor-layout__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.5rem;
}

.public-monitor-layout__icon-button,
.public-monitor-layout__login {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--glass-bg-interactive);
  color: var(--color-text-secondary);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs)) saturate(var(--glass-saturate));
}

.public-monitor-layout__icon-button {
  width: 2.25rem;
  padding: 0;
  cursor: pointer;
}

.public-monitor-layout__login {
  padding: 0.45rem 0.85rem;
  border-color: var(--glass-border-active);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
  font-size: var(--font-size-sm);
  font-weight: 600;
  text-decoration: none;
}

.public-monitor-layout__icon-button:hover {
  border-color: var(--glass-border-hover);
  background: var(--glass-bg-interactive);
  color: var(--color-text-primary);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
}

.public-monitor-layout__login:hover {
  background: var(--glass-tint-brand);
  border-color: var(--glass-border-hover);
  box-shadow: var(--glass-shadow-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate-hover));
}

.public-monitor-layout__main {
  width: 100%;
}

.public-monitor-layout__footer {
  padding: 1.5rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  text-align: center;
}

@media (max-width: 600px) {
  .public-monitor-layout__header {
    padding: 0.75rem 0.75rem 0;
  }

  .public-monitor-layout__bar {
    gap: 0.5rem;
  }

  .public-monitor-layout__brand span {
    display: none;
  }
}
</style>
