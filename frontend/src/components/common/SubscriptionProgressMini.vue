<template>
  <div v-if="hasActiveSubscriptions" class="components-common-subscription-progress-mini__panel" ref="containerRef">
    <!-- Mini Progress Display -->
    <button
      ref="triggerRef"
      @click="toggleTooltip"
      class="components-common-subscription-progress-mini__action"
      :title="t('subscriptionProgress.viewDetails')"
    >
      <Icon name="creditCard" size="sm" class="components-common-subscription-progress-mini__icon" />
      <div class="components-common-subscription-progress-mini__panel-2">
        <!-- Combined progress indicator -->
        <div class="components-common-subscription-progress-mini__panel-3">
          <div
            v-for="(sub, index) in displaySubscriptions.slice(0, 3)"
            :key="index"
            class="components-common-subscription-progress-mini__panel-4"
            :class="getProgressDotClass(sub)"
          ></div>
        </div>
        <span class="components-common-subscription-progress-mini__text">
          {{ activeSubscriptions.length }}
        </span>
      </div>
    </button>

    <!-- Hover/Click Tooltip -->
    <Teleport to="body">
      <transition name="dropdown">
        <div
          v-if="tooltipOpen"
          ref="panelRef"
          :style="panelStyle"
          class="components-common-subscription-progress-mini__panel-5 subscription-progress__popover"
          @click.stop
        >
        <div class="components-common-subscription-progress-mini__panel-6">
          <h3 class="components-common-subscription-progress-mini__heading">
            {{ t('subscriptionProgress.title') }}
          </h3>
          <p class="components-common-subscription-progress-mini__description">
            {{ t('subscriptionProgress.activeCount', { count: activeSubscriptions.length }) }}
          </p>
        </div>

        <div class="components-common-subscription-progress-mini__panel-7">
          <div
            v-for="subscription in displaySubscriptions"
            :key="subscription.id"
            class="components-common-subscription-progress-mini__panel-8"
          >
            <div class="components-common-subscription-progress-mini__panel-9">
              <span class="components-common-subscription-progress-mini__text-2">
                {{ subscription.group?.name || `Group #${subscription.group_id}` }}
              </span>
              <span
                v-if="subscription.expires_at"
                class="components-common-subscription-progress-mini__text-3"
                :class="getDaysRemainingClass(subscription.expires_at)"
              >
                {{ formatDaysRemaining(subscription.expires_at) }}
              </span>
            </div>

            <!-- Progress bars or Unlimited badge -->
            <div class="components-common-subscription-progress-mini__panel-10">
              <!-- Unlimited subscription badge -->
              <div
                v-if="isUnlimited(subscription)"
                class="components-common-subscription-progress-mini__panel-11"
              >
                <span class="components-common-subscription-progress-mini__text-4">∞</span>
                <span class="components-common-subscription-progress-mini__text-5">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <!-- Progress bars for limited subscriptions -->
              <template v-else>
                <div v-if="subscriptionLimitPoints(subscription, 'daily')" class="components-common-subscription-progress-mini__panel-12">
                  <span class="components-common-subscription-progress-mini__text-6">{{
                    t('subscriptionProgress.daily')
                  }}</span>
                  <div class="components-common-subscription-progress-mini__panel-13">
                    <div
                      class="components-common-subscription-progress-mini__panel-14"
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
                  <span class="components-common-subscription-progress-mini__text-7">
                    {{
                      formatUsage(subscriptionUsagePoints(subscription, 'daily'), subscriptionLimitPoints(subscription, 'daily'))
                    }}
                  </span>
                </div>

                <div v-if="subscriptionLimitPoints(subscription, 'weekly')" class="components-common-subscription-progress-mini__panel-12">
                  <span class="components-common-subscription-progress-mini__text-6">{{
                    t('subscriptionProgress.weekly')
                  }}</span>
                  <div class="components-common-subscription-progress-mini__panel-13">
                    <div
                      class="components-common-subscription-progress-mini__panel-14"
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
                  <span class="components-common-subscription-progress-mini__text-7">
                    {{
                      formatUsage(subscriptionUsagePoints(subscription, 'weekly'), subscriptionLimitPoints(subscription, 'weekly'))
                    }}
                  </span>
                </div>

                <div v-if="subscriptionLimitPoints(subscription, 'monthly')" class="components-common-subscription-progress-mini__panel-12">
                  <span class="components-common-subscription-progress-mini__text-6">{{
                    t('subscriptionProgress.monthly')
                  }}</span>
                  <div class="components-common-subscription-progress-mini__panel-13">
                    <div
                      class="components-common-subscription-progress-mini__panel-14"
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
                  <span class="components-common-subscription-progress-mini__text-7">
                    {{
                      formatUsage(
                        subscriptionUsagePoints(subscription, 'monthly'),
                        subscriptionLimitPoints(subscription, 'monthly')
                      )
                    }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="components-common-subscription-progress-mini__panel-15">
          <router-link
            to="/subscriptions"
            @click="closeTooltip"
            class="components-common-subscription-progress-mini__router-link"
          >
            {{ t('subscriptionProgress.viewAll') }}
          </router-link>
        </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useSubscriptionStore } from '@/stores'
