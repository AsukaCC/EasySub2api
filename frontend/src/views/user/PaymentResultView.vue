<template>
  <div class="views-user-payment-result-view__panel">
    <div class="views-user-payment-result-view__panel-2">
      <!-- Loading -->
      <div v-if="loading" class="views-user-payment-result-view__panel-3">
        <div class="views-user-payment-result-view__panel-4"></div>
      </div>
      <template v-else>
        <!-- Status Icon -->
        <div class="views-user-payment-result-view__panel-5">
          <div v-if="isSuccess"
            class="views-user-payment-result-view__panel-6">
            <svg class="views-user-payment-result-view__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor"
              stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <div v-else-if="isPending"
            class="views-user-payment-result-view__panel-7">
            <div class="views-user-payment-result-view__panel-8"></div>
          </div>
          <div v-else
            class="views-user-payment-result-view__panel-9">
            <svg class="views-user-payment-result-view__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="views-user-payment-result-view__heading">
            {{ statusTitle }}
          </h2>
          <p v-if="isPending" class="views-user-payment-result-view__description">
            {{ t('payment.result.processingHint') }}
          </p>
        </div>
        <!-- Order Info -->
        <div v-if="order" class="views-user-payment-result-view__panel-10">
          <div class="views-user-payment-result-view__panel-11">
            <div v-if="hasOrderId(order)" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.orderId') }}</span>
              <span class="views-user-payment-result-view__text-2">#{{ order.id }}</span>
            </div>
            <div v-if="order.out_trade_no" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.orderNo') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ order.out_trade_no }}</span>
            </div>
            <div v-if="isRechargeOrder" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.baseAmount') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ formatCNY(baseAmount, localeCode) }}</span>
            </div>
            <div v-if="isRechargeOrder && feeAmount > 0" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.fee') }} ({{ rechargeFeeRate }}%)</span>
              <span class="views-user-payment-result-view__text-2">{{ formatCNY(feeAmount, localeCode) }}</span>
            </div>
            <div v-if="isRechargeOrder" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.payAmount') }}</span>
              <span class="views-user-payment-result-view__text-3">{{ formatCNY(rechargePayAmount, localeCode) }}</span>
            </div>
            <div v-if="isRechargeOrder" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.basePoints') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ formatPoints(basePoints, localeCode) }}</span>
            </div>
            <div v-if="isRechargeOrder && bonusPoints > 0" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.bonusPoints') }}</span>
              <span class="views-user-payment-result-view__text-2">+{{ formatPoints(bonusPoints, localeCode) }}</span>
            </div>
            <div v-if="isRechargeOrder" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.creditedPoints') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ formatPoints(creditedPoints, localeCode) }}</span>
            </div>
            <div v-if="isSubscriptionPayment" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.pointsPaid') }}</span>
              <span class="views-user-payment-result-view__text-3">{{ formatPoints(subscriptionPoints, localeCode) }}</span>
            </div>
            <div v-if="hasPaymentType(order)" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.paymentMethod') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ t(paymentMethodI18nKey(order.payment_type), normalizedOrderPaymentType(order.payment_type)) }}</span>
            </div>
            <div class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.status') }}</span>
              <OrderStatusBadge :status="displayOrderStatus(order.status)" />
            </div>
          </div>
        </div>
        <!-- EasyPay return info (when no order loaded) -->
        <div v-else-if="returnInfo" class="views-user-payment-result-view__panel-10">
          <div class="views-user-payment-result-view__panel-11">
            <div v-if="returnInfo.outTradeNo" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.orderId') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ returnInfo.outTradeNo }}</span>
            </div>
            <div v-if="returnInfo.money" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.payAmount') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ formatCNY(Number(returnInfo.money) || 0, localeCode) }}</span>
            </div>
            <div v-if="returnInfo.type" class="views-user-payment-result-view__panel-12">
              <span class="views-user-payment-result-view__text">{{ t('payment.orders.paymentMethod') }}</span>
              <span class="views-user-payment-result-view__text-2">{{ t(paymentMethodI18nKey(returnInfo.type), normalizedOrderPaymentType(returnInfo.type)) }}</span>
            </div>
          </div>
        </div>
        <!-- Actions -->
        <div class="views-user-payment-result-view__panel-13">
          <button class="views-user-payment-result-view__action btn btn-secondary" @click="router.push(backDestination)">{{ t(backLabelKey) }}</button>
          <button class="views-user-payment-result-view__action btn btn-primary" @click="router.push('/orders')">{{ t('payment.result.viewOrders') }}</button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  readPaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { usePaymentStore } from '@/stores/payment'
