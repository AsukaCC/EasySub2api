<template>
  <header class="floating-nav-wrap">
    <div class="floating-nav-shell">
      <!-- 左:站点 logo + 名称 -->
      <div class="components-model-plaza-plaza-nav-bar__panel-2">
        <template v-if="settings">
          <span
            class="components-model-plaza-plaza-nav-bar__text"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="components-model-plaza-plaza-nav-bar__image" />
          </span>
          <span class="components-model-plaza-plaza-nav-bar__text-2">
            {{ siteName }}
          </span>
        </template>
        <template v-else>
          <span class="components-model-plaza-plaza-nav-bar__text-3" aria-hidden="true"></span>
          <span class="components-model-plaza-plaza-nav-bar__text-4" aria-hidden="true"></span>
        </template>
      </div>

      <!-- 右:登录 / 回到后台 -->
      <RouterLink
        v-if="isAuthenticated"
        :to="backTarget"
        class="components-model-plaza-plaza-nav-bar__router-link"
      >
        {{ t('modelPlaza.nav.backToDashboard') }}
      </RouterLink>
      <RouterLink
        v-else
        :to="{ path: '/login', query: { redirect: '/model-plaza' } }"
        class="components-model-plaza-plaza-nav-bar__router-link-2"
      >
        {{ t('modelPlaza.nav.login') }}
      </RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => settings.value?.site_name || 'EasySub2api')
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || '', { allowRelative: true, allowDataUrl: true })
)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const backTarget = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
</script>
