<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.usageStatistics')"
    width="extra-wide"
    @close="handleClose"
  >
    <div class="components-account-account-stats-modal__panel">
      <!-- Account Info Header -->
      <div
        v-if="account"
        class="components-account-account-stats-modal__panel-2"
      >
        <div class="components-account-account-stats-modal__panel-3">
          <div
            class="components-account-account-stats-modal__panel-4"
          >
            <Icon name="chartBar" size="md" class="components-account-account-stats-modal__icon" :stroke-width="2" />
          </div>
          <div>
            <div class="components-account-account-stats-modal__panel-5">{{ account.name }}</div>
            <div class="components-account-account-stats-modal__panel-6">
              {{ t('admin.accounts.last30DaysUsage') }}
            </div>
          </div>
        </div>
        <span
          :class="[
            'components-account-account-stats-modal__text-6',
            account.status === 'active'
              ? 'components-account-account-stats-modal__text-8'
              : 'components-account-account-stats-modal__text-9'
          ]"
        >
          {{ account.status }}
        </span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="components-account-account-stats-modal__panel-7">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Main Stats Cards -->
        <div class="components-account-account-stats-modal__panel-8">
          <!-- 30-Day Total Cost -->
          <div
            class="components-account-account-stats-modal__panel-9 card"
          >
            <div class="components-account-account-stats-modal__panel-10">
              <span class="components-account-account-stats-modal__text">{{
                t('admin.accounts.stats.totalCost')
              }}</span>
              <div class="components-account-account-stats-modal__panel-11">
                <svg
                  class="components-account-account-stats-modal__icon-2"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
            </div>
            <p class="components-account-account-stats-modal__description">
              ${{ formatCost(stats.summary.total_cost) }}
            </p>
            <p class="components-account-account-stats-modal__description-2">
              {{ t('admin.accounts.stats.accumulatedCost') }}
              <span class="components-account-account-stats-modal__text-2">
                ({{ t('usage.userBilled') }}: {{ formatPoints(stats.summary.total_user_cost) }} ·
                {{ t('admin.accounts.stats.standardCost') }}: ${{
                  formatCost(stats.summary.total_standard_cost)
                }})
              </span>
            </p>
          </div>

          <!-- 30-Day Total Requests -->
          <div
            class="components-account-account-stats-modal__panel-12 card"
          >
            <div class="components-account-account-stats-modal__panel-10">
              <span class="components-account-account-stats-modal__text">{{
                t('admin.accounts.stats.totalRequests')
              }}</span>
              <div class="components-account-account-stats-modal__panel-13">
                <Icon name="bolt" size="sm" class="components-account-account-stats-modal__icon-3" :stroke-width="2" />
              </div>
            </div>
            <p class="components-account-account-stats-modal__description">
              {{ formatNumber(stats.summary.total_requests) }}
            </p>
            <p class="components-account-account-stats-modal__description-2">
              {{ t('admin.accounts.stats.totalCalls') }}
            </p>
          </div>

          <!-- Daily Average Cost -->
          <div
            class="components-account-account-stats-modal__panel-14 card"
          >
            <div class="components-account-account-stats-modal__panel-10">
              <span class="components-account-account-stats-modal__text">{{
                t('admin.accounts.stats.avgDailyCost')
              }}</span>
              <div class="components-account-account-stats-modal__panel-15">
                <Icon
                  name="calculator"
                  size="sm"
                  class="components-account-account-stats-modal__icon-4"
                  :stroke-width="2"
                />
              </div>
            </div>
            <p class="components-account-account-stats-modal__description">
              ${{ formatCost(stats.summary.avg_daily_cost) }}
            </p>
             <p class="components-account-account-stats-modal__description-2">
              {{
                t('admin.accounts.stats.basedOnActualDays', {
                  days: stats.summary.actual_days_used
                })
              }}
              <span class="components-account-account-stats-modal__text-2">
                ({{ t('usage.userBilled') }}: {{ formatPoints(stats.summary.avg_daily_user_cost) }})
              </span>
            </p>
          </div>

          <!-- Daily Average Requests -->
          <div
            class="components-account-account-stats-modal__panel-16 card"
          >
            <div class="components-account-account-stats-modal__panel-10">
              <span class="components-account-account-stats-modal__text">{{
                t('admin.accounts.stats.avgDailyRequests')
              }}</span>
              <div class="components-account-account-stats-modal__panel-17">
                <svg
                  class="components-account-account-stats-modal__icon-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"
                  />
                </svg>
              </div>
            </div>
            <p class="components-account-account-stats-modal__description">
              {{ formatNumber(Math.round(stats.summary.avg_daily_requests)) }}
            </p>
            <p class="components-account-account-stats-modal__description-2">
              {{ t('admin.accounts.stats.avgDailyUsage') }}
            </p>
          </div>
        </div>

        <!-- Row 2: Today, Highest Cost, Highest Requests -->
        <div class="components-account-account-stats-modal__panel-18">
          <!-- Today Overview -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-21">
                <svg
                  class="components-account-account-stats-modal__icon-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.todayOverview')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.accountBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >${{ formatCost(stats.summary.today?.cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.userBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >{{ formatPoints(stats.summary.today?.user_cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatNumber(stats.summary.today?.requests || 0)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.tokens')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatTokens(stats.summary.today?.tokens || 0)
                }}</span>
              </div>
            </div>
          </div>

          <!-- Highest Cost Day -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-24">
                <Icon
                  name="fire"
                  size="sm"
                  class="components-account-account-stats-modal__icon-7"
                  :stroke-width="2"
                />
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.highestCostDay')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.date')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  stats.summary.highest_cost_day?.label || '-'
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.accountBilled') }}</span>
                <span class="components-account-account-stats-modal__text-4"
                  >${{ formatCost(stats.summary.highest_cost_day?.cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.userBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >{{ formatPoints(stats.summary.highest_cost_day?.user_cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatNumber(stats.summary.highest_cost_day?.requests || 0)
                }}</span>
              </div>
            </div>
          </div>

          <!-- Highest Request Day -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-25">
                <Icon
                  name="trendingUp"
                  size="sm"
                  class="components-account-account-stats-modal__icon-8"
                  :stroke-width="2"
                />
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.highestRequestDay')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.date')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  stats.summary.highest_request_day?.label || '-'
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.requests')
                }}</span>
                <span class="components-account-account-stats-modal__text-5">{{
                  formatNumber(stats.summary.highest_request_day?.requests || 0)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.accountBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >${{ formatCost(stats.summary.highest_request_day?.cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.userBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >{{ formatPoints(stats.summary.highest_request_day?.user_cost || 0) }}</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- Row 3: Token Stats -->
        <div class="components-account-account-stats-modal__panel-18">
          <!-- Accumulated Tokens -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-26">
                <Icon name="cube" size="sm" class="components-account-account-stats-modal__icon-9" :stroke-width="2" />
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.accumulatedTokens')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.totalTokens')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatTokens(stats.summary.total_tokens)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.dailyAvgTokens')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatTokens(Math.round(stats.summary.avg_daily_tokens))
                }}</span>
              </div>
            </div>
          </div>

          <!-- Performance -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-27">
                <Icon name="bolt" size="sm" class="components-account-account-stats-modal__icon-10" :stroke-width="2" />
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.performance')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.avgResponseTime')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatDuration(stats.summary.avg_duration_ms)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.daysActive')
                }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >{{ stats.summary.actual_days_used }} / {{ stats.summary.days }}</span
                >
              </div>
            </div>
          </div>

          <!-- Recent Activity -->
          <div class="components-account-account-stats-modal__panel-19 card">
            <div class="components-account-account-stats-modal__panel-20">
              <div class="components-account-account-stats-modal__panel-28">
                <Icon
                  name="clipboard"
                  size="sm"
                  class="components-account-account-stats-modal__icon-11"
                  :stroke-width="2"
                />
              </div>
              <span class="components-account-account-stats-modal__text-3">{{
                t('admin.accounts.stats.recentActivity')
              }}</span>
            </div>
            <div class="components-account-account-stats-modal__panel-22">
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.todayRequests')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatNumber(stats.summary.today?.requests || 0)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{
                  t('admin.accounts.stats.todayTokens')
                }}</span>
                <span class="components-account-account-stats-modal__text-3">{{
                  formatTokens(stats.summary.today?.tokens || 0)
                }}</span>
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.accountBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >${{ formatCost(stats.summary.today?.cost || 0) }}</span
                >
              </div>
              <div class="components-account-account-stats-modal__panel-23">
                <span class="components-account-account-stats-modal__panel-6">{{ t('usage.userBilled') }}</span>
                <span class="components-account-account-stats-modal__text-3"
                  >{{ formatPoints(stats.summary.today?.user_cost || 0) }}</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- Usage Trend Chart -->
        <div class="components-account-account-stats-modal__panel-19 card">
          <h3 class="components-account-account-stats-modal__heading">
            {{ t('admin.accounts.stats.usageTrend') }}
          </h3>
          <div class="components-account-account-stats-modal__panel-29">
            <D3LineChart v-if="trendChartData" :data="trendChartData" :options="lineChartOptions" />
            <div
              v-else
              class="components-account-account-stats-modal__panel-30"
            >
              {{ t('admin.dashboard.noDataAvailable') }}
            </div>
          </div>
        </div>

        <!-- Model Distribution -->
        <ModelDistributionChart :model-stats="stats.models" :loading="false" />

        <EndpointDistributionChart
          :endpoint-stats="stats.endpoints || []"
          :loading="false"
          :title="t('usage.inboundEndpoint')"
        />

        <EndpointDistributionChart
          :endpoint-stats="stats.upstream_endpoints || []"
          :loading="false"
          :title="t('usage.upstreamEndpoint')"
        />
      </template>

      <!-- No Data State -->
      <div
        v-else-if="!loading"
        class="components-account-account-stats-modal__panel-31"
      >
        <Icon name="chartBar" size="xl" class="components-account-account-stats-modal__icon-12" :stroke-width="1.5" />
        <p class="components-account-account-stats-modal__description-3">{{ t('admin.accounts.stats.noData') }}</p>
      </div>
    </div>

    <template #footer>
      <div class="components-account-account-stats-modal__panel-32">
        <button
          @click="handleClose"
          class="components-account-account-stats-modal__action"
        >
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import EndpointDistributionChart from '@/components/charts/EndpointDistributionChart.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { Account, AccountUsageStatsResponse } from '@/types'
import { formatPoints, formatUSD } from '@/utils/format'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const loading = ref(false)
const stats = ref<AccountUsageStatsResponse | null>(null)

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#2e2e33',
  grid: isDarkMode.value ? '#2e2e33' : '#e5e7eb'
}))