import { paymentAPI } from '@/api/payment'
import type { PublicOrderVerifyResult } from '@/api/payment'
import type { OrderStatus, PaymentOrder } from '@/types/payment'
import {
  rechargeBasePoints,
  rechargeBonusPoints,
  rechargeCreditedPoints,
  rechargeFeeAmount,
  rechargePrincipalAmount,
} from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'
import { normalizePaymentMethodForDisplay, paymentMethodI18nKey } from './paymentUx'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const paymentStore = usePaymentStore()

type ResolvedOrder = PaymentOrder | PublicOrderVerifyResult

const order = ref<ResolvedOrder | null>(null)
const loading = ref(true)

interface ReturnInfo {
  outTradeNo: string
  money: string
  type: string
  tradeStatus: string
}
const returnInfo = ref<ReturnInfo | null>(null)

const SUCCESS_STATUSES = new Set(['COMPLETED', 'PAID', 'RECHARGING'])
const PENDING_STATUSES = new Set(['PENDING', 'CREATED', 'WAITING', 'PROCESSING'])
const STATUS_REFRESH_INTERVAL_MS = 2000
const STATUS_REFRESH_MAX_ATTEMPTS = 15

let statusRefreshTimer: ReturnType<typeof setTimeout> | null = null
const refreshAttempts = ref(0)

const baseAmount = computed(() => {
  if (!hasAmountFields(order.value)) return 0
  return rechargePrincipalAmount(order.value)
})

const feeAmount = computed(() => {
  if (!hasAmountFields(order.value)) return 0
  return rechargeFeeAmount(order.value)
})
const basePoints = computed(() => hasAmountFields(order.value) ? rechargeBasePoints(order.value) : 0)
const bonusPoints = computed(() => hasAmountFields(order.value) ? rechargeBonusPoints(order.value) : 0)
const creditedPoints = computed(() => hasAmountFields(order.value) ? rechargeCreditedPoints(order.value) : 0)
const rechargeFeeRate = computed(() => hasAmountFields(order.value) ? order.value.fee_rate : 0)
const rechargePayAmount = computed(() => hasAmountFields(order.value) ? order.value.pay_amount : 0)
const subscriptionPoints = computed(() => {
  if (!hasAmountFields(order.value)) return 0
  return order.value.wallet_amount > 0 ? order.value.wallet_amount : order.value.amount
})
const isRechargeOrder = computed(() => hasAmountFields(order.value) && order.value.order_type === 'balance')
const isSubscriptionPayment = computed(() => hasAmountFields(order.value) && order.value.order_type === 'subscription')

const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

const isSuccess = computed(() => {
  return isSuccessStatus(order.value?.status)
})

const isPending = computed(() => {
  return isPendingStatus(order.value?.status)
})

function isSubscriptionOrder(value: ResolvedOrder | null): boolean {
  return !!value && 'order_type' in value && value.order_type === 'subscription'
}

const backDestination = computed(() =>
  isSubscriptionOrder(order.value) ? '/subscriptions' : '/purchase',
)
const backLabelKey = computed(() =>
  isSubscriptionOrder(order.value)
    ? 'payment.result.backToSubscriptions'
    : 'payment.result.backToRecharge',
)

