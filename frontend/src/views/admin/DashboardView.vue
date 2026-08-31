<template>
  <AppLayout>
    <div class="page-stack">
      <!-- Loading State -->
      <div v-if="loading" class="views-admin-dashboard-view__panel-2">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Core Stats -->
        <div class="views-admin-dashboard-view__panel-3">
          <!-- Total API Keys -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-6">
                <Icon name="key" size="md" class="views-admin-dashboard-view__icon" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.apiKeys') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ stats.total_api_keys }}
                </p>
                <p class="views-admin-dashboard-view__description-3">
                  {{ stats.active_api_keys }} {{ t('common.active') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Service Accounts -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-7">
                <Icon name="server" size="md" class="views-admin-dashboard-view__icon-2" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.accounts') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ stats.total_accounts }}
                </p>
                <p class="views-admin-dashboard-view__description-4">
                  <span class="views-admin-dashboard-view__text"
                    >{{ stats.normal_accounts }} {{ t('common.active') }}</span
                  >
                  <span v-if="stats.error_accounts > 0" class="views-admin-dashboard-view__text-2"
                    >{{ stats.error_accounts }} {{ t('common.error') }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Today Requests -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-8">
                <Icon name="chart" size="md" class="views-admin-dashboard-view__text" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.todayRequests') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ stats.today_requests }}
                </p>
                <p class="views-admin-dashboard-view__description-5">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
                </p>
              </div>
            </div>
          </div>

          <!-- New Users Today -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-9">
                <Icon name="userPlus" size="md" class="views-admin-dashboard-view__icon-3" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.users') }}
                </p>
                <p class="views-admin-dashboard-view__description-6">
                  +{{ stats.today_new_users }}
                </p>
                <p class="views-admin-dashboard-view__description-5">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_users) }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Account weekly quota -->
        <section class="quota-dashboard card">
          <div class="quota-dashboard__header">
            <div>
              <h2 class="quota-dashboard__title">{{ t('admin.dashboard.accountQuotaTitle') }}</h2>
              <p class="quota-dashboard__hint">{{ t('admin.dashboard.accountQuotaHint') }}</p>
            </div>
            <button type="button" class="btn btn-secondary" :disabled="quotaLoading" @click="loadAccountQuotas">
              <LoadingSpinner v-if="quotaLoading" size="sm" />
              <span v-else>{{ t('common.refresh') }}</span>
            </button>
          </div>
          <div v-if="quotaLoading && !quotaDashboard" class="quota-dashboard__loading">
            <LoadingSpinner />
          </div>
          <div v-else-if="quotaDashboard?.platforms.length" class="quota-dashboard__platforms">
            <article v-for="platform in quotaDashboard.platforms" :key="platform.platform" class="quota-platform">
              <div class="quota-platform__header">
                <div>
                  <h3 class="quota-platform__title">{{ platformLabel(platform.platform) }}</h3>
                  <span class="quota-platform__meta">
                    {{ platform.summary.account_count }} {{ t('admin.dashboard.quotaAccounts') }} ·
                    {{ platform.summary.rate_limited_count }} {{ t('admin.dashboard.quotaRateLimited') }}
                  </span>
                </div>
                <span class="quota-platform__percent">{{ formatPercent(platform.summary.available_percent) }}</span>
              </div>
              <div class="quota-bar" :aria-label="t('admin.dashboard.accountQuotaTitle')">
                <span class="quota-bar__used" :style="barStyle(platform.summary.used_percent)" />
                <span class="quota-bar__available" :style="barStyle(platform.summary.available_percent)" />
                <span class="quota-bar__limited" :style="barStyle(platform.summary.rate_limited_percent)" />
                <span class="quota-bar__unknown" :style="barStyle(platform.summary.unknown_percent)" />
                <span class="quota-bar__unavailable" :style="barStyle(platform.summary.unavailable_percent)" />
              </div>
              <div class="quota-legend">
                <span><i class="quota-legend__dot quota-legend__dot--available" />{{ t('admin.dashboard.quotaAvailable') }}</span>
                <span><i class="quota-legend__dot quota-legend__dot--limited" />{{ t('admin.dashboard.quotaRateLimited') }}</span>
                <span><i class="quota-legend__dot quota-legend__dot--unknown" />{{ t('admin.dashboard.quotaUnknown') }}</span>
              </div>
              <div class="quota-groups">
                <div v-for="group in platform.groups" :key="`${platform.platform}-${group.id}`" class="quota-group">
                  <button type="button" class="quota-group__toggle" @click="toggleQuotaGroup(platform.platform, group.id)">
                    <span class="quota-group__name">{{ group.name }}</span>
                    <span class="quota-group__summary">
                      {{ formatPercent(group.summary.available_percent) }} · {{ group.summary.account_count }}
                    </span>
                    <Icon :name="isQuotaGroupExpanded(platform.platform, group.id) ? 'chevronUp' : 'chevronDown'" size="sm" />
                  </button>
                  <div class="quota-bar quota-bar--group">
                    <span class="quota-bar__used" :style="barStyle(group.summary.used_percent)" />
                    <span class="quota-bar__available" :style="barStyle(group.summary.available_percent)" />
                    <span class="quota-bar__limited" :style="barStyle(group.summary.rate_limited_percent)" />
                    <span class="quota-bar__unknown" :style="barStyle(group.summary.unknown_percent)" />
                    <span class="quota-bar__unavailable" :style="barStyle(group.summary.unavailable_percent)" />
                  </div>
                  <div v-if="isQuotaGroupExpanded(platform.platform, group.id)" class="quota-accounts">
                    <div v-if="quotaAccountsLoading[groupKey(platform.platform, group.id)]" class="quota-accounts__loading">
                      <LoadingSpinner size="sm" />
                    </div>
                    <div v-else-if="quotaAccounts[groupKey(platform.platform, group.id)]?.length" class="quota-account-list">
                      <div v-for="account in quotaAccounts[groupKey(platform.platform, group.id)]" :key="account.id" class="quota-account">
                        <div class="quota-account__info">
                          <span class="quota-account__name">{{ account.name }}</span>
                          <span v-if="account.rate_limited" class="quota-account__badge quota-account__badge--limited">429</span>
                          <span v-if="account.model_rate_limit_count" class="quota-account__badge quota-account__badge--model">{{ account.model_rate_limit_count }} {{ t('admin.dashboard.quotaModelsLimited') }}</span>
                        </div>
                        <div class="quota-account__meta">
                          <span>{{ account.weekly.known ? `${formatPercent(account.weekly.remaining_percent)} ${t('admin.dashboard.quotaRemaining')}` : t('admin.dashboard.quotaUnknown') }}</span>
                          <span v-if="account.weekly.reset_at">{{ formatQuotaReset(account.weekly.reset_at) }}</span>
                        </div>
                        <div class="quota-bar quota-bar--account">
                          <span class="quota-bar__used" :style="barStyle(account.weekly.known ? account.weekly.used_percent : 0)" />
                          <span v-if="account.state === 'available'" class="quota-bar__available" :style="barStyle(account.weekly.remaining_percent)" />
                          <span v-else-if="account.state === 'rate_limited'" class="quota-bar__limited" :style="barStyle(account.weekly.remaining_percent)" />
                          <span v-else-if="account.state === 'unavailable'" class="quota-bar__unavailable" :style="barStyle(account.weekly.remaining_percent || 100)" />
                          <span v-else class="quota-bar__unknown" :style="barStyle(100)" />
                        </div>
                      </div>
                    </div>
                    <span v-else class="quota-accounts__empty">{{ t('admin.dashboard.quotaNoAccounts') }}</span>
                  </div>
                </div>
              </div>
            </article>
          </div>
          <p v-else class="quota-dashboard__empty">{{ t('admin.dashboard.quotaNoAccounts') }}</p>
          <p v-if="quotaDashboard?.latest_observed_at" class="quota-dashboard__updated">
            {{ t('admin.dashboard.quotaLastObserved') }}: {{ formatQuotaReset(quotaDashboard.latest_observed_at) }}
          </p>
        </section>

        <!-- Row 2: Token Stats -->
        <div class="views-admin-dashboard-view__panel-3">
          <!-- Today Tokens -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-10">
                <Icon name="cube" size="md" class="views-admin-dashboard-view__icon-4" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.todayTokens') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ formatTokens(stats.today_tokens) }}
                </p>
                <p class="views-admin-dashboard-view__description-4">
                  <span
                    class="views-admin-dashboard-view__text"
                    :title="t('admin.dashboard.actual')"
                    >{{ formatPoints(stats.today_actual_cost) }}</span
                  >
                  <span class="views-admin-dashboard-view__text-3"> / </span>
                  <span
                    class="views-admin-dashboard-view__text-4"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.today_account_cost) }}</span
                  >
                  <span class="views-admin-dashboard-view__text-3"> / </span>
                  <span
                    class="views-admin-dashboard-view__text-3"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.today_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Total Tokens -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-11">
                <Icon name="database" size="md" class="views-admin-dashboard-view__icon-5" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.totalTokens') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ formatTokens(stats.total_tokens) }}
                </p>
                <p class="views-admin-dashboard-view__description-4">
                  <span
                    class="views-admin-dashboard-view__text"
                    :title="t('admin.dashboard.actual')"
                    >{{ formatPoints(stats.total_actual_cost) }}</span
                  >
                  <span class="views-admin-dashboard-view__text-3"> / </span>
                  <span
                    class="views-admin-dashboard-view__text-4"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.total_account_cost) }}</span
                  >
                  <span class="views-admin-dashboard-view__text-3"> / </span>
                  <span
                    class="views-admin-dashboard-view__text-3"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.total_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Performance (RPM/TPM) -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-12">
                <Icon name="bolt" size="md" class="views-admin-dashboard-view__icon-6" :stroke-width="2" />
              </div>
              <div class="views-admin-dashboard-view__panel-13">
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.performance') }}
                </p>
                <div class="views-admin-dashboard-view__panel-14">
                  <p class="views-admin-dashboard-view__description-2">
                    {{ formatTokens(stats.rpm) }}
                  </p>
                  <span class="views-admin-dashboard-view__description-5">RPM</span>
                </div>
                <div class="views-admin-dashboard-view__panel-14">
                  <p class="views-admin-dashboard-view__description-7">
                    {{ formatTokens(stats.tpm) }}
                  </p>
                  <span class="views-admin-dashboard-view__description-5">TPM</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Avg Response Time -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-5">
              <div class="views-admin-dashboard-view__panel-15">
                <Icon name="clock" size="md" class="views-admin-dashboard-view__icon-7" :stroke-width="2" />
              </div>
              <div>
                <p class="views-admin-dashboard-view__description">
                  {{ t('admin.dashboard.avgResponse') }}
                </p>
                <p class="views-admin-dashboard-view__description-2">
                  {{ formatDuration(stats.average_duration_ms) }}
                </p>
                <p class="views-admin-dashboard-view__description-5">
                  {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="views-admin-dashboard-view__panel-4 card">
          <div class="views-admin-dashboard-view__panel-16">
            <h2 class="views-admin-dashboard-view__heading">
              {{ t('admin.dashboard.quickActions') }}
            </h2>
          </div>
          <div class="views-admin-dashboard-view__panel-17">
            <button
              type="button"
              class="views-admin-dashboard-view__action-2"
              @click="router.push('/admin/groups')"
            >
              <span class="views-admin-dashboard-view__text-9">
                <Icon name="grid" size="md" :stroke-width="2" />
              </span>
              <span class="views-admin-dashboard-view__text-6">
                <span class="views-admin-dashboard-view__text-7">
                  {{ t('admin.dashboard.groupPricing') }}
                </span>
                <span class="views-admin-dashboard-view__text-8">
                  {{ t('admin.dashboard.groupPricingDesc') }}
                </span>
              </span>
              <Icon name="chevronRight" size="sm" class="views-admin-dashboard-view__icon-9" />
            </button>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="page-stack">
          <!-- Date Range Filter -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <div class="views-admin-dashboard-view__panel-18">
              <div class="views-admin-dashboard-view__panel-19">
                <span class="views-admin-dashboard-view__text-10"
                  >{{ t('admin.dashboard.timeRange') }}:</span
                >
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <button @click="loadDashboardStats" :disabled="chartsLoading" class="btn btn-secondary">
                {{ t('common.refresh') }}
              </button>
              <div class="views-admin-dashboard-view__panel-20">
                <span class="views-admin-dashboard-view__text-10"
                  >{{ t('admin.dashboard.granularity') }}:</span
                >
                <div class="views-admin-dashboard-view__panel-21">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Charts Grid -->
          <div class="views-admin-dashboard-view__panel-22">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <!-- User Usage Trend (Full Width) -->
          <div class="views-admin-dashboard-view__panel-4 card">
            <h3 class="views-admin-dashboard-view__heading-2">
              {{ t('admin.dashboard.recentUsage') }} (Top 12)
            </h3>
            <div class="views-admin-dashboard-view__panel-23">
              <div v-if="userTrendLoading" class="views-admin-dashboard-view__panel-24">
                <LoadingSpinner size="md" />
              </div>
              <D3LineChart v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
              <div
                v-else
                class="views-admin-dashboard-view__panel-25"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem
} from '@/types'
import type { AccountQuotaAccount, AccountQuotaDashboard } from '@/api/admin/dashboard'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import D3LineChart from '@/components/charts/d3/D3LineChart.vue'
import { formatPoints } from '@/utils/format'

