<template>
  <component :is="embedded ? 'div' : AppLayout">
    <div :class="['page-stack', { 'page-stack--narrow': !embedded }]">
      <div v-if="loading" class="views-user-payment-view__panel-2">
        <div class="views-user-payment-view__panel-3"></div>
      </div>
      <template v-else>
        <!-- Tab Switcher (hide during payment and subscription confirm) -->
        <div v-if="tabs.length > 1 && paymentPhase === 'select' && !selectedPlan" class="views-user-payment-view__panel-4">
          <button v-for="tab in tabs" :key="tab.key"
            class="views-user-payment-view__action"
            :class="activeTab === tab.key ? 'views-user-payment-view__action-4' : 'views-user-payment-view__action-5'"
            @click="activeTab = tab.key">{{ tab.label }}</button>
        </div>
        <!-- Payment in progress (shared by recharge and subscription) -->
        <template v-if="paymentPhase === 'paying'">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :amount="paymentState.amount"
            :pay-amount="paymentState.payAmount"
            :qr-code="paymentState.qrCode"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            :currency="paymentState.currency || selectedCurrency"
            :out-trade-no="paymentState.outTradeNo"
            :mobile-alipay-deep-link="paymentState.alipayMobilePrecreateDeepLink"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
            @settled="onPaymentSettled"
          />
        </template>
        <!-- Tab content (select phase) -->
        <template v-else>
          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <div class="wallet-summary card">
              <div class="wallet-summary__header">
                <div>
                  <p class="wallet-summary__eyebrow">{{ t('payment.balanceDetails') }}</p>
                  <p class="wallet-summary__account">{{ user?.username || user?.email || '' }}</p>
                </div>
                <div class="wallet-summary__available">
                  <span>{{ t('payment.wallet.availablePoints') }}</span>
                  <strong>{{ formatWalletAmount(checkout.wallet.available_balance) }}</strong>
                </div>
              </div>
              <div class="wallet-summary__rows">
                <div class="wallet-summary__row">
                  <span>{{ t('payment.wallet.rechargePoints') }}</span>
                  <strong>{{ formatWalletAmount(checkout.wallet.recharge_balance) }}</strong>
                </div>
                <div class="wallet-summary__row">
                  <span>{{ t('payment.wallet.bonusPoints') }}</span>
                  <strong>{{ formatWalletAmount(checkout.wallet.bonus_balance) }}</strong>
                </div>
                <div v-if="checkout.wallet.overdraft_amount > 0" class="wallet-summary__row wallet-summary__row--warning">
                  <span>{{ t('payment.overdraftAmount') }}</span>
                  <strong>{{ formatWalletAmount(checkout.wallet.overdraft_amount) }}</strong>
                </div>
              </div>
              <p v-if="checkout.wallet.next_bonus_expires_at && checkout.wallet.bonus_balance > 0" class="wallet-summary__expiry">
                {{ t('payment.wallet.expiresAt', { date: formatWalletExpiry(checkout.wallet.next_bonus_expires_at) }) }}
              </p>
            </div>
            <div v-if="checkout.balance_disabled || enabledMethods.length === 0" class="views-user-payment-view__panel-6 card">
              <p class="views-user-payment-view__description-4">{{ t('payment.notAvailable') }}</p>
            </div>
            <button
              v-else
              type="button"
              class="wallet-summary__recharge btn btn-primary"
              @click="openRechargeDialog"
            >
              <Icon name="plus" size="sm" />
              {{ t('payment.rechargeNow') }}
            </button>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <div v-if="checkout.plans.length === 0" class="views-user-payment-view__panel-6 card">
              <Icon name="gift" size="xl" class="views-user-payment-view__icon" />
              <p class="views-user-payment-view__description-4">{{ t('payment.noPlans') }}</p>
            </div>
            <div v-else :class="planGridClass">
              <SubscriptionPlanCard
                v-for="plan in checkout.plans"
                :key="plan.id"
                :plan="plan"
                :active-subscriptions="activeSubscriptions"
                :pending-platforms="pendingPlatforms"
                @select="selectPlan"
              />
            </div>
            <!-- Active subscriptions (compact, below plan list) -->
            <div v-if="showActiveSubscriptions && activeSubscriptions.length > 0">
              <p class="views-user-payment-view__description-8">{{ t('payment.activeSubscription') }}</p>
              <div class="views-user-payment-view__panel-17">
                <div v-for="sub in activeSubscriptions" :key="sub.id"
                  class="views-user-payment-view__panel-18">
                  <div :class="['views-user-payment-view__panel-29', platformAccentBarClass(sub.group?.platform || '')]" />
                  <div class="views-user-payment-view__panel-19">
                    <div class="views-user-payment-view__panel-20">
                      <span class="views-user-payment-view__text-9">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                      <span :class="['views-user-payment-view__text-14', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                    </div>
                    <div class="views-user-payment-view__panel-21">
                      <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                      <span v-if="subscriptionHasPeakRate(sub)">{{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(sub) }}</span>
                      <span v-if="!subscriptionGroupHasQuota(sub)">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                      <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                      <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                    </div>
                  </div>
                  <span class="views-user-payment-view__text-10 badge badge-success">{{ t('userSubscriptions.status.active') }}</span>
                </div>
              </div>
            </div>
          </template>
        </template>
        <div v-if="(checkout.help_text || checkout.help_image_url) && paymentPhase === 'select'" class="views-user-payment-view__panel-22 card">
          <div class="views-user-payment-view__panel-23">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="views-user-payment-view__image"
              @click="previewImage = checkout.help_image_url" />
            <p v-if="checkout.help_text" class="views-user-payment-view__description-9">{{ checkout.help_text }}</p>
          </div>
        </div>
      </template>
    </div>
    <!-- Recharge Points Dialog -->
    <BaseDialog
      :show="showRechargeDialog"
      :title="t('payment.rechargeDialogTitle')"
      width="normal"
      :close-on-click-outside="true"
      @close="closeRechargeDialog"
    >
      <div class="recharge-dialog-form">
        <div class="recharge-dialog-form__account">
          <span class="recharge-dialog-form__account-label">{{ t('payment.rechargeAccount') }}</span>
          <strong class="recharge-dialog-form__account-name">{{ user?.username || user?.email || '' }}</strong>
        </div>
        <div class="recharge-dialog-form__section">
          <p class="recharge-rate-preview" data-testid="recharge-points-conversion">
            {{ t('payment.rechargeRatePreview') }}
          </p>
          <AmountInput
            v-model="amount"
            :amounts="[10, 20, 50, 100, 200, 500, 1000, 2000, 5000]"
            :min="globalMinAmount"
            :max="globalMaxAmount"
          />
          <p v-if="amountError" class="recharge-error-text">{{ amountError }}</p>
        </div>
        <div class="recharge-dialog-form__section">
          <PaymentMethodSelector
            :methods="methodOptions"
            :selected="selectedMethod"
            @select="selectedMethod = $event"
          />
        </div>
        <div class="recharge-dialog-form__summary">
          <div class="recharge-summary-row">
            <span class="recharge-summary-label">{{ t('payment.rechargePrincipal') }}</span>
            <span class="recharge-summary-val">{{ formatCNY(validAmount, localeCode) }}</span>
          </div>
          <div class="recharge-summary-row">
            <span class="recharge-summary-label">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
            <span class="recharge-summary-val">{{ formatCNY(feeAmount, localeCode) }}</span>
          </div>
          <div v-if="feeRate > 0" class="recharge-summary-row recharge-summary-row--highlight">
            <span class="recharge-summary-label">{{ t('payment.actualPay') }}</span>
            <span class="recharge-summary-val recharge-summary-val--pay">{{ formatCNY(totalAmount, localeCode) }}</span>
          </div>
          <div class="recharge-summary-row" :class="{ 'recharge-summary-row--divider': feeRate <= 0 }">
            <span class="recharge-summary-label">{{ t('payment.basePoints') }}</span>
            <span class="recharge-summary-val">{{ formatPoints(rechargeBasePoints, localeCode) }}</span>
          </div>
          <div class="recharge-summary-row">
            <span class="recharge-summary-label">{{ t('payment.bonusPoints') }}</span>
            <span class="recharge-summary-val recharge-summary-val--bonus">+{{ formatPoints(rechargeBonusPoints, localeCode) }}</span>
          </div>
          <div class="recharge-summary-row recharge-summary-row--total">
            <span class="recharge-summary-label">{{ t('payment.creditedPoints') }}</span>
            <span class="recharge-summary-val recharge-summary-val--total">{{ formatPoints(rechargeCreditedPoints, localeCode) }}</span>
          </div>
        </div>
        <p v-if="errorMessage" class="recharge-error-text">{{ errorMessage }}</p>
        <p v-if="errorHintMessage" class="recharge-hint-text">{{ errorHintMessage }}</p>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="submitting" @click="closeRechargeDialog">
          {{ t('common.cancel') }}
        </button>
        <button :class="['btn', paymentButtonClass]" :disabled="!canSubmit || submitting" @click="handleSubmitRecharge">
          <span v-if="submitting" class="views-user-payment-view__text-4">
            <span class="views-user-payment-view__text-5"></span>
            {{ t('common.processing') }}
          </span>
          <span v-else>{{ t('payment.createOrder') }} {{ formatCNY(totalAmount, localeCode) }}</span>
        </button>
      </template>
    </BaseDialog>

    <!-- Subscription Purchase / Renewal Confirm Dialog -->
    <BaseDialog
      :show="showSubscriptionDialog"
      :title="isRenewalSelectedPlan ? t('payment.renewTitle') : t('payment.subscribeTitle')"
      width="normal"
      :close-on-click-outside="true"
      @close="closeSubscriptionDialog"
    >
      <div v-if="selectedPlan" class="subscription-dialog-form">
        <!-- Header: platform badge + plan name -->
        <div class="subscription-dialog__header">
          <span :class="['subscription-dialog__badge', planBadgeClass]">
            {{ platformLabel(selectedPlan.group_platform || '') }}
          </span>
          <h3 class="subscription-dialog__title">{{ selectedPlan.name }}</h3>
        </div>
        <!-- Price -->
        <div class="subscription-dialog__price-row">
          <span v-if="selectedPlanOriginalPricePoints != null && selectedPlanOriginalPricePoints > 0" class="subscription-dialog__orig-price">
            {{ formatPoints(selectedPlanOriginalPricePoints, localeCode) }}
          </span>
          <span :class="['subscription-dialog__price', planTextClass]">{{ formatPoints(selectedPlanPricePoints, localeCode) }}</span>
          <span class="subscription-dialog__validity">/ {{ planValiditySuffix }}</span>
        </div>
        <!-- Description -->
        <p v-if="selectedPlan.description" class="subscription-dialog__desc">
          {{ selectedPlan.description }}
        </p>
        <p v-if="selectedPlatformHasActive" class="subscription-dialog__queue-note">
          {{ t('payment.subscriptionWillQueue') }}
        </p>
        <!-- Rate + Limits grid -->
        <div class="subscription-dialog__limits-grid">
          <div class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.rate') }}</span>
            <div class="subscription-dialog__limit-value">
              <span :class="planTextClass">×{{ selectedPlan.rate_multiplier ?? 1 }}</span>
            </div>
          </div>
          <div v-if="planHasPeakRate(selectedPlan)" class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.peakRate') }}</span>
            <div class="subscription-dialog__limit-value">
              {{ planPeakRateLabel(selectedPlan) }}
            </div>
          </div>
          <div v-if="selectedDailyLimitPoints != null" class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.dailyLimit') }}</span>
            <div class="subscription-dialog__limit-value">{{ formatPoints(selectedDailyLimitPoints, localeCode) }}</div>
          </div>
          <div v-if="selectedWeeklyLimitPoints != null" class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.weeklyLimit') }}</span>
            <div class="subscription-dialog__limit-value">{{ formatPoints(selectedWeeklyLimitPoints, localeCode) }}</div>
          </div>
          <div v-if="selectedMonthlyLimitPoints != null" class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.monthlyLimit') }}</span>
            <div class="subscription-dialog__limit-value">{{ formatPoints(selectedMonthlyLimitPoints, localeCode) }}</div>
          </div>
          <div v-if="!selectedPlanHasQuota" class="subscription-dialog__limit-item">
            <span class="subscription-dialog__limit-label">{{ t('payment.planCard.quota') }}</span>
            <div class="subscription-dialog__limit-value">{{ t('payment.planCard.unlimited') }}</div>
          </div>
        </div>

        <!-- Wallet points check -->
        <div class="subscription-dialog__wallet card">
          <div class="subscription-dialog__wallet-inner">
            <div class="subscription-dialog__wallet-row">
              <span class="subscription-dialog__wallet-label">{{ t('payment.wallet.availablePoints') }}</span>
              <span class="subscription-dialog__wallet-val">{{ formatPoints(checkout.wallet.available_balance, localeCode) }}</span>
            </div>
            <div class="subscription-dialog__wallet-row subscription-dialog__wallet-row--price">
              <span class="subscription-dialog__wallet-label">{{ t('payment.wallet.planPricePoints') }}</span>
              <span class="subscription-dialog__wallet-val">{{ formatPoints(selectedPlanPricePoints, localeCode) }}</span>
            </div>
            <div v-if="hasSufficientSubscriptionPoints" class="subscription-dialog__wallet-row subscription-dialog__wallet-row--remaining">
              <span class="subscription-dialog__wallet-label">{{ t('payment.wallet.remainingAfterPay') }}</span>
              <span class="subscription-dialog__wallet-val">{{ formatPoints(checkout.wallet.available_balance - selectedPlanPricePoints, localeCode) }}</span>
            </div>
            <div v-else class="subscription-insufficient-points">
              <p class="subscription-insufficient-text">
                {{ t('payment.wallet.insufficientPoints') }}
              </p>
              <button
                v-if="!checkout.balance_disabled && enabledMethods.length > 0"
                type="button"
                class="btn btn-secondary btn-sm"
                @click="openRechargeFromSubscription"
              >
                {{ t('payment.rechargeNow') }}
              </button>
            </div>
          </div>
        </div>
        <p v-if="errorMessage" class="recharge-error-text">{{ errorMessage }}</p>
      </div>
      <template #footer>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="submitting"
          @click="closeSubscriptionDialog"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="!canSubmitSubscription || submitting"
          @click="confirmSubscribe"
        >
          <span v-if="submitting" class="views-user-payment-view__text-4">
            <span class="views-user-payment-view__text-5"></span>
            {{ t('common.processing') }}
          </span>
          <span v-else>
            {{ isRenewalSelectedPlan ? t('payment.confirmRenew') : t('payment.wallet.payWithPoints') }} ({{ formatPoints(selectedPlanPricePoints, localeCode) }})
          </span>
        </button>
      </template>
    </BaseDialog>

    <!-- Renewal Plan Selection Modal -->
    <BaseDialog
      :show="showRenewalModal"
      :title="t('payment.selectPlan')"
      width="wide"
      :close-on-click-outside="true"
      @close="closeRenewalModal"
    >
      <div class="renewal-modal-grid">
        <SubscriptionPlanCard
          v-for="plan in renewalPlans"
          :key="plan.id"
          :plan="plan"
          :active-subscriptions="activeSubscriptions"
          :pending-platforms="pendingPlatforms"
          @select="selectPlanFromModal"
        />
      </div>
    </BaseDialog>

    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="views-user-payment-view__panel-27" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="views-user-payment-view__image-2" />
        </div>
      </Transition>
    </Teleport>
  </component>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel, type PeakRateFields } from '@/utils/peak-rate'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeatures, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { platformAccentBarClass, platformBadgeLightClass, platformBadgeClass, platformTextClass, platformLabel } from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import { normalizePaymentCurrency } from '@/components/payment/currency'
