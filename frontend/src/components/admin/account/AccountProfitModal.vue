<template>
  <BaseDialog :show="show" :title="t('admin.accounts.profit.title')" width="full" @close="emit('close')">
    <div v-if="account" class="account-profit-modal">
      <div class="account-profit-modal__header">
        <div>
          <h4>{{ account.name }}</h4>
          <span>{{ account.id }} · {{ platformLabel(account.platform) }} · {{ account.subscription_tier || t('admin.accounts.profit.noTier') }}</span>
          <small>{{ t('admin.accounts.profit.accountCreated') }}: {{ formatDate(account.created_at) }} · {{ t('admin.accounts.profit.accountExpires') }}: {{ formatExpiry(account.expires_at) }}</small>
        </div>
        <span class="account-profit-modal__unit">{{ t('admin.accounts.profit.unitHint') }}</span>
      </div>

      <LoadingState v-if="loading && !stats" variant="section" />
      <div v-else-if="error" class="account-profit-modal__error">{{ error }}</div>
      <template v-else-if="stats">
        <div class="account-profit-modal__periods">
          <section v-for="period in periods" :key="period.key" :class="['account-profit-modal__period', `is-${period.key}`]">
            <div class="account-profit-modal__period-title">{{ period.label }}</div>
            <div class="account-profit-modal__metrics">
              <div><span>{{ t('admin.accounts.profit.revenuePoints') }}</span><strong class="is-revenue">{{ formatPoints(period.value.revenue_points) }}</strong></div>
              <div><span>{{ t('admin.accounts.profit.upstreamCost') }}</span><strong class="is-cost">{{ formatUSD(period.value.cost_usd) }}</strong></div>
              <div><span>{{ t('admin.accounts.profit.profit') }}</span><strong class="is-profit">{{ formatPoints(period.value.profit_points) }}</strong></div>
              <div><span>{{ t('admin.accounts.profit.tokens') }}</span><strong class="is-tokens">{{ formatNumber(period.value.tokens) }}</strong></div>
              <div><span>{{ t('admin.accounts.profit.requests') }}</span><strong>{{ formatNumber(period.value.requests) }}</strong></div>
            </div>
          </section>
        </div>

        <div class="account-profit-modal__history-head">
          <div>
            <h4>{{ t('admin.accounts.profit.dailyChart') }}</h4>
            <span>{{ t('admin.accounts.profit.dailyChartHint') }}</span>
          </div>
          <div class="account-profit-modal__range-tabs">
            <button
              v-for="option in rangeOptions"
              :key="option.value"
              type="button"
              :class="['account-profit-modal__range-tab', { 'is-active': selectedRange === option.value }]"
              :disabled="loading"
              @click="selectedRange = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </div>
        <div v-if="selectedRange === 'custom'" class="account-profit-modal__custom-range">
          <label>
            <span>{{ t('admin.accounts.profit.rangeFrom') }}</span>
            <input v-model="customFrom" type="date" :max="customTo || undefined" />
          </label>
          <label>
            <span>{{ t('admin.accounts.profit.rangeTo') }}</span>
            <input v-model="customTo" type="date" :min="customFrom || undefined" />
          </label>
          <span v-if="customFrom && customTo && !customRangeValid" class="account-profit-modal__range-error">
            {{ t('admin.accounts.profit.rangeInvalid') }}
          </span>
        </div>

        <div class="account-profit-modal__charts">
          <section class="account-profit-modal__chart-panel">
            <h5>{{ t('admin.accounts.profit.moneyTrend') }}</h5>
            <div v-if="!chartRows.length" class="account-profit-modal__chart-empty">{{ t('admin.accounts.profit.noHistory') }}</div>
            <D3LineChart v-else :data="moneyChartData" :options="moneyChartOptions" :aria-label="t('admin.accounts.profit.moneyTrend')" />
          </section>
          <section class="account-profit-modal__chart-panel">
            <h5>{{ t('admin.accounts.profit.usageTrend') }}</h5>
            <div v-if="!chartRows.length" class="account-profit-modal__chart-empty">{{ t('admin.accounts.profit.noHistory') }}</div>
            <D3LineChart v-else :data="usageChartData" :options="usageChartOptions" :aria-label="t('admin.accounts.profit.usageTrend')" />
          </section>
        </div>

        <div class="account-profit-modal__history-head">
          <div>
            <h4>{{ t('admin.accounts.profit.history') }}</h4>
            <span>{{ t('admin.accounts.profit.historyHint') }}</span>
          </div>
          <span>{{ t('admin.accounts.profit.days', { count: stats.total }) }}</span>
        </div>
        <div v-if="!historyRows.length" class="account-profit-modal__empty">{{ t('admin.accounts.profit.noHistory') }}</div>
        <div v-else class="account-profit-modal__table-wrap">
          <table class="account-profit-modal__table">
            <thead>
              <tr>
                <th>{{ t('admin.accounts.profit.date') }}</th>
                <th>{{ t('admin.accounts.profit.revenuePoints') }}</th>
                <th>{{ t('admin.accounts.profit.upstreamCost') }}</th>
                <th>{{ t('admin.accounts.profit.profit') }}</th>
                <th>{{ t('admin.accounts.profit.tokens') }}</th>
                <th>{{ t('admin.accounts.profit.requests') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in historyRows" :key="row.date">
                <td>{{ row.date }}</td>
                <td class="is-revenue">{{ formatPoints(row.revenue_points) }}</td>
                <td class="is-cost">{{ formatUSD(row.cost_usd) }}</td>
                <td class="is-profit">{{ formatPoints(row.profit_points) }}</td>
                <td class="is-tokens">{{ formatNumber(row.tokens) }}</td>
                <td>{{ formatNumber(row.requests) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <button v-if="stats.has_more" type="button" class="account-profit-modal__load-more" :disabled="loading" @click="loadMore">
          {{ loading ? t('common.loading') : t('admin.accounts.profit.loadMore') }}
        </button>
      </template>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import { adminAPI } from '@/api/admin'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime, formatNumber, formatPoints, formatUSD } from '@/utils/format'
import { useThemeColors } from '@/composables/useThemeColors'
import type { AccountPlatform, AccountProfitDailyRecord, AccountProfitListItem, AccountProfitPeriod, AccountProfitResponse } from '@/types'
import type { D3ChartData, D3ChartOptions, D3LineTooltipItem } from '@/components/charts/d3/chartTypes'

type ProfitRange = '7d' | '30d' | '90d' | 'all' | 'custom'

const props = defineProps<{ show: boolean; account: AccountProfitListItem | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const themeColors = useThemeColors()
const loading = ref(false)
const error = ref('')
const stats = ref<AccountProfitResponse | null>(null)
const historyRows = ref<AccountProfitDailyRecord[]>([])
const historyPage = ref(1)
const selectedRange = ref<ProfitRange>('30d')
const customFrom = ref('')
const customTo = ref('')

const rangeOptions = computed(() => [
  { value: '7d' as const, label: t('admin.accounts.profit.range7d') },
  { value: '30d' as const, label: t('admin.accounts.profit.range30d') },
  { value: '90d' as const, label: t('admin.accounts.profit.range90d') },
  { value: 'all' as const, label: t('admin.accounts.profit.rangeAll') },
  { value: 'custom' as const, label: t('admin.accounts.profit.rangeCustom') }
])

const customRangeValid = computed(() => Boolean(customFrom.value && customTo.value && customFrom.value <= customTo.value))

const periods = computed<Array<{ key: string; label: string; value: AccountProfitPeriod }>>(() => {
  const current = stats.value
  if (!current) return []
  const result: Array<{ key: string; label: string; value: AccountProfitPeriod }> = [
    { key: 'period_7d', label: t('admin.accounts.profit.period7d'), value: current.period_7d },
    { key: 'lifetime', label: t('admin.accounts.profit.lifetime'), value: current.lifetime }
  ]
  if (current.expiry_30d) {
    result.splice(1, 0, {
      key: 'expiry_30d',
      label: t('admin.accounts.profit.expiry30d'),
      value: current.expiry_30d
    })
  }
  return result
})

const chartRows = computed(() => [...historyRows.value].sort((a, b) => a.date.localeCompare(b.date)))
const moneyChartData = computed<D3ChartData>(() => ({
  labels: chartRows.value.map(row => row.date),
  datasets: [
    { label: t('admin.accounts.profit.revenuePoints'), data: chartRows.value.map(row => row.revenue_points), borderColor: '#218739', backgroundColor: '#21873918', fill: true, tension: 0.25, pointRadius: 2 },
    { label: t('admin.accounts.profit.upstreamCost'), data: chartRows.value.map(row => row.cost_usd), borderColor: '#b45309', backgroundColor: '#b4530912', fill: false, tension: 0.25, pointRadius: 2 },
    { label: t('admin.accounts.profit.profit'), data: chartRows.value.map(row => row.profit_points), borderColor: '#2563eb', backgroundColor: '#2563eb12', fill: false, tension: 0.25, pointRadius: 2 }
  ]
}))
const usageChartData = computed<D3ChartData>(() => ({
  labels: chartRows.value.map(row => row.date),
  datasets: [
    { label: t('admin.accounts.profit.tokens'), data: chartRows.value.map(row => row.tokens), borderColor: '#7c3aed', backgroundColor: '#7c3aed12', fill: true, tension: 0.25, pointRadius: 2 },
    { label: t('admin.accounts.profit.requests'), data: chartRows.value.map(row => row.requests), borderColor: '#0891b2', backgroundColor: '#0891b212', fill: false, tension: 0.25, pointRadius: 2, yAxisID: 'count' }
  ]
}))

const baseChartOptions = computed<D3ChartOptions>(() => ({
  scales: {
    x: { grid: { color: themeColors.value.grid }, ticks: { color: themeColors.value.textTertiary, maxTicksLimit: 10 } },
    y: { grid: { color: themeColors.value.grid }, ticks: { color: themeColors.value.textTertiary, callback: (value: string | number) => formatNumber(Number(value)) } }
  },
  plugins: {
    legend: { position: 'top', labels: { color: themeColors.value.textTertiary, font: { size: 11 }, padding: 12 } },
    tooltip: {
      callbacks: {
        title: (items: D3LineTooltipItem[]) => items[0]?.label || '',
        label: (item: D3LineTooltipItem) => {
          const row = chartRows.value[item.dataIndex]
          if (!row) return `${item.dataset.label}: ${formatNumber(Number(item.parsed.y ?? 0))}`
          if (item.dataset.label === t('admin.accounts.profit.upstreamCost')) return `${item.dataset.label}: ${formatUSD(row.cost_usd)}`
          if (item.dataset.label === t('admin.accounts.profit.revenuePoints')) return `${item.dataset.label}: ${formatPoints(row.revenue_points)}`
          if (item.dataset.label === t('admin.accounts.profit.profit')) return `${item.dataset.label}: ${formatPoints(row.profit_points)}`
          if (item.dataset.label === t('admin.accounts.profit.tokens')) return `${item.dataset.label}: ${formatNumber(row.tokens)}`
          return `${item.dataset.label}: ${formatNumber(row.requests)}`
        },
        footer: (items: D3LineTooltipItem[]) => {
          const row = chartRows.value[items[0]?.dataIndex ?? -1]
          return row ? `${t('admin.accounts.profit.tokens')}: ${formatNumber(row.tokens)} | ${t('admin.accounts.profit.requests')}: ${formatNumber(row.requests)}` : ''
        }
      }
    }
  }
}))
const moneyChartOptions = computed(() => baseChartOptions.value)
const usageChartOptions = computed<D3ChartOptions>(() => {
  const baseOptions = baseChartOptions.value
  const baseScales = (baseOptions.scales as Record<string, unknown> | undefined) ?? {}
  return {
    ...baseOptions,
    scales: {
      ...baseScales,
      count: { position: 'right', grid: { drawOnChartArea: false }, ticks: { color: '#0891b2', callback: (value: string | number) => formatNumber(Number(value)) } }
    }
  }
})

function dateOnly(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function rangeFrom(): string | undefined {
  if (selectedRange.value === 'all') {
    return props.account ? dateOnly(new Date(props.account.created_at)) : undefined
  }
  if (selectedRange.value === 'custom') return customFrom.value || undefined
  const days = selectedRange.value === '7d' ? 7 : selectedRange.value === '90d' ? 90 : 30
  const date = new Date()
  date.setDate(date.getDate() - days + 1)
  return dateOnly(date)
}

function rangeTo(): string | undefined {
  return selectedRange.value === 'custom' ? customTo.value || undefined : undefined
}

async function load(page = 1) {
  if (!props.account) return
  loading.value = true
  error.value = ''
  try {
    const response = await adminAPI.accounts.getProfit(props.account.id, {
      from: rangeFrom(),
      to: rangeTo(),
      page,
      page_size: 100
    })
    stats.value = response
    historyRows.value = page === 1 ? response.history : [...historyRows.value, ...response.history]
    historyPage.value = page
  } catch (err) {
    error.value = extractApiErrorMessage(err, t('admin.accounts.profit.loadFailed'))
  } finally {
    loading.value = false
  }
}

function loadMore() {
  void load(historyPage.value + 1)
}

function platformLabel(value: AccountPlatform) {
  return t(`admin.accounts.platforms.${value}`)
}

function formatDate(value: string) {
  return formatDateTime(value, { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function formatExpiry(timestamp?: number | null) {
  return timestamp ? formatDateTime(new Date(timestamp * 1000), { year: 'numeric', month: '2-digit', day: '2-digit' }) : t('admin.accounts.profit.noExpiry')
}

watch([() => props.show, () => props.account?.id, selectedRange], ([visible]) => {
  if (!visible) return
  if (selectedRange.value === 'custom' && !customRangeValid.value) {
    stats.value = null
    historyRows.value = []
    error.value = ''
    return
  }
  void load(1)
}, { immediate: true })

watch([customFrom, customTo], () => {
  if (props.show && selectedRange.value === 'custom' && customRangeValid.value) void load(1)
})
</script>

<style scoped>
.account-profit-modal {
  display: grid;
  gap: 1rem;
  min-width: 0;
  color: var(--color-text-secondary);
}

.account-profit-modal__header,
.account-profit-modal__history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.account-profit-modal__header {
  padding-bottom: 0.875rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.account-profit-modal__header > div:first-child {
  min-width: 0;
}

.account-profit-modal h4 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-md);
  font-weight: 650;
  line-height: 1.25;
}

.account-profit-modal__header span,
.account-profit-modal__header small,
.account-profit-modal__history-head span,
.account-profit-modal__unit {
  display: block;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  line-height: 1.4;
}

.account-profit-modal__header span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-profit-modal__header small {
  margin-top: 0.25rem;
  color: var(--color-text-quaternary);
  font-size: var(--font-size-2xs);
}

.account-profit-modal__unit {
  max-width: 18rem;
  flex: 0 1 18rem;
  color: var(--color-text-quaternary);
  font-size: var(--font-size-2xs);
  text-align: right;
}

.account-profit-modal__periods {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.account-profit-modal__period {
  position: relative;
  min-width: 0;
  padding: 0.75rem;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.account-profit-modal__period::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 0.1875rem;
  background: var(--period-accent, var(--theme-accent));
}

.account-profit-modal__period.is-period_7d { --period-accent: var(--color-success); }
.account-profit-modal__period.is-expiry_30d { --period-accent: var(--color-warning); }
.account-profit-modal__period.is-lifetime { --period-accent: var(--theme-accent); }

.account-profit-modal__period-title {
  margin-bottom: 0.75rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 650;
}

.account-profit-modal__metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.625rem;
}

.account-profit-modal__metrics div {
  display: grid;
  min-width: 0;
  gap: 0.125rem;
}

.account-profit-modal__metrics span {
  overflow: hidden;
  color: var(--color-text-quaternary);
  font-size: var(--font-size-2xs);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-profit-modal__metrics strong {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-profit-modal__range-tabs {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.125rem;
  padding: 0.1875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-subtle);
}

.account-profit-modal__range-tab {
  min-height: 1.875rem;
  padding: 0.25rem 0.625rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: 600;
  cursor: pointer;
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease;
}

.account-profit-modal__range-tab:hover:not(:disabled) {
  color: var(--color-text-primary);
  background: var(--glass-bg-interactive-hover);
}

.account-profit-modal__range-tab.is-active {
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.account-profit-modal__range-tab:focus-visible,
.account-profit-modal__load-more:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 20%, transparent);
}

.account-profit-modal__range-tab:disabled {
  cursor: wait;
  opacity: 0.65;
}

.account-profit-modal__custom-range {
  display: flex;
  align-items: end;
  flex-wrap: wrap;
  gap: 0.625rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--glass-bg-subtle);
}

.account-profit-modal__custom-range label {
  display: grid;
  gap: 0.25rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: 600;
}

.account-profit-modal__custom-range input {
  min-height: 2rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-sm);
  background: var(--glass-field-bg);
  color: var(--color-text-primary);
  font-size: var(--font-size-xs);
}

.account-profit-modal__custom-range input:focus {
  border-color: var(--color-primary);
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 18%, transparent);
}

.account-profit-modal__range-error {
  color: var(--color-text-danger);
  font-size: var(--font-size-2xs);
}

.account-profit-modal__charts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.account-profit-modal__chart-panel {
  min-width: 0;
  padding: 0.75rem 0.875rem 0.5rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-subtle);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.account-profit-modal__chart-panel h5 {
  margin: 0 0 0.25rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-xs);
  font-weight: 650;
}

.account-profit-modal__chart-panel :deep(.d3-line-chart__frame) {
  min-height: 15rem;
}

.account-profit-modal__chart-empty {
  display: grid;
  min-height: 15rem;
  place-items: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.account-profit-modal__table-wrap {
  max-height: 20rem;
  overflow: auto;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  scrollbar-width: thin;
}

.account-profit-modal__table {
  width: 100%;
  border-collapse: collapse;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-variant-numeric: tabular-nums;
}

.account-profit-modal__table th,
.account-profit-modal__table td {
  padding: 0.5625rem 0.625rem;
  border-bottom: 1px solid var(--color-border-subtle);
  text-align: right;
  white-space: nowrap;
}

.account-profit-modal__table th:first-child,
.account-profit-modal__table td:first-child {
  text-align: left;
}

.account-profit-modal__table th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--color-surface-elevated);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: 650;
}

.account-profit-modal__table tbody tr:hover td {
  background: color-mix(in srgb, var(--theme-accent) 5%, transparent);
}

.account-profit-modal__table tr:last-child td {
  border-bottom: 0;
}

.account-profit-modal__load-more {
  justify-self: center;
  min-height: 2rem;
  padding: 0.25rem 0.875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-interactive);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-weight: 600;
  cursor: pointer;
}

.account-profit-modal__load-more:hover:not(:disabled) {
  border-color: var(--glass-border-hover);
  background: var(--glass-bg-interactive-hover);
  color: var(--color-text-primary);
}

.account-profit-modal__load-more:disabled {
  cursor: wait;
  opacity: 0.65;
}

.account-profit-modal__empty,
.account-profit-modal__error {
  padding: 2rem 1rem;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-subtle);
  color: var(--color-text-muted);
  text-align: center;
  font-size: var(--font-size-xs);
}

.account-profit-modal__error {
  border-color: var(--color-danger-border);
  background: var(--glass-tint-danger);
  color: var(--color-text-danger);
}

.is-revenue { color: var(--color-text-success) !important; }
.is-cost { color: var(--color-text-warning) !important; }
.is-profit { color: var(--color-text-brand) !important; }
.is-tokens { color: #7c3aed !important; }

@media (max-width: 960px) {
  .account-profit-modal__periods,
  .account-profit-modal__charts {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .account-profit-modal__header,
  .account-profit-modal__history-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .account-profit-modal__unit {
    max-width: none;
    text-align: left;
  }

  .account-profit-modal__range-tabs {
    width: 100%;
  }

  .account-profit-modal__range-tab {
    flex: 1 1 auto;
  }
}

@media (max-width: 480px) {
  .account-profit-modal__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
