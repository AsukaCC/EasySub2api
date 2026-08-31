<template>
  <AppLayout>
    <main class="feature-management">
      <header class="feature-management__header">
        <div>
          <h1>{{ t('admin.settings.featureManagement.title') }}</h1>
          <p>{{ t('admin.settings.featureManagement.description') }}</p>
        </div>
        <button class="btn btn-secondary" type="button" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" />
          {{ t('common.refresh') }}
        </button>
      </header>

      <div v-if="loading && !loaded" class="feature-management__loading">
        <Icon name="refresh" size="lg" />
        <span>{{ t('common.loading') }}</span>
      </div>

      <section v-else class="feature-management__list" :aria-busy="loading">
        <article v-for="feature in features" :key="feature.id" class="feature-management__row">
          <div class="feature-management__identity">
            <span class="feature-management__icon"><Icon :name="feature.icon" size="md" /></span>
            <div>
              <div class="feature-management__title-line">
                <h2>{{ t(feature.titleKey) }}</h2>
                <span class="feature-management__status" :class="state[feature.enabledKey] ? 'is-on' : 'is-off'">
                  {{ state[feature.enabledKey] ? t('admin.settings.featureManagement.running') : t('admin.settings.featureManagement.stopped') }}
                </span>
                <span
                  v-if="feature.visibleKey"
                  class="feature-management__status"
                  :class="state[feature.visibleKey] ? 'is-published' : 'is-draft'"
                >
                  {{ state[feature.visibleKey] ? t('admin.settings.featureManagement.published') : t('admin.settings.featureManagement.hidden') }}
                </span>
              </div>
              <p>{{ t(feature.descriptionKey) }}</p>
            </div>
          </div>

          <div class="feature-management__controls">
            <label class="feature-management__switch">
              <span>{{ t('admin.settings.featureManagement.enableFeature') }}</span>
              <Toggle
                :model-value="state[feature.enabledKey]"
                :disabled="saving.has(feature.id)"
                @update:model-value="updateFeature(feature, 'enabled', $event)"
              />
            </label>
            <label v-if="feature.visibleKey" class="feature-management__switch">
              <span>{{ t('admin.settings.featureManagement.userVisible') }}</span>
              <Toggle
                :model-value="state[feature.visibleKey]"
                :disabled="saving.has(feature.id) || !state[feature.enabledKey]"
                @update:model-value="updateFeature(feature, 'visible', $event)"
              />
            </label>
            <label v-for="secondary in feature.secondaryKeys" :key="secondary.key" class="feature-management__switch">
              <span>{{ t(secondary.labelKey) }}</span>
              <Toggle
                :model-value="state[secondary.key]"
                :disabled="saving.has(feature.id) || !state[feature.enabledKey]"
                @update:model-value="updateFeature(feature, secondary.key, $event)"
              />
            </label>
            <RouterLink class="btn btn-secondary feature-management__configure" :to="feature.configPath">
              <Icon name="cog" size="sm" />
              {{ t('admin.settings.featureManagement.configure') }}
            </RouterLink>
          </div>
        </article>
      </section>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { UpdateSettingsRequest } from '@/api/admin/settings'
import { useAdminSettingsStore, useAppStore } from '@/stores'

type EnabledKey =
  | 'channel_monitor_enabled'
  | 'available_channels_enabled'
  | 'model_plaza_enabled'
  | 'payment_enabled'
  | 'affiliate_enabled'
  | 'risk_control_enabled'
  | 'ops_monitoring_enabled'
  | 'support_tickets_enabled'
  | 'support_ticket_account_enabled'
  | 'support_ticket_refund_enabled'
type VisibleKey =
  | 'channel_monitor_user_visible'
  | 'available_channels_user_visible'
  | 'model_plaza_user_visible'
  | 'payment_user_visible'
  | 'affiliate_user_visible'
  | 'support_tickets_user_visible'
