<template>
  <AppLayout>
    <div class="views-admin-orders-admin-payment-dashboard-view__panel">
      <!-- Header with Day Switcher -->
      <div class="views-admin-orders-admin-payment-dashboard-view__panel-2">
        <div class="views-admin-orders-admin-payment-dashboard-view__panel-3">
          <div class="views-admin-orders-admin-payment-dashboard-view__panel-4">
            <button
              v-for="d in DAYS_OPTIONS"
              :key="d"
              type="button"
              class="views-admin-orders-admin-payment-dashboard-view__action"
              :class="days === d
                ? 'views-admin-orders-admin-payment-dashboard-view__action-2'
                : 'views-admin-orders-admin-payment-dashboard-view__action-3'"
              @click="days = d"
            >
              {{ d }}{{ t('payment.admin.daySuffix') }}
            </button>
          </div>
          <button @click="loadDashboard" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
            <Icon name="refresh" size="md" :class="loading ? 'views-admin-orders-admin-payment-dashboard-view__icon' : ''" />
          </button>
        </div>
      </div>

      <!-- Dashboard Content -->
      <LoadingState v-if="loading" variant="page" />
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="views-admin-orders-admin-payment-dashboard-view__panel-6">
          <div class="views-admin-orders-admin-payment-dashboard-view__panel-7 card">
            <h3 class="views-admin-orders-admin-payment-dashboard-view__heading">{{ t('payment.admin.paymentDistribution') }}</h3>
            <div v-if="!stats.payment_methods?.length" class="views-admin-orders-admin-payment-dashboard-view__panel-8">{{ t('payment.admin.noData') }}</div>
            <div v-else class="views-admin-orders-admin-payment-dashboard-view__panel-9">
              <div v-for="method in stats.payment_methods" :key="method.type" class="views-admin-orders-admin-payment-dashboard-view__panel-10">
                <div class="views-admin-orders-admin-payment-dashboard-view__panel-3">
                  <span :class="['views-admin-orders-admin-payment-dashboard-view__text-5', methodColor(method.type)]"></span>
                  <span class="views-admin-orders-admin-payment-dashboard-view__text">{{ t('payment.methods.' + method.type, method.type) }}</span>
                </div>
                <div class="views-admin-orders-admin-payment-dashboard-view__panel-11">
                  <span v-for="[currency, amount] in sortedAmounts(method.amount)" :key="currency" class="views-admin-orders-admin-payment-dashboard-view__text-2">{{ formatMoney(currency, amount) }}</span>
                  <span class="views-admin-orders-admin-payment-dashboard-view__text-3">({{ method.count }})</span>
                </div>
              </div>
            </div>
          </div>
          <div class="views-admin-orders-admin-payment-dashboard-view__panel-7 card">
            <h3 class="views-admin-orders-admin-payment-dashboard-view__heading">{{ t('payment.admin.topUsers') }}</h3>
            <div v-if="!hasTopUsers(stats.top_users)" class="views-admin-orders-admin-payment-dashboard-view__panel-8">{{ t('payment.admin.noData') }}</div>
            <div v-else class="views-admin-orders-admin-payment-dashboard-view__panel-12">
              <div v-for="[currency, users] in sortedTopUsers(stats.top_users)" :key="currency" class="views-admin-orders-admin-payment-dashboard-view__panel-12">
                <p class="views-admin-orders-admin-payment-dashboard-view__description">{{ currency }}</p>
                <div v-for="(user, idx) in users" :key="user.user_id" class="views-admin-orders-admin-payment-dashboard-view__panel-13">
                  <div class="views-admin-orders-admin-payment-dashboard-view__panel-14">
                    <span :class="['views-admin-orders-admin-payment-dashboard-view__text-6', rankClass(idx)]">{{ idx + 1 }}</span>
                    <span class="views-admin-orders-admin-payment-dashboard-view__text">{{ user.email }}</span>
                  </div>
                  <span class="views-admin-orders-admin-payment-dashboard-view__text-4">{{ formatMoney(currency, user.amount) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { CurrencyAmounts, DashboardStats, TopUserPaymentStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

function methodColor(type: string): string {
  const c: Record<string, string> = {
    alipay: 'status-fill--info', wxpay: 'status-fill--success',
    alipay_direct: 'status-fill--info', wxpay_direct: 'status-fill--success',
    stripe: 'status-fill--accent',
  }
  return c[type] || 'status-fill--neutral'
}

function rankClass(idx: number): string {
  if (idx === 0) return 'views-admin-orders-admin-payment-dashboard-view__state'
  if (idx === 1) return 'views-admin-orders-admin-payment-dashboard-view__state-2'
  if (idx === 2) return 'views-admin-orders-admin-payment-dashboard-view__state-3'
  return 'views-admin-orders-admin-payment-dashboard-view__state-4'
}

function sortedAmounts(amounts: CurrencyAmounts): [string, number][] {
  return Object.entries(amounts).sort(([left], [right]) => left.localeCompare(right))
}

function sortedTopUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): [string, TopUserPaymentStats[]][] {
  return Object.entries(usersByCurrency).sort(([left], [right]) => left.localeCompare(right))
}

function hasTopUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): boolean {
  return Object.values(usersByCurrency).some(users => users.length > 0)
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}

async function loadDashboard() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getDashboard(days.value)
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

watch(days, () => loadDashboard())
onMounted(() => loadDashboard())
</script>
