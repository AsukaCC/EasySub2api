<template>
  <div class="components-admin-usage-usage-stats-cards__panel">
    <div class="components-admin-usage-usage-stats-cards__panel-2 card">
      <div class="components-admin-usage-usage-stats-cards__panel-3">
        <Icon name="document" size="md" />
      </div>
      <div>
        <p class="components-admin-usage-usage-stats-cards__description">{{ t('usage.totalRequests') }}</p>
        <p class="components-admin-usage-usage-stats-cards__description-2">{{ stats?.total_requests?.toLocaleString() || '0' }}</p>
        <p class="components-admin-usage-usage-stats-cards__description-3">{{ t('usage.inSelectedRange') }}</p>
      </div>
    </div>
    <div class="components-admin-usage-usage-stats-cards__panel-2 card">
      <div class="components-admin-usage-usage-stats-cards__panel-4"><svg class="components-admin-usage-usage-stats-cards__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg></div>
      <div>
        <p class="components-admin-usage-usage-stats-cards__description">{{ t('usage.totalTokens') }}</p>
        <p class="components-admin-usage-usage-stats-cards__description-2">{{ formatTokens(stats?.total_tokens || 0) }}</p>
        <p class="components-admin-usage-usage-stats-cards__description-4">
          <span>{{ t('usage.in') }}: {{ formatTokens(stats?.total_input_tokens || 0) }}</span>
          <span>/</span>
          <span>{{ t('usage.out') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}</span>
          <span>/</span>
          <span class="components-admin-usage-usage-stats-cards__text">
            <span>{{ cacheLabel() }}: {{ formatTokens(stats?.total_cache_tokens || 0) }}</span>
            <HelpTooltip width-class="w-64">
              <template #trigger>
                <svg
                  class="components-admin-usage-usage-stats-cards__icon-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </template>
              <span class="usage-cache-popover">
              <span class="components-admin-usage-usage-stats-cards__text-3">
                {{ cacheDetailLabel() }}
              </span>
              <span class="components-admin-usage-usage-stats-cards__text-4">
                <span>{{ t('usage.cacheCreationTokensLabel') }}</span>
                <span class="components-admin-usage-usage-stats-cards__text-5">
                  {{ formatTokens(stats?.total_cache_creation_tokens || 0) }}
                </span>
              </span>
              <span class="components-admin-usage-usage-stats-cards__text-6">
                <span>{{ t('usage.cacheReadTokensLabel') }}</span>
                <span class="components-admin-usage-usage-stats-cards__text-5">
                  {{ formatTokens(stats?.total_cache_read_tokens || 0) }}
                </span>
              </span>
              </span>
            </HelpTooltip>
          </span>
        </p>
      </div>
    </div>
    <div class="components-admin-usage-usage-stats-cards__panel-2 card">
      <div class="components-admin-usage-usage-stats-cards__panel-5">
        <Icon name="dollar" size="md" />
      </div>
      <div class="components-admin-usage-usage-stats-cards__panel-6">
        <p class="components-admin-usage-usage-stats-cards__description">{{ t('usage.totalCost') }}</p>
        <p class="components-admin-usage-usage-stats-cards__description-5">
          {{ formatPoints(stats?.total_actual_cost || 0) }}
        </p>
        <p class="components-admin-usage-usage-stats-cards__description-3">
          <template v-if="showAccountCost && totalAccountCost != null">
            <span class="components-admin-usage-usage-stats-cards__text-7">{{ t('usage.accountCost') }} ${{ totalAccountCost.toFixed(4) }}</span>
            <span> · </span>
          </template>
          <span>
            {{ t('usage.standardCost') }}
            <span :class="{ 'components-admin-usage-usage-stats-cards__text-8': strikeStandardCost }">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
          </span>
        </p>
      </div>
    </div>
    <div class="components-admin-usage-usage-stats-cards__panel-2 card">
      <div class="components-admin-usage-usage-stats-cards__panel-7">
        <Icon name="clock" size="md" />
      </div>
      <div><p class="components-admin-usage-usage-stats-cards__description">{{ t('usage.avgDuration') }}</p><p class="components-admin-usage-usage-stats-cards__description-2">{{ formatDuration(stats?.average_duration_ms || 0) }}</p></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import type { UsageStatsResponse } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import { formatPoints } from '@/utils/format'

const props = withDefaults(defineProps<{
  stats: (AdminUsageStatsResponse | UsageStatsResponse) | null
  showAccountCost?: boolean
  strikeStandardCost?: boolean
}>(), {
  showAccountCost: true,
  strikeStandardCost: false,
})

const { t } = useI18n()

const totalAccountCost = computed(() => {
  const stats = props.stats as (AdminUsageStatsResponse & { total_account_cost?: number }) | null
  return stats?.total_account_cost ?? null
})
const showAccountCost = computed(() => props.showAccountCost)
const strikeStandardCost = computed(() => props.strikeStandardCost)

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}

const cacheLabel = () => t('usage.cacheTotal')
const cacheDetailLabel = () => t('usage.cacheBreakdown')
</script>

<style scoped>
.usage-cache-popover {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
</style>
