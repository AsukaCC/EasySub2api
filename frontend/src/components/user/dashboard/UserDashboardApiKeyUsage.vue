<template>
  <section class="dashboard-key-usage card" :aria-label="t('dashboard.apiKeyUsage.title')">
    <header class="dashboard-key-usage__header">
      <div>
        <h2 class="dashboard-key-usage__title">{{ t('dashboard.apiKeyUsage.title') }}</h2>
        <p class="dashboard-key-usage__description">{{ t('dashboard.apiKeyUsage.description') }}</p>
      </div>
      <router-link to="/keys" class="dashboard-key-usage__link">
        {{ t('dashboard.apiKeyUsage.manage') }}
        <Icon name="arrowRight" size="sm" />
      </router-link>
    </header>

    <div v-if="loading" class="dashboard-key-usage__state">
      <LoadingSpinner size="lg" />
    </div>
    <div v-else-if="keyCount === 0" class="dashboard-key-usage__state dashboard-key-usage__state--empty">
      <Icon name="key" size="xl" />
      <div>
        <p class="dashboard-key-usage__empty-title">{{ t('dashboard.apiKeyUsage.emptyTitle') }}</p>
        <p class="dashboard-key-usage__description">{{ t('dashboard.apiKeyUsage.emptyDescription') }}</p>
      </div>
    </div>
    <template v-else>
      <div class="dashboard-key-usage__totals">
        <div class="dashboard-key-usage__total">
          <span>{{ t('dashboard.apiKeyUsage.keys') }}</span>
          <strong>{{ formatNumber(keyCount) }}</strong>
        </div>
        <div class="dashboard-key-usage__total">
          <span>{{ t('dashboard.apiKeyUsage.requests') }}</span>
          <strong>{{ formatNumber(totals.requests) }}</strong>
        </div>
        <div class="dashboard-key-usage__total">
          <span>{{ t('dashboard.apiKeyUsage.tokens') }}</span>
          <strong>{{ formatTokens(totals.tokens) }}</strong>
        </div>
        <div class="dashboard-key-usage__total">
          <span>{{ t('dashboard.apiKeyUsage.cost') }}</span>
          <strong>{{ formatPoints(totals.cost) }}</strong>
        </div>
      </div>

      <div v-if="apiKeys.length > 0" class="dashboard-key-usage__key-filter">
        <span class="dashboard-key-usage__key-filter-label">
          {{ t('dashboard.apiKeyUsage.keyFilter') }}
        </span>
        <Select
          :model-value="selectedApiKeyId"
          :options="keyFilterOptions"
          :placeholder="t('dashboard.apiKeyUsage.keyPlaceholder')"
          searchable="auto"
          class="dashboard-key-usage__key-select"
          :aria-label="t('dashboard.apiKeyUsage.keyFilter')"
          @update:model-value="onKeyFilterChange"
        />
      </div>

      <div class="dashboard-key-usage__chart-toolbar">
        <p v-if="!hasUsage" class="dashboard-key-usage__chart-empty">
          {{ t('dashboard.apiKeyUsage.noUsageTitle') }}
        </p>
        <div class="dashboard-key-usage__segments" role="group" :aria-label="t('dashboard.apiKeyUsage.metric')">
          <button
            v-for="option in metricOptions"
            :key="option.value"
            type="button"
            :class="['dashboard-key-usage__segment', { 'dashboard-key-usage__segment--active': selectedMetric === option.value }]"
            :aria-pressed="selectedMetric === option.value"
            @click="selectedMetric = option.value"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="dashboard-key-usage__chart">
        <D3LineChart :data="chartData" :options="chartOptions" />
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import type { D3ChartData, D3ChartOptions, D3LineTooltipItem } from '@/components/charts/d3/chartTypes'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import type { ApiKey, TrendDataPoint } from '@/types'
import { formatPointAmount, formatPoints } from '@/utils/format'

type UsageMetric = 'requests' | 'tokens' | 'cost'

const props = defineProps<{
  trend: TrendDataPoint[]
  keyCount: number
  loading: boolean
  apiKeys?: ApiKey[]
  selectedApiKeyId?: string | null
}>()

const emit = defineEmits<{
  (event: 'select-key', apiKeyId: string | null): void
}>()

const { t } = useI18n()
const selectedMetric = ref<UsageMetric>('requests')

