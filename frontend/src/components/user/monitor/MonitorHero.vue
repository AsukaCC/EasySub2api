<template>
  <section class="components-user-monitor-monitor-hero__section">
    <div class="components-user-monitor-monitor-hero__panel">
      <div
        role="tablist"
        class="components-user-monitor-monitor-hero__panel-2"
      >
        <button
          v-for="opt in windowOptions"
          :key="opt.value"
          type="button"
          role="tab"
          :aria-selected="window === opt.value"
          class="components-user-monitor-monitor-hero__action"
          :class="window === opt.value
            ? 'components-user-monitor-monitor-hero__action-3'
            : 'components-user-monitor-monitor-hero__action-4'"
          @click="emit('update:window', opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <span
        class="components-user-monitor-monitor-hero__text"
        :class="overallChipClass"
      >
        <span
          class="components-user-monitor-monitor-hero__text-2"
          :class="overallDotClass"
        ></span>
        {{ overallLabel }}
      </span>

      <button
        type="button"
        class="components-user-monitor-monitor-hero__action-2"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'components-user-monitor-monitor-hero__icon' : ''" />
      </button>

      <AutoRefreshButton
        v-if="autoRefresh"
        :enabled="autoRefresh.enabled.value"
        :interval-seconds="autoRefresh.intervalSeconds.value"
        :countdown="autoRefresh.countdown.value"
        :intervals="autoRefresh.intervals"
        @update:enabled="autoRefresh.setEnabled"
        @update:interval="autoRefresh.setInterval"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AutoRefreshButton from '@/components/common/AutoRefreshButton.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

const props = defineProps<{
  overallStatus: OverallStatus
  intervalSeconds: number
  window: MonitorWindow
  loading: boolean
  autoRefresh?: {
    enabled: { value: boolean }
    intervalSeconds: { value: number }
    countdown: { value: number }
    intervals: readonly number[]
    setEnabled: (v: boolean) => void
    setInterval: (v: number) => void
  }
}>()

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])

const overallLabel = computed(() => t(`channelStatus.overall.${props.overallStatus}`))

const overallChipClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'components-user-monitor-monitor-hero__state'
    case 'degraded':
    default:
      return 'components-user-monitor-monitor-hero__state-2'
  }
})

const overallDotClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'components-user-monitor-monitor-hero__state-3'
    case 'degraded':
    default:
      return 'components-user-monitor-monitor-hero__state-4'
  }
})

</script>