import { planValiditySuffix as validitySuffixOf } from '@/components/payment/validity'
import {
  subscriptionPlanHasQuota,
  subscriptionPlanLimitPoints,
  subscriptionPlanOriginalPricePoints,
  subscriptionPlanPricePoints,
} from '@/components/payment/planPoints'
import { rechargeBonusPointsForAmount } from '@/components/payment/rechargeBonus'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { formatCNY, formatPoints } from '@/utils/format'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const props = withDefaults(defineProps<{
  checkoutMode?: 'all' | 'recharge' | 'subscription'
  embedded?: boolean
  showActiveSubscriptions?: boolean
}>(), {
  checkoutMode: 'all',
  embedded: false,
  showActiveSubscriptions: true,
})

const emit = defineEmits<{
  subscriptionUpdated: []
}>()

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const pendingSubscriptions = computed(() => subscriptionStore.pendingSubscriptions)
const pendingPlatforms = computed(() => pendingSubscriptions.value.map(item => item.platform))

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

function subscriptionHasPeakRate(sub: { group?: PeakRateFields | null }): boolean {
  return hasPeakRate(sub.group)
}

function subscriptionPeakRateLabel(sub: { group?: PeakRateFields | null }): string {
  return formatPeakRateWindow(sub.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

function subscriptionGroupHasQuota(sub: {
  group?: {
    daily_limit_points?: number | null
    daily_limit_usd?: number | null
    weekly_limit_points?: number | null
    weekly_limit_usd?: number | null
    monthly_limit_points?: number | null
    monthly_limit_usd?: number | null
  } | null
}): boolean {
  const group = sub.group
  return [
    group?.daily_limit_points ?? group?.daily_limit_usd,
    group?.weekly_limit_points ?? group?.weekly_limit_usd,
    group?.monthly_limit_points ?? group?.monthly_limit_usd,
  ].some((limit) => limit != null)
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const activeTab = ref<'recharge' | 'subscription'>(
  props.checkoutMode === 'subscription' ? 'subscription' : 'recharge',
)
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')
const showRechargeDialog = ref(false)

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: '',
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    alipayMobilePrecreateDeepLink: false,
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId) {
    query.order_id = state.orderId
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: string; orderAmount: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = props.checkoutMode === 'subscription'
      ? '/subscriptions'
      : targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
    emit('subscriptionUpdated')
  }
}

async function onPaymentSuccess() {
  const completedPayment = { ...paymentState.value }
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
    emit('subscriptionUpdated')
  }
  await redirectToPaymentResult(completedPayment)
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_bonus_tiers: [], subscription_usd_to_cny_rate: 0, recharge_fee_rate: 0, help_text: '', help_image_url: '', stripe_publishable_key: '',
  wallet: { balance: 0, available_balance: 0, recharge_balance: 0, bonus_balance: 0, overdraft_amount: 0, frozen_balance: 0, frozen_recharge_balance: 0, frozen_bonus_balance: 0, total_balance: 0, next_expiring_bonus_amount: 0 },
})

