<template>
  <div class="payment-method-selector-root">
    <label class="components-payment-payment-method-selector__label payment-method-selector-label">
      {{ t('payment.paymentMethod') }}
    </label>
    <div
      data-testid="payment-method-grid"
      class="components-payment-payment-method-selector__panel payment-method-grid"
    >
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :title="methodLabel(method)"
        :disabled="!method.available"
        :class="[
          'components-payment-payment-method-selector__action payment-method-btn',
          !method.available
            ? 'components-payment-payment-method-selector__action-2 payment-method-btn--disabled'
            : selected === method.type
              ? ['payment-method-btn--selected', methodSelectedClass(method.type)]
              : 'components-payment-payment-method-selector__action-3',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="components-payment-payment-method-selector__text payment-method-btn-content">
          <img :src="methodIcon(method.type)" :alt="methodLabel(method)" class="components-payment-payment-method-selector__image payment-method-icon" />
          <span class="components-payment-payment-method-selector__text-2 payment-method-info">
            <span data-testid="payment-method-label" class="components-payment-payment-method-selector__text-3 payment-method-name">
              {{ methodLabel(method) }}
            </span>
            <span
              v-if="method.fee_rate > 0"
              class="components-payment-payment-method-selector__text-4 payment-method-fee"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { METHOD_ORDER, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import paymentIcon from '@/assets/icons/payment.svg'

export interface PaymentMethodOption {
  type: string
  display_name?: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
  credit_card: paymentIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodIcon(type: string): string {
  if (isBuiltInAlipayMethod(type)) return METHOD_ICONS.alipay
  if (isBuiltInWxpayMethod(type)) return METHOD_ICONS.wxpay
  if (type === 'airwallex') return METHOD_ICONS.airwallex
  return METHOD_ICONS[type] || paymentIcon
}

function methodLabel(method: PaymentMethodOption): string {
  return method.display_name || t(`payment.methods.${method.type}`, method.type)
}

function methodSelectedClass(type: string): string {
  if (isBuiltInAlipayMethod(type)) return 'components-payment-payment-method-selector__state method-btn--alipay'
  if (isBuiltInWxpayMethod(type)) return 'components-payment-payment-method-selector__state-2 method-btn--wxpay'
  if (type === 'stripe') return 'components-payment-payment-method-selector__state-3 method-btn--stripe'
  if (type === 'airwallex') return 'components-payment-payment-method-selector__state-4 method-btn--airwallex'
  return 'components-payment-payment-method-selector__state-5 method-btn--default'
}
</script>

<style scoped>
.payment-method-selector-root {
  display: grid;
  gap: 0.5rem;
}

.payment-method-selector-label {
  display: block;
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.375rem;
}

.payment-method-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(135px, 1fr)) !important;
  gap: 0.625rem;
}

.payment-method-btn {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease-in-out;
  min-height: 3.25rem;
}

.payment-method-btn:hover:not(:disabled) {
  background: var(--color-bg-tertiary);
  border-color: var(--color-border-hover, var(--color-primary));
}

.payment-method-btn-content {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  width: 100%;
  min-width: 0;
}

.payment-method-icon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  object-fit: contain;
}

.payment-method-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  line-height: 1.25;
}

.payment-method-name {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.payment-method-fee {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  margin-top: 0.125rem;
  white-space: nowrap;
}

.payment-method-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  background: var(--color-bg-secondary);
}

.payment-method-btn--selected {
  font-weight: 600;
  border-width: 1.5px;
}

.method-btn--alipay {
  border-color: #1677ff !important;
  background: rgba(22, 119, 255, 0.08) !important;
  box-shadow: 0 0 0 1px #1677ff;
}

.method-btn--wxpay {
  border-color: #07c160 !important;
  background: rgba(7, 193, 96, 0.08) !important;
  box-shadow: 0 0 0 1px #07c160;
}

.method-btn--stripe {
  border-color: #635bff !important;
  background: rgba(99, 91, 255, 0.08) !important;
  box-shadow: 0 0 0 1px #635bff;
}

.method-btn--airwallex {
  border-color: #612fff !important;
  background: rgba(97, 47, 255, 0.08) !important;
  box-shadow: 0 0 0 1px #612fff;
}

.method-btn--default {
  border-color: var(--color-primary) !important;
  background: var(--color-primary-subtle, rgba(99, 102, 241, 0.08)) !important;
  box-shadow: 0 0 0 1px var(--color-primary);
}
</style>