const appStore = useAppStore()
const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)
const quotaDashboard = ref<AccountQuotaDashboard | null>(null)
const quotaLoading = ref(false)
const quotaAccounts = ref<Record<string, AccountQuotaAccount[]>>({})
const quotaAccountsLoading = ref<Record<string, boolean>>({})
const expandedQuotaGroups = ref<Record<string, boolean>>({})
let quotaRefreshTimer: number | undefined

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12
const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const platformLabel = (platform: string): string => {
  const labels: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    gemini: 'Gemini',
    antigravity: 'Antigravity',
    grok: 'Grok',
    kimi: 'Kimi',
    zhipu: 'Zhipu',
    deepseek: 'DeepSeek'
  }
  return labels[platform] || platform
}

const formatPercent = (value: number): string => `${Math.round(Math.max(0, Math.min(100, value || 0)))}%`
const barStyle = (value: number) => ({ width: `${Math.max(0, Math.min(100, value || 0))}%` })
const formatQuotaReset = (value: string): string => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}
const groupKey = (platform: string, groupId: string): string => `${platform}:${groupId}`
const isQuotaGroupExpanded = (platform: string, groupId: string): boolean => expandedQuotaGroups.value[groupKey(platform, groupId)] === true

const loadAccountQuotas = async () => {
  quotaLoading.value = true
  try {
    quotaDashboard.value = await adminAPI.dashboard.getAccountQuotas()
  } catch (error) {
    console.error('Error loading account quota dashboard:', error)
  } finally {
    quotaLoading.value = false
  }
}

