<template>
  <div class="components-admin-payment-payment-method-chart__panel card">
    <h3 class="components-admin-payment-payment-method-chart__heading">
      {{ t('payment.admin.paymentDistribution') }}
    </h3>
    <div
      v-if="!methods?.length"
      class="components-admin-payment-payment-method-chart__panel-2"
    >
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else class="components-admin-payment-payment-method-chart__panel-3">
      <div v-for="method in methods" :key="method.type" class="components-admin-payment-payment-method-chart__panel-4">
        <div class="components-admin-payment-payment-method-chart__panel-5">
          <div class="components-admin-payment-payment-method-chart__panel-6">
            <span :class="['components-admin-payment-payment-method-chart__text-5', colorMap[method.type] || 'components-admin-payment-payment-method-chart__text-6']"></span>
            <span class="components-admin-payment-payment-method-chart__text">
              {{ t('payment.methods.' + method.type, method.type) }}
            </span>
          </div>
          <div class="components-admin-payment-payment-method-chart__panel-7">
            <span v-for="[currency, amount] in sortedAmounts(method.amount)" :key="currency" class="components-admin-payment-payment-method-chart__text-2">
              {{ formatMoney(currency, amount) }}
            </span>
            <span class="components-admin-payment-payment-method-chart__text-3">
              ({{ method.count }})
            </span>
          </div>
        </div>
        <div v-for="[currency, amount] in sortedAmounts(method.amount)" :key="currency" class="components-admin-payment-payment-method-chart__panel-6">
          <span class="components-admin-payment-payment-method-chart__text-4">{{ currency }}</span>
          <div class="components-admin-payment-payment-method-chart__panel-8">
            <div
              :class="['components-admin-payment-payment-method-chart__panel-9', barColorMap[method.type] || 'components-admin-payment-payment-method-chart__text-6']"
              :style="{ width: barWidth(currency, amount) + '%' }"
            ></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CurrencyAmounts, PaymentMethodStats } from '@/types/payment'

const { t } = useI18n()

const props = defineProps<{
  methods: PaymentMethodStats[]
}>()

const colorMap: Record<string, string> = {
  alipay: 'status-fill--info',
  wxpay: 'status-fill--success',
  alipay_direct: 'status-fill--info',
  wxpay_direct: 'status-fill--success',
  stripe: 'status-fill--accent',
}

const barColorMap: Record<string, string> = {
  alipay: 'status-fill--info',
  wxpay: 'status-fill--success',
  alipay_direct: 'status-fill--info',
  wxpay_direct: 'status-fill--success',
  stripe: 'status-fill--accent',
}

const maxAmounts = computed<CurrencyAmounts>(() => {
  return props.methods.reduce<CurrencyAmounts>((maximums, method) => {
    for (const [currency, amount] of Object.entries(method.amount)) {
      maximums[currency] = Math.max(maximums[currency] || 0, amount)
    }
    return maximums
  }, {})
})

function sortedAmounts(amounts: CurrencyAmounts): [string, number][] {
  return Object.entries(amounts).sort(([left], [right]) => left.localeCompare(right))
}

function barWidth(currency: string, amount: number): number {
  return Math.min((amount / (maxAmounts.value[currency] || 1)) * 100, 100)
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}
</script>