const tabs = computed(() => {
  const result: { key: 'recharge' | 'subscription'; label: string }[] = []
  if (props.checkoutMode !== 'subscription' && !checkout.value.balance_disabled) {
    result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  }
  if (props.checkoutMode !== 'recharge') {
    result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  }
  return result
})

const visibleMethods = computed(() => Object.fromEntries(
  Object.entries(getVisibleMethods(checkout.value.methods))
    .filter(([, limit]) => normalizePaymentCurrency(limit.currency) === 'CNY'),
))
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const validAmount = computed(() => amount.value ?? 0)
const rechargeBasePoints = computed(() => Math.max(validAmount.value, 0))
const rechargeBonusPoints = computed(() => rechargeBonusPointsForAmount(
  checkout.value.recharge_bonus_tiers,
  validAmount.value,
))
const rechargeCreditedPoints = computed(() => rechargeBasePoints.value + rechargeBonusPoints.value)

// Adaptive grid: center single card, 2-col for 2 plans, 3-col for 3+
const planGridClass = computed(() => {
  const n = checkout.value.plans.length
  if (n <= 2) return 'views-user-payment-view__state'
  return 'views-user-payment-view__state-2'
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => limit.single_min))
})
const globalMaxAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => limit.single_max))
})

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])
const selectedCurrency = computed(() => normalizePaymentCurrency(selectedLimit.value?.currency))
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function formatWalletAmount(value: number): string {
  return formatPoints(value, localeCode.value)
}

