<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsErrorTrendPoint } from '@/api/admin/ops'
import type { ChartState } from '../types'
import { formatHistoryLabel, sumNumbers } from '../utils/opsFormatters'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'

interface Props {
  points: OpsErrorTrendPoint[]
  loading: boolean
  timeRange: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'openRequestErrors'): void
  (e: 'openUpstreamErrors'): void
}>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const colors = computed(() => ({
  red: '#ef4444',
  redAlpha: '#ef444420',
  purple: '#8b5cf6',
  purpleAlpha: '#8b5cf620',
  gray: '#9ca3af',
  grid: isDarkMode.value ? '#2e2e33' : '#f3f4f6',
  text: isDarkMode.value ? '#9ca3af' : '#6b7280'
}))

const totalRequestErrors = computed(() => sumNumbers(props.points.map((p) => p.error_count_sla ?? 0)))

const totalUpstreamErrors = computed(() =>
  sumNumbers(
    props.points.map((p) => (p.upstream_error_count_excl_429_529 ?? 0) + (p.upstream_429_count ?? 0) + (p.upstream_529_count ?? 0))
  )
)

const totalDisplayed = computed(() =>
  sumNumbers(props.points.map((p) => (p.error_count_sla ?? 0) + (p.upstream_error_count_excl_429_529 ?? 0) + (p.business_limited_count ?? 0)))
)

const hasRequestErrors = computed(() => totalRequestErrors.value > 0)
const hasUpstreamErrors = computed(() => totalUpstreamErrors.value > 0)

const chartData = computed(() => {
  if (!props.points.length || totalDisplayed.value <= 0) return null
  return {
    labels: props.points.map((p) => formatHistoryLabel(p.bucket_start, props.timeRange)),
    datasets: [
      {
        label: t('admin.ops.errorsSla'),
        data: props.points.map((p) => p.error_count_sla ?? 0),
        borderColor: colors.value.red,
        backgroundColor: colors.value.redAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.upstreamExcl429529'),
        data: props.points.map((p) => p.upstream_error_count_excl_429_529 ?? 0),
        borderColor: colors.value.purple,
        backgroundColor: colors.value.purpleAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.businessLimited'),
        data: props.points.map((p) => p.business_limited_count ?? 0),
        borderColor: colors.value.gray,
        backgroundColor: 'transparent',
        borderDash: [6, 6],
        fill: false,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      }
    ]
  }
})

const state = computed<ChartState>(() => {
  if (chartData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' as const },
    plugins: {
      legend: {
        position: 'top' as const,
        align: 'end' as const,
        labels: { color: c.text, usePointStyle: true, boxWidth: 6, font: { size: 10 } }
      },
      tooltip: {
        backgroundColor: isDarkMode.value ? '#1b1b1f' : '#ffffff',
        titleColor: isDarkMode.value ? '#f3f4f6' : '#121214',
        bodyColor: isDarkMode.value ? '#d1d5db' : '#4b5563',
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true
      }
    },
    scales: {
      x: {
        type: 'category' as const,
        grid: { display: false },
        ticks: {
          color: c.text,
          font: { size: 10 },
          maxTicksLimit: 8,
          autoSkip: true,
          autoSkipPadding: 10
        }
      },
      y: {
        type: 'linear' as const,
        display: true,
        position: 'left' as const,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 }, precision: 0 }
      }
    }
  }
})
</script>

<template>
  <div class="views-admin-ops-components-ops-error-trend-chart__panel card-body">
    <div class="views-admin-ops-components-ops-error-trend-chart__panel-2">
      <h3 class="views-admin-ops-components-ops-error-trend-chart__heading">
        <svg class="views-admin-ops-components-ops-error-trend-chart__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"
          />
        </svg>
        {{ t('admin.ops.errorTrend') }}
        <HelpTooltip :content="t('admin.ops.tooltips.errorTrend')" />
      </h3>
      <div class="views-admin-ops-components-ops-error-trend-chart__panel-3">
        <button
          type="button"
          class="views-admin-ops-components-ops-error-trend-chart__action"
          :disabled="!hasRequestErrors"
          @click="emit('openRequestErrors')"
        >
          {{ t('admin.ops.errorDetails.requestErrors') }}
        </button>
        <button
          type="button"
          class="views-admin-ops-components-ops-error-trend-chart__action"
          :disabled="!hasUpstreamErrors"
          @click="emit('openUpstreamErrors')"
        >
          {{ t('admin.ops.errorDetails.upstreamErrors') }}
        </button>
      </div>
    </div>

    <div class="views-admin-ops-components-ops-error-trend-chart__panel-4">
      <D3LineChart v-if="state === 'ready' && chartData" :data="chartData" :options="options" />
      <div v-else class="views-admin-ops-components-ops-error-trend-chart__panel-5">
        <div v-if="state === 'loading'" class="views-admin-ops-components-ops-error-trend-chart__panel-6">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyError')" />
      </div>
    </div>
  </div>
</template>
