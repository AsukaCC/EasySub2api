<template>
  <div class="stat-grid">
    <div v-if="!isSimple" class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--balance">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
          </svg>
        </div>
        <div>
          <p class="stat-card__label">{{ t('dashboard.balance') }}</p>
          <p class="stat-card__value">{{ formatPoints(balance) }}</p>
          <p class="stat-card__hint">{{ t('common.available') }}</p>
          <p class="stat-card__hint">{{ t('common.bonusBalance') }}: {{ formatPoints(bonusBalance) }}</p>
        </div>
      </div>
    </div>

    <div class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--keys">
          <Icon name="key" size="md" :stroke-width="2" />
        </div>
        <div>
          <p class="stat-card__label">{{ t('dashboard.apiKeys') }}</p>
          <p class="stat-card__value">{{ stats?.total_api_keys || 0 }}</p>
          <p class="stat-card__hint">{{ stats?.active_api_keys || 0 }} {{ t('common.active') }}</p>
        </div>
      </div>
    </div>

    <div class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--bolt">
          <Icon name="bolt" size="md" :stroke-width="2" />
        </div>
        <div>
          <p class="stat-card__label">{{ t('dashboard.concurrency') }}</p>
          <p class="stat-card__value">
            {{ formatNumber(stats?.current_concurrency || 0) }} / {{ formatNumber(stats?.concurrency || 0) }}
          </p>
          <p class="stat-card__hint">{{ t('dashboard.concurrencyUsage') }}</p>
        </div>
      </div>
    </div>
  </div>

  <div class="stat-grid">
    <div class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--cube">
          <Icon name="cube" size="md" :stroke-width="2" />
        </div>
        <div>
          <p class="stat-card__label">{{ t('dashboard.todayTokens') }}</p>
          <p class="stat-card__value">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <p class="stat-card__hint">
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }} /
            {{ t('dashboard.cache') }}: {{ formatTokens((stats?.today_cache_creation_tokens || 0) + (stats?.today_cache_read_tokens || 0)) }}
          </p>
        </div>
      </div>
    </div>

    <div class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--data">
          <Icon name="database" size="md" :stroke-width="2" />
        </div>
        <div>
          <p class="stat-card__label">{{ t('dashboard.totalTokens') }}</p>
          <p class="stat-card__value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="stat-card__hint">
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }} /
            {{ t('dashboard.cache') }}: {{ formatTokens((stats?.total_cache_creation_tokens || 0) + (stats?.total_cache_read_tokens || 0)) }}
          </p>
        </div>
      </div>
    </div>

    <div class="stat-card card">
      <div class="stat-card__row">
        <div class="stat-card__icon stat-card__icon--bolt">
          <Icon name="bolt" size="md" :stroke-width="2" />
        </div>
        <div class="stat-card__metrics">
          <p class="stat-card__label">{{ t('dashboard.performance') }}</p>
          <div class="stat-card__metric">
            <p class="stat-card__value">{{ formatTokens(stats?.rpm || 0) }}</p>
            <span class="stat-card__hint">RPM</span>
          </div>
          <div class="stat-card__metric">
            <p class="stat-card__value">{{ formatTokens(stats?.tpm || 0) }}</p>
            <span class="stat-card__hint">TPM</span>
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
  bonusBalance: number
  isSimple: boolean
}>()
const { t } = useI18n()

const formatNumber = (n: number) => n.toLocaleString()
const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toString()
}
</script>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
}

@media (min-width: 640px) {
  .stat-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .stat-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.stat-card {
  padding: 0.75rem 0.875rem;
}

.stat-card__row {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
}

.stat-card__icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-lg);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.stat-card__icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.stat-card__label {
  margin: 0;
  color: var(--color-text-tertiary);
  font-size: var(--type-caption-size);
}

.stat-card__value {
  margin: 0.125rem 0 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  line-height: 1.5rem;
}

.stat-card__hint {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--type-caption-size);
}

.stat-card__metrics {
  min-width: 0;
}

.stat-card__metric {
  display: flex;
  align-items: baseline;
  gap: 0.375rem;
}
</style>
