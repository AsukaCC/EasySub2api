<template>
  <AppLayout>
    <main class="ops-feature-settings">
      <header class="ops-feature-settings__header">
        <div>
          <p>{{ t('admin.settings.featureManagement.opsSettings.description') }}</p>
        </div>
        <RouterLink class="btn btn-secondary" to="/admin/ops">
          <Icon name="chart" size="sm" />
          {{ t('admin.settings.featureManagement.opsSettings.openDashboard') }}
        </RouterLink>
      </header>

      <div v-if="loading" class="ops-feature-settings__loading">{{ t('common.loading') }}</div>
      <template v-else>
        <section class="card ops-feature-settings__card">
          <div class="ops-feature-settings__row">
            <div>
              <h2>{{ t('admin.settings.featureManagement.opsSettings.realtime') }}</h2>
              <p>{{ t('admin.settings.featureManagement.opsSettings.realtimeHint') }}</p>
            </div>
            <Toggle v-model="form.ops_realtime_monitoring_enabled" />
          </div>

          <div class="ops-feature-settings__field-grid">
            <label>
              <span>{{ t('admin.settings.featureManagement.opsSettings.queryMode') }}</span>
              <Select v-model="form.ops_query_mode_default" :options="queryModeOptions" />
            </label>
            <label>
              <span>{{ t('admin.settings.featureManagement.opsSettings.metricsInterval') }}</span>
              <input v-model.number="form.ops_metrics_interval_seconds" class="input" type="number" min="1" max="3600" />
            </label>
          </div>

          <div class="ops-feature-settings__actions">
            <button type="button" class="btn btn-secondary" @click="showAdvanced = true">
              <Icon name="cog" size="sm" />
              {{ t('admin.settings.featureManagement.opsSettings.advanced') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
              {{ saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </section>

        <OpsRuntimeSettingsCard />
      </template>

      <OpsSettingsDialog :show="showAdvanced" @close="showAdvanced = false" @saved="showAdvanced = false" />
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import OpsRuntimeSettingsCard from '@/views/admin/ops/components/OpsRuntimeSettingsCard.vue'
import OpsSettingsDialog from '@/views/admin/ops/components/OpsSettingsDialog.vue'
import { adminAPI } from '@/api'
import { useAdminSettingsStore, useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const loading = ref(true)
const saving = ref(false)
const showAdvanced = ref(false)
const form = reactive({
  ops_realtime_monitoring_enabled: true,
  ops_query_mode_default: 'auto',
  ops_metrics_interval_seconds: 5,
})

const queryModeOptions = computed(() => [
  { value: 'auto', label: t('admin.ops.queryMode.auto') },
  { value: 'raw', label: t('admin.ops.queryMode.raw') },
  { value: 'preagg', label: t('admin.ops.queryMode.preagg') },
])

async function load() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    form.ops_realtime_monitoring_enabled = settings.ops_realtime_monitoring_enabled ?? true
    form.ops_query_mode_default = settings.ops_query_mode_default || 'auto'
    form.ops_metrics_interval_seconds = settings.ops_metrics_interval_seconds || 5
  } catch {
    appStore.showError(t('admin.settings.featureManagement.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function save() {
  const interval = Math.floor(Number(form.ops_metrics_interval_seconds))
  if (!Number.isInteger(interval) || interval < 1 || interval > 3600) {
    appStore.showError(t('admin.settings.featureManagement.opsSettings.intervalError'))
    return
  }
  saving.value = true
  try {
    await adminAPI.settings.updateSettings({
      ops_realtime_monitoring_enabled: form.ops_realtime_monitoring_enabled,
      ops_query_mode_default: form.ops_query_mode_default,
      ops_metrics_interval_seconds: interval,
    })
    await adminSettingsStore.fetch(true)
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
.ops-feature-settings { width: min(1100px, 100%); margin: 0 auto; padding: 2rem; }
.ops-feature-settings__header { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; margin-bottom: 1.5rem; }
.ops-feature-settings__header p { margin: 0; color: var(--color-text-secondary); }
.ops-feature-settings__header .btn, .ops-feature-settings__actions .btn { display: inline-flex; align-items: center; gap: .45rem; }
.ops-feature-settings__card { padding: 1.25rem; margin-bottom: 1.5rem; }
.ops-feature-settings__row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding-bottom: 1.25rem; border-bottom: 1px solid var(--glass-border); }
.ops-feature-settings__row h2 { margin: 0; font-size: var(--font-size-base); }
.ops-feature-settings__row p { margin: .35rem 0 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.ops-feature-settings__field-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; padding: 1.25rem 0; }
.ops-feature-settings__field-grid label { display: grid; gap: .5rem; font-size: var(--font-size-sm); font-weight: 600; }
.ops-feature-settings__actions { display: flex; justify-content: space-between; gap: .75rem; }
.ops-feature-settings__loading { min-height: 18rem; display: grid; place-items: center; color: var(--color-text-secondary); }
@media (max-width: 700px) { .ops-feature-settings { padding: 1rem; } .ops-feature-settings__header { flex-direction: column; } .ops-feature-settings__field-grid { grid-template-columns: 1fr; } }
</style>
