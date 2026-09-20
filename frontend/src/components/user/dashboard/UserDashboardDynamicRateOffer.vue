<template>
  <section
    v-if="activeOffers.length > 0"
    class="dashboard-dynamic-rate-offer card"
    :aria-label="t('dashboard.dynamicRateOffer.title')"
  >
    <header class="dashboard-dynamic-rate-offer__header">
      <Icon name="sparkles" size="sm" class="dashboard-dynamic-rate-offer__icon" />
      <h2 class="dashboard-dynamic-rate-offer__title">{{ t('dashboard.dynamicRateOffer.title') }}</h2>
      <span class="dashboard-dynamic-rate-offer__count">{{ t('dashboard.dynamicRateOffer.groups', { count: activeOffers.length }) }}</span>
    </header>
    <ul class="dashboard-dynamic-rate-offer__list">
      <li v-for="offer in activeOffers" :key="offer.group_id" class="dashboard-dynamic-rate-offer__item">
        <strong class="dashboard-dynamic-rate-offer__group">{{ offer.group_name || offer.rule_name }}</strong>
        <span class="dashboard-dynamic-rate-offer__discount" :title="t('dashboard.dynamicRateOffer.discountDetail', discountValues(offer))">
          {{ t('dashboard.dynamicRateOffer.discount', discountValues(offer)) }}
        </span>
        <div class="dashboard-dynamic-rate-offer__participation">
          <span
            class="dashboard-dynamic-rate-offer__status"
            :class="{ 'dashboard-dynamic-rate-offer__status--active': offer.status === 'participating' }"
          >{{ t(`dashboard.dynamicRateOffer.status.${offer.status}`) }}</span>
          <span>{{ t('dashboard.dynamicRateOffer.conditions') }}</span>
          <span>{{ t('dashboard.dynamicRateOffer.rechargeOnly') }}</span>
          <span>{{ offer.activation_spend > 0
            ? t('dashboard.dynamicRateOffer.spendRequired', { amount: formatAmount(offer.activation_spend) })
            : t('dashboard.dynamicRateOffer.noSpendRequired') }}</span>
          <span v-if="offer.activation_spend > 0">{{ t('dashboard.dynamicRateOffer.currentSpend', { amount: formatAmount(offer.usage_7d) }) }}</span>
          <span v-if="offer.personal_quota_amount > 0">{{ t('dashboard.dynamicRateOffer.personalQuota', {
            used: formatAmount(offer.personal_used_amount), amount: formatAmount(offer.personal_quota_amount),
          }) }}</span>
          <span v-if="participationRequirement(offer)">{{ participationRequirement(offer) }}</span>
        </div>
        <div class="dashboard-dynamic-rate-offer__expiry">
          <Icon name="clock" size="xs" />
          <span>{{ t('dashboard.dynamicRateOffer.endsAt') }}</span>
          <time :datetime="offer.end_at">{{ formatDateTimeToMinute(offer.end_at, locale) }}</time>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getDynamicRateOffers } from '@/api/userLevel'
import Icon from '@/components/icons/Icon.vue'
import type { DynamicRateOffer } from '@/types'
import { formatDateTimeToMinute } from '@/utils/format'

const { t, locale } = useI18n()
const offers = ref<DynamicRateOffer[]>([])
const now = ref(Date.now())
const activeOffers = computed(() => offers.value.filter(offer =>
  Date.parse(offer.start_at) <= now.value && now.value < Date.parse(offer.end_at)
  && Number.isFinite(offer.discount_coefficient)
  && offer.discount_coefficient > 0 && offer.discount_coefficient < 1
))

function discountValues(offer: DynamicRateOffer) {
  const number = new Intl.NumberFormat(locale.value, { maximumFractionDigits: 3 })
  return {
    percent: number.format((1 - offer.discount_coefficient) * 100),
    coefficient: String(offer.discount_coefficient),
  }
}

function formatAmount(value: number) {
  return new Intl.NumberFormat(locale.value, { maximumFractionDigits: 8 }).format(value)
}

function participationRequirement(offer: DynamicRateOffer) {
  switch (offer.status) {
    case 'group_unavailable':
    case 'subscription_required':
    case 'subscription_limited':
    case 'level_required':
      return t(`dashboard.dynamicRateOffer.requirement.${offer.status}`)
    default:
      return ''
  }
}

let clock: ReturnType<typeof setInterval> | undefined
let refresh: ReturnType<typeof setInterval> | undefined
let disposed = false
let fetching = false
async function loadOffers() {
  if (fetching || disposed) return
  fetching = true
  now.value = Date.now()
  try {
    const result = await getDynamicRateOffers()
    if (!disposed) offers.value = result
  } catch {
    if (!disposed) offers.value = []
  } finally { fetching = false }
}
function refreshOnVisible() {
  if (document.visibilityState === 'visible') void loadOffers()
}

onMounted(() => {
  void loadOffers()
  clock = setInterval(() => { now.value = Date.now() }, 1000)
  refresh = setInterval(() => { if (document.visibilityState === 'visible') void loadOffers() }, 60_000)
  document.addEventListener('visibilitychange', refreshOnVisible)
})
onUnmounted(() => {
  disposed = true
  clearInterval(clock)
  clearInterval(refresh)
  document.removeEventListener('visibilitychange', refreshOnVisible)
})
</script>

<style scoped>
.dashboard-dynamic-rate-offer { overflow: hidden; }
.dashboard-dynamic-rate-offer__header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  border-bottom: 1px solid var(--color-border-subtle);
}
.dashboard-dynamic-rate-offer__icon { flex: 0 0 auto; color: var(--color-text-success); }
.dashboard-dynamic-rate-offer__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 650;
}
.dashboard-dynamic-rate-offer__count {
  margin-left: auto;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  white-space: nowrap;
}
.dashboard-dynamic-rate-offer__list { margin: 0; padding: 0; list-style: none; }
.dashboard-dynamic-rate-offer__item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: baseline;
  gap: 0.25rem 0.75rem;
  padding: 0.625rem 1rem;
}
.dashboard-dynamic-rate-offer__item + .dashboard-dynamic-rate-offer__item { border-top: 1px solid var(--color-border-subtle); }
.dashboard-dynamic-rate-offer__group {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}
.dashboard-dynamic-rate-offer__discount {
  color: var(--color-text-success);
  font-size: var(--font-size-sm);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.dashboard-dynamic-rate-offer__expiry {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.25rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}
.dashboard-dynamic-rate-offer__participation {
  grid-column: 1 / -1;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.25rem 0.5rem;
  overflow-wrap: anywhere;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}
.dashboard-dynamic-rate-offer__status {
  color: var(--color-text-warning);
  font-weight: 600;
}
.dashboard-dynamic-rate-offer__status--active { color: var(--color-text-success); }
.dashboard-dynamic-rate-offer__expiry time {
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
}
</style>