type FeatureState = Record<EnabledKey | VisibleKey, boolean>
type FeatureDefinition = {
  id: string
  titleKey: string
  descriptionKey: string
  enabledKey: EnabledKey
  visibleKey?: VisibleKey
  secondaryKeys?: Array<{ key: EnabledKey; labelKey: string }>
  configPath: string
  icon: 'chart' | 'globe' | 'grid' | 'creditCard' | 'gift' | 'shield' | 'server' | 'clipboard'
}

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const loading = ref(false)
const loaded = ref(false)
const saving = ref(new Set<string>())
const state = reactive<FeatureState>({
  channel_monitor_enabled: true,
  channel_monitor_user_visible: true,
  available_channels_enabled: false,
  available_channels_user_visible: false,
  model_plaza_enabled: false,
  model_plaza_user_visible: false,
  payment_enabled: false,
  payment_user_visible: false,
  affiliate_enabled: false,
  affiliate_user_visible: false,
  risk_control_enabled: false,
  ops_monitoring_enabled: true,
  support_tickets_enabled: false,
  support_tickets_user_visible: false,
  support_ticket_account_enabled: true,
  support_ticket_refund_enabled: true,
})

const features: FeatureDefinition[] = [
  { id: 'channel-monitor', titleKey: 'admin.settings.features.channelMonitor.title', descriptionKey: 'admin.settings.features.channelMonitor.description', enabledKey: 'channel_monitor_enabled', visibleKey: 'channel_monitor_user_visible', configPath: '/admin/channels/monitor/settings', icon: 'chart' },
  { id: 'available-channels', titleKey: 'admin.settings.features.availableChannels.title', descriptionKey: 'admin.settings.features.availableChannels.description', enabledKey: 'available_channels_enabled', visibleKey: 'available_channels_user_visible', configPath: '/admin/channels/pricing', icon: 'globe' },
  { id: 'model-plaza', titleKey: 'admin.settings.features.modelPlaza.title', descriptionKey: 'admin.settings.features.modelPlaza.description', enabledKey: 'model_plaza_enabled', visibleKey: 'model_plaza_user_visible', configPath: '/admin/model-plaza/settings', icon: 'grid' },
  { id: 'payment', titleKey: 'admin.settings.featureManagement.modules.payment', descriptionKey: 'admin.settings.featureManagement.moduleDescriptions.payment', enabledKey: 'payment_enabled', visibleKey: 'payment_user_visible', configPath: '/admin/orders/settings', icon: 'creditCard' },
  {
    id: 'support-tickets',
    titleKey: 'admin.settings.featureManagement.modules.supportTickets',
    descriptionKey: 'admin.settings.featureManagement.moduleDescriptions.supportTickets',
    enabledKey: 'support_tickets_enabled',
    visibleKey: 'support_tickets_user_visible',
    configPath: '/admin/tickets',
    icon: 'clipboard',
  },
  { id: 'affiliate', titleKey: 'admin.settings.features.affiliate.title', descriptionKey: 'admin.settings.features.affiliate.description', enabledKey: 'affiliate_enabled', visibleKey: 'affiliate_user_visible', configPath: '/admin/affiliates/settings', icon: 'gift' },
  { id: 'risk-control', titleKey: 'admin.settings.features.riskControl.title', descriptionKey: 'admin.settings.features.riskControl.description', enabledKey: 'risk_control_enabled', configPath: '/admin/risk-control', icon: 'shield' },
  { id: 'ops-monitoring', titleKey: 'admin.settings.featureManagement.modules.opsMonitoring', descriptionKey: 'admin.settings.featureManagement.moduleDescriptions.opsMonitoring', enabledKey: 'ops_monitoring_enabled', configPath: '/admin/ops/settings', icon: 'server' },
]

