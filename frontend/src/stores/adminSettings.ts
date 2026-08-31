import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminAPI } from '@/api'
import type { CustomMenuItem } from '@/types'

export const useAdminSettingsStore = defineStore('adminSettings', () => {
  const loaded = ref(false)
  const loading = ref(false)

  const readCachedBool = (key: string, defaultValue: boolean): boolean => {
    try {
      const raw = localStorage.getItem(key)
      if (raw === 'true') return true
      if (raw === 'false') return false
    } catch {
      // ignore localStorage failures
    }
    return defaultValue
  }

  const writeCachedBool = (key: string, value: boolean) => {
    try {
      localStorage.setItem(key, value ? 'true' : 'false')
    } catch {
      // ignore localStorage failures
    }
  }

  const readCachedString = (key: string, defaultValue: string): string => {
    try {
      const raw = localStorage.getItem(key)
      if (typeof raw === 'string' && raw.length > 0) return raw
    } catch {
      // ignore localStorage failures
    }
    return defaultValue
  }

  const writeCachedString = (key: string, value: string) => {
    try {
      localStorage.setItem(key, value)
    } catch {
      // ignore localStorage failures
    }
  }

  // Default open, but honor cached value to reduce UI flicker on first paint.
  const opsMonitoringEnabled = ref(readCachedBool('ops_monitoring_enabled_cached', true))
  const opsRealtimeMonitoringEnabled = ref(readCachedBool('ops_realtime_monitoring_enabled_cached', true))
  const opsQueryModeDefault = ref(readCachedString('ops_query_mode_default_cached', 'auto'))
  const paymentEnabled = ref(readCachedBool('payment_enabled_cached', false))
  const channelMonitorEnabled = ref(readCachedBool('channel_monitor_enabled_admin_cached', true))
  const availableChannelsEnabled = ref(readCachedBool('available_channels_enabled_admin_cached', false))
  const modelPlazaEnabled = ref(readCachedBool('model_plaza_enabled_admin_cached', false))
  const riskControlEnabled = ref(readCachedBool('risk_control_enabled_admin_cached', false))
  const affiliateEnabled = ref(readCachedBool('affiliate_enabled_admin_cached', false))
  const supportTicketsEnabled = ref(readCachedBool('support_tickets_enabled_admin_cached', false))
  const customMenuItems = ref<CustomMenuItem[]>([])

  async function fetch(force = false): Promise<void> {
    if (loaded.value && !force) return
    if (loading.value) return

    loading.value = true
    try {
      const [settings, paymentConfigResp] = await Promise.all([
        adminAPI.settings.getSettings(),
        adminAPI.payment.getConfig()
      ])
      opsMonitoringEnabled.value = settings.ops_monitoring_enabled ?? true
      writeCachedBool('ops_monitoring_enabled_cached', opsMonitoringEnabled.value)

      opsRealtimeMonitoringEnabled.value = settings.ops_realtime_monitoring_enabled ?? true
      writeCachedBool('ops_realtime_monitoring_enabled_cached', opsRealtimeMonitoringEnabled.value)

      opsQueryModeDefault.value = settings.ops_query_mode_default || 'auto'
      writeCachedString('ops_query_mode_default_cached', opsQueryModeDefault.value)

      customMenuItems.value = Array.isArray(settings.custom_menu_items) ? settings.custom_menu_items : []

      paymentEnabled.value = paymentConfigResp.data?.enabled ?? false
      writeCachedBool('payment_enabled_cached', paymentEnabled.value)

      channelMonitorEnabled.value = settings.channel_monitor_enabled ?? true
      availableChannelsEnabled.value = settings.available_channels_enabled ?? false
      modelPlazaEnabled.value = settings.model_plaza_enabled ?? false
      riskControlEnabled.value = settings.risk_control_enabled ?? false
      affiliateEnabled.value = settings.affiliate_enabled ?? false
      supportTicketsEnabled.value = settings.support_tickets_enabled ?? false
      writeCachedBool('channel_monitor_enabled_admin_cached', channelMonitorEnabled.value)
      writeCachedBool('available_channels_enabled_admin_cached', availableChannelsEnabled.value)
      writeCachedBool('model_plaza_enabled_admin_cached', modelPlazaEnabled.value)
      writeCachedBool('risk_control_enabled_admin_cached', riskControlEnabled.value)
      writeCachedBool('affiliate_enabled_admin_cached', affiliateEnabled.value)
      writeCachedBool('support_tickets_enabled_admin_cached', supportTicketsEnabled.value)

      loaded.value = true
    } catch (err) {
      // Keep cached/default value: do not "flip" the UI based on a transient fetch failure.
      loaded.value = true
      console.error('[adminSettings] Failed to fetch settings:', err)
    } finally {
      loading.value = false
    }
  }

  function setOpsMonitoringEnabledLocal(value: boolean) {
    opsMonitoringEnabled.value = value
    writeCachedBool('ops_monitoring_enabled_cached', value)
    loaded.value = true
  }

  function setOpsRealtimeMonitoringEnabledLocal(value: boolean) {
    opsRealtimeMonitoringEnabled.value = value
    writeCachedBool('ops_realtime_monitoring_enabled_cached', value)
    loaded.value = true
  }

  function setPaymentEnabledLocal(value: boolean) {
    paymentEnabled.value = value
    writeCachedBool('payment_enabled_cached', value)
    loaded.value = true
  }

  function setFeatureEnabledLocal(feature: 'channelMonitor' | 'availableChannels' | 'modelPlaza' | 'riskControl' | 'affiliate', value: boolean) {
    const refs = {
      channelMonitor: channelMonitorEnabled,
      availableChannels: availableChannelsEnabled,
      modelPlaza: modelPlazaEnabled,
      riskControl: riskControlEnabled,
      affiliate: affiliateEnabled,
    }
    const cacheKeys = {
      channelMonitor: 'channel_monitor_enabled_admin_cached',
      availableChannels: 'available_channels_enabled_admin_cached',
      modelPlaza: 'model_plaza_enabled_admin_cached',
      riskControl: 'risk_control_enabled_admin_cached',
      affiliate: 'affiliate_enabled_admin_cached',
    }
    refs[feature].value = value
    writeCachedBool(cacheKeys[feature], value)
    loaded.value = true
  }

  function setOpsQueryModeDefaultLocal(value: string) {
    opsQueryModeDefault.value = value || 'auto'
    writeCachedString('ops_query_mode_default_cached', opsQueryModeDefault.value)
    loaded.value = true
  }

  // Keep UI consistent if we learn that ops is disabled via feature-gated 404s.
  // (event is dispatched from the axios interceptor)
  let eventHandlerCleanup: (() => void) | null = null

  function initializeEventListeners() {
    if (eventHandlerCleanup) return

    try {
      const handler = () => {
        setOpsMonitoringEnabledLocal(false)
      }
      window.addEventListener('ops-monitoring-disabled', handler)
      eventHandlerCleanup = () => {
        window.removeEventListener('ops-monitoring-disabled', handler)
      }
    } catch {
      // ignore window access failures (SSR)
    }
  }

  if (typeof window !== 'undefined') {
    initializeEventListeners()
  }

  return {
    loaded,
    loading,
    opsMonitoringEnabled,
    opsRealtimeMonitoringEnabled,
    opsQueryModeDefault,
    paymentEnabled,
    channelMonitorEnabled,
    availableChannelsEnabled,
    modelPlazaEnabled,
    riskControlEnabled,
    affiliateEnabled,
    supportTicketsEnabled,
    customMenuItems,
    fetch,
    setOpsMonitoringEnabledLocal,
    setOpsRealtimeMonitoringEnabledLocal,
    setPaymentEnabledLocal,
    setFeatureEnabledLocal,
    setOpsQueryModeDefaultLocal
  }
})
