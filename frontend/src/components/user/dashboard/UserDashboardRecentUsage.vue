<template>
  <div class="card">
    <div class="components-user-dashboard-user-dashboard-recent-usage__panel">
      <div>
        <h2 class="components-user-dashboard-user-dashboard-recent-usage__heading">{{ t('dashboard.requestSummary') }}</h2>
        <p class="components-user-dashboard-user-dashboard-recent-usage__description-2">{{ t('dashboard.requestSummaryHint') }}</p>
      </div>
    </div>
    <div class="components-user-dashboard-user-dashboard-recent-usage__panel-2">
      <div v-if="loading" class="components-user-dashboard-user-dashboard-recent-usage__panel-3">
        <LoadingSpinner size="lg" />
      </div>
      <div v-else-if="stats.total_requests === 0" class="components-user-dashboard-user-dashboard-recent-usage__panel-4">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else class="components-user-dashboard-user-dashboard-recent-usage__panel-5">
        <div class="components-user-dashboard-user-dashboard-recent-usage__panel-6">
          <div class="components-user-dashboard-user-dashboard-recent-usage__panel-7">
            <div class="components-user-dashboard-user-dashboard-recent-usage__panel-8">
              <Icon name="chart" size="md" class="components-user-dashboard-user-dashboard-recent-usage__icon" />
            </div>
            <div>
              <p class="components-user-dashboard-user-dashboard-recent-usage__description">{{ t('dashboard.todayRequests') }}</p>
              <p class="components-user-dashboard-user-dashboard-recent-usage__description-3">{{ formatNumber(stats.today_requests) }}</p>
            </div>
          </div>
          <div class="components-user-dashboard-user-dashboard-recent-usage__panel-9">
            <p class="components-user-dashboard-user-dashboard-recent-usage__description">{{ t('dashboard.totalRequests') }}</p>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description-3">{{ formatNumber(stats.total_requests) }}</p>
          </div>
        </div>

        <div class="components-user-dashboard-user-dashboard-recent-usage__panel-6">
          <div>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description">{{ t('dashboard.todayCost') }}</p>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description-3">{{ formatPoints(stats.today_actual_cost) }}</p>
          </div>
          <div class="components-user-dashboard-user-dashboard-recent-usage__panel-9">
            <p class="components-user-dashboard-user-dashboard-recent-usage__description">{{ t('dashboard.todayTokens') }}</p>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description-3">{{ formatTokens(stats.today_tokens) }}</p>
          </div>
        </div>

        <div v-if="latest" class="components-user-dashboard-user-dashboard-recent-usage__panel-6">
          <div>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description">{{ t('dashboard.latestRequest') }}</p>
            <p class="components-user-dashboard-user-dashboard-recent-usage__description-2" :title="latest.model">{{ latest.model || '-' }}</p>
          </div>
          <p class="components-user-dashboard-user-dashboard-recent-usage__description-2">{{ formatDateTime(latest.created_at) }}</p>
        </div>

        <router-link to="/usage" class="components-user-dashboard-user-dashboard-recent-usage__router-link">
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime, formatPoints } from '@/utils/format'
import type { UsageLog } from '@/types'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

defineProps<{
  stats: UserStatsType
  latest: UsageLog | null
  loading: boolean
}>()
const { t } = useI18n()
const formatNumber = (n: number) => n.toLocaleString()
const formatTokens = (n: number) => {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`
  return n.toLocaleString()
}
</script>
