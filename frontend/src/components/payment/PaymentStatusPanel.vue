<template>
  <div class="components-payment-payment-status-panel__panel">
    <!-- ═══ Terminal States: show result, user clicks to return ═══ -->

    <!-- Success -->
    <template v-if="outcome === 'success'">
      <div class="components-payment-payment-status-panel__panel-2 card">
        <div class="components-payment-payment-status-panel__panel-3">
          <div class="components-payment-payment-status-panel__panel-4">
            <Icon name="check" size="lg" class="components-payment-payment-status-panel__icon" />
          </div>
          <p class="components-payment-payment-status-panel__description">{{ props.orderType === 'subscription' ? t('payment.result.subscriptionSuccess') : t('payment.result.success') }}</p>
          <div v-if="paidOrder" class="components-payment-payment-status-panel__panel-5">
            <div class="components-payment-payment-status-panel__panel-6">
              <div class="components-payment-payment-status-panel__panel-7">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.orders.orderId') }}</span>
                <span class="components-payment-payment-status-panel__text-2">#{{ paidOrder.id }}</span>
              </div>
              <div v-if="paidOrder.out_trade_no" class="components-payment-payment-status-panel__panel-7">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.orders.orderNo') }}</span>
                <span class="components-payment-payment-status-panel__text-2">{{ paidOrder.out_trade_no }}</span>
              </div>
              <div class="components-payment-payment-status-panel__panel-7">
                <span class="components-payment-payment-status-panel__text">
                  {{ paidOrder.order_type === 'balance' ? t('payment.orders.creditedPoints') : t('payment.orders.pointsPaid') }}
                </span>
                <span class="components-payment-payment-status-panel__text-2">{{ formatPoints(orderPoints(paidOrder), localeCode) }}</span>
              </div>
              <div v-if="paidOrder.order_type === 'balance'" class="components-payment-payment-status-panel__panel-7">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.orders.payAmount') }}</span>
                <span class="components-payment-payment-status-panel__text-2">{{ formatCNY(paidOrder.pay_amount, localeCode) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Cancelled -->
    <template v-else-if="outcome === 'cancelled'">
      <div class="components-payment-payment-status-panel__panel-2 card">
        <div class="components-payment-payment-status-panel__panel-3">
          <div class="components-payment-payment-status-panel__panel-8">
            <svg class="components-payment-payment-status-panel__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <p class="components-payment-payment-status-panel__description">{{ t('payment.qr.cancelled') }}</p>
          <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.cancelledDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- Expired / Failed -->
    <template v-else-if="outcome === 'expired'">
      <div class="components-payment-payment-status-panel__panel-2 card">
        <div class="components-payment-payment-status-panel__panel-3">
          <div class="components-payment-payment-status-panel__panel-9">
            <svg class="components-payment-payment-status-panel__icon-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="components-payment-payment-status-panel__description">{{ t('payment.qr.expired') }}</p>
          <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.expiredDesc') }}</p>
          <button class="btn btn-primary" @click="handleDone">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>

    <!-- ═══ Active States: QR or Popup waiting ═══ -->

    <!-- Mobile Alipay app handoff. The QR fallback stays hidden until launch timeout. -->
    <template v-else-if="isMobileAlipayDeepLink">
      <template v-if="!deepLinkFallbackVisible">
        <div class="components-payment-payment-status-panel__panel-2 card">
          <div class="components-payment-payment-status-panel__panel-10">
            <div
              v-if="deepLinkState === 'launching'"
              class="components-payment-payment-status-panel__panel-11"
            ></div>
            <div
              v-else
              class="components-payment-payment-status-panel__panel-12"
            >
              <Icon name="checkCircle" size="lg" class="components-payment-payment-status-panel__icon-4" />
            </div>
            <p class="components-payment-payment-status-panel__description-3">
              {{ deepLinkState === 'backgrounded' ? t('payment.qr.alipayContinueInApp') : t('payment.qr.alipayOpening') }}
            </p>
            <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.alipayWaitingHint') }}</p>
            <button
              v-if="deepLinkState === 'backgrounded'"
              data-test="reopen-alipay"
              class="components-payment-payment-status-panel__action btn btn-alipay"
              @click="reopenAlipay"
            >
              <Icon name="externalLink" size="sm" />
              {{ t('payment.qr.reopenAlipay') }}
            </button>
          </div>
        </div>
        <div class="components-payment-payment-status-panel__panel-13 card">
          <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.expiresIn') }}</p>
          <p class="components-payment-payment-status-panel__description-4">{{ countdownDisplay }}</p>
          <p class="components-payment-payment-status-panel__description-5">{{ t('payment.qr.waitingPayment') }}</p>
        </div>
      </template>
      <template v-else>
        <div data-test="alipay-qr-fallback" class="components-payment-payment-status-panel__panel-2 card">
          <div class="components-payment-payment-status-panel__panel-14">
            <div class="components-payment-payment-status-panel__panel-15">
              <p class="components-payment-payment-status-panel__description-3">{{ t('payment.qr.alipayFallbackTitle') }}</p>
              <p class="components-payment-payment-status-panel__description-6">{{ t('payment.qr.alipayFallbackHint') }}</p>
            </div>
            <div class="components-payment-payment-status-panel__panel-16">
              <div class="components-payment-payment-status-panel__panel-17">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.orders.payAmount') }}</span>
                <span class="components-payment-payment-status-panel__text-3">{{ displayPaymentAmount }}</span>
              </div>
              <div class="components-payment-payment-status-panel__panel-17">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.orders.orderNo') }}</span>
                <span class="components-payment-payment-status-panel__text-4">
                  {{ displayOrderNumber }}
                </span>
              </div>
              <div class="components-payment-payment-status-panel__panel-17">
                <span class="components-payment-payment-status-panel__text">{{ t('payment.qr.expiresIn') }}</span>
                <span class="components-payment-payment-status-panel__text-5">{{ countdownDisplay }}</span>
              </div>
            </div>
            <div :class="['components-payment-payment-status-panel__panel-21', qrBorderClass]">
              <canvas ref="qrCanvas" class="components-payment-payment-status-panel__canvas"></canvas>
              <div class="components-payment-payment-status-panel__panel-18">
                <span :class="['components-payment-payment-status-panel__text-6', qrLogoBgClass]">
                  <img :src="qrLogoIcon" alt="" class="components-payment-payment-status-panel__image" />
                </span>
              </div>
            </div>
            <p class="components-payment-payment-status-panel__description-7">
              {{ t('payment.qr.alipaySaveAndScanHint') }}
            </p>
            <div class="components-payment-payment-status-panel__panel-19">
              <button
                data-test="reopen-alipay"
                class="components-payment-payment-status-panel__action-2 btn btn-alipay"
                @click="reopenAlipay"
              >
                <Icon name="externalLink" size="sm" />
                {{ t('payment.qr.reopenAlipay') }}
              </button>
              <button
                data-test="save-alipay-qr"
                class="components-payment-payment-status-panel__action-2 btn btn-secondary"
                @click="saveQRCode"
              >
                <Icon name="download" size="sm" />
                {{ t('payment.qr.saveQRCode') }}
              </button>
            </div>
            <button class="components-payment-payment-status-panel__action-3 btn btn-secondary" @click="handleDone">
              {{ t('payment.result.backToRecharge') }}
            </button>
          </div>
        </div>
      </template>
    </template>

    <!-- QR Code Mode -->
    <template v-else-if="showQRCode">
      <div class="components-payment-payment-status-panel__panel-2 card">
        <div class="components-payment-payment-status-panel__panel-14">
          <p class="components-payment-payment-status-panel__description-3">{{ scanTitle }}</p>
          <div :class="['components-payment-payment-status-panel__panel-21', qrBorderClass]">
            <canvas ref="qrCanvas" class="components-payment-payment-status-panel__canvas"></canvas>
            <!-- Brand logo overlay -->
            <div class="components-payment-payment-status-panel__panel-18">
              <span :class="['components-payment-payment-status-panel__text-6', qrLogoBgClass]">
                <img :src="qrLogoIcon" alt="" class="components-payment-payment-status-panel__image" />
              </span>
            </div>
          </div>
          <p v-if="scanHint" class="components-payment-payment-status-panel__description-8">{{ scanHint }}</p>
          <button v-if="payUrl" class="components-payment-payment-status-panel__action-4 btn btn-secondary" @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
        </div>
      </div>
      <div class="components-payment-payment-status-panel__panel-13 card">
        <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.expiresIn') }}</p>
        <p class="components-payment-payment-status-panel__description-4">{{ countdownDisplay }}</p>
        <p class="components-payment-payment-status-panel__description-5">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button class="components-payment-payment-status-panel__action-3 btn btn-secondary" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>

    <!-- Waiting for Popup/Redirect Mode -->
    <template v-else>
      <div class="components-payment-payment-status-panel__panel-2 card">
        <div class="components-payment-payment-status-panel__panel-3">
          <div class="components-payment-payment-status-panel__panel-20"></div>
          <p class="components-payment-payment-status-panel__description-2">{{ t('payment.qr.payInNewWindowHint') }}</p>
          <button v-if="payUrl" class="components-payment-payment-status-panel__action-4 btn btn-secondary" @click="reopenPopup">
            {{ t('payment.qr.openPayWindow') }}
          </button>
        </div>
      </div>
      <div class="components-payment-payment-status-panel__panel-13 card">
        <p class="components-payment-payment-status-panel__description-4">{{ countdownDisplay }}</p>
        <p class="components-payment-payment-status-panel__description-5">{{ t('payment.qr.waitingPayment') }}</p>
      </div>
      <button class="components-payment-payment-status-panel__action-3 btn btn-secondary" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePaymentStore } from '@/stores/payment'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { getPaymentPopupFeatures, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import type { PaymentOrder } from '@/types/payment'
import { rechargeCreditedPoints } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import QRCode from 'qrcode'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import paymentIcon from '@/assets/icons/payment.svg'
import {
  createAlipayDeepLinkLauncher,
  type AlipayDeepLinkLauncher,
  type AlipayDeepLinkState,
} from './alipayDeepLink'

const props = defineProps<{
  orderId: string
  amount?: number
  payAmount?: number
  qrCode: string
  expiresAt: string
  paymentType: string
  payUrl?: string
  orderType?: string
  currency?: string
  outTradeNo?: string
  mobileAlipayDeepLink?: boolean
}>()

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const emit = defineEmits<{ done: []; success: []; settled: [outcome: PaymentOutcome] }>()

const i18n = useI18n()
const { t } = i18n
const paymentStore = usePaymentStore()
const appStore = useAppStore()

const qrCanvas = ref<HTMLCanvasElement | null>(null)
const qrUrl = ref('')
const remainingSeconds = ref(0)
const cancelling = ref(false)
const paidOrder = ref<PaymentOrder | null>(null)
const deepLinkState = ref<AlipayDeepLinkState>('idle')
const deepLinkFallbackVisible = ref(false)
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

// Terminal outcome: null = still active, 'success' | 'cancelled' | 'expired'
const outcome = ref<PaymentOutcome | null>(null)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null
let verifyAttempts = 0
let lastVerifyAt = 0
let alipayLauncher: AlipayDeepLinkLauncher | null = null

const VERIFY_RETRY_INTERVAL_MS = 15000
const VERIFY_RETRY_MAX_ATTEMPTS = 6

const isAlipay = computed(() => isBuiltInAlipayMethod(props.paymentType))
const isWxpay = computed(() => isBuiltInWxpayMethod(props.paymentType))
const isMobileAlipayDeepLink = computed(() => props.mobileAlipayDeepLink === true && isAlipay.value && !!qrUrl.value)
const showQRCode = computed(() => !!qrUrl.value && (!isMobileAlipayDeepLink.value || deepLinkFallbackVisible.value))

const qrBorderClass = computed(() => {
  if (isAlipay.value) return 'components-payment-payment-status-panel__state'
  if (isWxpay.value) return 'components-payment-payment-status-panel__state-2'
  return 'components-payment-payment-status-panel__state-3'
})

const qrLogoBgClass = computed(() => {
  if (isAlipay.value) return 'payment-logo--alipay'
  if (isWxpay.value) return 'payment-logo--wechat'
  return 'status-fill--neutral'
})

const qrLogoIcon = computed(() => {
  if (isAlipay.value) return alipayIcon
  if (isWxpay.value) return wxpayIcon
  return paymentIcon
})

const scanTitle = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipay')
  if (isWxpay.value) return t('payment.qr.scanWxpay')
  return t('payment.qr.scanToPay')
})