// Line chart data
const trendChartData = computed(() => {
  if (!stats.value?.history?.length) return null

  return {
    labels: stats.value.history.map((h) => h.label),
    datasets: [
      {
        label: t('usage.accountBilled'),
        data: stats.value.history.map((h) => h.account_cost),
        unit: 'usd',
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        fill: true,
        tension: 0.3,
        yAxisID: 'y'
      },
      {
        label: t('usage.userBilled'),
        data: stats.value.history.map((h) => h.actual_cost),
        unit: 'points',
        borderColor: '#10b981',
        backgroundColor: 'rgba(16, 185, 129, 0.08)',
        fill: false,
        tension: 0.3,
        borderDash: [5, 5],
        yAxisID: 'y'
      },
      {
        label: t('admin.accounts.stats.requests'),
        data: stats.value.history.map((h) => h.requests),
        unit: 'count',
        borderColor: '#f97316',
        backgroundColor: 'rgba(249, 115, 22, 0.1)',
        fill: false,
        tension: 0.3,
        yAxisID: 'y1'
      }
    ]
  }
})

// Line chart options with dual Y-axis
const lineChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const label = context.dataset.label || ''
          const value = context.raw
          if (context.dataset.unit === 'usd') {
            return `${label}: ${formatUSD(value)}`
          }
          if (context.dataset.unit === 'points') {
            return `${label}: ${formatPoints(value)}`
          }
          return `${label}: ${formatNumber(value)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        maxRotation: 45,
        minRotation: 0
      }
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: '#3b82f6',
        font: {
          size: 10
        },
        callback: (value: string | number) => formatNumber(Number(value))
      },
      title: {
        display: true,
        text: t('admin.accounts.stats.costAndPoints'),
        color: '#3b82f6',
        font: {
          size: 11
        }
      }
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      grid: {
        drawOnChartArea: false
      },
      ticks: {
        color: '#f97316',
        font: {
          size: 10
        },
        callback: (value: string | number) => formatNumber(Number(value))
      },
      title: {
        display: true,
        text: t('admin.accounts.stats.requests'),
        color: '#f97316',
        font: {
          size: 11
        }
      }
    }
  }
}))

// Load stats when modal opens
watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      await loadStats()
    } else {
      stats.value = null
    }
  }
)

const loadStats = async () => {
  if (!props.account) return

  loading.value = true
  try {
    stats.value = await adminAPI.accounts.getStats(props.account.id, 30)
  } catch (error) {
    console.error('Failed to load account stats:', error)
    stats.value = null
  } finally {
    loading.value = false
  }
}

const handleClose = () => {
  emit('close')
}

// Format helpers
const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}

const formatNumber = (value: number): string => {
  if (value >= 1_000_000) {
    return (value / 1_000_000).toFixed(2) + 'M'
  } else if (value >= 1_000) {
    return (value / 1_000).toFixed(2) + 'K'
  }
  return value.toLocaleString()
}

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

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}
</script>
