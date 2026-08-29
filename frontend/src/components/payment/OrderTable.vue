<template>
  <DataTable :columns="columns" :data="orders" :loading="loading">
    <template #cell-id="{ value }">
      <span class="components-payment-order-table__text">#{{ value }}</span>
    </template>
    <template #cell-out_trade_no="{ value }">
      <span class="components-payment-order-table__text-2">{{ value }}</span>
    </template>
    <template v-if="showUser" #cell-user_email="{ value, row }">
      <div class="components-payment-order-table__panel">
        <span class="components-payment-order-table__text-3">{{ value || row.user_name || '#' + row.user_id }}</span>
        <span v-if="row.user_notes" class="components-payment-order-table__text-4">({{ row.user_notes }})</span>
      </div>
    </template>
    <template #cell-pay_amount="{ row }">
      <div class="components-payment-order-table__panel">
        <span class="components-payment-order-table__text-5">
          {{ row.order_type === 'balance' ? formatCNY(row.pay_amount, localeCode) : formatPoints(subscriptionPoints(row), localeCode) }}
        </span>
        <span v-if="row.order_type === 'balance' && row.fee_rate > 0" class="components-payment-order-table__text-4" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
          ({{ t('payment.orders.fee') }} {{ row.fee_rate }}%)
        </span>
        <div v-if="row.order_type === 'balance'" class="components-payment-order-table__panel-2">
          {{ t('payment.orders.creditedPoints') }}: {{ formatPoints(rechargeCreditedPoints(row), localeCode) }}
        </div>
      </div>
    </template>
    <template #cell-payment_type="{ value }">
      <span class="components-payment-order-table__text-6">{{ t('payment.methods.' + value, value) }}</span>
    </template>
    <template #cell-status="{ value }">
      <OrderStatusBadge :status="value" />
    </template>
    <template #cell-created_at="{ value }">
      <span class="components-payment-order-table__text-7">{{ formatDate(value) }}</span>
    </template>
    <template #cell-actions="{ row }">
      <slot name="actions" :row="row" />
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import { rechargeCreditedPoints } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))

const props = defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  showUser?: boolean
}>()

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString() }

function subscriptionPoints(order: PaymentOrder): number {
  return order.wallet_amount > 0 ? order.wallet_amount : order.amount
}

const columns = computed((): Column[] => {
  const cols: Column[] = [
    { key: 'id', label: t('payment.orders.orderId') },
    { key: 'out_trade_no', label: t('payment.orders.orderNo') },
  ]
  if (props.showUser) {
    cols.push({ key: 'user_email', label: t('payment.admin.colUser') })
  }
  cols.push(
    { key: 'pay_amount', label: t('payment.orders.payAmount') },
    { key: 'payment_type', label: t('payment.orders.paymentMethod') },
    { key: 'status', label: t('payment.orders.status') },
    { key: 'created_at', label: t('payment.orders.createdAt') },
    { key: 'actions', label: t('common.actions') },
  )
  return cols
})
</script>