const scanHint = computed(() => {
  if (isAlipay.value) return t('payment.qr.scanAlipayHint')
  if (isWxpay.value) return t('payment.qr.scanWxpayHint')
  return ''
})

const countdownDisplay = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return m.toString().padStart(2, '0') + ':' + s.toString().padStart(2, '0')
})

const displayPaymentAmount = computed(() => formatGatewayAmount(props.payAmount || props.amount || 0))
const displayOrderNumber = computed(() => props.outTradeNo || `#${props.orderId}`)

function formatGatewayAmount(value: number): string {
  return formatCNY(value, localeCode.value)
}

function orderPoints(order: PaymentOrder): number {
  if (order.order_type === 'balance') return rechargeCreditedPoints(order)
  return order.wallet_amount > 0 ? order.wallet_amount : order.amount
}

function isSuccessStatus(status: string | null | undefined): boolean {
  return status === 'COMPLETED' || status === 'PAID' || status === 'RECHARGING'
}

function reopenPopup() {
  if (props.payUrl) {
    const win = window.open(props.payUrl, 'paymentPopup', getPaymentPopupFeatures())
    if (!win || win.closed) {
      window.location.href = props.payUrl
    }
  }
}

function setOutcome(next: PaymentOutcome) {
  if (outcome.value === next) return
  outcome.value = next
  emit('settled', next)
}

