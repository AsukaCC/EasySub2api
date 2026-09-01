<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsThroughputGroupBreakdownItem, OpsThroughputPlatformBreakdownItem, OpsThroughputTrendPoint } from '@/api/admin/ops'
import type { ChartState } from '../types'
import { formatHistoryLabel, sumNumbers } from '../utils/opsFormatters'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import type { D3LineChartHandle } from '@/components/charts/d3/chartTypes'
import { formatNumber } from '@/utils/format'

interface Props {
  points: OpsThroughputTrendPoint[]
  loading: boolean
  timeRange: string
  byPlatform?: OpsThroughputPlatformBreakdownItem[]
  topGroups?: OpsThroughputGroupBreakdownItem[]
  fullscreen?: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()
const emit = defineEmits<{
  (e: 'selectPlatform', platform: string): void
  (e: 'selectGroup', groupId: string): void
  (e: 'openDetails'): void
}>()

const throughputChartRef = ref<D3LineChartHandle | null>(null)
watch(
  () => props.timeRange,
  () => {
    setTimeout(() => {
      throughputChartRef.value?.resetZoom()
    }, 100)
  }
)

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const colors = computed(() => ({
  blue: '#3b82f6',
  blueAlpha: '#3b82f620',
  green: '#10b981',
  greenAlpha: '#10b98120',
  grid: isDarkMode.value ? '#2e2e33' : '#e4e4e7',
  text: isDarkMode.value ? '#9ca3af' : '#71717a'
}))

const totalRequests = computed(() => sumNumbers(props.points.map((p) => p.request_count)))

const chartData = computed(() => {
  if (!props.points.length || totalRequests.value <= 0) return null
  return {
    labels: props.points.map((p) => formatHistoryLabel(p.bucket_start, props.timeRange)),
    datasets: [
      {
        label: 'QPS',
        data: props.points.map((p) => p.qps ?? 0),
        borderColor: colors.value.blue,
        backgroundColor: colors.value.blueAlpha,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.tpsK'),
        data: props.points.map((p) => (p.tps ?? 0) / 1000),
        borderColor: colors.value.green,
        backgroundColor: colors.value.greenAlpha,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHitRadius: 10,
        yAxisID: 'y1'
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
        titleColor: isDarkMode.value ? '#f3f4f6' : '#09090b',
        bodyColor: isDarkMode.value ? '#d1d5db' : '#3f3f46',
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label: (context: any) => {
            let label = context.dataset.label || ''
            if (label) label += ': '
            if (context.raw !== null) label += context.parsed.y.toFixed(1)
            return label
          }
        }
      },
      zoom: {
        pan: { enabled: true, mode: 'x' as const, modifierKey: 'ctrl' as const },
        zoom: { wheel: { enabled: true }, pinch: { enabled: true }, mode: 'x' as const }
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
        ticks: { color: c.text, font: { size: 10 } }
      },
      y1: {
        type: 'linear' as const,
        display: true,
        position: 'right' as const,
        grid: { display: false },
        ticks: { color: c.green, font: { size: 10 } }
      }
    }
  }
})

function resetZoom() {
  throughputChartRef.value?.resetZoom()
}

function downloadChart() {
  const url = throughputChartRef.value?.toDataUrl()
  if (!url) return
  const a = document.createElement('a')
  a.href = url
  a.download = `ops-throughput-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.svg`
  a.click()
}
</script>

<template>
  <div class="views-admin-ops-components-ops-throughput-trend-chart__panel card-body">
    <div
      data-testid="throughput-chart-header"
      class="views-admin-ops-components-ops-throughput-trend-chart__panel-2"
    >
      <h3 class="views-admin-ops-components-ops-throughput-trend-chart__heading">
        <svg class="views-admin-ops-components-ops-throughput-trend-chart__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        {{ t('admin.ops.throughputTrend') }}
        <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.throughputTrend')" />
      </h3>
      <div
        data-testid="throughput-chart-toolbar"
        class="views-admin-ops-components-ops-throughput-trend-chart__panel-3"
      >
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text"><span class="views-admin-ops-components-ops-throughput-trend-chart__text-2"></span>QPS</span>
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text"><span class="views-admin-ops-components-ops-throughput-trend-chart__text-3"></span>{{ t('admin.ops.tpsK') }}</span>
        <template v-if="!props.fullscreen">
          <button
            type="button"
            class="views-admin-ops-components-ops-throughput-trend-chart__action"
            :disabled="state !== 'ready'"
            :title="t('admin.ops.requestDetails.title')"
            @click="emit('openDetails')"
          >
            {{ t('admin.ops.requestDetails.details') }}
          </button>
          <button
            type="button"
            class="views-admin-ops-components-ops-throughput-trend-chart__action"
            :disabled="state !== 'ready'"
            :title="t('admin.ops.charts.resetZoomHint')"
            @click="resetZoom"
          >
            {{ t('admin.ops.charts.resetZoom') }}
          </button>
          <button
            type="button"
            class="views-admin-ops-components-ops-throughput-trend-chart__action"
            :disabled="state !== 'ready'"
            :title="t('admin.ops.charts.downloadChartHint')"
            @click="downloadChart"
          >
            {{ t('admin.ops.charts.downloadChart') }}
          </button>
        </template>
      </div>
    </div>

    <!-- Drilldown chips (baseline interaction: click to set global filter) -->
    <div v-if="(props.topGroups?.length ?? 0) > 0" class="views-admin-ops-components-ops-throughput-trend-chart__panel-4">
      <button
        v-for="g in props.topGroups"
        :key="g.group_id"
        type="button"
        class="views-admin-ops-components-ops-throughput-trend-chart__action-2"
        @click="emit('selectGroup', g.group_id)"
      >
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text-4">{{ g.group_name || `#${g.group_id}` }}</span>
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text-5">{{ formatNumber(g.request_count) }}</span>
      </button>
    </div>

    <div v-else-if="(props.byPlatform?.length ?? 0) > 0" class="views-admin-ops-components-ops-throughput-trend-chart__panel-4">
      <button
        v-for="p in props.byPlatform"
        :key="p.platform"
        type="button"
        class="views-admin-ops-components-ops-throughput-trend-chart__action-2"
        @click="emit('selectPlatform', p.platform)"
      >
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text-6">{{ p.platform }}</span>
        <span class="views-admin-ops-components-ops-throughput-trend-chart__text-5">{{ formatNumber(p.request_count) }}</span>
      </button>
    </div>

    <div class="views-admin-ops-components-ops-throughput-trend-chart__panel-5">
      <D3LineChart v-if="state === 'ready' && chartData" ref="throughputChartRef" :data="chartData" :options="options" />
      <div v-else class="views-admin-ops-components-ops-throughput-trend-chart__panel-6">
        <div v-if="state === 'loading'" class="views-admin-ops-components-ops-throughput-trend-chart__panel-7">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyRequest')" />
      </div>
    </div>
  </div>
</template>