const toggleQuotaGroup = async (platform: string, groupId: string) => {
  const key = groupKey(platform, groupId)
  expandedQuotaGroups.value[key] = !expandedQuotaGroups.value[key]
  if (!expandedQuotaGroups.value[key] || quotaAccounts.value[key]) return
  quotaAccountsLoading.value[key] = true
  try {
    const response = await adminAPI.dashboard.getAccountQuotaAccounts({ platform, group_id: groupId, limit: 100 })
    quotaAccounts.value[key] = response.accounts || []
  } catch (error) {
    console.error('Error loading account quota details:', error)
    quotaAccounts.value[key] = []
  } finally {
    quotaAccountsLoading.value[key] = false
  }
}

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end)
  }
}

// Date range
const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#2e2e33',
  grid: isDarkMode.value ? '#2e2e33' : '#e5e7eb'
}))

// Line chart options (for user trend chart)
const lineOptions = computed(() => ({
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
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
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
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

// User trend chart data
const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) {
      return username
    }

    const email = point.email?.trim()
    if (email) {
      return email
    }

    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  // Group by user_id to avoid merging different users with the same display name
  const userGroups = new Map<string, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.user_id
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = [
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

  const datasets = Array.from(userGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const toFiniteNumber = (value: unknown): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
}

const formatNumber = (value: number | null | undefined): string => {
  return toFiniteNumber(value).toLocaleString()
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

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: browserTimezone,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: browserTimezone,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) {
      userTrendLoading.value = false
    }
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      timezone: browserTimezone,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

onMounted(() => {
  loadDashboardStats()
  loadAccountQuotas()
  quotaRefreshTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') {
      loadAccountQuotas()
    }
  }, 30_000)
})

