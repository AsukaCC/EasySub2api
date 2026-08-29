<template>
  <AppLayout>
    <main class="channel-monitor-feature-settings">
      <section class="card channel-monitor-feature-settings__privacy">
        <div class="channel-monitor-feature-settings__copy">
          <h2>{{ t('admin.settings.features.channelMonitor.hideThroughput') }}</h2>
          <p>{{ t('admin.settings.features.channelMonitor.hideThroughputHint') }}</p>
        </div>
        <div class="channel-monitor-feature-settings__actions">
          <RouterLink class="btn btn-secondary" to="/admin/channels/monitor">
            <Icon name="chart" size="sm" />
            {{ t('admin.settings.features.channelMonitor.openDashboard') }}
          </RouterLink>
          <Toggle
            :model-value="hideThroughput"
            :disabled="loading || saving"
            @update:model-value="updateHideThroughput"
          />
        </div>
      </section>

      <MonitorSettingsPanel />
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import MonitorSettingsPanel from '@/features/channel-monitor-v2/MonitorSettingsPanel.vue'
import { adminAPI } from '@/api'
import { useAdminSettingsStore, useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const hideThroughput = ref(false)
const loading = ref(true)
const saving = ref(false)

async function load() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    hideThroughput.value = Boolean(settings.channel_monitor_hide_throughput)
  } catch {
    appStore.showError(t('admin.settings.featureManagement.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function updateHideThroughput(next: boolean) {
  if (saving.value) return
  const previous = hideThroughput.value
  hideThroughput.value = next
  saving.value = true
  try {
    const settings = await adminAPI.settings.updateSettings({
      channel_monitor_hide_throughput: next,
    })
    hideThroughput.value = Boolean(settings.channel_monitor_hide_throughput)
    await Promise.all([appStore.fetchPublicSettings(true), adminSettingsStore.fetch(true)])
    appStore.showSuccess(t('common.saved'))
  } catch {
    hideThroughput.value = previous
    appStore.showError(t('admin.settings.featureManagement.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.channel-monitor-feature-settings {
  width: min(1100px, 100%);
  margin: 0 auto;
  padding: 2rem;
}

.channel-monitor-feature-settings__privacy {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.25rem;
  padding: 1.15rem 1.25rem;
  margin-bottom: 1.5rem;
}

.channel-monitor-feature-settings__copy h2 {
  margin: 0;
  font-size: var(--font-size-base);
}

.channel-monitor-feature-settings__copy p {
  max-width: 48rem;
  margin: .4rem 0 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.5;
}

.channel-monitor-feature-settings__actions {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  gap: 1rem;
}

.channel-monitor-feature-settings__actions .btn {
  display: inline-flex;
  align-items: center;
  gap: .45rem;
}

@media (max-width: 700px) {
  .channel-monitor-feature-settings {
    padding: 1rem;
  }

  .channel-monitor-feature-settings__privacy {
    align-items: flex-start;
    flex-direction: column;
  }

  .channel-monitor-feature-settings__actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
