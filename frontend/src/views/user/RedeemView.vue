<template>
  <AppLayout>
    <div class="redeem-page">
      <!-- Redeem Form -->
      <section class="redeem-card card" :aria-label="t('redeem.title')">
        <header class="redeem-card__header">
          <div class="redeem-card__title-group">
            <span class="redeem-card__icon redeem-card__icon--primary">
              <Icon name="gift" size="md" />
            </span>
            <div>
              <h2 class="redeem-card__title">{{ t('redeem.title') }}</h2>
              <p class="redeem-card__description">{{ t('redeem.description') }}</p>
            </div>
          </div>
        </header>

        <form class="redeem-form" @submit.prevent="handleRedeem">
          <label for="code" class="input-label">
            {{ t('redeem.redeemCodeLabel') }}
          </label>
          <div class="redeem-form__row">
            <div class="redeem-form__field">
              <Icon name="gift" size="md" class="redeem-form__field-icon" />
              <input
                id="code"
                v-model="redeemCode"
                type="text"
                required
                :placeholder="t('redeem.redeemCodePlaceholder')"
                :disabled="submitting"
                class="redeem-form__input input"
              />
            </div>
            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="redeem-form__submit btn btn-primary"
            >
              <svg v-if="submitting" class="redeem-form__spinner" fill="none" viewBox="0 0 24 24">
                <circle class="redeem-form__spinner-track" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path
                  class="redeem-form__spinner-head"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="checkCircle" size="md" />
              {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
            </button>
          </div>
          <p class="input-hint">{{ t('redeem.redeemCodeHint') }}</p>
        </form>
      </section>

      <!-- Recent redemptions -->
      <section class="redeem-card card" :aria-label="t('redeem.recentActivity')">
        <header class="redeem-card__header">
          <div class="redeem-card__title-group">
            <span class="redeem-card__icon redeem-card__icon--neutral">
              <Icon name="clock" size="md" />
            </span>
            <h2 class="redeem-card__title">{{ t('redeem.recentActivity') }}</h2>
          </div>
          <span v-if="history.length > 0" class="badge badge-gray">{{ history.length }}</span>
        </header>

        <div class="redeem-card__body">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="redeem-history__state">
            <LoadingSpinner size="md" />
          </div>

          <!-- History List -->
          <div v-else-if="history.length > 0" class="redeem-history">
            <div v-for="item in history" :key="item.id" class="redeem-history__item">
              <div class="redeem-history__main">
                <span class="redeem-history__chip" :class="`redeem-history__chip--${historyTone(item)}`">
                  <Icon v-if="isBalanceType(item.type)" name="sparkles" size="md" />
                  <Icon v-else-if="isSubscriptionType(item.type)" name="badge" size="md" />
                  <Icon v-else name="bolt" size="md" />
                </span>
                <div class="redeem-history__text">
                  <p class="redeem-history__title">{{ getHistoryItemTitle(item) }}</p>
                  <p class="redeem-history__time">{{ formatDateTime(item.used_at) }}</p>
                </div>
              </div>
              <div class="redeem-history__meta">
                <p class="redeem-history__value" :class="`redeem-history__value--${historyTone(item)}`">
                  {{ formatHistoryValue(item) }}
                </p>
                <p class="redeem-history__code">{{ item.code.slice(0, 8) }}...</p>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state">
            <Icon name="clock" size="xl" class="empty-state-icon" />
            <p class="empty-state-description">{{ t('redeem.historyWillAppear') }}</p>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime, formatPoints } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const redeemCode = ref('')
const submitting = ref(false)

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

// 条目语义色:余额 = success/danger、订阅 = purple、并发 = info/warning
const historyTone = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) return item.value >= 0 ? 'success' : 'danger'
  if (isSubscriptionType(item.type)) return 'purple'
  return item.value >= 0 ? 'info' : 'warning'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${formatPoints(item.value)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = (await redeemAPI.getHistory()).filter((item) =>
      ['balance', 'concurrency', 'subscription'].includes(item.type)
    )
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('redeem.failedToRedeem'))
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchHistory()
})
</script>

<style scoped>
.redeem-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: 100%;
  max-width: 48rem;
  margin: 0 auto;
}