async function load() {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getSettings()
    for (const feature of features) {
      state[feature.enabledKey] = Boolean(settings[feature.enabledKey])
      if (feature.visibleKey) state[feature.visibleKey] = Boolean(settings[feature.visibleKey])
      for (const secondary of feature.secondaryKeys || []) state[secondary.key] = Boolean(settings[secondary.key])
    }
    loaded.value = true
  } catch {
    appStore.showError(t('admin.settings.featureManagement.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function updateFeature(feature: FeatureDefinition, kind: 'enabled' | 'visible' | EnabledKey, next: boolean) {
  if (saving.value.has(feature.id)) return
  const beforeEnabled = state[feature.enabledKey]
  const beforeVisible = feature.visibleKey ? state[feature.visibleKey] : false
  const beforeSecondary = Object.fromEntries((feature.secondaryKeys || []).map(item => [item.key, state[item.key]])) as Partial<FeatureState>
  const payload: UpdateSettingsRequest = {}
  if (kind === 'enabled') {
    state[feature.enabledKey] = next
    payload[feature.enabledKey] = next
    if (!next && feature.visibleKey) {
      state[feature.visibleKey] = false
      payload[feature.visibleKey] = false
    }
  } else if (feature.visibleKey) {
    if (kind === 'visible') {
      state[feature.visibleKey] = next
      payload[feature.visibleKey] = next
    } else {
      state[kind] = next
      payload[kind] = next
    }
  } else if (kind !== 'visible') {
    state[kind] = next
    payload[kind] = next
  }

  saving.value = new Set(saving.value).add(feature.id)
  try {
    const updated = await adminAPI.settings.updateSettings(payload)
    state[feature.enabledKey] = Boolean(updated[feature.enabledKey])
    if (feature.visibleKey) state[feature.visibleKey] = Boolean(updated[feature.visibleKey])
    for (const secondary of feature.secondaryKeys || []) state[secondary.key] = Boolean(updated[secondary.key])
    await Promise.all([appStore.fetchPublicSettings(true), adminSettingsStore.fetch(true)])
  } catch {
    state[feature.enabledKey] = beforeEnabled
    if (feature.visibleKey) state[feature.visibleKey] = beforeVisible
    for (const secondary of feature.secondaryKeys || []) state[secondary.key] = Boolean(beforeSecondary[secondary.key])
    appStore.showError(t('admin.settings.featureManagement.saveFailed'))
  } finally {
    const nextSaving = new Set(saving.value)
    nextSaving.delete(feature.id)
    saving.value = nextSaving
  }
}

onMounted(load)
</script>

<style scoped>
.feature-management { width: min(1180px, 100%); margin: 0 auto; padding: 2rem; }
.feature-management__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1.5rem; }
.feature-management__header h1 { margin: 0; font-size: var(--font-size-2xl); }
.feature-management__header p { margin: .45rem 0 0; color: var(--color-text-secondary); }
.feature-management__header .btn, .feature-management__configure { display: inline-flex; align-items: center; gap: .45rem; }
.feature-management__list { border: 1px solid var(--glass-border); border-radius: 8px; background: var(--glass-bg); overflow: hidden; }
.feature-management__row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 1.5rem; align-items: center; padding: 1.15rem 1.25rem; border-bottom: 1px solid var(--glass-border); }
.feature-management__row:last-child { border-bottom: 0; }
.feature-management__identity { display: flex; min-width: 0; gap: .9rem; align-items: flex-start; }
.feature-management__icon { display: grid; place-items: center; width: 2.25rem; height: 2.25rem; flex: 0 0 auto; border-radius: 6px; background: var(--color-surface-muted); color: var(--color-primary); }
.feature-management__title-line { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; }
.feature-management__title-line h2 { margin: 0; font-size: var(--font-size-base); }
.feature-management__identity p { margin: .35rem 0 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); line-height: 1.5; }
.feature-management__status { padding: .18rem .45rem; border-radius: 4px; font-size: var(--font-size-xs); font-weight: 600; background: var(--color-surface-muted); color: var(--color-text-secondary); }
.feature-management__status.is-on, .feature-management__status.is-published { color: var(--color-success); }
.feature-management__status.is-off { color: var(--color-danger); }
.feature-management__controls { display: flex; align-items: center; justify-content: flex-end; gap: 1rem; }
.feature-management__switch { display: grid; justify-items: center; gap: .35rem; color: var(--color-text-secondary); font-size: var(--font-size-xs); white-space: nowrap; }
.feature-management__loading { min-height: 18rem; display: grid; place-content: center; justify-items: center; gap: .75rem; color: var(--color-text-secondary); }
.feature-management__loading svg { animation: feature-spin 1s linear infinite; }
@keyframes feature-spin { to { transform: rotate(360deg); } }
@media (max-width: 900px) { .feature-management { padding: 1rem; } .feature-management__row { grid-template-columns: 1fr; } .feature-management__controls { justify-content: flex-start; flex-wrap: wrap; padding-left: 3.15rem; } }
@media (max-width: 560px) { .feature-management__header { align-items: stretch; flex-direction: column; } .feature-management__controls { padding-left: 0; } .feature-management__configure { width: 100%; justify-content: center; } }
</style>
