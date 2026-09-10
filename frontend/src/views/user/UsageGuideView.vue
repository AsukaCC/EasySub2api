<template>
  <AppLayout>
    <main class="usage-guide-page">
      <section class="usage-guide-card card">
        <header class="usage-guide-header">
          <div>
            <h1>{{ t('usageGuide.title') }}</h1>
            <p>{{ t('usageGuide.description') }}</p>
          </div>
          <time v-if="updatedAt" :datetime="updatedAt">{{ formatUpdatedAt(updatedAt) }}</time>
        </header>

        <div v-if="loading" class="usage-guide-state">{{ t('common.loading') }}</div>
        <div v-else-if="error" class="usage-guide-state usage-guide-state--error">
          <p>{{ t('usageGuide.loadFailed') }}</p>
          <button type="button" class="btn btn-secondary" @click="load">{{ t('common.retry') }}</button>
        </div>
        <div v-else-if="!renderedHtml" class="usage-guide-state">{{ t('usageGuide.empty') }}</div>
        <div v-else class="usage-guide-content markdown-body" v-html="renderedHtml"></div>
      </section>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import AppLayout from '@/components/layout/AppLayout.vue'
import { usageGuideAPI } from '@/api/usageGuide'

const { t, locale } = useI18n()
const loading = ref(false)
const error = ref(false)
const renderedHtml = ref('')
const updatedAt = ref('')

function formatUpdatedAt(value: string): string {
  try {
    return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(value))
  } catch {
    return value
  }
}

async function load() {
  loading.value = true
  error.value = false
  try {
    const response = await usageGuideAPI.get()
    updatedAt.value = response.updated_at || ''
    const html = marked.parse(response.content_md || '') as string
    renderedHtml.value = DOMPurify.sanitize(html)
  } catch {
    error.value = true
    renderedHtml.value = ''
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.usage-guide-page { width: min(1080px, 100%); margin: 0 auto; padding: 2rem; }
.usage-guide-card { overflow: hidden; }
.usage-guide-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; padding: 1.5rem 1.75rem; border-bottom: 1px solid var(--glass-border); }
.usage-guide-header h1 { margin: 0; font-size: 1.5rem; }
.usage-guide-header p { margin: .45rem 0 0; color: var(--color-text-secondary); }
.usage-guide-header time { flex: 0 0 auto; color: var(--color-text-secondary); font-size: .85rem; }
.usage-guide-content { padding: 1.75rem; line-height: 1.75; }
.usage-guide-state { display: grid; place-items: center; gap: .75rem; min-height: 220px; padding: 2rem; color: var(--color-text-secondary); text-align: center; }
.usage-guide-state--error { color: var(--color-text-danger); }
@media (max-width: 700px) { .usage-guide-page { padding: 1rem; } .usage-guide-header { flex-direction: column; padding: 1.25rem; } .usage-guide-content { padding: 1.25rem; } }
</style>