const statusTitle = computed(() => {
  if (isSuccess.value) {
    return t('payment.result.success')
  }
  if (isPending.value) {
    return t('payment.result.processing')
  }
  return t('payment.result.failed')
})

function normalizedOrderPaymentType(paymentType: string): string {
  return normalizePaymentMethodForDisplay(paymentType || '') || paymentType || ''
}

function setResolvedOrder(nextOrder: ResolvedOrder | null): void {
  order.value = nextOrder
}

function hasOrderId(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'id' in nextOrder && typeof nextOrder.id === 'string' && nextOrder.id !== ''
}

function hasAmountFields(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'pay_amount' in nextOrder && typeof nextOrder.pay_amount === 'number' && 'amount' in nextOrder && typeof nextOrder.amount === 'number'
}

function hasPaymentType(nextOrder: ResolvedOrder | null): nextOrder is PaymentOrder {
  return !!nextOrder && 'payment_type' in nextOrder && typeof nextOrder.payment_type === 'string' && nextOrder.payment_type.trim() !== ''
}

function normalizeOrderStatus(status: string | null | undefined): string {
  return String(status || '').trim().toUpperCase()
}

function displayOrderStatus(status: string): OrderStatus {
  return normalizeOrderStatus(status) as OrderStatus
}

function isSuccessStatus(status: string | null | undefined): boolean {
  return SUCCESS_STATUSES.has(normalizeOrderStatus(status))
}

function isPendingStatus(status: string | null | undefined): boolean {
  return PENDING_STATUSES.has(normalizeOrderStatus(status))
}

function readRouteQueryString(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function restoreRecoverySnapshot(context: {
  resumeToken: string
  routeOrderId: string
  routeOutTradeNo: string
}) {
  if (typeof window === 'undefined') {
    return null
  }

  const rawSnapshot = window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY)
  if (!rawSnapshot) {
    return null
  }

  if (context.resumeToken) {
    return readPaymentRecoverySnapshot(rawSnapshot, {
      resumeToken: context.resumeToken,
    })
  }

  if (!context.routeOrderId && !context.routeOutTradeNo) {
    return null
  }

  const restored = readPaymentRecoverySnapshot(rawSnapshot)
  if (!restored) {
    return null
  }

  if (context.routeOrderId && restored.orderId !== context.routeOrderId) {
    return null
  }

  if (context.routeOutTradeNo && restored.outTradeNo !== context.routeOutTradeNo) {
    return null
  }

  return restored
}

async function resolveOrderFromResumeToken(resumeToken: string): Promise<ResolvedOrder | null> {
  try {
    const result = await paymentAPI.resolveOrderPublicByResumeToken(resumeToken)
    return result.data
  } catch (_err: unknown) {
    return null
  }
}

async function resolveOrderFromOutTradeNo(outTradeNo: string): Promise<ResolvedOrder | null> {
  try {
    const result = await paymentAPI.verifyOrder(outTradeNo)
    return result.data
  } catch (_err: unknown) {
    try {
      const result = await paymentAPI.verifyOrderPublic(outTradeNo)
      return result.data
    } catch (_innerErr: unknown) {
      return null
    }
  }
}

function clearStatusRefreshTimer(): void {
  if (statusRefreshTimer !== null) {
    clearTimeout(statusRefreshTimer)
    statusRefreshTimer = null
  }
}

