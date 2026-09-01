<template>
  <section
    class="features-channel-monitor-v2-monitor-trend-chart__section card"
  >
    <div class="features-channel-monitor-v2-monitor-trend-chart__panel card-header">
      <div class="features-channel-monitor-v2-monitor-trend-chart__panel-2">
        <h2 class="features-channel-monitor-v2-monitor-trend-chart__heading">
          <span class="features-channel-monitor-v2-monitor-trend-chart__text" aria-hidden="true">
            <Icon name="chart" size="sm" />
          </span>
          {{ t('channelMonitorV2.chart.title') }}
        </h2>
        <p class="features-channel-monitor-v2-monitor-trend-chart__description">
          {{ t('channelMonitorV2.chart.description') }}
        </p>
      </div>
      <div class="features-channel-monitor-v2-monitor-trend-chart__panel-3">
        <span class="features-channel-monitor-v2-monitor-trend-chart__text-2">
          <span class="features-channel-monitor-v2-monitor-trend-chart__text-3"></span>{{ t('channelMonitorV2.chart.errorLegend') }}
        </span>
        <span class="features-channel-monitor-v2-monitor-trend-chart__text-2">
          <span class="features-channel-monitor-v2-monitor-trend-chart__text-4"></span>{{ t('channelMonitorV2.chart.cacheLegend') }}
        </span>
        <span class="features-channel-monitor-v2-monitor-trend-chart__text-2">
          <span class="features-channel-monitor-v2-monitor-trend-chart__text-5"></span>{{ t('channelMonitorV2.chart.ttftLegend') }}
        </span>
        <span class="features-channel-monitor-v2-monitor-trend-chart__text-6 badge badge-gray">{{ bucketLabel }}</span>
        <button
          type="button"
          class="features-channel-monitor-v2-monitor-trend-chart__action"
          :disabled="!zoomed"
          @click="resetChartZoom"
        >
          {{ t('channelMonitorV2.chart.resetZoom') }}
        </button>
      </div>
    </div>
    <div class="features-channel-monitor-v2-monitor-trend-chart__panel-4 card-body">
      <div v-if="loading" class="features-channel-monitor-v2-monitor-trend-chart__panel-5">
        <div class="features-channel-monitor-v2-monitor-trend-chart__panel-6">{{ t('common.loading') }}</div>
      </div>
      <div
        v-else-if="chartData"
        ref="chartRef"
        class="features-channel-monitor-v2-monitor-trend-chart__panel-7"
        @wheel="onChartWheel"
      >
        <D3LineChart :data="chartData" :options="chartOptions" />
      </div>
      <div v-else class="features-channel-monitor-v2-monitor-trend-chart__panel-5">
        <EmptyState
          :title="t('channelMonitorV2.chart.emptyTitle')"
          :description="t('channelMonitorV2.empty.description')"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref, watch } from 'vue'
import EmptyState from '@/components/common/EmptyState.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import Icon from '@/components/icons/Icon.vue'
import { useThemeColors } from '@/composables/useThemeColors'
import type { MonitorCoverage, MonitorMetric, MonitorHealth } from '@/api/channelMonitorV2'
import { formatMonitorMs, formatMonitorPercent } from '@/features/channel-monitor-v2/monitorFormat'
import {
  applyWheelZoom,
  clientXRatio,
  isZoomed,
  resetZoom,
  sliceByZoom,
  type ZoomState,
} from '@/features/channel-monitor-v2/monitorZoom'

const { t, locale } = useI18n()

const props = defineProps<{
  trend: Array<{ bucket_start: string; metrics: MonitorMetric; health: MonitorHealth }>
  coverage: MonitorCoverage | null
  loading?: boolean
}>()

const chartRef = ref<HTMLElement | null>(null)
const zoom = ref<ZoomState>(resetZoom())
const zoomed = computed(() => isZoomed(zoom.value))
const themeColors = useThemeColors()

const bucketLabel = computed(() => {
  const seconds = props.coverage?.bucket_seconds || 60
  const minutes = seconds / 60
  if (minutes < 60) return t('channelMonitorV2.bucket.minutes', { count: minutes })
  const hours = minutes / 60
  if (hours < 24) return t('channelMonitorV2.bucket.hours', { count: hours })
  return t('channelMonitorV2.bucket.days', { count: hours / 24 })
})

