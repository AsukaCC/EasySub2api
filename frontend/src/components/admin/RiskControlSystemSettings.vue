<template>
  <section class="card risk-system-settings">
    <div class="risk-system-settings__heading">
      <div>
        <h2>{{ t('admin.settings.features.riskControl.cyberSessionBlock') }}</h2>
        <p>{{ t('admin.settings.features.riskControl.cyberSessionBlockHint') }}</p>
      </div>
      <Toggle v-model="form.enabled" :disabled="loading || saving" />
    </div>
    <label v-if="form.enabled" class="risk-system-settings__field">
      <span>{{ t('admin.settings.features.riskControl.cyberSessionBlockTTL') }}</span>
      <input v-model.number="form.ttl" class="input" type="number" min="1" max="604800" />
    </label>
    <div class="risk-system-settings__actions">
      <button type="button" class="btn btn-primary" :disabled="loading || saving" @click="save">
        {{ saving ? t('common.saving') : t('common.save') }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const form = reactive({ enabled: false, ttl: 3600 })

async function load() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    form.enabled = Boolean(settings.cyber_session_block_enabled)
    form.ttl = settings.cyber_session_block_ttl_seconds || 3600
  } catch {
    appStore.showError(t('admin.settings.featureManagement.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function save() {
  const ttl = Math.floor(Number(form.ttl))
  if (form.enabled && (!Number.isInteger(ttl) || ttl < 1 || ttl > 604800)) {
    appStore.showError(t('admin.settings.featureManagement.riskSettings.ttlError'))
    return
  }
  saving.value = true
  try {
    await adminAPI.settings.updateSettings({
      cyber_session_block_enabled: form.enabled,
      cyber_session_block_ttl_seconds: ttl || 3600,
    })
    appStore.showSuccess(t('common.saved'))
  } catch {
    appStore.showError(t('admin.settings.featureManagement.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.risk-system-settings { margin-bottom: 1.5rem; padding: 1.25rem; }
.risk-system-settings__heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.risk-system-settings__heading h2 { margin: 0; font-size: var(--font-size-base); }
.risk-system-settings__heading p { margin: .4rem 0 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); line-height: 1.5; }
.risk-system-settings__field { display: grid; gap: .5rem; max-width: 320px; margin-top: 1.25rem; font-size: var(--font-size-sm); font-weight: 600; }
.risk-system-settings__actions { display: flex; justify-content: flex-end; margin-top: 1rem; }
</style>
