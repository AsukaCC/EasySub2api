<template>
  <span
    class="components-payment-order-status-badge__text"
    :class="statusClass"
  >
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OrderStatus } from '@/types/payment'

const props = defineProps<{
  status: OrderStatus
}>()

const { t } = useI18n()

const statusMap: Record<OrderStatus, { key: string; class: string }> = {
  PENDING: { key: 'payment.status.pending', class: 'components-payment-order-status-badge__render' },
  PAID: { key: 'payment.status.paid', class: 'components-payment-order-status-badge__render-2' },
  RECHARGING: { key: 'payment.status.recharging', class: 'components-payment-order-status-badge__render-2' },
  COMPLETED: { key: 'payment.status.completed', class: 'components-payment-order-status-badge__render-3' },
  EXPIRED: { key: 'payment.status.expired', class: 'components-payment-order-status-badge__render-4' },
  CANCELLED: { key: 'payment.status.cancelled', class: 'components-payment-order-status-badge__render-4' },
  FAILED: { key: 'payment.status.failed', class: 'components-payment-order-status-badge__render-5' },
  REFUND_REQUESTED: { key: 'payment.status.refund_requested', class: 'components-payment-order-status-badge__render-6' },
  REFUNDING: { key: 'payment.status.refunding', class: 'components-payment-order-status-badge__render-6' },
  REFUND_PENDING: { key: 'payment.status.refund_pending', class: 'components-payment-order-status-badge__render-6' },
  REFUNDED: { key: 'payment.status.refunded', class: 'components-payment-order-status-badge__render-7' },
  PARTIALLY_REFUNDED: { key: 'payment.status.partially_refunded', class: 'components-payment-order-status-badge__render-7' },
  REFUND_FAILED: { key: 'payment.status.refund_failed', class: 'components-payment-order-status-badge__render-5' },
}

const statusLabel = computed(() => {
  const entry = statusMap[props.status]
  return entry ? t(entry.key) : props.status
})

const statusClass = computed(() => {
  const entry = statusMap[props.status]
  return entry?.class ?? 'components-payment-order-status-badge__render-4'
})
</script>