async function renderQR() {
  await nextTick()
  if (!showQRCode.value || !qrCanvas.value || !qrUrl.value) return
  await QRCode.toCanvas(qrCanvas.value, qrUrl.value, {
    width: 220, margin: 2,
    errorCorrectionLevel: 'M',
  })
}

function updateDeepLinkState(state: AlipayDeepLinkState) {
  deepLinkState.value = state
  if (state === 'fallback') {
    deepLinkFallbackVisible.value = true
    renderQR()
  } else if (state === 'backgrounded') {
    deepLinkFallbackVisible.value = false
  }
}

function reopenAlipay() {
  alipayLauncher?.launch()
}

function saveQRCode() {
  const canvas = qrCanvas.value
  if (!canvas) return
  const link = document.createElement('a')
  link.href = canvas.toDataURL('image/png')
  link.download = `alipay-${props.outTradeNo || props.orderId}.png`
  document.body.appendChild(link)
  link.click()
  link.remove()
}

async function tryRecoverPendingOrder(order: PaymentOrder): Promise<PaymentOrder> {
  if (!isWxpay.value && !isMobileAlipayDeepLink.value) return order
  const outTradeNo = String(order.out_trade_no || '').trim()
  if (!outTradeNo) return order
  const normalizedStatus = String(order.status || '').trim().toUpperCase()
  if (normalizedStatus !== 'PENDING') return order
  const now = Date.now()
  if (verifyAttempts >= VERIFY_RETRY_MAX_ATTEMPTS || now - lastVerifyAt < VERIFY_RETRY_INTERVAL_MS) {
    return order
  }

  lastVerifyAt = now
  verifyAttempts += 1
  try {
    const result = await paymentAPI.verifyOrder(outTradeNo)
    return result.data ?? order
  } catch {
    return order
  }
}

