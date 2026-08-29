<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.refundOrder')"
    width="normal"
    @close="emit('cancel')"
  >
    <form id="refund-form" @submit.prevent="handleSubmit" class="components-admin-payment-admin-refund-dialog__form">
      <!-- Refund Request Info -->
      <div
        v-if="order?.refund_requested_at || order?.refund_request_reason"
        class="components-admin-payment-admin-refund-dialog__panel"
      >
        <div class="components-admin-payment-admin-refund-dialog__panel-2">
          <svg class="components-admin-payment-admin-refund-dialog__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ t('payment.admin.refundRequestInfo') }}
        </div>
        <div v-if="order?.refund_requested_at" class="components-admin-payment-admin-refund-dialog__panel-3">
          <span class="components-admin-payment-admin-refund-dialog__text">{{ t('payment.admin.refundRequestedAt') }}</span>
          <span class="components-admin-payment-admin-refund-dialog__text-2">{{ formatDateTime(order.refund_requested_at) }}</span>
        </div>
        <div v-if="order?.refund_request_reason" class="components-admin-payment-admin-refund-dialog__panel-4">
          <span class="components-admin-payment-admin-refund-dialog__text">{{ t('payment.admin.refundRequestReason') }}:</span>
          <span class="components-admin-payment-admin-refund-dialog__text-3">{{ order.refund_request_reason }}</span>
        </div>
      </div>

      <!-- Order Info -->
      <div class="components-admin-payment-admin-refund-dialog__panel-5">
        <div class="components-admin-payment-admin-refund-dialog__panel-6">
          <span class="components-admin-payment-admin-refund-dialog__text-4">{{ t('payment.orders.orderId') }}</span>
          <span class="components-admin-payment-admin-refund-dialog__text-5">#{{ order?.id }}</span>
        </div>
        <div class="components-admin-payment-admin-refund-dialog__panel-7">
          <span class="components-admin-payment-admin-refund-dialog__text-4">{{ t('payment.orders.creditedPoints') }}</span>
          <span class="components-admin-payment-admin-refund-dialog__text-6">{{ formatPoints(orderCreditedPoints, localeCode) }}</span>
        </div>
        <div class="components-admin-payment-admin-refund-dialog__panel-7">
          <span class="components-admin-payment-admin-refund-dialog__text-4">{{ t('payment.orders.payAmount') }}</span>
          <span class="components-admin-payment-admin-refund-dialog__text-6">{{ formatCNY(order?.pay_amount ?? 0, localeCode) }}</span>
        </div>
        <div v-if="actuallyRefunded > 0" class="components-admin-payment-admin-refund-dialog__panel-7">
          <span class="components-admin-payment-admin-refund-dialog__text-4">{{ t('payment.admin.alreadyRefunded') }}</span>
          <span class="components-admin-payment-admin-refund-dialog__text-7">{{ formatCNY(actuallyRefunded, localeCode) }}</span>
        </div>
      </div>

      <div class="components-admin-payment-admin-refund-dialog__panel-9">
        <div class="components-admin-payment-admin-refund-dialog__panel-10">
          <div class="components-admin-payment-admin-refund-dialog__text-4">{{ t('payment.admin.pointsToReverse') }}</div>
          <div class="components-admin-payment-admin-refund-dialog__panel-11">{{ formatPoints(pointsToReverse, localeCode) }}</div>
        </div>
        <p class="components-admin-payment-admin-refund-dialog__text-8">{{ t('payment.admin.pointsReverseAutomaticHint') }}</p>
      </div>

      <!-- Refund Amount -->
      <div>
        <label class="input-label">{{ t('payment.admin.refundAmount') }}</label>
        <div class="components-admin-payment-admin-refund-dialog__panel-14">
          <span class="components-admin-payment-admin-refund-dialog__text-9">¥</span>
          <input
            v-model.number="form.principal_amount"
            type="number"
            step="0.01"
            min="0.01"
            :max="maxRefundable"
            class="components-admin-payment-admin-refund-dialog__field-2 input"
            required
          />
        </div>
        <p class="components-admin-payment-admin-refund-dialog__description">
          {{ t('payment.admin.maxRefundable') }}: {{ formatCNY(maxRefundable, localeCode) }}
        </p>
      </div>

      <!-- Reason -->
      <div>
        <label class="input-label">{{ t('payment.admin.refundReason') }}</label>
        <textarea
          v-model="form.reason"
          rows="3"
          class="input"
          :placeholder="t('payment.admin.refundReasonPlaceholder')"
          required
        ></textarea>
      </div>

    </form>

    <template #footer>
      <div class="components-admin-payment-admin-refund-dialog__panel-16">
        <button type="button" @click="emit('cancel')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="refund-form"
          :disabled="submitting || form.principal_amount <= 0"
          class="components-admin-payment-admin-refund-dialog__element"
        >
          {{ submitting ? t('common.processing') : t('payment.admin.confirmRefund') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import { formatOrderDateTime, rechargeCreditedPoints, rechargePrincipalAmount } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'confirm', data: { principal_amount: number; reason: string; idempotency_key: string }): void
  (e: 'cancel'): void
}>()

const form = reactive({
  principal_amount: 0,
  reason: '',
})
const idempotencyKey = ref('')

// In REFUND_REQUESTED / REFUND_PENDING status, refund_amount is requested/pending, not actually refunded.
// Only PARTIALLY_REFUNDED / REFUNDED have real refund amounts.
const actuallyRefunded = computed(() => {
  if (!props.order) return 0
  if (props.order.refunded_principal_amount != null) return props.order.refunded_principal_amount
  const s = props.order.status
  if (s === 'PARTIALLY_REFUNDED' || s === 'REFUNDED') return props.order.refund_amount || 0
  return 0
})

const maxRefundable = computed(() => {
  if (!props.order) return 0
  return Math.max(rechargePrincipalAmount(props.order) - actuallyRefunded.value, 0)
})
const orderCreditedPoints = computed(() => props.order ? rechargeCreditedPoints(props.order) : 0)
const pointsToReverse = computed(() => {
  if (!props.order) return 0
  const principal = rechargePrincipalAmount(props.order)
  if (principal <= 0) return 0
  return orderCreditedPoints.value * Math.min(Math.max(form.principal_amount / principal, 0), 1)
})

function createAdminRefundIdempotencyKey(orderId: string): string {
  const randomPart = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`
  return `admin-refund-${orderId}-${randomPart}`
}

watch(() => props.show, (val) => {
  if (val && props.order) {
    idempotencyKey.value = createAdminRefundIdempotencyKey(props.order.id)
    // For REFUND_REQUESTED, pre-fill with the requested amount
    if (props.order.status === 'REFUND_REQUESTED' && props.order.refund_amount) {
      form.principal_amount = props.order.refund_amount
    } else {
      form.principal_amount = maxRefundable.value
    }
    form.reason = props.order.refund_request_reason || ''
  } else if (!val) {
    idempotencyKey.value = ''
  }
}, { immediate: true })

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}

function handleSubmit() {
  if (form.principal_amount <= 0 || form.principal_amount > maxRefundable.value || !props.order) return
  if (!idempotencyKey.value) idempotencyKey.value = createAdminRefundIdempotencyKey(props.order.id)
  emit('confirm', { ...form, idempotency_key: idempotencyKey.value })
}
</script>
