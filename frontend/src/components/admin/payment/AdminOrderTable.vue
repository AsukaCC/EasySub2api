<template>
  <div class="components-admin-payment-admin-order-table__panel">
    <div class="components-admin-payment-admin-order-table__panel-2 card">
      <div class="components-admin-payment-admin-order-table__panel-3">
        <div class="components-admin-payment-admin-order-table__panel-4">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('payment.admin.searchOrders')"
            class="input"
            @input="handleSearch"
          />
        </div>
        <Select
          v-model="filters.status"
          :options="statusFilterOptions"
          class="components-admin-payment-admin-order-table__field"
          @change="emitFiltersChanged"
        />
        <Select
          v-model="filters.payment_type"
          :options="paymentTypeFilterOptions"
          class="components-admin-payment-admin-order-table__field-2"
          @change="emitFiltersChanged"
        />
        <Select
          v-model="filters.order_type"
          :options="orderTypeFilterOptions"
          class="components-admin-payment-admin-order-table__field"
          @change="emitFiltersChanged"
        />
        <div class="components-admin-payment-admin-order-table__panel-5">
          <button
            @click="emit('refresh')"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'components-admin-payment-admin-order-table__icon' : ''" />
          </button>
        </div>
      </div>
    </div>

    <DataTable :columns="columns" :data="orders" :loading="loading">
      <template #cell-id="{ value }">
        <span class="components-admin-payment-admin-order-table__text">#{{ value }}</span>
      </template>

      <template #cell-user_id="{ value }">
        <span class="components-admin-payment-admin-order-table__text-2">#{{ value }}</span>
      </template>

      <template #cell-pay_amount="{ row }">
        <div class="components-admin-payment-admin-order-table__panel-6">
          <span class="components-admin-payment-admin-order-table__text-3">
            {{ row.order_type === 'balance' ? formatCNY(row.pay_amount, localeCode) : formatPoints(subscriptionPoints(row), localeCode) }}
          </span>
          <span v-if="row.order_type === 'balance' && row.fee_rate > 0" class="components-admin-payment-admin-order-table__text-4" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
            ({{ row.fee_rate }}%)
          </span>
          <div v-if="row.order_type === 'balance'" class="components-admin-payment-admin-order-table__panel-7">
            {{ t('payment.orders.creditedPoints') }}: {{ formatPoints(rechargeCreditedPoints(row), localeCode) }}
          </div>
        </div>
      </template>

      <template #cell-payment_type="{ value }">
        <span class="components-admin-payment-admin-order-table__text-5">
          {{ t('payment.methods.' + value, value) }}
        </span>
      </template>

      <template #cell-status="{ value }">
        <span :class="['badge', statusBadgeClass(value)]">
          {{ t('payment.status.' + value.toLowerCase(), value) }}
        </span>
      </template>

      <template #cell-order_type="{ value }">
        <span class="components-admin-payment-admin-order-table__text-5">
          {{ t('payment.admin.' + value + 'Order', value) }}
        </span>
      </template>

      <template #cell-created_at="{ value }">
        <span class="components-admin-payment-admin-order-table__text-6">{{ formatDateTime(value) }}</span>
      </template>

      <template #cell-actions="{ row }">
        <div class="components-admin-payment-admin-order-table__panel-8">
          <button
            @click="emit('detail', row)"
            class="components-admin-payment-admin-order-table__action"
          >
            <Icon name="eye" size="sm" />
            <span class="components-admin-payment-admin-order-table__text-7">{{ t('common.view') }}</span>
          </button>
          <button
            v-if="row.status === 'PENDING'"
            @click="emit('cancel', row)"
            class="components-admin-payment-admin-order-table__action-2"
          >
            <Icon name="x" size="sm" />
            <span class="components-admin-payment-admin-order-table__text-7">{{ t('payment.orders.cancel') }}</span>
          </button>
          <button
            v-if="row.status === 'FAILED'"
            @click="emit('retry', row)"
            class="components-admin-payment-admin-order-table__action-3"
          >
            <Icon name="refresh" size="sm" />
            <span class="components-admin-payment-admin-order-table__text-7">{{ t('payment.admin.retry') }}</span>
          </button>
          <button
            v-if="canRefundRow(row)"
            @click="emit('refund', row)"
            class="components-admin-payment-admin-order-table__action-4"
          >
            <Icon name="dollar" size="sm" />
            <span class="components-admin-payment-admin-order-table__text-7">{{ t('payment.admin.refund') }}</span>
          </button>
        </div>
      </template>
    </DataTable>

    <Pagination
      v-if="total > 0"
      :page="page"
      :total="total"
      :page-size="pageSize"
      @update:page="emit('update:page', $event)"
      @update:pageSize="emit('update:pageSize', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { statusBadgeClass, canDirectRefund, formatOrderDateTime, rechargeCreditedPoints } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))

defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'detail', order: PaymentOrder): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
  (e: 'refresh'): void
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', size: number): void
  (e: 'filter', filters: { keyword?: string; status?: string; payment_type?: string; order_type?: string }): void
}>()

const searchQuery = ref('')
const filters = reactive({ status: '', payment_type: '', order_type: '' })
function subscriptionPoints(order: PaymentOrder): number {
  return order.wallet_amount > 0 ? order.wallet_amount : order.amount
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function handleSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => emitFiltersChanged(), 300)
}

function emitFiltersChanged() {
  emit('filter', {
    keyword: searchQuery.value || undefined,
    status: filters.status || undefined,
    payment_type: filters.payment_type || undefined,
    order_type: filters.order_type || undefined,
  })
}

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('payment.orders.orderId') },
  { key: 'user_id', label: t('payment.orders.userId') },
  { key: 'pay_amount', label: t('payment.orders.payAmount') },
  { key: 'payment_type', label: t('payment.orders.paymentMethod') },
  { key: 'status', label: t('payment.orders.status') },
  { key: 'order_type', label: t('payment.orders.orderType') },
  { key: 'created_at', label: t('payment.orders.createdAt') },
  { key: 'actions', label: t('payment.orders.actions') },
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_PENDING', label: t('payment.status.refund_pending') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

function canRefundRow(order: PaymentOrder): boolean {
  return canDirectRefund(order)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>
