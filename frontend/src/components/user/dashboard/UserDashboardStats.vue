<template>
  <!-- Row 1: Core Stats -->
  <div class="components-user-dashboard-user-dashboard-stats__panel">
    <!-- Balance -->
    <div v-if="!isSimple" class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-4">
          <svg class="components-user-dashboard-user-dashboard-stats__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
          </svg>
        </div>
        <div>
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.balance') }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-2">{{ formatPoints(balance) }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-3">{{ t('common.available') }}</p>
        </div>
      </div>
    </div>

    <!-- API Keys -->
    <div class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-5">
          <Icon name="key" size="md" class="components-user-dashboard-user-dashboard-stats__icon-2" :stroke-width="2" />
        </div>
        <div>
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.apiKeys') }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-4">{{ stats?.total_api_keys || 0 }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-5">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
        </div>
      </div>
    </div>

    <!-- User concurrency -->
    <div class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-10">
          <Icon name="bolt" size="md" class="components-user-dashboard-user-dashboard-stats__icon-7" :stroke-width="2" />
        </div>
        <div>
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.concurrency') }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-4">
            {{ formatNumber(stats?.current_concurrency || 0) }} / {{ formatNumber(stats?.concurrency || 0) }}
          </p>
          <p class="components-user-dashboard-user-dashboard-stats__description-3">{{ t('dashboard.concurrencyUsage') }}</p>
        </div>
      </div>
    </div>

  </div>

  <!-- Row 2: Token Stats -->
  <div class="components-user-dashboard-user-dashboard-stats__panel">
    <!-- Today Tokens -->
    <div class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-8">
          <Icon name="cube" size="md" class="components-user-dashboard-user-dashboard-stats__icon-5" :stroke-width="2" />
        </div>
        <div>
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.todayTokens') }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-4">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-3">{{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }} / {{ t('dashboard.cache') }}: {{ formatTokens((stats?.today_cache_creation_tokens || 0) + (stats?.today_cache_read_tokens || 0)) }}</p>
        </div>
      </div>
    </div>

    <!-- Total Tokens -->
    <div class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-9">
          <Icon name="database" size="md" class="components-user-dashboard-user-dashboard-stats__icon-6" :stroke-width="2" />
        </div>
        <div>
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.totalTokens') }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-4">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="components-user-dashboard-user-dashboard-stats__description-3">{{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} / {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }} / {{ t('dashboard.cache') }}: {{ formatTokens((stats?.total_cache_creation_tokens || 0) + (stats?.total_cache_read_tokens || 0)) }}</p>
        </div>
      </div>
    </div>

    <!-- Performance (RPM/TPM) -->
    <div class="components-user-dashboard-user-dashboard-stats__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-stats__panel-3">
        <div class="components-user-dashboard-user-dashboard-stats__panel-10">
          <Icon name="bolt" size="md" class="components-user-dashboard-user-dashboard-stats__icon-7" :stroke-width="2" />
        </div>
        <div class="components-user-dashboard-user-dashboard-stats__panel-11">
          <p class="components-user-dashboard-user-dashboard-stats__description">{{ t('dashboard.performance') }}</p>
          <div class="components-user-dashboard-user-dashboard-stats__panel-12">
            <p class="components-user-dashboard-user-dashboard-stats__description-4">{{ formatTokens(stats?.rpm || 0) }}</p>
            <span class="components-user-dashboard-user-dashboard-stats__description-3">RPM</span>
          </div>
          <div class="components-user-dashboard-user-dashboard-stats__panel-12">
            <p class="components-user-dashboard-user-dashboard-stats__description-7">{{ formatTokens(stats?.tpm || 0) }}</p>
            <span class="components-user-dashboard-user-dashboard-stats__description-3">TPM</span>
          </div>
        </div>
      </div>
    </div>

  </div>

</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import { formatPoints } from '@/utils/format'

defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()
const { t } = useI18n()

const formatNumber = (n: number) => n.toLocaleString()
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
</script>
