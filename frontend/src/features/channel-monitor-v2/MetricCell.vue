<template>
  <div
    class="features-channel-monitor-v2-metric-cell__panel stat-card"
    :title="title || undefined"
  >
    <div
      v-if="state"
      class="features-channel-monitor-v2-metric-cell__panel-2"
      :class="dotClass"
      aria-hidden="true"
    ></div>
    <div class="features-channel-monitor-v2-metric-cell__panel-3">
      <span class="features-channel-monitor-v2-metric-cell__text stat-label">{{ label }}</span>
      <strong
        class="features-channel-monitor-v2-metric-cell__strong stat-value"
        :class="stateClass"
      >{{ value }}</strong>
      <div
        v-if="detailParts.length > 1"
        class="features-channel-monitor-v2-metric-cell__panel-4"
      >
        <span
          v-for="(part, index) in detailParts"
          :key="`${index}:${part}`"
          class="features-channel-monitor-v2-metric-cell__text-2"
        >{{ part }}</span>
      </div>
      <small
        v-else-if="detail"
        class="features-channel-monitor-v2-metric-cell__small"
      >{{ detail }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { HealthState } from '@/api/channelMonitorV2'

const props = defineProps<{
  label: string
  value: string
  detail: string
  state?: HealthState
  /** Exact numeric tooltip (e.g. uncompacted RPM/TPM). */
  title?: string
}>()

/** Split "AVG 475ms · P90 800ms" into chips so nothing is ellipsized. */
const detailParts = computed(() => {
  const raw = (props.detail || '').trim()
  if (!raw || raw === '-') return []
  return raw
    .split(/\s*[·|]\s*/)
    .map((part) => part.trim())
    .filter(Boolean)
})

const stateClass = computed(() => {
  if (!props.state) return 'features-channel-monitor-v2-metric-cell__state'
  if (props.state === 'healthy') return 'features-channel-monitor-v2-metric-cell__state-2'
  if (props.state === 'warning') return 'features-channel-monitor-v2-metric-cell__state-3'
  if (props.state === 'critical') return 'features-channel-monitor-v2-metric-cell__state-4'
  return 'features-channel-monitor-v2-metric-cell__state-5'
})

const dotClass = computed(() => {
  if (props.state === 'healthy') return 'status-fill--success'
  if (props.state === 'warning') return 'status-fill--warning'
  if (props.state === 'critical') return 'status-fill--danger'
  return 'features-channel-monitor-v2-metric-cell__state-6'
})
</script>
