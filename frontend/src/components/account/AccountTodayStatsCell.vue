<template>
  <div>
    <!-- Loading state -->
    <div v-if="props.loading && !props.stats" class="components-account-account-today-stats-cell__panel">
      <div class="components-account-account-today-stats-cell__panel-2"></div>
      <div class="components-account-account-today-stats-cell__panel-3"></div>
      <div class="components-account-account-today-stats-cell__panel-4"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="props.error && !props.stats" class="components-account-account-today-stats-cell__panel-5">
      {{ props.error }}
    </div>

    <!-- Stats data -->
    <div v-else-if="props.stats" class="components-account-account-today-stats-cell__panel-6">
      <!-- Requests -->
      <div class="components-account-account-today-stats-cell__panel-7">
        <span class="components-account-account-today-stats-cell__text"
          >{{ t('admin.accounts.stats.requests') }}:</span
        >
        <span class="components-account-account-today-stats-cell__text-2">{{
          formatNumber(props.stats.requests)
        }}</span>
      </div>
      <!-- Tokens -->
      <div class="components-account-account-today-stats-cell__panel-7">
        <span class="components-account-account-today-stats-cell__text"
          >{{ t('admin.accounts.stats.tokens') }}:</span
        >
        <span class="components-account-account-today-stats-cell__text-2">{{
          formatTokens(props.stats.tokens)
        }}</span>
      </div>
      <!-- Cost (Account) -->
      <div class="components-account-account-today-stats-cell__panel-7">
        <span class="components-account-account-today-stats-cell__text">{{ t('usage.accountBilled') }}:</span>
        <span class="components-account-account-today-stats-cell__text-3">{{
          formatUSD(props.stats.cost)
        }}</span>
      </div>
      <!-- Cost (User/API Key) -->
      <div v-if="props.stats.user_cost != null" class="components-account-account-today-stats-cell__panel-7">
        <span class="components-account-account-today-stats-cell__text">{{ t('usage.userBilled') }}:</span>
        <span class="components-account-account-today-stats-cell__text-2">{{
          formatPoints(props.stats.user_cost)
        }}</span>
      </div>
    </div>

    <!-- No data -->
    <div v-else class="components-account-account-today-stats-cell__panel-8">-</div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { WindowStats } from '@/types'
import { formatNumber, formatPoints, formatUSD } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    stats?: WindowStats | null
    loading?: boolean
    error?: string | null
  }>(),
  {
    stats: null,
    loading: false,
    error: null
  }
)

const { t } = useI18n()

// Format large token numbers (e.g., 1234567 -> 1.23M)
const formatTokens = (tokens: number): string => {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(2)}M`
  } else if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}K`
  }
  return tokens.toString()
}
</script>
