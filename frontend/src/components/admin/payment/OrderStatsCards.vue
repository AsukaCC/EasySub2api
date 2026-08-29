<template>
  <div class="components-admin-payment-order-stats-cards__panel">
    <!-- Today Revenue -->
    <div class="components-admin-payment-order-stats-cards__panel-2 card">
      <div class="components-admin-payment-order-stats-cards__panel-3">
        <div class="components-admin-payment-order-stats-cards__panel-4">
          <Icon name="dollar" size="md" class="components-admin-payment-order-stats-cards__icon" :stroke-width="2" />
        </div>
        <div>
          <p class="components-admin-payment-order-stats-cards__description">{{ t('payment.admin.todayRevenue') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.today_amount)" :key="currency" class="components-admin-payment-order-stats-cards__description-2">
            {{ formatMoney(currency, amount) }}
          </p>
          <p class="components-admin-payment-order-stats-cards__description-3">
            {{ stats.today_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Total Revenue -->
    <div class="components-admin-payment-order-stats-cards__panel-2 card">
      <div class="components-admin-payment-order-stats-cards__panel-3">
        <div class="components-admin-payment-order-stats-cards__panel-5">
          <Icon name="creditCard" size="md" class="components-admin-payment-order-stats-cards__icon-2" :stroke-width="2" />
        </div>
        <div>
          <p class="components-admin-payment-order-stats-cards__description">{{ t('payment.admin.totalRevenue') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.total_amount)" :key="currency" class="components-admin-payment-order-stats-cards__description-2">
            {{ formatMoney(currency, amount) }}
          </p>
          <p class="components-admin-payment-order-stats-cards__description-3">
            {{ stats.total_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Today Orders -->
    <div class="components-admin-payment-order-stats-cards__panel-2 card">
      <div class="components-admin-payment-order-stats-cards__panel-3">
        <div class="components-admin-payment-order-stats-cards__panel-6">
          <Icon name="chart" size="md" class="components-admin-payment-order-stats-cards__icon-3" :stroke-width="2" />
        </div>
        <div>
          <p class="components-admin-payment-order-stats-cards__description">{{ t('payment.admin.todayOrders') }}</p>
          <p class="components-admin-payment-order-stats-cards__description-2">{{ stats.today_count }}</p>
        </div>
      </div>
    </div>

    <!-- Average Amount -->
    <div class="components-admin-payment-order-stats-cards__panel-2 card">
      <div class="components-admin-payment-order-stats-cards__panel-3">
        <div class="components-admin-payment-order-stats-cards__panel-7">
          <Icon name="chart" size="md" class="components-admin-payment-order-stats-cards__icon-4" :stroke-width="2" />
        </div>
        <div>
          <p class="components-admin-payment-order-stats-cards__description">{{ t('payment.admin.avgAmount') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.avg_amount)" :key="currency" class="components-admin-payment-order-stats-cards__description-2">
            {{ formatMoney(currency, amount) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { CurrencyAmounts, DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function sortedAmounts(amounts: CurrencyAmounts): [string, number][] {
  return Object.entries(amounts).sort(([left], [right]) => left.localeCompare(right))
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}
</script>