import type { UserSubscription } from '@/types'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { formatPointAmount, formatPoints } from '@/utils/format'

const { t } = useI18n()

const subscriptionStore = useSubscriptionStore()

const containerRef = ref<HTMLElement | null>(null)
const tooltipOpen = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, tooltipOpen, {
  maxWidth: 340,
  minComfortableHeight: 240
})

// Use store data instead of local state
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const hasActiveSubscriptions = computed(() => subscriptionStore.hasActiveSubscriptions)

const displaySubscriptions = computed(() => {
  // Sort by most usage (highest percentage first)
  return [...activeSubscriptions.value].sort((a, b) => {
    const aMax = getMaxUsagePercentage(a)
    const bMax = getMaxUsagePercentage(b)
    return bMax - aMax
  })
})

type QuotaWindow = 'daily' | 'weekly' | 'monthly'

function subscriptionUsagePoints(sub: UserSubscription, window: QuotaWindow): number {
  return sub[`${window}_usage_points`] ?? sub[`${window}_usage_usd`] ?? 0
}

function subscriptionLimitPoints(sub: UserSubscription, window: QuotaWindow): number | null {
  return sub.group?.[`${window}_limit_points`] ?? sub.group?.[`${window}_limit_usd`] ?? null
}

function getMaxUsagePercentage(sub: UserSubscription): number {
  const percentages: number[] = []
  for (const window of ['daily', 'weekly', 'monthly'] as const) {
    const limit = subscriptionLimitPoints(sub, window)
    if (limit) {
      percentages.push((subscriptionUsagePoints(sub, window) / limit) * 100)
    }
  }
  return percentages.length > 0 ? Math.max(...percentages) : 0
}

function isUnlimited(sub: UserSubscription): boolean {
  return (
    !subscriptionLimitPoints(sub, 'daily') &&
    !subscriptionLimitPoints(sub, 'weekly') &&
    !subscriptionLimitPoints(sub, 'monthly')
  )
}

function getProgressDotClass(sub: UserSubscription): string {
  // Unlimited subscriptions get a special color
  if (isUnlimited(sub)) {
    return 'status-fill--success'
  }
  const maxPercentage = getMaxUsagePercentage(sub)
  if (maxPercentage >= 90) return 'status-fill--danger'
  if (maxPercentage >= 70) return 'status-fill--warning'
  return 'status-fill--success'
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'status-fill--neutral'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'status-fill--danger'
  if (percentage >= 70) return 'status-fill--warning'
  return 'status-fill--success'
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function formatUsage(used: number | undefined, limit: number | null | undefined): string {
  if (limit == null) return `${formatPointAmount(used)} / ∞`
  return `${formatPointAmount(used)} / ${formatPoints(limit)}`
}

function formatDaysRemaining(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  if (diff < 0) return t('subscriptionProgress.expired')
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return t('subscriptionProgress.expiresToday')
  if (days === 1) return t('subscriptionProgress.expiresTomorrow')
  return t('subscriptionProgress.daysRemaining', { days })
}

function getDaysRemainingClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'components-common-subscription-progress-mini__state'
  if (days <= 7) return 'components-common-subscription-progress-mini__state-2'
  return 'components-common-subscription-progress-mini__state-3'
}

function toggleTooltip() {
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip() {
  tooltipOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  if (containerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  closeTooltip()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  // Trigger initial fetch if not already loaded
  // The actual data loading is handled by App.vue globally
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions in SubscriptionProgressMini:', error)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.subscription-progress__popover {
  border-color: var(--glass-border-hover);
  background-color: var(--glass-layer-floating-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow:
    var(--glass-shadow-hover),
    0 1px 0 var(--glass-highlight) inset;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