function formatWalletExpiry(value: string): string {
  return new Intl.DateTimeFormat(localeCode.value || undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      display_name: ml?.display_name,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: formatCNY(ml.single_min, localeCode.value) })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: formatCNY(ml.single_max, localeCode.value) })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

const selectedPlanPricePoints = computed(() => selectedPlan.value
  ? subscriptionPlanPricePoints(selectedPlan.value)
  : 0)
const selectedPlanOriginalPricePoints = computed(() => selectedPlan.value
  ? subscriptionPlanOriginalPricePoints(selectedPlan.value)
  : null)
const selectedDailyLimitPoints = computed(() => selectedPlan.value
  ? subscriptionPlanLimitPoints(selectedPlan.value, 'daily')
  : null)
const selectedWeeklyLimitPoints = computed(() => selectedPlan.value
  ? subscriptionPlanLimitPoints(selectedPlan.value, 'weekly')
  : null)
const selectedMonthlyLimitPoints = computed(() => selectedPlan.value
  ? subscriptionPlanLimitPoints(selectedPlan.value, 'monthly')
  : null)
const selectedPlanHasQuota = computed(() => selectedPlan.value
  ? subscriptionPlanHasQuota(selectedPlan.value)
  : false)
const hasSufficientSubscriptionPoints = computed(() =>
  checkout.value.wallet.available_balance + 0.00000001 >= selectedPlanPricePoints.value)
