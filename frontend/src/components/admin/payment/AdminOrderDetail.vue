<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.orderDetail')"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="order" class="components-admin-payment-admin-order-detail__panel">
      <div class="components-admin-payment-admin-order-detail__panel-2">
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.orderId') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-2">#{{ order.id }}</p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.status') }}</p>
          <span :class="['badge', statusBadgeClass(order.status)]">
            {{ t('payment.status.' + order.status.toLowerCase(), order.status) }}
          </span>
        </div>
        <div v-if="order.order_type === 'balance'">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.baseAmount') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatCNY(baseAmount, localeCode) }}</p>
        </div>
        <div v-if="order.order_type === 'balance' && feeAmount > 0">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatCNY(feeAmount, localeCode) }}</p>
        </div>
        <div v-if="order.order_type === 'balance'">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.payAmount') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatCNY(order.pay_amount, localeCode) }}</p>
        </div>
        <div v-if="order.order_type === 'balance'">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.basePoints') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatPoints(basePoints, localeCode) }}</p>
        </div>
        <div v-if="order.order_type === 'balance' && bonusPoints > 0">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.bonusPoints') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatPoints(bonusPoints, localeCode) }}</p>
        </div>
        <div v-if="order.order_type === 'balance'">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.creditedPoints') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatPoints(creditedPoints, localeCode) }}</p>
        </div>
        <div v-else>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.pointsPaid') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-3">{{ formatPoints(subscriptionPoints, localeCode) }}</p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.paymentMethod') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">
            {{ t('payment.methods.' + order.payment_type, order.payment_type) }}
          </p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.admin.orderType') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">
            {{ t('payment.admin.' + order.order_type + 'Order', order.order_type) }}
          </p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.userId') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">#{{ order.user_id }}</p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.orders.createdAt') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">{{ formatDateTime(order.created_at) }}</p>
        </div>
        <div>
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.admin.expiresAt') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">{{ formatDateTime(order.expires_at) }}</p>
        </div>
        <div v-if="order.paid_at">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.admin.paidAt') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">{{ formatDateTime(order.paid_at) }}</p>
        </div>
        <div v-if="order.completed_at">
          <p class="components-admin-payment-admin-order-detail__description">{{ t('payment.admin.completedAt') }}</p>
          <p class="components-admin-payment-admin-order-detail__description-4">{{ formatDateTime(order.completed_at) }}</p>
        </div>
      </div>

      <div
        v-if="hasRefundSnapshot"
        class="components-admin-payment-admin-order-detail__panel-3"
      >
        <h4 class="components-admin-payment-admin-order-detail__heading">
          {{ t('payment.admin.refundInfo') }}
        </h4>
        <div class="components-admin-payment-admin-order-detail__panel-4">
          <div v-if="refundedGatewayAmount > 0">
            <span class="components-admin-payment-admin-order-detail__text">{{ t('payment.admin.gatewayRefundAmount') }}:</span>
            <span class="components-admin-payment-admin-order-detail__text-2">{{ formatCNY(refundedGatewayAmount, localeCode) }}</span>
          </div>
          <div v-if="reversedPoints > 0">
            <span class="components-admin-payment-admin-order-detail__text">{{ t('payment.admin.reversedPoints') }}:</span>
            <span class="components-admin-payment-admin-order-detail__text-2">{{ formatPoints(reversedPoints, localeCode) }}</span>
          </div>
          <div v-if="order.refund_reason" class="components-admin-payment-admin-order-detail__panel-5">
            <span class="components-admin-payment-admin-order-detail__text">{{ t('payment.admin.refundReason') }}:</span>
            <span class="components-admin-payment-admin-order-detail__text-3">{{ order.refund_reason }}</span>
          </div>
        </div>
      </div>

      <div class="components-admin-payment-admin-order-detail__panel-6">
        <button
          v-if="order.status === 'PENDING'"
          @click="emit('cancel', order)"
          class="components-admin-payment-admin-order-detail__action btn btn-sm"
        >
          {{ t('payment.orders.cancel') }}
        </button>
        <button
          v-if="order.status === 'FAILED'"
          @click="emit('retry', order)"
          class="btn btn-sm btn-secondary"
        >
          {{ t('payment.admin.retry') }}
        </button>
        <button
          v-if="canRefund(order)"
          @click="emit('refund', order)"
          class="components-admin-payment-admin-order-detail__action-2 btn btn-sm"
        >
          {{ t('payment.admin.refund') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import {
  statusBadgeClass,
  canDirectRefund,
  formatOrderDateTime,
  rechargeBasePoints,
  rechargeBonusPoints,
  rechargeCreditedPoints,
  rechargeFeeAmount,
  rechargePrincipalAmount,
} from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
}>()

const baseAmount = computed(() => props.order ? rechargePrincipalAmount(props.order) : 0)
const feeAmount = computed(() => props.order ? rechargeFeeAmount(props.order) : 0)
const basePoints = computed(() => props.order ? rechargeBasePoints(props.order) : 0)
const bonusPoints = computed(() => props.order ? rechargeBonusPoints(props.order) : 0)
const creditedPoints = computed(() => props.order ? rechargeCreditedPoints(props.order) : 0)
const subscriptionPoints = computed(() => {
  if (!props.order) return 0
  return props.order.wallet_amount > 0 ? props.order.wallet_amount : props.order.amount
})
const refundedGatewayAmount = computed(() => props.order?.refunded_gateway_amount ?? 0)
const reversedPoints = computed(() => {
  if (!props.order) return 0
  return (props.order.reversed_base_points ?? 0) + (props.order.reversed_bonus_points ?? 0)
})
const hasRefundSnapshot = computed(() => refundedGatewayAmount.value > 0
  || reversedPoints.value > 0
  || !!props.order?.refund_reason)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
}>()

function canRefund(order: PaymentOrder): boolean {
  return canDirectRefund(order)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>