function clearRecoverySnapshot(): void {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function clearRecoverySnapshotForTerminalStatus(status: string | null | undefined): void {
  if (!status) return
  if (!isPendingStatus(status)) {
    clearRecoverySnapshot()
  }
}

function scheduleStatusRefresh(refreshOrder: (() => Promise<ResolvedOrder | null>) | null): void {
  clearStatusRefreshTimer()
  if (!refreshOrder || !isPending.value || refreshAttempts.value >= STATUS_REFRESH_MAX_ATTEMPTS) {
    return
  }

  statusRefreshTimer = setTimeout(async () => {
    refreshAttempts.value += 1
    const refreshedOrder = await refreshOrder()
    if (refreshedOrder) {
      setResolvedOrder(refreshedOrder)
      clearRecoverySnapshotForTerminalStatus(refreshedOrder.status)
    }

    if (isPendingStatus(order.value?.status)) {
      scheduleStatusRefresh(refreshOrder)
    }
  }, STATUS_REFRESH_INTERVAL_MS)
}

onMounted(async () => {
  const resumeToken = readRouteQueryString('resume_token')
  const routeOrderId = readRouteQueryString('order_id')
  let outTradeNo = readRouteQueryString('out_trade_no')
  let orderId = ''
  let resumeTokenLookupFailed = false

  const restored = restoreRecoverySnapshot({
    resumeToken,
    routeOrderId,
    routeOutTradeNo: outTradeNo,
  })
  if (restored?.orderId) {
    orderId = restored.orderId
  }
  if (!outTradeNo && restored?.outTradeNo) {
    outTradeNo = restored.outTradeNo
  }

  if (resumeToken) {
    const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
    if (resolvedOrder) {
      setResolvedOrder(resolvedOrder)
      if (!orderId) {
        orderId = hasOrderId(resolvedOrder) ? resolvedOrder.id : ''
      }
    } else if (routeOrderId) {
      resumeTokenLookupFailed = true
      orderId = routeOrderId
    } else {
      resumeTokenLookupFailed = true
    }
  } else if (routeOrderId) {
    orderId = routeOrderId
  }

  const hasLegacyFallbackContext = readRouteQueryString('trade_status').trim() !== ''
  const shouldUsePublicOutTradeNo = outTradeNo !== '' && (hasLegacyFallbackContext || !!routeOrderId || !!orderId)

  if (!order.value && orderId && (!resumeToken || !!routeOrderId)) {
    try {
      setResolvedOrder(await paymentStore.pollOrderStatus(orderId))
    } catch (_err: unknown) {
      // Order lookup failed, will try legacy fallback below when possible.
    }
  }

  if (!order.value && shouldUsePublicOutTradeNo && (!resumeToken || resumeTokenLookupFailed)) {
    const legacyOrder = await resolveOrderFromOutTradeNo(outTradeNo)
    if (legacyOrder) {
      setResolvedOrder(legacyOrder)
      if (!orderId) {
        orderId = hasOrderId(legacyOrder) ? legacyOrder.id : ''
      }
    }
  }

  if (!order.value && !orderId && outTradeNo && hasLegacyFallbackContext) {
    returnInfo.value = {
      outTradeNo,
      money: String(route.query.money || ''),
      type: String(route.query.type || ''),
      tradeStatus: String(route.query.trade_status || ''),
    }
  }

  const refreshOrder = async (): Promise<ResolvedOrder | null> => {
    if (resumeToken) {
      const resolvedOrder = await resolveOrderFromResumeToken(resumeToken)
      if (resolvedOrder) {
        return resolvedOrder
      }
    }

    if (orderId) {
      try {
        return await paymentStore.pollOrderStatus(orderId)
      } catch (_err: unknown) {
        // Fall through to legacy public verification when order polling is unavailable.
      }
    }

    if (shouldUsePublicOutTradeNo) {
      return await resolveOrderFromOutTradeNo(outTradeNo)
    }

    return null
  }

  if (isPendingStatus(order.value?.status)) {
    scheduleStatusRefresh(refreshOrder)
  } else if (order.value) {
    clearRecoverySnapshotForTerminalStatus(order.value.status)
  } else if (returnInfo.value) {
    clearRecoverySnapshot()
  }
  loading.value = false
})

onBeforeUnmount(() => {
  clearStatusRefreshTimer()
})
</script>