const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && hasSufficientSubscriptionPoints.value
    && !selectedPlanSoldOut.value
    && !selectedPlatformHasPending.value)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (!m) return 'btn-primary'
  if (isBuiltInAlipayMethod(m)) return 'btn-alipay'
  if (isBuiltInWxpayMethod(m)) return 'btn-wxpay'
  if (m === 'stripe') return 'btn-stripe'
  if (m === 'airwallex') return 'btn-airwallex'
  return 'btn-primary'
})

// Subscription confirm: platform accent colors (clean card, no gradient)
const planBadgeClass = computed(() => platformBadgeClass(selectedPlan.value?.group_platform || ''))
const planTextClass = computed(() => platformTextClass(selectedPlan.value?.group_platform || ''))

// Subscription dialog and renewal state
const showSubscriptionDialog = ref(false)
const showRenewalModal = ref(false)
const renewGroupId = ref<string | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const isRenewalSelectedPlan = computed(() => {
  if (!selectedPlan.value) return false
  return activeSubscriptions.value.some(
    s => s.group_id === selectedPlan.value?.group_id && s.status === 'active'
  )
})

const selectedPlanSoldOut = computed(() => selectedPlan.value?.stock_enabled === true
  && (selectedPlan.value.stock_available ?? 0) <= 0)