const apiKeys = computed(() => props.apiKeys ?? [])
const selectedApiKeyId = computed(() => props.selectedApiKeyId ?? null)
const keyFilterOptions = computed(() => [
  { value: null, label: t('dashboard.apiKeyUsage.allKeys') },
  ...apiKeys.value.map((apiKey) => ({
    value: apiKey.id,
    label: apiKey.name?.trim() || `#${apiKey.id}`,
  })),
])

function onKeyFilterChange(value: string | number | boolean | null) {
  emit('select-key', typeof value === 'string' && value ? value : null)
}

const normalizedTrend = computed<TrendDataPoint[]>(() => {
  const pointsByDate = new Map(props.trend.map((point) => [point.date.slice(0, 10), point]))
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(today)
    date.setDate(today.getDate() - 6 + index)
    const dateKey = formatLocalDate(date)
    return pointsByDate.get(dateKey) ?? {
      date: dateKey,
      requests: 0,
      input_tokens: 0,
      output_tokens: 0,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      total_tokens: 0,
      cost: 0,
      actual_cost: 0,
    }
  })
})

const totals = computed(() => normalizedTrend.value.reduce(
  (result, point) => ({
    requests: result.requests + point.requests,
    tokens: result.tokens + point.total_tokens,
    cost: result.cost + point.actual_cost,
  }),
  { requests: 0, tokens: 0, cost: 0 },
))

const hasUsage = computed(() => (
  totals.value.requests > 0 || totals.value.tokens > 0 || totals.value.cost > 0
))

const metricOptions = computed<Array<{ value: UsageMetric; label: string }>>(() => [
  { value: 'requests', label: t('dashboard.requests') },
  { value: 'tokens', label: t('dashboard.tokens') },
  { value: 'cost', label: t('dashboard.apiKeyUsage.chartCost') },
])

const metricColor = computed(() => ({
  requests: '#2563eb',
  tokens: '#059669',
  cost: '#d97706',
})[selectedMetric.value])

const metricValues = computed(() => normalizedTrend.value.map((point) => {
  if (selectedMetric.value === 'tokens') return point.total_tokens
  if (selectedMetric.value === 'cost') return point.actual_cost
  return point.requests
}))

const metricLabel = computed(() => metricOptions.value.find(
  (option) => option.value === selectedMetric.value,
)?.label ?? '')

const chartData = computed<D3ChartData>(() => ({
  labels: normalizedTrend.value.map((point) => formatDateLabel(point.date)),
  datasets: [{
    label: metricLabel.value,
    data: metricValues.value,
    borderColor: metricColor.value,
    backgroundColor: `${metricColor.value}1f`,
    borderWidth: 2,
    fill: true,
    pointBackgroundColor: metricColor.value,
    pointBorderColor: '#ffffff',
    pointBorderWidth: 2,
    pointRadius: 3,
    pointHoverRadius: 5,
    tension: 0.32,
  }],
}))

const chartOptions = computed<D3ChartOptions>(() => {
  const isDark = document.documentElement.classList.contains('dark')
  const textColor = isDark ? '#cbd5e1' : '#64748b'
  const gridColor = isDark ? 'rgba(148, 163, 184, 0.16)' : 'rgba(148, 163, 184, 0.2)'

  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' },
    plugins: {
      legend: { display: false },
      tooltip: {
        displayColors: false,
        callbacks: {
          label: (context: D3LineTooltipItem) => `${metricLabel.value}: ${formatMetric(Number(context.raw))}`,
        },
      },
    },
    scales: {
      x: {
        border: { display: false },
        grid: { display: false },
        ticks: { color: textColor, maxRotation: 0 },
      },
      y: {
        beginAtZero: true,
        border: { display: false },
        grid: { color: gridColor },
        ticks: {
          color: textColor,
          precision: selectedMetric.value === 'cost' ? undefined : 0,
          callback: (value: string | number) => formatAxisValue(Number(value)),
        },
      },
    },
  }
})

function formatDateLabel(value: string) {
  const parts = value.slice(0, 10).split('-')
  if (parts.length !== 3) return value
  return `${Number(parts[1])}/${Number(parts[2])}`
}

