<template>
  <button
    type="button"
    class="components-user-monitor-monitor-card__action"
    @click="emit('click')"
  >
    <!-- Header: icon + name/model + status chip -->
    <div class="components-user-monitor-monitor-card__panel">
      <span
        class="components-user-monitor-monitor-card__text"
        :class="[providerGradient(item.provider), providerTintClass]"
      >
        <ProviderIcon :provider="item.provider" :size="20" />
      </span>
      <div class="components-user-monitor-monitor-card__panel-2">
        <div class="components-user-monitor-monitor-card__panel-3">
          {{ item.name }}
        </div>
        <div class="components-user-monitor-monitor-card__panel-4">
          <span
            class="components-user-monitor-monitor-card__text-2"
            :class="providerBadgeClass(item.provider)"
          >
            {{ providerLabel(item.provider) }}
          </span>
          <span class="components-user-monitor-monitor-card__text-3">
            {{ item.primary_model }}
          </span>
          <span
            v-if="item.group_name"
            class="components-user-monitor-monitor-card__text-4"
          >
            {{ item.group_name }}
          </span>
        </div>
      </div>
      <span
        class="components-user-monitor-monitor-card__text-5"
        :class="statusBadgeClass(item.primary_status)"
      >
        {{ statusLabel(item.primary_status) }}
      </span>
    </div>

    <!-- Metrics -->
    <MonitorMetricPair
      primary-icon="bolt"
      :primary-label="t('monitorCommon.dialogLatency')"
      :primary-value="formatLatency(item.primary_latency_ms)"
      primary-unit="ms"
      secondary-icon="globe"
      :secondary-label="t('monitorCommon.endpointPing')"
      :secondary-value="formatLatency(item.primary_ping_latency_ms)"
      secondary-unit="ms"
    />

    <!-- 配额模式：最新用量/余额快照（服务端已按系统开关剥离，此处 flag 为纵深防御） -->
    <MonitorQuotaView v-if="quotaVisible" :snapshot="item.latest_quota" class="components-user-monitor-monitor-card__monitor-quota-view" />

    <!-- Divider -->
    <div class="components-user-monitor-monitor-card__panel-5"></div>

    <!-- Availability row -->
    <MonitorAvailabilityRow
      :window-label="availabilityLabel"
      :value="availabilityValue"
      :samples-label="extraModelsCountLabel"
    />

    <!-- Timeline -->
    <MonitorTimeline
      :buckets="item.timeline"
      :countdown-seconds="countdownSeconds"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import {
  useChannelMonitorFormat,
  providerGradient,
} from '@/composables/useChannelMonitorFormat'
import { isChannelMonitorQuotaVisible } from '@/utils/featureFlags'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'
import MonitorQuotaView from '@/components/common/MonitorQuotaView.vue'

// 图标配色与 utils/platformColors.ts 的平台色对齐（新 4 家）。
const PROVIDER_TINT: Record<string, string> = {
  openai: 'components-user-monitor-monitor-card__state',
  anthropic: 'components-user-monitor-monitor-card__state-2',
  grok: 'components-user-monitor-monitor-card__state-4',
  kimi: 'components-user-monitor-monitor-card__state-6',
  zhipu: 'components-user-monitor-monitor-card__state-7',
  deepseek: 'components-user-monitor-monitor-card__state-8',
}

const props = defineProps<{
  item: UserMonitorView
  window: '7d' | '15d' | '30d'
  availabilityValue: number | null
  countdownSeconds: number
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const { t } = useI18n()
const {
  statusLabel,
  statusBadgeClass,
  providerLabel,
  providerBadgeClass,
  formatLatency,
} = useChannelMonitorFormat()

const providerTintClass = computed(() =>
  PROVIDER_TINT[props.item.provider] ?? 'components-user-monitor-monitor-card__state-9'
)

const quotaVisible = computed(
  () => isChannelMonitorQuotaVisible() && !!props.item.latest_quota
)

const availabilityLabel = computed(() => {
  const win = t(`channelStatus.windowTab.${props.window}`)
  return `${t('monitorCommon.availabilityPrefix')} · ${win}`
})

const extraModelsCountLabel = computed(() => {
  const count = props.item.extra_models?.length ?? 0
  if (count === 0) return undefined
  return t('monitorCommon.extraModelsCount', { n: count })
})
</script>
