<template>
  <AppLayout>
    <main class="feature-unavailable">
      <div class="feature-unavailable__icon"><Icon name="lock" size="lg" /></div>
      <h1>{{ t('featureUnavailable.title') }}</h1>
      <p>{{ t('featureUnavailable.description') }}</p>
      <RouterLink class="btn btn-primary" :to="homePath">
        <Icon name="home" size="sm" />
        {{ t('featureUnavailable.back') }}
      </RouterLink>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore } from '@/stores'

const { t } = useI18n()
const authStore = useAuthStore()
const homePath = computed(() => authStore.isAuthenticated ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/home')
</script>

<style scoped>
.feature-unavailable { min-height: min(70vh, 620px); display: grid; place-content: center; justify-items: center; padding: 2rem; text-align: center; }
.feature-unavailable__icon { display: grid; place-items: center; width: 3.5rem; height: 3.5rem; border-radius: 8px; background: var(--color-surface-muted); color: var(--color-text-secondary); }
.feature-unavailable h1 { margin: 1rem 0 .5rem; font-size: var(--font-size-2xl); }
.feature-unavailable p { max-width: 32rem; margin: 0 0 1.25rem; color: var(--color-text-secondary); }
.feature-unavailable .btn { display: inline-flex; align-items: center; gap: .45rem; }
</style>