onUnmounted(() => {
  if (quotaRefreshTimer !== undefined) {
    window.clearInterval(quotaRefreshTimer)
  }
})
</script>

<style scoped>
.views-admin-dashboard-view__panel-4 {
  border-radius: var(--radius-xl);
}

.quota-dashboard {
  padding: 1.25rem;
}

.quota-dashboard__header,
.quota-platform__header,
.quota-group__toggle,
.quota-account__info,
.quota-account__meta {
  display: flex;
  align-items: center;
}

.quota-dashboard__header,
.quota-platform__header,
.quota-group__toggle {
  justify-content: space-between;
  gap: 1rem;
}

.quota-dashboard__title,
.quota-platform__title {
  margin: 0;
  color: var(--color-text-primary);
}

.quota-dashboard__title {
  font-size: var(--type-dialog-title-size);
}

.quota-platform__title {
  font-size: var(--type-card-size);
}

.quota-dashboard__hint,
.quota-platform__meta,
.quota-account__meta,
.quota-dashboard__updated,
.quota-accounts__empty {
  color: var(--color-text-secondary);
  font-size: var(--type-caption-size);
}

.quota-dashboard__hint {
  margin: 0.25rem 0 0;
}

.quota-dashboard__platforms {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.quota-platform {
  padding: 1rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-inset-bg, var(--color-surface-muted));
}