const selectedPlatformHasPending = computed(() => {
  const platform = selectedPlan.value?.group_platform || ''
  return platform !== '' && pendingPlatforms.value.includes(platform)
})
const selectedPlatformHasActive = computed(() => {
  const platform = selectedPlan.value?.group_platform || ''
  return platform !== '' && activeSubscriptions.value.some(
    subscription => (subscription.status === 'active' || subscription.status === 'suspended')
      && subscription.group?.platform === platform,
  )
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  return validitySuffixOf(selectedPlan.value, t)
})

function planHasPeakRate(plan: SubscriptionPlan): boolean {
  return hasPeakRate(plan)
}

function planPeakRateLabel(plan: SubscriptionPlan): string {
  return formatPeakRateWindow(plan, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

function selectPlan(plan: SubscriptionPlan) {
  if (plan.stock_enabled === true && (plan.stock_available ?? 0) <= 0) {
    appStore.showWarning(t('payment.errors.PLAN_SOLD_OUT'))
    return
  }
  if (pendingPlatforms.value.includes(plan.group_platform || '')) {
    appStore.showWarning(t('payment.errors.SUBSCRIPTION_PLATFORM_PENDING_EXISTS'))
    return
  }
  selectedPlan.value = plan
  errorMessage.value = ''
  showSubscriptionDialog.value = true
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  if (plan.stock_enabled === true && (plan.stock_available ?? 0) <= 0) {
    appStore.showWarning(t('payment.errors.PLAN_SOLD_OUT'))
    return
  }
  if (pendingPlatforms.value.includes(plan.group_platform || '')) {
    appStore.showWarning(t('payment.errors.SUBSCRIPTION_PLATFORM_PENDING_EXISTS'))
    return
  }
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
  showSubscriptionDialog.value = true
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

function closeSubscriptionDialog() {
  if (submitting.value) return
  showSubscriptionDialog.value = false
  selectedPlan.value = null
}

function openRechargeFromSubscription() {
  showSubscriptionDialog.value = false
  openRechargeDialog()
}

function openRenewalForGroup(groupId: string | number) {
  const gid = String(groupId)
  const groupPlans = checkout.value.plans.filter(p => String(p.group_id) === gid)
  if (groupPlans.length === 1) {
    selectPlan(groupPlans[0])
  } else if (groupPlans.length > 1) {
    renewGroupId.value = gid
    showRenewalModal.value = true
  } else {
    appStore.showWarning(t('payment.noPlansForGroup'))
  }
}

function openRechargeDialog() {
  errorMessage.value = ''
  errorHintMessage.value = ''
  showRechargeDialog.value = true
}

function closeRechargeDialog() {
  if (submitting.value) return
  showRechargeDialog.value = false
}

watch(paymentPhase, (phase) => {
  if (phase === 'paying') {
    showRechargeDialog.value = false
  }
})

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || !canSubmitSubscription.value || submitting.value) return
  await createOrder(selectedPlanPricePoints.value, 'subscription', selectedPlan.value.id, {
    paymentType: 'wallet',
  })
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: string, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = orderType === 'subscription'
    ? 'wallet'
    : normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && normalizeVisibleMethod(requestType) === 'alipay'),
      mobilePrecreateDeepLink: checkout.value.alipay_mobile_precreate_deep_link === true,
      useBalance: orderType === 'subscription',
    })
    if (orderType === 'balance' && options.openid) {
      payload.openid = options.openid
    }
    if (orderType === 'balance' && options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
	if (orderType === 'subscription' && !result.wallet_only) {
	  throw new Error('Subscription checkout must be completed with platform points')
    }
    if (result.wallet_only) {
      const wasRenewal = isRenewalSelectedPlan.value
      await authStore.refreshUser()
      await subscriptionStore.fetchActiveSubscriptions(true)
      await subscriptionStore.fetchPendingSubscriptions(true)
      try {
        const res = await paymentAPI.getCheckoutInfo()
        checkout.value = res.data
      } catch {
        // The successful wallet payment should not be reversed by a refresh failure.
      }
      showSubscriptionDialog.value = false
      selectedPlan.value = null
      emit('subscriptionUpdated')
      appStore.showSuccess(result.activation_status === 'pending'
        ? t('payment.subscriptionQueued')
        : wasRenewal ? t('userSubscriptions.renewSuccess') : t('payment.wallet.paymentSuccess'))
      return
    }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && visibleMethod === 'alipay'),
      mobilePrecreateDeepLink: checkout.value.alipay_mobile_precreate_deep_link === true,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'PLAN_SOLD_OUT') {
      errorMessage.value = t('payment.errors.PLAN_SOLD_OUT')
      errorHintMessage.value = ''
      try {
        const res = await paymentAPI.getCheckoutInfo()
        checkout.value = res.data
      } catch {
        // Preserve the inventory conflict as the primary error.
      }
    } else if (apiErr.reason === 'SUBSCRIPTION_PLATFORM_PENDING_EXISTS') {
      errorMessage.value = t('payment.errors.SUBSCRIPTION_PLATFORM_PENDING_EXISTS')
      errorHintMessage.value = ''
      await subscriptionStore.fetchPendingSubscriptions(true).catch(() => {})
    } else if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: string
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    selectedMethod.value = resume.paymentType
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  // Subscription purchases are point-only. Legacy/provider callback parameters may
  // select the plan for confirmation, but must never resume a payment channel.
  if (resume.orderType === 'subscription') {
    return
  }

  if (resume.wechatResumeToken) {
    await createOrder(0, 'balance', undefined, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, 'balance', undefined, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
    })
  }
}

