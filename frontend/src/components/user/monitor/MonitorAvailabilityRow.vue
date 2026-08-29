<template>
  <div class="monitor-availability">
    <div class="monitor-availability__window">
      {{ windowLabel }}
    </div>
    <div class="monitor-availability__value-group">
      <span
        class="monitor-availability__value"
        :style="colorStyle"
      >
        {{ displayValue }}
      </span>
      <span
        class="monitor-availability__unit"
        :style="colorStyle"
      >%</span>
    </div>
  </div>
  <div
    v-if="samplesLabel"
    class="monitor-availability__samples"
  >
    {{ samplesLabel }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { hslForPct } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  windowLabel: string
  value: number | null
  samplesLabel?: string
}>()

const { t } = useI18n()

const displayValue = computed(() => {
  if (props.value === null || Number.isNaN(props.value)) return t('monitorCommon.latencyEmpty')
  return props.value.toFixed(2)
})

const colorStyle = computed(() => {
  const colour = hslForPct(props.value)
  return colour ? { color: colour } : { color: 'var(--color-text-tertiary)' }
})
</script>

<style scoped>
.monitor-availability {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-top: 0.75rem;
}

.monitor-availability__window,
.monitor-availability__samples {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
}

.monitor-availability__window { text-transform: uppercase; }

.monitor-availability__value-group {
  display: flex;
  align-items: baseline;
  gap: 0.125rem;
}

.monitor-availability__value {
  font-size: var(--type-display-size);
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  line-height: var(--line-height-none);
}

.monitor-availability__unit {
  font-size: var(--type-card-size);
  font-weight: var(--font-weight-semibold);
  line-height: var(--line-height-none);
}

.monitor-availability__samples {
  margin-top: 0.25rem;
  text-align: right;
}
</style>
