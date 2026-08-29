<template>
  <div
    :class="[
      'subscription-plan-card',
      borderClass,
    ]"
  >
    <!-- Colored top accent bar -->
    <div :class="['subscription-plan-card__accent', accentClass]" />

    <div class="subscription-plan-card__body">
      <!-- Header: name + badge + price -->
      <div class="subscription-plan-card__header">
        <div class="subscription-plan-card__intro" data-testid="plan-card-intro">
          <h3
            :title="plan.name"
            class="subscription-plan-card__title"
            data-testid="plan-card-title"
          >
            {{ plan.name }}
          </h3>
          <p v-if="plan.description" class="subscription-plan-card__description">
            {{ plan.description }}
          </p>
        </div>
        <div class="subscription-plan-card__pricing" data-testid="plan-card-pricing">
          <div class="subscription-plan-card__price-row" data-testid="plan-card-price-row">
            <span :class="['subscription-plan-card__price', textClass]">{{ formatPoints(planPricePoints, localeCode) }}</span>
          </div>
          <div class="subscription-plan-card__period-row" data-testid="plan-card-period-row">
            <span :class="['subscription-plan-card__platform', badgeLightClass]" data-testid="plan-card-platform">
              {{ pLabel }}
            </span>
            <span class="subscription-plan-card__period">/ {{ validitySuffix }}</span>
          </div>
          <div v-if="planOriginalPricePoints != null && planOriginalPricePoints > 0" class="subscription-plan-card__discount-row">
            <span class="subscription-plan-card__original-price">{{ formatPoints(planOriginalPricePoints, localeCode) }}</span>
            <span :class="['subscription-plan-card__discount', discountClass]">{{ discountText }}</span>
          </div>
        </div>
      </div>

      <!-- Group quota info (compact) -->
      <div class="subscription-plan-card__quota-grid">
        <div class="subscription-plan-card__quota-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.rate') }}</span>
          <span class="subscription-plan-card__quota-value">{{ rateDisplay }}</span>
        </div>
        <div v-if="hasPeakRate" class="subscription-plan-card__peak-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.peakRate') }}</span>
          <span class="subscription-plan-card__peak-value">{{ peakRateDisplay }}</span>
        </div>
        <div v-if="dailyLimitPoints != null" class="subscription-plan-card__quota-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="subscription-plan-card__quota-value">{{ formatPoints(dailyLimitPoints, localeCode) }}</span>
        </div>
        <div v-if="weeklyLimitPoints != null" class="subscription-plan-card__quota-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="subscription-plan-card__quota-value">{{ formatPoints(weeklyLimitPoints, localeCode) }}</span>
        </div>
        <div v-if="monthlyLimitPoints != null" class="subscription-plan-card__quota-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="subscription-plan-card__quota-value">{{ formatPoints(monthlyLimitPoints, localeCode) }}</span>
        </div>
        <div v-if="!planHasQuota" class="subscription-plan-card__quota-row">
          <span class="subscription-plan-card__quota-label">{{ t('payment.planCard.quota') }}</span>
          <span class="subscription-plan-card__quota-value">{{ t('payment.planCard.unlimited') }}</span>
        </div>
      </div>

      <!-- Features list (compact) -->
      <div v-if="plan.features.length > 0" class="subscription-plan-card__feature-list">
        <div v-for="feature in plan.features" :key="feature" class="subscription-plan-card__feature">
          <svg :class="['subscription-plan-card__feature-icon', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="subscription-plan-card__feature-text">{{ feature }}</span>
        </div>
      </div>

      <div class="subscription-plan-card__spacer" />

      <!-- Subscribe Button -->
      <button
        type="button"
        :class="['subscription-plan-card__subscribe', btnClass]"
        @click="emit('select', plan)"
      >
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { useAppStore } from '@/stores/app'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { formatPoints } from '@/utils/format'
import { planValiditySuffix } from './validity'
import {
  subscriptionPlanHasQuota,
  subscriptionPlanLimitPoints,
  subscriptionPlanOriginalPricePoints,
  subscriptionPlanPricePoints,
} from './planPoints'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBorderClass,
  platformTextClass,
  platformIconClass,
  platformButtonClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const accentClass = computed(() => platformAccentBarClass(platform.value))
const borderClass = computed(() => platformBorderClass(platform.value))
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const btnClass = computed(() => platformButtonClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))
const planPricePoints = computed(() => subscriptionPlanPricePoints(props.plan))
const planOriginalPricePoints = computed(() => subscriptionPlanOriginalPricePoints(props.plan))
const dailyLimitPoints = computed(() => subscriptionPlanLimitPoints(props.plan, 'daily'))
const weeklyLimitPoints = computed(() => subscriptionPlanLimitPoints(props.plan, 'weekly'))
const monthlyLimitPoints = computed(() => subscriptionPlanLimitPoints(props.plan, 'monthly'))
const planHasQuota = computed(() => subscriptionPlanHasQuota(props.plan))