.quota-platform__percent,
.quota-group__summary {
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}

.quota-bar {
  display: flex;
  overflow: hidden;
  height: 0.5rem;
  margin-top: 0.75rem;
  border-radius: 999px;
  background: var(--color-surface-muted);
}

.quota-bar > span {
  display: block;
  min-width: 0;
  transition: width 180ms ease;
}

.quota-bar__used { background: var(--color-text-tertiary); }
.quota-bar__available { background: var(--color-success); }
.quota-bar__limited { background: var(--color-warning); }
.quota-bar__unknown { background: repeating-linear-gradient(135deg, var(--color-text-quaternary) 0 3px, transparent 3px 6px); }
.quota-bar__unavailable { background: var(--color-text-quaternary); }

.quota-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.65rem;
  color: var(--color-text-secondary);
  font-size: var(--type-micro-size);
}

.quota-legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.quota-legend__dot {
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 50%;
}
.quota-legend__dot--available { background: var(--color-success); }
.quota-legend__dot--limited { background: var(--color-warning); }
.quota-legend__dot--unknown { border: 1px dashed var(--color-text-quaternary); }

.quota-groups {
  display: grid;
  gap: 0.5rem;
  margin-top: 1rem;
}

.quota-group__toggle {
  width: 100%;
  padding: 0.4rem 0;
  border: 0;
  color: var(--color-text-primary);
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.quota-group__name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quota-bar--group { height: 0.35rem; margin-top: 0; }
.quota-accounts { padding: 0.25rem 0 0.5rem 0.5rem; }
.quota-account-list { display: grid; gap: 0.6rem; }
.quota-account { min-width: 0; }
.quota-account__info { gap: 0.4rem; min-width: 0; }
.quota-account__name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--color-text-primary); font-size: var(--type-control-size); }
.quota-account__badge { padding: 0.1rem 0.35rem; border-radius: 999px; font-size: var(--type-micro-size); white-space: nowrap; }
.quota-account__badge--limited { color: var(--color-warning); background: color-mix(in srgb, var(--color-warning) 14%, transparent); }
.quota-account__badge--model { color: var(--color-text-secondary); background: var(--color-surface-muted); }
.quota-account__meta { justify-content: space-between; gap: 0.5rem; margin-top: 0.2rem; }
.quota-bar--account { height: 0.3rem; margin-top: 0.3rem; }
.quota-dashboard__updated { margin: 1rem 0 0; }
.quota-dashboard__loading,
.quota-dashboard__empty,
.quota-accounts__loading { display: flex; justify-content: center; padding: 1.5rem; }

@media (max-width: 640px) {
  .quota-dashboard { padding: 1rem; }
  .quota-dashboard__header { align-items: flex-start; }
  .quota-dashboard__header .btn { flex-shrink: 0; }
}
</style>