.redeem-card {
  overflow: hidden;
}

.redeem-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.redeem-card__title-group {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.75rem;
}

.redeem-card__icon {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-md);
}

.redeem-card__icon--primary {
  color: var(--color-text-brand);
  background: var(--color-primary-subtle);
}

.redeem-card__icon--neutral {
  color: var(--color-text-tertiary);
  background: var(--color-surface-muted);
}

.redeem-card__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.redeem-card__description {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.redeem-card__body {
  padding: 1.25rem 1.5rem;
}

/* ---- 兑换表单 ---- */
.redeem-form {
  padding: 1.25rem 1.5rem 1.5rem;
}

.redeem-form__row {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.25rem;
}

.redeem-form__field {
  position: relative;
  flex: 1;
  min-width: 0;
}

.redeem-form__field-icon {
  position: absolute;
  top: 50%;
  left: 1rem;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.redeem-form__input {
  padding-left: 3rem;
}

.redeem-form__submit {
  flex: 0 0 auto;
  min-width: 8rem;
}

.redeem-form__spinner {
  width: 1.25rem;
  height: 1.25rem;
  animation: redeem-spin 1s linear infinite;
}

.redeem-form__spinner-track {
  opacity: 0.25;
}

.redeem-form__spinner-head {
  opacity: 0.75;
}

@keyframes redeem-spin {
  to {
    transform: rotate(360deg);
  }
}

/* ---- 兑换记录 ---- */
.redeem-history {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.redeem-history__state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 6rem;
}

.redeem-history__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.875rem 1rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-interactive);
  box-shadow: var(--glass-shadow);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs)) saturate(var(--glass-saturate));
  transition: border-color 160ms ease, background-color 160ms ease;
}

.redeem-history__item:hover {
  border-color: var(--glass-border-hover);
  background: var(--glass-bg-interactive);
  box-shadow: var(--glass-shadow-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate-hover));
}

.dark .redeem-history__item {
  background: var(--glass-bg-interactive);
}

.dark .redeem-history__item:hover {
  background: var(--glass-bg-interactive);
}

.redeem-history__main {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.875rem;
}

.redeem-history__chip {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--radius-lg);
}

.redeem-history__chip--success { color: #059669; background: rgba(16, 185, 129, 0.14); }
.redeem-history__chip--danger { color: #dc2626; background: rgba(239, 68, 68, 0.14); }
.redeem-history__chip--purple { color: #7e22ce; background: rgba(168, 85, 247, 0.14); }
.redeem-history__chip--info { color: var(--color-text-brand); background: var(--color-primary-subtle); }
.redeem-history__chip--warning { color: #d97706; background: rgba(245, 158, 11, 0.16); }

.dark .redeem-history__chip--success { color: #34d399; background: rgba(6, 95, 70, 0.4); }
.dark .redeem-history__chip--danger { color: #f87171; background: rgba(153, 27, 27, 0.4); }
.dark .redeem-history__chip--purple { color: #c084fc; background: rgba(88, 28, 135, 0.4); }
.dark .redeem-history__chip--warning { color: #fbbf24; background: rgba(120, 53, 15, 0.4); }

.redeem-history__text {
  min-width: 0;
}

.redeem-history__title {
  margin: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.redeem-history__time {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.redeem-history__meta {
  flex: 0 0 auto;
  text-align: right;
}

.redeem-history__value {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.redeem-history__value--success { color: #059669; }
.redeem-history__value--danger { color: #dc2626; }
.redeem-history__value--purple { color: #7e22ce; }
.redeem-history__value--info { color: var(--color-text-brand); }
.redeem-history__value--warning { color: #d97706; }

.dark .redeem-history__value--success { color: #34d399; }
.dark .redeem-history__value--danger { color: #f87171; }
.dark .redeem-history__value--purple { color: #c084fc; }
.dark .redeem-history__value--warning { color: #fbbf24; }

.redeem-history__code {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--font-size-xs);
}

/* ---- 移动端 ---- */
@media (max-width: 560px) {
  .redeem-form__row {
    flex-direction: column;
  }

  .redeem-form__submit {
    width: 100%;
  }

  .redeem-card__header,
  .redeem-card__body,
  .redeem-form {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
