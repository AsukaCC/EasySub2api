<template>
  <div class="components-charts-model-distribution-chart__panel card">
    <div class="components-charts-model-distribution-chart__panel-2">
      <h3 class="components-charts-model-distribution-chart__heading">
        {{ !enableRankingView || activeView === 'model_distribution'
          ? t('admin.dashboard.modelDistribution')
          : t('admin.dashboard.spendingRankingTitle') }}
      </h3>
      <div class="components-charts-model-distribution-chart__panel-3">
        <div
          v-if="showSourceToggle"
          class="components-charts-model-distribution-chart__panel-4"
        >
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="source === 'requested'
              ? 'components-charts-model-distribution-chart__action-2'
              : 'components-charts-model-distribution-chart__action-3'"
            @click="emit('update:source', 'requested')"
          >
            {{ t('usage.requestedModel') }}
          </button>
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="source === 'upstream'
              ? 'components-charts-model-distribution-chart__action-2'
              : 'components-charts-model-distribution-chart__action-3'"
            @click="emit('update:source', 'upstream')"
          >
            {{ t('usage.upstreamModel') }}
          </button>
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="source === 'mapping'
              ? 'components-charts-model-distribution-chart__action-2'
              : 'components-charts-model-distribution-chart__action-3'"
            @click="emit('update:source', 'mapping')"
          >
            {{ t('usage.mapping') }}
          </button>
        </div>
        <div
          v-if="showMetricToggle"
          class="components-charts-model-distribution-chart__panel-4"
        >
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="metric === 'tokens'
              ? 'components-charts-model-distribution-chart__action-2'
              : 'components-charts-model-distribution-chart__action-3'"
            @click="emit('update:metric', 'tokens')"
          >
            {{ t('admin.dashboard.metricTokens') }}
          </button>
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="metric === 'actual_cost'
              ? 'components-charts-model-distribution-chart__action-2'
              : 'components-charts-model-distribution-chart__action-3'"
            @click="emit('update:metric', 'actual_cost')"
          >
            {{ t('admin.dashboard.metricActualCost') }}
          </button>
        </div>
        <div v-if="enableRankingView" class="components-charts-model-distribution-chart__panel-5">
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="
              activeView === 'model_distribution'
                ? 'components-charts-model-distribution-chart__action-2'
                : 'components-charts-model-distribution-chart__action-3'
            "
            @click="activeView = 'model_distribution'"
          >
            {{ t('admin.dashboard.viewModelDistribution') }}
          </button>
          <button
            type="button"
            class="components-charts-model-distribution-chart__action"
            :class="
              activeView === 'spending_ranking'
                ? 'components-charts-model-distribution-chart__action-2'
                : 'components-charts-model-distribution-chart__action-3'
            "
            @click="activeView = 'spending_ranking'"
          >
            {{ t('admin.dashboard.viewSpendingRanking') }}
          </button>
        </div>
      </div>
    </div>

    <LoadingState v-if="activeView === 'model_distribution' && loading" variant="section" class="components-charts-model-distribution-chart__panel-6" />
    <div
      v-else-if="activeView === 'model_distribution' && displayModelStats.length > 0 && chartData"
      class="components-charts-model-distribution-chart__panel-7"
    >
      <div class="components-charts-model-distribution-chart__panel-8">
        <D3DonutChart :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="components-charts-model-distribution-chart__panel-9">
        <table class="components-charts-model-distribution-chart__table">
          <thead>
            <tr class="components-charts-model-distribution-chart__row">
              <th class="components-charts-model-distribution-chart__heading-2">{{ t('admin.dashboard.model') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.requests') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.tokens') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.actual') }}</th>
              <th v-if="showAccountCost" class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.accountCost') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="model in displayModelStats" :key="model.model">
              <tr
                class="components-charts-model-distribution-chart__row-2"
                :class="enableBreakdown ? 'components-charts-model-distribution-chart__row-3' : ''"
                @click="enableBreakdown && toggleBreakdown('model', model.model)"
              >
                <td
                  class="components-charts-model-distribution-chart__cell"
                  :class="enableBreakdown ? 'components-charts-model-distribution-chart__cell-8' : 'components-charts-model-distribution-chart__cell-9'"
                  :title="model.model"
                >
                  <span class="components-charts-model-distribution-chart__text">
                    <svg v-if="enableBreakdown && expandedKey === `model-${model.model}`" class="components-charts-model-distribution-chart__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else-if="enableBreakdown" class="components-charts-model-distribution-chart__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ model.model }}
                  </span>
                </td>
                <td class="components-charts-model-distribution-chart__cell-2">
                  {{ formatNumber(model.requests) }}
                </td>
                <td class="components-charts-model-distribution-chart__cell-2">
                  {{ formatTokens(model.total_tokens) }}
                </td>
                <td class="components-charts-model-distribution-chart__cell-3">
                  {{ formatPoints(model.actual_cost) }}
                </td>
                <td v-if="showAccountCost" class="components-charts-model-distribution-chart__cell-4">
                  ${{ formatCost(model.account_cost) }}
                </td>
                <td class="components-charts-model-distribution-chart__cell-5">
                  ${{ formatCost(model.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === `model-${model.model}`">
                <td :colspan="distributionColspan" class="components-charts-model-distribution-chart__cell-6">
                  <UserBreakdownSubTable
                    :items="breakdownItems"
                    :loading="breakdownLoading"
                    :show-account-cost="showAccountCost"
                  />
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else-if="activeView === 'model_distribution'"
      class="components-charts-model-distribution-chart__panel-10"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>

    <div v-else-if="rankingLoading" class="components-charts-model-distribution-chart__panel-6">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="rankingError"
      class="components-charts-model-distribution-chart__panel-10"
    >
      {{ t('admin.dashboard.failedToLoad') }}
    </div>
    <div v-else-if="rankingDisplayItems.length > 0 && rankingChartData" class="components-charts-model-distribution-chart__panel-7">
      <div class="components-charts-model-distribution-chart__panel-8">
        <D3DonutChart :data="rankingChartData" :options="rankingDoughnutOptions" />
      </div>
      <div class="components-charts-model-distribution-chart__panel-9">
        <table class="components-charts-model-distribution-chart__table">
          <thead>
            <tr class="components-charts-model-distribution-chart__row">
              <th class="components-charts-model-distribution-chart__heading-2">{{ t('admin.dashboard.spendingRankingUser') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
              <th class="components-charts-model-distribution-chart__heading-3">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in rankingDisplayItems"
              :key="item.isOther ? 'others' : `${item.user_id}-${index}`"
              class="components-charts-model-distribution-chart__row-2"
              :class="item.isOther
                ? 'components-charts-model-distribution-chart__row-4'
                : 'components-charts-model-distribution-chart__row-3'"
              @click="item.isOther ? undefined : emit('ranking-click', item)"
            >
              <td class="components-charts-model-distribution-chart__cell-7">
                <div class="components-charts-model-distribution-chart__panel-11">
                  <span class="components-charts-model-distribution-chart__text-2">
                    {{ item.isOther ? 'Σ' : `#${index + 1}` }}
                  </span>
                  <span
                    class="components-charts-model-distribution-chart__text-3"
                    :title="getRankingRowLabel(item)"
                  >
                    {{ getRankingRowLabel(item) }}
                  </span>
                </div>
              </td>
              <td class="components-charts-model-distribution-chart__cell-2">
                {{ formatNumber(item.requests) }}
              </td>
              <td class="components-charts-model-distribution-chart__cell-2">
                {{ formatTokens(item.tokens) }}
              </td>
              <td class="components-charts-model-distribution-chart__cell-3">
                {{ formatPoints(item.actual_cost) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else
      class="components-charts-model-distribution-chart__panel-10"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import D3DonutChart from '@/components/charts/d3/D3DonutChart.vue'
import UserBreakdownSubTable from './UserBreakdownSubTable.vue'
import type { ModelStat, UserSpendingRankingItem, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import { formatPoints } from '@/utils/format'

const { t } = useI18n()

type DistributionMetric = 'tokens' | 'actual_cost'
type ModelSource = 'requested' | 'upstream' | 'mapping'
type RankingDisplayItem = UserSpendingRankingItem & { isOther?: boolean }
const props = withDefaults(defineProps<{
  modelStats: ModelStat[]
  upstreamModelStats?: ModelStat[]
  mappingModelStats?: ModelStat[]
  source?: ModelSource
  enableRankingView?: boolean
  rankingItems?: UserSpendingRankingItem[]
  rankingTotalActualCost?: number
  rankingTotalRequests?: number
  rankingTotalTokens?: number
  loading?: boolean
  metric?: DistributionMetric
  showSourceToggle?: boolean
  showMetricToggle?: boolean
  enableBreakdown?: boolean
  showAccountCost?: boolean
  rankingLoading?: boolean
  rankingError?: boolean
  startDate?: string
  endDate?: string
  filters?: Record<string, any>
}>(), {
  upstreamModelStats: () => [],
  mappingModelStats: () => [],
  source: 'requested',
  enableRankingView: false,
  rankingItems: () => [],
  rankingTotalActualCost: 0,
  rankingTotalRequests: 0,
  rankingTotalTokens: 0,
  loading: false,
  metric: 'tokens',
  showSourceToggle: false,
  showMetricToggle: false,
  enableBreakdown: true,
  showAccountCost: true,
  rankingLoading: false,
  rankingError: false
})

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)

const toggleBreakdown = async (type: string, id: string) => {
  const key = `${type}-${id}`
  if (expandedKey.value === key) {
    expandedKey.value = null
    return
  }
  expandedKey.value = key
  breakdownLoading.value = true
  breakdownItems.value = []
  try {
    const res = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      model: id,
      model_source: props.source,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
  'update:source': [value: ModelSource]
  'ranking-click': [item: UserSpendingRankingItem]
}>()

const enableRankingView = computed(() => props.enableRankingView)
const showAccountCost = computed(() => props.showAccountCost)
const distributionColspan = computed(() => showAccountCost.value ? 6 : 5)
const activeView = ref<'model_distribution' | 'spending_ranking'>('model_distribution')

const chartColors = [
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
  '#f97316',
  '#6366f1',
  '#84cc16',
  '#06b6d4',
  '#a855f7'
]

const displayModelStats = computed(() => {
  const sourceStats = props.source === 'upstream'
    ? props.upstreamModelStats
    : props.source === 'mapping'
      ? props.mappingModelStats
      : props.modelStats
  if (!sourceStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...sourceStats].sort((a, b) => toFiniteNumber(b[metricKey]) - toFiniteNumber(a[metricKey]))
})

const chartData = computed(() => {
  if (!displayModelStats.value.length) return null

  return {
    labels: displayModelStats.value.map((m) => m.model),
    datasets: [
      {
        data: displayModelStats.value.map((m) => toFiniteNumber(props.metric === 'actual_cost' ? m.actual_cost : m.total_tokens)),
        backgroundColor: chartColors.slice(0, displayModelStats.value.length),
        borderWidth: 0
      }
    ]
  }
})

const rankingChartData = computed(() => {
  if (!props.rankingItems?.length) return null

  const labels = props.rankingItems.map((item, index) => `#${index + 1} ${getRankingUserLabel(item)}`)
  const data = props.rankingItems.map((item) => toFiniteNumber(item.actual_cost))
  const backgroundColor = chartColors.slice(0, props.rankingItems.length)

  if (otherRankingItem.value) {
    labels.push(t('admin.dashboard.spendingRankingOther'))
    data.push(otherRankingItem.value.actual_cost)
    backgroundColor.push('#94a3b8')
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor,
        borderWidth: 0
      }
    ]
  }
})

const otherRankingItem = computed<RankingDisplayItem | null>(() => {
  if (!props.rankingItems?.length) return null

  const rankedActualCost = props.rankingItems.reduce((sum, item) => sum + toFiniteNumber(item.actual_cost), 0)
  const rankedRequests = props.rankingItems.reduce((sum, item) => sum + toFiniteNumber(item.requests), 0)
  const rankedTokens = props.rankingItems.reduce((sum, item) => sum + toFiniteNumber(item.tokens), 0)

  const otherActualCost = Math.max((props.rankingTotalActualCost || 0) - rankedActualCost, 0)
  const otherRequests = Math.max((props.rankingTotalRequests || 0) - rankedRequests, 0)
  const otherTokens = Math.max((props.rankingTotalTokens || 0) - rankedTokens, 0)

  if (otherActualCost <= 0.000001 && otherRequests <= 0 && otherTokens <= 0) return null

  return {
    user_id: '',
    email: '',
    username: '',
    actual_cost: otherActualCost,
    requests: otherRequests,
    tokens: otherTokens,
    isOther: true
  }
})

const rankingDisplayItems = computed<RankingDisplayItem[]>(() => {
  if (!props.rankingItems?.length) return []
  return otherRankingItem.value
    ? [...props.rankingItems, otherRankingItem.value]
    : [...props.rankingItems]
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          const formattedValue = props.metric === 'actual_cost'
            ? formatPoints(value)
            : formatTokens(value)
          return `${context.label}: ${formattedValue} (${percentage}%)`
        }
      }
    }
  }
}))

const rankingDoughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          return `${context.label}: ${formatPoints(value)} (${percentage}%)`
        }
      }
    }
  }
}))

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatNumber = (value: number): string => {
  return toFiniteNumber(value).toLocaleString()
}

const getRankingUserLabel = (item: UserSpendingRankingItem): string => {
  if (item.username?.trim()) return item.username.trim()
  if (item.email?.trim()) return item.email.trim()
  return t('admin.redeem.userPrefix', { id: item.user_id })
}

const getRankingRowLabel = (item: RankingDisplayItem): string => {
  if (item.isOther) return t('admin.dashboard.spendingRankingOther')
  return getRankingUserLabel(item)
}

const toFiniteNumber = (value: unknown): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
}

const formatCost = (value: number | null | undefined): string => {
  const safeValue = toFiniteNumber(value)
  if (safeValue >= 1000) {
    return (safeValue / 1000).toFixed(2) + 'K'
  } else if (safeValue >= 1) {
    return safeValue.toFixed(2)
  } else if (safeValue >= 0.01) {
    return safeValue.toFixed(3)
  }
  return safeValue.toFixed(4)
}
</script>