const chartData = computed(() => {
  const points = visibleTrend.value
  if (!points.length) return null
  const labels = points.map((p) =>
    new Intl.DateTimeFormat(locale.value || undefined, {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(p.bucket_start))
  )
  const errorRates = smoothTrend(points.map((p) => (p.metrics.error_rate || 0) * 100))
  const cacheRates = smoothTrend(points.map((p) => (p.metrics.cache_rate || 0) * 100))
  const ttftP50 = smoothTrend(points.map((p) => p.metrics.ttft?.p50_ms ?? null))
  return {
    labels,
    datasets: [
      {
        label: t('channelMonitorV2.chart.errorDataset'),
        data: errorRates,
        borderColor: '#ef4444',
        backgroundColor: 'rgba(239, 67, 67, 0.10)',
        yAxisID: 'yPct',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: 'origin' as const,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
      },
      {
        label: t('channelMonitorV2.chart.cacheDataset'),
        data: cacheRates,
        borderColor: '#10b981',
        backgroundColor: 'rgba(16, 185, 129, 0.08)',
        yAxisID: 'yPct',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: false,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
      },
      {
        label: t('channelMonitorV2.chart.ttftDataset'),
        data: ttftP50,
        borderColor: '#0ea5e9',
        backgroundColor: 'rgba(14, 165, 233, 0.08)',
        yAxisID: 'yTtft',
        tension: 0.4,
        cubicInterpolationMode: 'monotone' as const,
        fill: false,
        pointRadius: 0,
        pointHoverRadius: 4,
        pointHitRadius: 10,
        borderWidth: 2,
        spanGaps: true,
      },
    ],
  }
})

/** Window the series by zoom state around the cursor — not always the last N points. */
const visibleTrend = computed(() => sliceByZoom(props.trend || [], zoom.value))

function onChartWheel(event: WheelEvent) {
  // Plain vertical wheel zooms X (narrower time range); shift/horizontal pans.
  event.preventDefault()
  const ratio = clientXRatio(event.clientX, chartRef.value)
  zoom.value = applyWheelZoom(zoom.value, event, ratio)
}

function resetChartZoom() {
  zoom.value = resetZoom()
}

watch(() => props.trend, () => {
  zoom.value = resetZoom()
})

function smoothTrend(values: Array<number | null>): Array<number | null> {
  if (values.length <= 2) return values
  return values.map((value, index) => {
    if (value == null) return null
    const neighbors = values.slice(Math.max(0, index - 1), Math.min(values.length, index + 2))
      .filter((item): item is number => item != null)
    if (!neighbors.length) return value
    return neighbors.reduce((sum, item) => sum + item, 0) / neighbors.length
  })
}

const chartOptions = computed(() => {
  const {
    textPrimary: tooltipTitle,
    textSecondary: tooltipBody,
    textTertiary: text,
    grid,
    elevatedSurface: tooltipBg
  } = themeColors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: 'index' as const, intersect: false },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: tooltipBg,
        titleColor: tooltipTitle,
        bodyColor: tooltipBody,
        borderColor: grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label(ctx: { dataset: { label?: string }; parsed: { y: number | null } }) {
            const label = ctx.dataset.label || ''
            const y = ctx.parsed.y
            if (y == null) return `${label}: -`
            if (label === t('channelMonitorV2.chart.errorDataset') || label === t('channelMonitorV2.chart.cacheDataset')) {
              return `${label}: ${formatMonitorPercent(y / 100)}`
            }
            return `${label}: ${formatMonitorMs(y)}`
          },
        },
      },
    },
    scales: {
      x: {
        ticks: { color: text, maxRotation: 0, autoSkip: true, maxTicksLimit: 8, autoSkipPadding: 10, font: { size: 10 } },
        grid: { display: false },
      },
      yPct: {
        type: 'linear' as const,
        position: 'left' as const,
        min: 0,
        suggestedMax: 100,
        ticks: {
          color: text,
          font: { size: 10 },
          callback: (v: string | number) => `${v}%`,
        },
        grid: { color: grid, borderDash: [4, 4] },
        title: { display: true, text: t('channelMonitorV2.chart.percentAxis'), color: text, font: { size: 11 } },
      },
      yTtft: {
        type: 'linear' as const,
        position: 'right' as const,
        min: 0,
        ticks: {
          color: '#0ea5e9',
          font: { size: 10 },
          callback: (v: string | number) => formatMonitorMs(Number(v)),
        },
        grid: { display: false },
        title: { display: true, text: t('channelMonitorV2.metrics.ttftP50'), color: '#0ea5e9', font: { size: 11 } },
      },
    },
  }
})
</script>
