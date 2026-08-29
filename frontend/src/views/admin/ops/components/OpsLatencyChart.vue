<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsLatencyHistogramResponse } from '@/api/admin/ops'
import type { ChartState } from '../types'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import D3BarChart from '@/components/charts/d3/D3BarChart.vue'

interface Props {
  latencyData: OpsLatencyHistogramResponse | null
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const colors = computed(() => ({
  blue: '#3b82f6',
  grid: isDarkMode.value ? '#2e2e33' : '#f3f4f6',
  text: isDarkMode.value ? '#9ca3af' : '#6b7280'
}))

const hasData = computed(() => (props.latencyData?.total_requests ?? 0) > 0)

const state = computed<ChartState>(() => {
  if (hasData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

const chartData = computed(() => {
  if (!props.latencyData || !hasData.value) return null
  const c = colors.value
  return {
    labels: props.latencyData.buckets.map((b) => b.range),
    datasets: [
      {
        label: t('admin.ops.requests'),
        data: props.latencyData.buckets.map((b) => b.count),
        backgroundColor: c.blue,
        borderRadius: 4,
        barPercentage: 0.6
      }
    ]
  }
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false }
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { color: c.text, font: { size: 10 } }
      },
      y: {
        beginAtZero: true,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 } }
      }
    }
  }
})
</script>

<template>
  <div class="views-admin-ops-components-ops-latency-chart__panel">
    <div class="views-admin-ops-components-ops-latency-chart__panel-2">
      <h3 class="views-admin-ops-components-ops-latency-chart__heading">
        <svg class="views-admin-ops-components-ops-latency-chart__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        {{ t('admin.ops.latencyHistogram') }}
        <HelpTooltip :content="t('admin.ops.tooltips.latencyHistogram')" />
      </h3>
    </div>

    <div class="views-admin-ops-components-ops-latency-chart__panel-3">
      <D3BarChart v-if="state === 'ready' && chartData" :data="chartData" :options="options" />
      <div v-else class="views-admin-ops-components-ops-latency-chart__panel-4">
        <div v-if="state === 'loading'" class="views-admin-ops-components-ops-latency-chart__panel-5">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyRequest')" />
      </div>
    </div>
  </div>
</template>