let pollInFlight = false
async function pollStatus() {
  if (!props.orderId || outcome.value) return
  // 防重入：接口（含 verifyOrder 二次确认）响应慢于 3 秒轮询间隔时避免并发重叠请求。
  if (pollInFlight) return
  pollInFlight = true
  try {
    let order = await paymentStore.pollOrderStatus(props.orderId)
    if (!order) return
    // 已进入终态则不再处理迟到的响应。
    if (outcome.value) return
    order = await tryRecoverPendingOrder(order)
    if (outcome.value) return
    if (isSuccessStatus(order.status)) {
      cleanup()
      paidOrder.value = order
      setOutcome('success')
      emit('success')
    } else if (order.status === 'CANCELLED') {
      cleanup()
      setOutcome('cancelled')
    } else if (order.status === 'EXPIRED' || order.status === 'FAILED') {
      cleanup()
      setOutcome('expired')
    }
  } finally {
    pollInFlight = false
  }
}

function startCountdown(seconds: number) {
  remainingSeconds.value = Math.max(0, seconds)
  if (remainingSeconds.value <= 0) { setOutcome('expired'); return }
  countdownTimer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) { setOutcome('expired'); cleanup() }
  }, 1000)
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    cleanup()
    setOutcome('cancelled')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}

function handleDone() { cleanup(); emit('done') }

function cleanup() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
  alipayLauncher?.dispose()
  alipayLauncher = null
}

// Initialize on mount
qrUrl.value = props.qrCode
verifyAttempts = 0
lastVerifyAt = 0
let seconds = 30 * 60
if (props.expiresAt) {
  seconds = Math.floor((new Date(props.expiresAt).getTime() - Date.now()) / 1000)
}
startCountdown(seconds)
pollTimer = setInterval(pollStatus, 3000)
renderQR()

watch([() => qrUrl.value, showQRCode], () => renderQR())
onMounted(() => {
  if (!isMobileAlipayDeepLink.value) return
  alipayLauncher = createAlipayDeepLinkLauncher({
    qrCode: qrUrl.value,
    document,
    lifecycleTarget: window,
    userAgent: window.navigator.userAgent,
    assignLocation: (url) => window.location.assign(url),
    onStateChange: updateDeepLinkState,
  })
  alipayLauncher.launch()
})
onUnmounted(() => cleanup())
</script>
