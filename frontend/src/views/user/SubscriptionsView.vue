<template>
  <AppLayout>
    <div class="views-user-subscriptions-view__panel">
      <section>
        <h2 class="views-user-subscriptions-view__heading">
          {{ t('userSubscriptions.currentSubscriptions') }}
        </h2>

        <!-- Loading State -->
        <div v-if="loading" class="views-user-subscriptions-view__panel-2">
          <div
            class="views-user-subscriptions-view__panel-3"
          ></div>
        </div>

        <!-- Empty State -->
        <div v-else-if="subscriptions.length === 0" class="views-user-subscriptions-view__panel-4 card">
          <div
            class="views-user-subscriptions-view__panel-5"
          >
            <Icon name="creditCard" size="xl" class="views-user-subscriptions-view__icon" />
          </div>
          <h3 class="views-user-subscriptions-view__heading">
            {{ t('userSubscriptions.noActiveSubscriptions') }}
          </h3>
          <p class="views-user-subscriptions-view__description">
            {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
          </p>
        </div>

        <!-- Subscriptions Grid -->
        <div v-else class="views-user-subscriptions-view__panel-6">
          <div
            v-for="subscription in subscriptions"
            :key="subscription.id"
            class="views-user-subscriptions-view__panel-7"
            :class="platformBorderClass(subscription.group?.platform || '')"
          >
          <!-- Header -->
          <div
            class="views-user-subscriptions-view__panel-8"
          >
            <div class="views-user-subscriptions-view__panel-9">
              <div :class="['views-user-subscriptions-view__panel-19', platformAccentDotClass(subscription.group?.platform || '')]" />
              <div>
                <div class="views-user-subscriptions-view__panel-10">
                  <h3 class="views-user-subscriptions-view__heading-2">
                    {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                  </h3>
                  <span :class="['views-user-subscriptions-view__text-6', platformBadgeClass(subscription.group?.platform || '')]">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="views-user-subscriptions-view__description-2">
                  {{ subscription.group.description }}
                </p>
                <div class="views-user-subscriptions-view__panel-11">
                  <span>{{ t('payment.planCard.rate') }}: ×{{ subscription.group?.rate_multiplier ?? 1 }}</span>
                  <span v-if="subscriptionHasPeakRate(subscription)" class="views-user-subscriptions-view__text">
                    {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                  </span>
                </div>
              </div>
            </div>
            <div class="views-user-subscriptions-view__panel-10">
              <span
                :class="[
                  'views-user-subscriptions-view__text-7',
                  subscription.status === 'active'
                    ? 'views-user-subscriptions-view__text-9'
                    : subscription.status === 'expired'
                      ? 'views-user-subscriptions-view__text-10'
                      : 'views-user-subscriptions-view__text-11'
                ]"
              >
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active' && paymentEnabled"
                :class="['views-user-subscriptions-view__action', platformButtonClass(subscription.group?.platform || '')]"
                @click="handleRenewSubscription(subscription)"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </div>

          <!-- Usage Progress -->
          <div class="views-user-subscriptions-view__panel-12">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="views-user-subscriptions-view__panel-13">
              <span class="views-user-subscriptions-view__description">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="views-user-subscriptions-view__panel-13">
              <span class="views-user-subscriptions-view__description">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="views-user-subscriptions-view__text-2">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="subscriptionLimitPoints(subscription, 'daily')" class="views-user-subscriptions-view__panel-14">
              <div class="views-user-subscriptions-view__panel-15">
                <span class="views-user-subscriptions-view__text-3">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="views-user-subscriptions-view__text-4">
                  {{ formatPointRange(subscriptionUsagePoints(subscription, 'daily'), subscriptionLimitPoints(subscription, 'daily')) }}
                </span>
              </div>
              <div class="views-user-subscriptions-view__panel-16">
                <div
                  class="views-user-subscriptions-view__panel-17"
                  :class="
                    getProgressBarClass(
                      subscriptionUsagePoints(subscription, 'daily'),
                      subscriptionLimitPoints(subscription, 'daily')
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscriptionUsagePoints(subscription, 'daily'),
                      subscriptionLimitPoints(subscription, 'daily')
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.daily_window_start"
                class="views-user-subscriptions-view__description-3"
              >
                {{ formatDailyUsageWindow(subscription) }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="subscriptionLimitPoints(subscription, 'weekly')" class="views-user-subscriptions-view__panel-14">
              <div class="views-user-subscriptions-view__panel-15">
                <span class="views-user-subscriptions-view__text-3">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="views-user-subscriptions-view__text-4">
                  {{ formatPointRange(subscriptionUsagePoints(subscription, 'weekly'), subscriptionLimitPoints(subscription, 'weekly')) }}
                </span>
              </div>
              <div class="views-user-subscriptions-view__panel-16">
                <div
                  class="views-user-subscriptions-view__panel-17"
                  :class="
                    getProgressBarClass(
                      subscriptionUsagePoints(subscription, 'weekly'),
                      subscriptionLimitPoints(subscription, 'weekly')
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscriptionUsagePoints(subscription, 'weekly'),
                      subscriptionLimitPoints(subscription, 'weekly')
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.weekly_window_start"
                class="views-user-subscriptions-view__description-3"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168)
                  })
                }}
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="subscriptionLimitPoints(subscription, 'monthly')" class="views-user-subscriptions-view__panel-14">
              <div class="views-user-subscriptions-view__panel-15">
                <span class="views-user-subscriptions-view__text-3">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="views-user-subscriptions-view__text-4">
                  {{ formatPointRange(subscriptionUsagePoints(subscription, 'monthly'), subscriptionLimitPoints(subscription, 'monthly')) }}
                </span>
              </div>
              <div class="views-user-subscriptions-view__panel-16">
                <div
                  class="views-user-subscriptions-view__panel-17"
                  :class="
                    getProgressBarClass(
                      subscriptionUsagePoints(subscription, 'monthly'),
                      subscriptionLimitPoints(subscription, 'monthly')
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscriptionUsagePoints(subscription, 'monthly'),
                      subscriptionLimitPoints(subscription, 'monthly')
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.monthly_window_start"
                class="views-user-subscriptions-view__description-3"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720)
                  })
                }}
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !subscriptionLimitPoints(subscription, 'daily') &&
                !subscriptionLimitPoints(subscription, 'weekly') &&
                !subscriptionLimitPoints(subscription, 'monthly')
              "
              class="views-user-subscriptions-view__panel-18"
            >
              <div class="views-user-subscriptions-view__panel-9">
                <span class="views-user-subscriptions-view__text-5">∞</span>
                <div>
                  <p class="views-user-subscriptions-view__description-4">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="views-user-subscriptions-view__description-5">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
          </div>
        </div>
      </section>

      <section v-if="paymentEnabled">
        <h2 class="views-user-subscriptions-view__heading">
          {{ t('userSubscriptions.availableSubscriptions') }}
        </h2>
        <PaymentView
          ref="paymentViewRef"
          checkout-mode="subscription"
          embedded
          :show-active-subscriptions="false"
          @subscription-updated="loadSubscriptions"
        />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import PaymentView from '@/views/user/PaymentView.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTimeToMinute, formatPointAmount, formatPoints } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformBorderClass, platformBadgeClass, platformButtonClass, platformLabel } from '@/utils/platformColors'
