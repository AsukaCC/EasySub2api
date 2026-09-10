<template>
  <AppLayout>
    <main class="usage-guide-settings-page">
      <header class="usage-guide-settings-header">
        <div>
          <h1>{{ t('admin.settings.usageGuide.title') }}</h1>
          <p>{{ t('admin.settings.usageGuide.description') }}</p>
        </div>
        <RouterLink class="btn btn-secondary" to="/admin/settings/features">{{ t('admin.settings.usageGuide.manageFeature') }}</RouterLink>
      </header>

      <section class="usage-guide-editor card">
        <div class="usage-guide-editor__toolbar">
          <span>{{ t('admin.settings.usageGuide.contentLabel') }}</span>
          <div class="usage-guide-editor__toolbar-actions">
            <label class="usage-guide-editor__switch">
              <span>{{ t('admin.settings.usageGuide.enabledLabel') }}</span>
              <Toggle v-model="enabled" :disabled="loading || saving" />
            </label>
            <span :class="contentBytes > maxBytes ? 'is-invalid' : 'muted'">{{ contentBytes }} / {{ maxBytes }} bytes</span>
          </div>
        </div>
        <div class="usage-guide-editor__grid">
          <textarea v-model="content" class="input usage-guide-editor__textarea" :placeholder="t('admin.settings.usageGuide.placeholder')" />
          <div class="usage-guide-preview markdown-body" v-html="previewHtml"></div>
        </div>
        <div class="usage-guide-editor__footer">
          <span class="muted">{{ enabled ? t('admin.settings.usageGuide.enabledHint') : t('admin.settings.usageGuide.disabledHint') }}</span>
          <button type="button" class="btn btn-primary" :disabled="saving || contentBytes > maxBytes" @click="save">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </section>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const content = ref('')
const enabled = ref(false)
const loading = ref(false)
const saving = ref(false)
const maxBytes = 1 << 20
const contentBytes = computed(() => new TextEncoder().encode(content.value).length)
const previewHtml = computed(() => DOMPurify.sanitize(marked.parse(content.value || '') as string))

async function load() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    content.value = settings.usage_guide_content_md || ''
    enabled.value = settings.usage_guide_enabled === true
  } catch {
    appStore.showError(t('admin.settings.usageGuide.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function save() {
  if (saving.value || contentBytes.value > maxBytes) return
  saving.value = true
  try {
    const updated = await adminAPI.settings.updateSettings({
      usage_guide_enabled: enabled.value,
      usage_guide_content_md: content.value,
    })
    content.value = updated.usage_guide_content_md || ''
    enabled.value = updated.usage_guide_enabled === true
    appStore.showSuccess(t('admin.settings.usageGuide.saved'))
    await appStore.fetchPublicSettings(true)
  } catch {
    appStore.showError(t('admin.settings.usageGuide.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.usage-guide-settings-page { width: min(1220px, 100%); margin: 0 auto; padding: 2rem; }
.usage-guide-settings-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1.5rem; }
.usage-guide-settings-header h1 { margin: 0; }
.usage-guide-settings-header p { margin: .45rem 0 0; color: var(--color-text-secondary); }
.usage-guide-editor { padding: 1.25rem; }
.usage-guide-editor__toolbar, .usage-guide-editor__footer, .usage-guide-editor__toolbar-actions { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.usage-guide-editor__toolbar { margin-bottom: .75rem; font-weight: 600; }
.usage-guide-editor__toolbar-actions { justify-content: flex-end; }
.usage-guide-editor__switch { display: inline-flex; align-items: center; gap: .5rem; font-weight: 500; }
.usage-guide-editor__grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 1rem; }
.usage-guide-editor__textarea { min-height: 560px; resize: vertical; font: 0.9rem/1.6 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.usage-guide-preview { min-height: 560px; max-height: 720px; overflow: auto; padding: 1rem 1.25rem; border: 1px solid var(--glass-border); border-radius: .5rem; background: var(--color-surface-muted); }
.usage-guide-editor__footer { margin-top: 1rem; }
.muted { color: var(--color-text-secondary); font-size: .85rem; }
.is-invalid { color: var(--color-text-danger); font-size: .85rem; }
@media (max-width: 800px) { .usage-guide-settings-page { padding: 1rem; } .usage-guide-settings-header { flex-direction: column; } .usage-guide-editor__grid { grid-template-columns: 1fr; } .usage-guide-editor__textarea, .usage-guide-preview { min-height: 360px; } .usage-guide-editor__footer { align-items: flex-start; flex-direction: column; } }
</style>