onMounted(async () => {
  if (props.checkoutMode === 'recharge' && route.query.tab === 'subscription') {
    await router.replace({
      path: '/subscriptions',
      query: route.query.group ? { group: route.query.group } : {},
    })
    return
  }
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    await Promise.allSettled([
      subscriptionStore.fetchActiveSubscriptions(),
      subscriptionStore.fetchPendingSubscriptions(),
    ])
    if (enabledMethods.value.length) {
      const order: readonly string[] = METHOD_ORDER
      const sorted = [...enabledMethods.value].sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
      selectedMethod.value = sorted[0]
    }
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
          || (visibleMethods.value[restored.paymentType] ? restored.paymentType : '')
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (checkout.value.balance_disabled && props.checkoutMode !== 'recharge') {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123 or ?plan_id=456
    if (props.checkoutMode === 'subscription' || route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (route.query.group) {
        openRenewalForGroup(String(route.query.group))
      } else if (route.query.plan_id) {
        const planId = typeof route.query.plan_id === 'string' ? route.query.plan_id : ''
        const targetPlan = checkout.value.plans.find(p => String(p.id) === planId)
        if (targetPlan) {
          selectPlan(targetPlan)
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
})

defineExpose({
  openRenewalForGroup,
  selectPlan,
  openRechargeDialog,
  checkout,
})
</script>

<style scoped>
.wallet-summary {
  overflow: hidden;
  padding: 1.25rem;
}

.wallet-summary__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.wallet-summary__eyebrow,
.wallet-summary__available span,
.wallet-summary__row span,
.recharge-dialog-form__account span {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.wallet-summary__account {
  margin-top: 0.25rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 600;
}

.wallet-summary__available {
  display: grid;
  justify-items: end;
  gap: 0.25rem;
  min-width: 8rem;
}

.wallet-summary__available strong {
  color: var(--color-primary);
  font-size: var(--font-size-2xl);
  font-weight: 700;
}

.wallet-summary__rows {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem 1.5rem;
  padding-top: 1rem;
}

.wallet-summary__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-width: 0;
}

.wallet-summary__row strong {
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  white-space: nowrap;
}

.wallet-summary__row--warning strong {
  color: var(--color-warning);
}

.wallet-summary__expiry {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
  color: var(--color-warning);
  font-size: var(--font-size-sm);
  line-height: 1.5;
}

.wallet-summary__recharge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  min-height: 2.75rem;
}

.recharge-dialog-form {
  display: grid;
  gap: 1.125rem;
}

.recharge-dialog-form__account {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.875rem;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}

.recharge-dialog-form__account-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-weight: 500;
}

.recharge-dialog-form__account-name {
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recharge-rate-preview {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin: 0 0 0.5rem;
  padding: 0.375rem 0.625rem;
  border-radius: var(--radius-sm);
  background: var(--color-primary-subtle, rgba(99, 102, 241, 0.08));
  color: var(--color-primary);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.recharge-dialog-form__section {
  display: grid;
  gap: 0.5rem;
}

.recharge-dialog-form__summary {
  display: grid;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  border-radius: var(--radius-lg);
  background: var(--color-bg-tertiary);
  border: 1px solid var(--color-border);
}

.recharge-summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.recharge-summary-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.recharge-summary-val {
  font-weight: 500;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.recharge-summary-row--highlight {
  padding: 0.375rem 0;
  border-top: 1px dashed var(--color-border);
  border-bottom: 1px dashed var(--color-border);
}

.recharge-summary-val--pay {
  color: var(--color-primary);
  font-weight: 600;
  font-size: var(--font-size-base);
}

.recharge-summary-row--divider {
  padding-top: 0.375rem;
  border-top: 1px solid var(--color-border);
}

.recharge-summary-val--bonus {
  color: var(--color-success, #10b981);
  font-weight: 500;
}

.recharge-summary-row--total {
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
}

.recharge-summary-val--total {
  color: var(--color-primary);
  font-size: var(--font-size-lg);
  font-weight: 700;
}

.recharge-error-text {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-danger, #ef4444);
}

.recharge-hint-text {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}

/* Subscription Dialog */
.subscription-dialog-form {
  display: grid;
  gap: 1rem;
}

.subscription-dialog__header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.subscription-dialog__badge {
  font-size: var(--font-size-xs);
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-full);
  font-weight: 600;
}

.subscription-dialog__title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text-primary);
}

.subscription-dialog__price-row {
  display: flex;
  align-items: baseline;
  gap: 0.375rem;
}

.subscription-dialog__orig-price {
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
  text-decoration: line-through;
}

.subscription-dialog__price {
  font-size: var(--font-size-2xl);
  font-weight: 700;
}

.subscription-dialog__validity {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.subscription-dialog__desc {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.subscription-dialog__queue-note {
  padding: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--color-text-warning) 30%, transparent);
  border-radius: var(--radius-md);
  background: var(--glass-tint-warning);
  color: var(--color-text-warning);
  font-size: var(--type-caption-size);
  line-height: var(--type-caption-line-height);
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.subscription-dialog__limits-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
  padding: 0.75rem;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.subscription-dialog__limit-item {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.subscription-dialog__limit-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}

.subscription-dialog__limit-value {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.subscription-dialog__wallet {
  padding: 0.875rem 1rem;
  border-radius: var(--radius-lg);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
}

.subscription-dialog__wallet-inner {
  display: grid;
  gap: 0.5rem;
}

.subscription-dialog__wallet-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.subscription-dialog__wallet-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.subscription-dialog__wallet-val {
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.subscription-dialog__wallet-row--price .subscription-dialog__wallet-val {
  color: var(--color-primary);
  font-size: var(--font-size-base);
}

.subscription-dialog__wallet-row--remaining {
  padding-top: 0.375rem;
  border-top: 1px solid var(--color-border);
}

.subscription-dialog__wallet-row--remaining .subscription-dialog__wallet-val {
  color: var(--color-success, #10b981);
}

.subscription-insufficient-points {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 0.5rem;
  border-top: 1px dashed var(--color-border);
}

.subscription-insufficient-text {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-warning, #f59e0b);
  font-weight: 500;
}

/* Renewal Modal Grid */
.renewal-modal-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

@media (max-width: 639px) {
  .wallet-summary__header {
    align-items: stretch;
    flex-direction: column;
  }

  .wallet-summary__available {
    align-items: flex-start;
  }

  .wallet-summary__rows {
    grid-template-columns: 1fr;
  }

  .subscription-dialog__limits-grid {
    grid-template-columns: 1fr;
  }
}
</style>
