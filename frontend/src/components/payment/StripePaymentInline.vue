<template>
  <div class="components-payment-stripe-payment-inline__panel">
    <div v-if="loading" class="components-payment-stripe-payment-inline__panel-2">
      <div class="components-payment-stripe-payment-inline__panel-3"></div>
    </div>
    <div v-else-if="initError" class="components-payment-stripe-payment-inline__panel-4 card-body card">
      <p class="components-payment-stripe-payment-inline__description">{{ initError }}</p>
      <button class="components-payment-stripe-payment-inline__action btn btn-secondary" @click="$emit('back')">{{ t('payment.result.backToRecharge') }}</button>
    </div>
    <!-- Success -->
    <template v-else-if="success">
      <div class="components-payment-stripe-payment-inline__panel-5 card-body card">
        <div class="components-payment-stripe-payment-inline__panel-6">
          <div class="components-payment-stripe-payment-inline__panel-7">
            <Icon name="check" size="lg" class="components-payment-stripe-payment-inline__icon" />
          </div>
          <p class="components-payment-stripe-payment-inline__description-2">{{ t('payment.result.success') }}</p>
          <div class="components-payment-stripe-payment-inline__panel-8">
            <div class="components-payment-stripe-payment-inline__panel-9">
              <div class="components-payment-stripe-payment-inline__panel-10">
                <span class="components-payment-stripe-payment-inline__text">{{ t('payment.orders.orderId') }}</span>
                <span class="components-payment-stripe-payment-inline__text-2">#{{ orderId }}</span>
              </div>
              <div v-if="amount > 0" class="components-payment-stripe-payment-inline__panel-10">
                <span class="components-payment-stripe-payment-inline__text">{{ t('payment.orders.creditedPoints') }}</span>
                <span class="components-payment-stripe-payment-inline__text-2">{{ formatPoints(amount) }}</span>
              </div>
              <div class="components-payment-stripe-payment-inline__panel-10">
                <span class="components-payment-stripe-payment-inline__text">{{ t('payment.orders.payAmount') }}</span>
                <span class="components-payment-stripe-payment-inline__text-2">{{ formatCNY(payAmount) }}</span>
              </div>
            </div>
          </div>
          <button class="btn btn-primary" @click="$emit('done')">{{ t('common.confirm') }}</button>
        </div>
      </div>
    </template>
    <template v-else>
      <!-- Amount -->
      <div class="components-payment-stripe-payment-inline__panel-11 card">
        <div class="components-payment-stripe-payment-inline__panel-12">
          <p class="components-payment-stripe-payment-inline__description-3">{{ t('payment.actualPay') }}</p>
          <p class="components-payment-stripe-payment-inline__description-4">{{ formatCNY(payAmount) }}</p>
        </div>
      </div>
      <!-- Stripe Payment Element -->
      <div class="components-payment-stripe-payment-inline__panel-5 card-body card">
        <div ref="stripeMount" class="components-payment-stripe-payment-inline__panel-13"></div>
        <p v-if="error" class="components-payment-stripe-payment-inline__description-5">{{ error }}</p>
        <button class="components-payment-stripe-payment-inline__action-2 btn btn-stripe" :disabled="submitting || !ready" @click="handlePay">
          <span v-if="submitting" class="components-payment-stripe-payment-inline__text-3">
            <span class="components-payment-stripe-payment-inline__text-4"></span>
            {{ t('common.processing') }}
          </span>
          <span v-else>{{ t('payment.stripePay') }}</span>
        </button>
      </div>
      <!-- Cancel order -->
      <button class="components-payment-stripe-payment-inline__action-3 btn btn-secondary" :disabled="cancelling" @click="handleCancel">
        {{ cancelling ? t('common.processing') : t('payment.qr.cancelOrder') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import { formatCNY, formatPoints } from '@/utils/format'
import type { Stripe, StripeElements } from '@stripe/stripe-js'
import Icon from '@/components/icons/Icon.vue'

// Stripe payment methods that open a popup (redirect or QR code)
const POPUP_METHODS = new Set(['alipay', 'wechat_pay'])

const props = defineProps<{
  orderId: string
  amount: number
  clientSecret: string
  orderType?: 'balance' | 'subscription'
  publishableKey: string
  payAmount: number
  currency?: string
}>()

const emit = defineEmits<{ success: []; done: []; back: []; redirect: [orderId: string, payUrl: string] }>()

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const stripeMount = ref<HTMLElement | null>(null)
const loading = ref(true)
const initError = ref('')
const error = ref('')
const submitting = ref(false)
const cancelling = ref(false)
const success = ref(false)
const ready = ref(false)
const selectedType = ref('')
let stripeInstance: Stripe | null = null
let elementsInstance: StripeElements | null = null

onMounted(async () => {
  try {
    const { loadStripe } = await import('@stripe/stripe-js/pure')
    const stripe = await loadStripe(props.publishableKey)
    if (!stripe) { initError.value = t('payment.stripeLoadFailed'); return }

    stripeInstance = stripe
    loading.value = false
    await nextTick()
    if (!stripeMount.value) return

    const isDark = document.documentElement.classList.contains('dark')
    const elements = stripe.elements({
      clientSecret: props.clientSecret,
      appearance: { theme: isDark ? 'night' : 'stripe', variables: { borderRadius: '8px' } },
    })
    elementsInstance = elements
    const paymentElement = elements.create('payment', {
      layout: 'tabs',
      paymentMethodOrder: ['alipay', 'wechat_pay', 'card', 'link'],
    } as Record<string, unknown>)
    paymentElement.mount(stripeMount.value)
    paymentElement.on('ready', () => { ready.value = true })
    paymentElement.on('change', (event: { value: { type: string } }) => {
      selectedType.value = event.value.type
    })
  } catch (err: unknown) {
    initError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.stripeLoadFailed'))
  } finally {
    loading.value = false
  }
})

async function handlePay() {
  if (!stripeInstance || !elementsInstance || submitting.value) return

  // Alipay / WeChat Pay: open popup for redirect or QR display
  if (POPUP_METHODS.has(selectedType.value)) {
    const popupUrl = router.resolve({
      path: '/payment/stripe-popup',
      query: {
        order_id: String(props.orderId),
        method: selectedType.value,
        amount: String(props.payAmount),
      },
    }).href
    const popup = window.open(popupUrl, 'paymentPopup', getPaymentPopupFeatures())

    const onReady = (event: MessageEvent) => {
      if (event.source !== popup || event.data?.type !== 'STRIPE_POPUP_READY') return
      window.removeEventListener('message', onReady)
      popup?.postMessage({
        type: 'STRIPE_POPUP_INIT',
        clientSecret: props.clientSecret,
        publishableKey: props.publishableKey,
      }, window.location.origin)
    }
    window.addEventListener('message', onReady)

    emit('redirect', props.orderId, popupUrl)
    return
  }

  // Card / Link: confirm inline
  submitting.value = true
  error.value = ''
  try {
    const { error: stripeError } = await stripeInstance.confirmPayment({
      elements: elementsInstance,
      confirmParams: {
        return_url: window.location.origin + '/payment/result?order_id=' + props.orderId + '&status=success',
      },
      redirect: 'if_required',
    })
    if (stripeError) {
      error.value = stripeError.message || t('payment.result.failed')
    } else {
      success.value = true
      emit('success')
    }
  } catch (err: unknown) {
    error.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.result.failed'))
  } finally {
    submitting.value = false
  }
}

async function handleCancel() {
  if (!props.orderId || cancelling.value) return
  cancelling.value = true
  try {
    await paymentAPI.cancelOrder(props.orderId)
    emit('back')
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    cancelling.value = false
  }
}
</script>
