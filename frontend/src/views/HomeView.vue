<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="views-home-view__panel">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="views-home-view__iframe"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="views-home-view__panel-2"
  >
    <header class="floating-nav-wrap">
      <nav class="floating-nav-shell views-home-view__floating-nav">
        <div class="views-home-view__panel-3">
          <img
            :src="siteLogo || '/logo.svg'"
            alt="Logo"
            class="views-home-view__image"
          />
          <span class="views-home-view__text">{{ siteName }}</span>
        </div>
        <div class="views-home-view__panel-4">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-home-view__link"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="showModelPlaza"
            to="/model-plaza"
            class="views-home-view__link views-home-view__plaza-button"
          >
            <Icon name="sparkles" size="md" />
            <span>{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <button
            class="views-home-view__link"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="views-home-view__router-link"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="views-home-view__main">
      <div class="views-home-view__panel-5">
        <img
          :src="siteLogo || '/logo.svg'"
          alt="Logo"
          class="views-home-view__image-2"
        />
        <h1 class="views-home-view__heading">{{ siteName }}</h1>
        <p class="views-home-view__description">{{ siteSubtitle }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="views-home-view__router-link-2"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="views-home-view__footer">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="views-home-view__panel-6"
  >
    <!-- Background Decorations -->
    <div class="views-home-view__panel-7">
      <div
        class="views-home-view__panel-8"
      ></div>
      <div
        class="views-home-view__panel-9"
      ></div>
      <div
        class="views-home-view__panel-10"
      ></div>
      <div
        class="views-home-view__panel-11"
      ></div>
      <div
        class="views-home-view__panel-12"
      ></div>
    </div>

    <!-- Header -->
    <header class="floating-nav-wrap">
      <nav class="floating-nav-shell views-home-view__floating-nav">
        <!-- Logo -->
        <div class="views-home-view__panel-13">
          <div class="views-home-view__panel-14">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="views-home-view__image-3" />
          </div>
        </div>

        <!-- Nav Actions -->
        <div class="views-home-view__panel-15">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-home-view__link-2"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <router-link
            v-if="showModelPlaza"
            to="/model-plaza"
            class="views-home-view__link-2 views-home-view__plaza-button"
          >
            <Icon name="sparkles" size="md" />
            <span>{{ t('nav.modelPlaza') }}</span>
          </router-link>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="views-home-view__link-2"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="views-home-view__router-link-3"
          >
            <span
              class="views-home-view__text-2"
            >
              {{ userInitial }}
            </span>
            <span class="views-home-view__text-3">{{ t('home.dashboard') }}</span>
            <svg
              class="views-home-view__icon"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="views-home-view__router-link-4"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="views-home-view__main-2">
      <div class="views-home-view__panel-16">
        <!-- Hero Section - Left/Right Layout -->
        <div class="views-home-view__panel-17">
          <!-- Left: Text Content -->
          <div class="views-home-view__panel-18">
            <h1
              class="views-home-view__heading-2"
            >
              {{ siteName }}
            </h1>
            <p class="views-home-view__description-2">
              {{ siteSubtitle }}
            </p>

            <!-- CTA Button -->
            <div>
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="views-home-view__router-link-5 btn btn-primary"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" class="views-home-view__icon-2" :stroke-width="2" />
              </router-link>
            </div>
          </div>

          <!-- Right: Terminal Animation -->
          <div class="views-home-view__panel-19">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">terminal</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="code-line line-1">
                    <span class="code-prompt">$</span>
                    <span class="code-cmd">curl</span>
                    <span class="code-flag">-X POST</span>
                    <span class="code-url">/v1/messages</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-comment"># Routing to upstream...</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">200 OK</span>
                    <span class="code-response">{ "content": "Hello!" }</span>
                  </div>
                  <div class="code-line line-4">
                    <span class="code-prompt">$</span>
                    <span class="views-home-view__text-4"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Feature Tags - Centered -->
        <div class="views-home-view__panel-20">
          <div
            class="views-home-view__panel-21"
          >
            <Icon name="swap" size="sm" class="views-home-view__icon-3" />
            <span class="views-home-view__text-5">{{
              t('home.tags.subscriptionToApi')
            }}</span>
          </div>
          <div
            class="views-home-view__panel-21"
          >
            <Icon name="shield" size="sm" class="views-home-view__icon-3" />
            <span class="views-home-view__text-5">{{
              t('home.tags.stickySession')
            }}</span>
          </div>
          <div
            class="views-home-view__panel-21"
          >
            <Icon name="chart" size="sm" class="views-home-view__icon-3" />
            <span class="views-home-view__text-5">{{
              t('home.tags.realtimeBilling')
            }}</span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="views-home-view__panel-22">
          <!-- Feature 1: Unified Gateway -->
          <div
            class="views-home-view__panel-23"
          >
            <div
              class="views-home-view__panel-24"
            >
              <Icon name="server" size="lg" class="views-home-view__icon-4" />
            </div>
            <h3 class="views-home-view__heading-3">
              {{ t('home.features.unifiedGateway') }}
            </h3>
            <p class="views-home-view__description-3">
              {{ t('home.features.unifiedGatewayDesc') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div
            class="views-home-view__panel-23"
          >
            <div
              class="views-home-view__panel-25"
            >
              <svg
                class="views-home-view__icon-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="views-home-view__heading-3">
              {{ t('home.features.multiAccount') }}
            </h3>
            <p class="views-home-view__description-3">
              {{ t('home.features.multiAccountDesc') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div
            class="views-home-view__panel-23"
          >
            <div
              class="views-home-view__panel-26"
            >
              <svg
                class="views-home-view__icon-5"
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
            </div>
            <h3 class="views-home-view__heading-3">
              {{ t('home.features.balanceQuota') }}
            </h3>
            <p class="views-home-view__description-3">
              {{ t('home.features.balanceQuotaDesc') }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="views-home-view__panel-27">
          <h2 class="views-home-view__heading-4">
            {{ t('home.providers.title') }}
          </h2>
          <p class="views-home-view__description-4">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="views-home-view__panel-28">
          <!-- Claude - Supported -->
          <div
            class="views-home-view__panel-29"
          >
            <div
              class="views-home-view__panel-30"
            >
              <span class="views-home-view__text-6">C</span>
            </div>
            <span class="views-home-view__text-5">{{ t('home.providers.claude') }}</span>
            <span
              class="views-home-view__text-7"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- GPT - Supported -->
          <div
            class="views-home-view__panel-29"
          >
            <div
              class="views-home-view__panel-31"
            >
              <span class="views-home-view__text-6">G</span>
            </div>
            <span class="views-home-view__text-5">GPT</span>
            <span
              class="views-home-view__text-7"
              >{{ t('home.providers.supported') }}</span
            >
          </div>
          <!-- More - Coming Soon -->
          <div
            class="views-home-view__panel-34"
          >
            <div
              class="views-home-view__panel-35"
            >
              <span class="views-home-view__text-6">+</span>
            </div>
            <span class="views-home-view__text-5">{{ t('home.providers.more') }}</span>
            <span
              class="views-home-view__text-8"
              >{{ t('home.providers.soon') }}</span
            >
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="views-home-view__footer-2">
      <div
        class="views-home-view__panel-36"
      >
        <p class="views-home-view__description-5">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="views-home-view__panel-37">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-home-view__link-3"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="views-home-view__link-3"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'EasySub2api')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
// 模型广场为 opt-in 功能:仅在设置明确开启时展示入口
const showModelPlaza = computed(() => appStore.cachedPublicSettings?.model_plaza_enabled === true)

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/AsukaCC/EasySub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* 首页悬浮导航：保持单行、居中和稳定高度，材质由全局 floating-nav-shell 提供。 */
.views-home-view__floating-nav {
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: min(calc(100% - 2rem), 80rem);
  min-height: 3.5rem;
  margin: 0.75rem auto 0;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  flex-wrap: nowrap;
}

.views-home-view__panel-3,
.views-home-view__panel-13 {
  width: auto;
  min-width: 0;
  flex: 0 0 auto;
}

.views-home-view__panel-4,
.views-home-view__panel-15 {
  width: auto;
  min-width: 0;
  margin-left: auto;
  flex: 0 1 auto;
}

@media (max-width: 639px) {
  .views-home-view__floating-nav {
    gap: 0.5rem;
    width: calc(100% - 1rem);
    min-height: 3rem;
    margin-top: 0.5rem;
    padding: 0.375rem 0.5rem;
  }

  .views-home-view__panel-3,
  .views-home-view__panel-13 {
    display: none;
  }

  .views-home-view__panel-4,
  .views-home-view__panel-15 {
    justify-content: space-between;
    gap: 0.5rem;
    width: 100%;
    margin-left: 0;
    flex: 1 1 auto;
  }
}

@media (max-width: 359px) {
  .views-home-view__plaza-button span {
    display: none;
  }

  .views-home-view__plaza-button {
    justify-content: center;
    min-width: 2.25rem;
    padding-inline: 0.5rem;
  }
}

/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
}

.views-home-view__plaza-button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  min-height: 2.25rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 6px;
  color: inherit;
  font-size: var(--font-size-sm);
  font-weight: 600;
  line-height: 1.2;
  text-decoration: none;
  white-space: nowrap;
}

.views-home-view__plaza-button:hover {
  border-color: rgba(96, 165, 250, 0.75);
  color: #60a5fa;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  background: linear-gradient(145deg, #17171a 0%, #0c0c0e 100%);
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(23, 23, 26, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: var(--font-size-xs);
  font-family: ui-monospace, monospace;
  color: #64748b;
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: var(--font-size-sm);
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #22c55e;
  font-weight: bold;
}
.code-cmd {
  color: #38bdf8;
}
.code-flag {
  color: #a78bfa;
}
.code-url {
  color: #0a84ff;
}
.code-comment {
  color: #64748b;
  font-style: italic;
}
.code-success {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.code-response {
  color: #fbbf24;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #22c55e;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}

/* Dark mode adjustments */
:deep(.dark) .terminal-window {
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(10, 132, 255, 0.2),
    0 0 40px rgba(10, 132, 255, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}
</style>