const discountText = computed(() => {
  if (!planOriginalPricePoints.value || planOriginalPricePoints.value <= 0) return ''
  const pct = Math.round((1 - planPricePoints.value / planOriginalPricePoints.value) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const appStore = useAppStore()

const hasPeakRate = computed(() => groupHasPeakRate(props.plan))

const peakRateDisplay = computed(() => {
  return formatPeakRateWindow(props.plan, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
})

const validitySuffix = computed(() => planValiditySuffix(props.plan, t))
</script>

<style scoped>
.subscription-plan-card {
  position: relative;
  isolation: isolate;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-width: 1px;
  border-style: solid;
  border-radius: var(--radius-xl);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
  transition: transform 180ms ease, border-color 180ms ease, box-shadow 180ms ease;
}

.subscription-plan-card::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -1;
  border-radius: inherit;
  background-color: var(--glass-layer-content-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate));
  pointer-events: none;
  transition: backdrop-filter 200ms ease, -webkit-backdrop-filter 200ms ease;
}

.subscription-plan-card:hover {
  transform: translateY(-0.125rem);
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight-hover) inset;
}

.subscription-plan-card:hover::before {
  -webkit-backdrop-filter: blur(var(--glass-layer-content-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-content-blur-hover)) saturate(var(--glass-saturate-hover));
}

.subscription-plan-card__accent {
  height: 0.375rem;
}

.subscription-plan-card__body {
  display: flex;
  flex: 1 1 0%;
  flex-direction: column;
  padding: 1rem;
}

.subscription-plan-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.subscription-plan-card__intro {
  flex: 1 1 0%;
  min-width: 0;
}

.subscription-plan-card__title {
  height: 3rem;
  min-width: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-bold);
  line-height: 1.5rem;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.subscription-plan-card__description {
  display: -webkit-box;
  margin-top: 0.125rem;
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1.625;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.subscription-plan-card__pricing {
  flex-shrink: 0;
  text-align: right;
}

.subscription-plan-card__price-row,
.subscription-plan-card__period-row,
.subscription-plan-card__discount-row {
  display: flex;
  justify-content: flex-end;
}

.subscription-plan-card__price-row {
  align-items: baseline;
  gap: 0.25rem;
}

.subscription-plan-card__price {
  font-size: var(--font-size-2xl);
  font-weight: 800;
  line-height: 2rem;
  letter-spacing: 0;
}

.subscription-plan-card__period-row {
  align-items: center;
  gap: 0.25rem;
}

.subscription-plan-card__platform {
  display: inline-flex;
  flex-shrink: 0;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-full);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-medium);
}

.subscription-plan-card__period,
.subscription-plan-card__original-price,
.subscription-plan-card__quota-label {
  color: var(--color-text-tertiary);
}

.subscription-plan-card__period {
  font-size: var(--font-size-2xs);
}

.subscription-plan-card__discount-row {
  align-items: center;
  gap: 0.375rem;
  margin-top: 0.125rem;
}

.subscription-plan-card__original-price {
  font-size: var(--font-size-xs);
  line-height: 1rem;
  text-decoration: line-through;
}

.subscription-plan-card__discount {
  padding: 0.125rem 0.25rem;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-3xs);
  font-weight: var(--font-weight-semibold);
}

.subscription-plan-card__quota-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  margin-bottom: 0.75rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background-color: var(--glass-layer-inset-bg);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1rem;
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.subscription-plan-card__quota-row,
.subscription-plan-card__peak-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.subscription-plan-card__peak-row {
  grid-column: span 2 / span 2;
  gap: 0.5rem;
}

.subscription-plan-card__quota-value {
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
}

.subscription-plan-card__peak-value {
  color: var(--color-text-warning);
  font-weight: var(--font-weight-medium);
  text-align: right;
}

.subscription-plan-card__feature-list {
  margin-bottom: 0.75rem;
}

.subscription-plan-card__feature-list > :not([hidden]) ~ :not([hidden]) {
  margin-top: 0.25rem;
}

.subscription-plan-card__feature {
  display: flex;
  align-items: flex-start;
  gap: 0.375rem;
}

.subscription-plan-card__feature-icon {
  flex-shrink: 0;
  width: 0.875rem;
  height: 0.875rem;
  margin-top: 0.125rem;
}

.subscription-plan-card__feature-text {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1rem;
}

.subscription-plan-card__spacer {
  flex: 1 1 0%;
}

.subscription-plan-card__subscribe {
  width: 100%;
  padding: 0.625rem 1rem;
  border-radius: var(--radius-lg);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  line-height: 1.25rem;
  transition: transform 150ms ease;
}

.subscription-plan-card__subscribe:active {
  transform: scale(0.98);
}

@media (prefers-reduced-motion: reduce) {
  .subscription-plan-card,
  .subscription-plan-card::before,
  .subscription-plan-card__subscribe {
    transition-duration: 1ms;
  }
}
</style>
