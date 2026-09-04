<template>
  <div class="components-charts-group-distribution-chart__panel card">
    <div class="components-charts-group-distribution-chart__panel-2">
      <h3 class="components-charts-group-distribution-chart__heading">
        {{ t('admin.dashboard.groupDistribution') }}
      </h3>
      <div
        v-if="showMetricToggle"
        class="components-charts-group-distribution-chart__panel-3"
      >
        <button
          type="button"
          class="components-charts-group-distribution-chart__action"
          :class="metric === 'tokens'
            ? 'components-charts-group-distribution-chart__action-2'
            : 'components-charts-group-distribution-chart__action-3'"
          @click="emit('update:metric', 'tokens')"
        >
          {{ t('admin.dashboard.metricTokens') }}
        </button>
        <button
          type="button"
          class="components-charts-group-distribution-chart__action"
          :class="metric === 'actual_cost'
            ? 'components-charts-group-distribution-chart__action-2'
            : 'components-charts-group-distribution-chart__action-3'"
          @click="emit('update:metric', 'actual_cost')"
        >
          {{ t('admin.dashboard.metricActualCost') }}
        </button>
      </div>
    </div>
    <LoadingState v-if="loading" variant="section" class="components-charts-group-distribution-chart__panel-4" />
    <div v-else-if="displayGroupStats.length > 0 && chartData" class="components-charts-group-distribution-chart__panel-5">
      <div class="components-charts-group-distribution-chart__panel-6">
        <D3DonutChart :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="components-charts-group-distribution-chart__panel-7">
        <table class="components-charts-group-distribution-chart__table">
          <thead>
            <tr class="components-charts-group-distribution-chart__row">
              <th class="components-charts-group-distribution-chart__heading-2">{{ t('admin.dashboard.group') }}</th>
              <th class="components-charts-group-distribution-chart__heading-3">{{ t('admin.dashboard.requests') }}</th>
              <th class="components-charts-group-distribution-chart__heading-3">{{ t('admin.dashboard.tokens') }}</th>
              <th class="components-charts-group-distribution-chart__heading-3">{{ t('admin.dashboard.actual') }}</th>
              <th v-if="showAccountCost" class="components-charts-group-distribution-chart__heading-3">{{ t('admin.dashboard.accountCost') }}</th>
              <th class="components-charts-group-distribution-chart__heading-3">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="group in displayGroupStats" :key="group.group_id">
              <tr
                class="components-charts-group-distribution-chart__row-2"
                :class="enableBreakdown && group.group_id ? 'components-charts-group-distribution-chart__row-3' : ''"
                @click="enableBreakdown && group.group_id && toggleBreakdown('group', group.group_id)"
              >
                <td
                  class="components-charts-group-distribution-chart__cell"
                  :class="enableBreakdown && group.group_id ? 'components-charts-group-distribution-chart__cell-7' : 'components-charts-group-distribution-chart__cell-8'"
                  :title="group.group_name || String(group.group_id)"
                >
                  <span class="components-charts-group-distribution-chart__text">
                    <svg v-if="enableBreakdown && group.group_id && expandedKey === `group-${group.group_id}`" class="components-charts-group-distribution-chart__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else-if="enableBreakdown && group.group_id" class="components-charts-group-distribution-chart__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ group.group_name || t('admin.dashboard.noGroup') }}
                  </span>
                </td>
                <td class="components-charts-group-distribution-chart__cell-2">
                  {{ formatNumber(group.requests) }}
                </td>
                <td class="components-charts-group-distribution-chart__cell-2">
                  {{ formatTokens(group.total_tokens) }}
                </td>
                <td class="components-charts-group-distribution-chart__cell-3">
                  {{ formatPoints(group.actual_cost) }}
                </td>
                <td v-if="showAccountCost" class="components-charts-group-distribution-chart__cell-4">
                  ${{ formatCost(group.account_cost) }}
                </td>
                <td class="components-charts-group-distribution-chart__cell-5">
                  ${{ formatCost(group.cost) }}
                </td>
              </tr>
              <!-- User breakdown sub-rows -->
              <tr v-if="expandedKey === `group-${group.group_id}`">
                <td :colspan="distributionColspan" class="components-charts-group-distribution-chart__cell-6">
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
      v-else
      class="components-charts-group-distribution-chart__panel-8"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import D3DonutChart from '@/components/charts/d3/D3DonutChart.vue'
import UserBreakdownSubTable from './UserBreakdownSubTable.vue'
import type { GroupStat, UserBreakdownItem } from '@/types'
import { getUserBreakdown } from '@/api/admin/dashboard'
import { formatPoints } from '@/utils/format'

const { t } = useI18n()

type DistributionMetric = 'tokens' | 'actual_cost'

const props = withDefaults(defineProps<{
  groupStats: GroupStat[]
  loading?: boolean
  metric?: DistributionMetric
  showMetricToggle?: boolean
  enableBreakdown?: boolean
  showAccountCost?: boolean
  startDate?: string
  endDate?: string
  filters?: Record<string, any>
}>(), {
  loading: false,
  metric: 'tokens',
  showMetricToggle: false,
  enableBreakdown: true,
  showAccountCost: true,
})

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
}>()

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)
const showAccountCost = computed(() => props.showAccountCost)
const distributionColspan = computed(() => showAccountCost.value ? 6 : 5)

const toggleBreakdown = async (type: string, id: string | string) => {
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
      group_id: id,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

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
  '#84cc16'
]

const displayGroupStats = computed(() => {
  if (!props.groupStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...props.groupStats].sort((a, b) => toFiniteNumber(b[metricKey]) - toFiniteNumber(a[metricKey]))
})

const chartData = computed(() => {
  if (!props.groupStats?.length) return null

  return {
    labels: displayGroupStats.value.map((g) => g.group_name || String(g.group_id)),
    datasets: [
      {
        data: displayGroupStats.value.map((g) => toFiniteNumber(props.metric === 'actual_cost' ? g.actual_cost : g.total_tokens)),
        backgroundColor: chartColors.slice(0, displayGroupStats.value.length),
        borderWidth: 0
      }
    ]
  }
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