function formatLocalDate(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function formatMetric(value: number) {
  if (selectedMetric.value === 'cost') return formatPoints(value)
  if (selectedMetric.value === 'tokens') return formatTokens(value)
  return formatNumber(value)
}

function formatAxisValue(value: number) {
  if (selectedMetric.value === 'cost') return formatPointAmount(value)
  return formatCompact(value)
}

const formatNumber = (value: number) => value.toLocaleString()
const formatCompact = (value: number) => Intl.NumberFormat(undefined, {
  notation: 'compact',
  maximumFractionDigits: 1,
}).format(value)
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toLocaleString()
}
</script>

<style scoped>
.dashboard-key-usage {
  overflow: hidden;
}

.dashboard-key-usage__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.dashboard-key-usage__title,
.dashboard-key-usage__empty-title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.dashboard-key-usage__description {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-key-usage__link {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.375rem;
  color: var(--color-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.dashboard-key-usage__totals {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-bottom: 1px solid var(--color-border-subtle);
  background: var(--glass-bg-subtle);
}

.dashboard-key-usage__total {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.875rem 1.5rem;
  border-right: 1px solid var(--color-border-subtle);
}

.dashboard-key-usage__total:last-child {
  border-right: 0;
}

.dashboard-key-usage__total span {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-key-usage__total strong {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  text-overflow: ellipsis;
}

.dashboard-key-usage__key-filter {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1.5rem 0;
}

.dashboard-key-usage__key-filter-label {
  flex: 0 0 auto;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.dashboard-key-usage__key-select {
  width: min(100%, 26rem);
  min-width: 12rem;
}

.dashboard-key-usage__chart-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.5rem 0;
}

.dashboard-key-usage__chart-empty {
  margin: 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-key-usage__segments {
  display: inline-grid;
  grid-template-columns: repeat(3, minmax(4.5rem, 1fr));
  padding: 0.1875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-subtle);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
}

.dashboard-key-usage__segment {
  min-height: 2rem;
  padding: 0.35rem 0.75rem;
  border: 1px solid transparent;
  border-radius: calc(var(--radius-md) - 2px);
  color: var(--color-text-tertiary);
  background: transparent;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  cursor: pointer;
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease;

  &:hover:not(.dashboard-key-usage__segment--active) {
    color: var(--color-text-primary);
  }
}

.dashboard-key-usage__segment--active {
  color: var(--color-text-primary);
  border-color: var(--glass-border-hover);
  background-color: var(--glass-bg-thick);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
  box-shadow:
    0 2px 8px rgba(10, 132, 255, 0.12),
    0 1px 0 var(--glass-highlight-hover) inset;

  .dark & {
    border-color: var(--glass-border-active);
    background-color: var(--glass-bg-thick);
    color: var(--color-text-primary);
    box-shadow:
      0 2px 8px rgba(0, 0, 0, 0.35),
      0 1px 0 var(--glass-highlight-hover) inset;
  }
}

.dashboard-key-usage__chart {
  position: relative;
  width: 100%;
  height: 19rem;
  padding: 0.875rem 1.5rem 1.5rem;
}

.dashboard-key-usage__state {
  display: flex;
  min-height: 11rem;
  align-items: center;
  justify-content: center;
  color: var(--color-text-tertiary);
}

.dashboard-key-usage__state--empty {
  gap: 0.875rem;
  padding: 1.25rem;
}

@media (max-width: 767px) {
  .dashboard-key-usage__header {
    align-items: flex-start;
  }

  .dashboard-key-usage__totals {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dashboard-key-usage__total:nth-child(2) {
    border-right: 0;
  }

  .dashboard-key-usage__total:nth-child(-n + 2) {
    border-bottom: 1px solid var(--color-border-subtle);
  }

  .dashboard-key-usage__chart-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .dashboard-key-usage__key-filter {
    align-items: stretch;
    flex-direction: column;
    gap: 0.4rem;
  }

  .dashboard-key-usage__key-select {
    width: 100%;
    min-width: 0;
  }

  .dashboard-key-usage__segments {
    width: 100%;
  }

  .dashboard-key-usage__chart {
    height: 16rem;
    padding-inline: 0.75rem;
  }
}
</style>

<style>
/* 暗色覆盖放在非 scoped 块:Vue scoped 编译器在生产构建中会丢弃
   `:global(.dark) ...` 规则(与 SettingsView 中的处理一致)。 */
.dark .dashboard-key-usage__segments {
  background: rgba(118, 118, 128, 0.24);
}

.dark .dashboard-key-usage__segment--active {
  background: rgba(38, 38, 43, 0.9);
  box-shadow:
    0 1px 4px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}
</style>