import {
  getExpirationDateRelation,
  getRemainingDurationParts,
  isOneTimeDailyQuota,
  type RemainingDurationParts
} from '@/utils/subscriptionQuota'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'status-fill--warning'
    case 'openai': return 'status-fill--success'
    default: return 'status-fill--neutral'
  }
}

const { t } = useI18n()
const appStore = useAppStore()

const paymentViewRef = ref<InstanceType<typeof PaymentView> | null>(null)
const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)
const paymentEnabled = computed(
  () => appStore.cachedPublicSettings?.payment_enabled === true,
)

function handleRenewSubscription(subscription: UserSubscription) {
  if (paymentViewRef.value) {
    paymentViewRef.value.openRenewalForGroup(subscription.group_id)
  }
}

type PointQuotaWindow = 'daily' | 'weekly' | 'monthly'

function subscriptionUsagePoints(subscription: UserSubscription, window: PointQuotaWindow): number {
  return subscription[`${window}_usage_points`] ?? subscription[`${window}_usage_usd`] ?? 0
}

function subscriptionLimitPoints(subscription: UserSubscription, window: PointQuotaWindow): number | null {
  return subscription.group?.[`${window}_limit_points`] ?? subscription.group?.[`${window}_limit_usd`] ?? null
}

function formatPointRange(used: number, limit: number | null): string {
  return limit == null ? `${formatPointAmount(used)} / —` : `${formatPointAmount(used)} / ${formatPoints(limit)}`
}

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(subscription.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'status-fill--neutral'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'status-fill--danger'
  if (percentage >= 70) return 'status-fill--warning'
  return 'status-fill--success'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const relation = getExpirationDateRelation(expires, now)

  if (relation === null) return ''

  if (relation === 'expired') {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateTimeToMinute(expires)

  if (relation === 'today') {
    return `${dateStr} (${t('common.today')})`
  }
  if (relation === 'tomorrow') {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (diff <= 0) return 'views-user-subscriptions-view__state'
  if (days <= 3) return 'views-user-subscriptions-view__state-2'
  if (days <= 7) return 'views-user-subscriptions-view__state-3'
  return 'views-user-subscriptions-view__text-2'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>
