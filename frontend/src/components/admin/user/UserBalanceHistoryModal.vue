<template>
  <BaseDialog :show="show" :title="t('admin.users.balanceHistoryTitle')" width="wide" :close-on-click-outside="true" :z-index="40" @close="$emit('close')">
    <div v-if="user" class="components-admin-user-user-balance-history-modal__panel">
      <!-- User header: two-row layout with full user info -->
      <div class="components-admin-user-user-balance-history-modal__panel-2">
        <!-- Row 1: avatar + email/username/created_at (left) + current balance (right) -->
        <div class="components-admin-user-user-balance-history-modal__panel-3">
          <div class="components-admin-user-user-balance-history-modal__panel-4">
            <span class="components-admin-user-user-balance-history-modal__text">
              {{ user.email.charAt(0).toUpperCase() }}
            </span>
          </div>
          <div class="components-admin-user-user-balance-history-modal__panel-5">
            <div class="components-admin-user-user-balance-history-modal__panel-6">
              <p class="components-admin-user-user-balance-history-modal__description">{{ user.email }}</p>
              <span v-if="user.deleted_at" class="components-admin-user-user-balance-history-modal__text-2">
                {{ t('admin.usage.userDeletedBadge') }}
              </span>
              <span
                v-if="user.username"
                class="components-admin-user-user-balance-history-modal__text-3"
              >
                {{ user.username }}
              </span>
            </div>
            <p class="components-admin-user-user-balance-history-modal__description-2">
              {{ t('admin.users.createdAt') }}: {{ formatDateTime(user.created_at) }}
            </p>
          </div>
          <!-- Current balance: prominent display on the right -->
          <div class="components-admin-user-user-balance-history-modal__panel-7">
            <p class="components-admin-user-user-balance-history-modal__description-3">{{ t('admin.users.currentBalance') }}</p>
            <p class="components-admin-user-user-balance-history-modal__description-4">
              {{ formatPoints(user.balance) }}
            </p>
          </div>
        </div>
        <!-- Row 2: notes + total recharged -->
        <div class="components-admin-user-user-balance-history-modal__panel-8">
          <p class="components-admin-user-user-balance-history-modal__description-5" :title="user.notes || ''">
            <template v-if="user.notes">{{ t('admin.users.notes') }}: {{ user.notes }}</template>
            <template v-else>&nbsp;</template>
          </p>
          <p class="components-admin-user-user-balance-history-modal__description-6">
            {{ t('admin.users.totalRecharged') }}: <span class="components-admin-user-user-balance-history-modal__text-4">{{ formatPoints(totalRecharged) }}</span>
          </p>
        </div>
      </div>

      <!-- Type filter + Action buttons -->
      <div class="components-admin-user-user-balance-history-modal__panel-3">
        <!-- Deposit button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('deposit')"
          class="components-admin-user-user-balance-history-modal__action"
        >
          <Icon name="plus" size="sm" class="components-admin-user-user-balance-history-modal__icon" :stroke-width="2" />
          {{ t('admin.users.deposit') }}
        </button>
        <!-- Withdraw button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('withdraw')"
          class="components-admin-user-user-balance-history-modal__action"
        >
          <svg class="components-admin-user-user-balance-history-modal__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
          </svg>
          {{ t('admin.users.withdraw') }}
        </button>
      </div>

      <!-- Loading -->
      <LoadingState v-if="loading" variant="section" size="sm" class="components-admin-user-user-balance-history-modal__panel-9" />

      <!-- Empty state -->
      <div v-else-if="history.length === 0" class="components-admin-user-user-balance-history-modal__panel-10">
        <p class="components-admin-user-user-balance-history-modal__description-7">{{ t('admin.users.noBalanceHistory') }}</p>
      </div>

      <!-- History list -->
      <div v-else class="components-admin-user-user-balance-history-modal__panel-11">
        <div
          v-for="item in history"
          :key="item.id"
          class="components-admin-user-user-balance-history-modal__panel-12"
        >
          <div class="components-admin-user-user-balance-history-modal__panel-13">
            <!-- Left: type icon + description -->
            <div class="components-admin-user-user-balance-history-modal__panel-14">
              <div
                :class="[
                  'components-admin-user-user-balance-history-modal__panel-17',
                  getIconBg(item)
                ]"
              >
                <Icon :name="getIconName(item)" size="sm" :class="getIconColor(item)" />
              </div>
              <div>
                <p class="components-admin-user-user-balance-history-modal__description-8">
                  {{ getItemTitle(item) }}
                </p>
                <!-- Notes (admin adjustment reason) -->
                <p
                  v-if="item.notes"
                  class="components-admin-user-user-balance-history-modal__description-9"
                  :title="item.notes"
                >
                  {{ item.notes.length > 60 ? item.notes.substring(0, 55) + '...' : item.notes }}
                </p>
                <p class="components-admin-user-user-balance-history-modal__description-10">
                  {{ formatDateTime(item.used_at || item.created_at) }}
                </p>
              </div>
            </div>
            <!-- Right: value -->
            <div class="components-admin-user-user-balance-history-modal__panel-15">
              <p :class="['components-admin-user-user-balance-history-modal__description-12', getValueColor(item)]">
                {{ formatValue(item) }}
              </p>
              <p
                v-if="isAdminType(item.type)"
                class="components-admin-user-user-balance-history-modal__description-2"
              >
                {{ t('redeem.adminAdjustment') }}
              </p>
              <p
                v-else
                class="components-admin-user-user-balance-history-modal__description-11"
              >
                {{ item.code.slice(0, 8) }}...
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="components-admin-user-user-balance-history-modal__panel-16">
        <button
          :disabled="currentPage <= 1"
          class="components-admin-user-user-balance-history-modal__element btn btn-secondary"
          @click="loadHistory(currentPage - 1)"
        >
          {{ t('pagination.previous') }}
        </button>
        <span class="components-admin-user-user-balance-history-modal__text-5">
          {{ currentPage }} / {{ totalPages }}
        </span>
        <button
          :disabled="currentPage >= totalPages"
          class="components-admin-user-user-balance-history-modal__element btn btn-secondary"
          @click="loadHistory(currentPage + 1)"
        >
          {{ t('pagination.next') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'

import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type BalanceHistoryItem } from '@/api/admin'
import { formatDateTime, formatPoints } from '@/utils/format'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null; hideActions?: boolean }>()
const emit = defineEmits(['close', 'deposit', 'withdraw'])
const { t } = useI18n()

const history = ref<BalanceHistoryItem[]>([])
const loading = ref(false)
const currentPage = ref(1)
const total = ref(0)
const totalRecharged = ref(0)
const pageSize = 15

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

// Watch modal open
watch(() => props.show, (v) => {
  if (v && props.user) {
    loadHistory(1)
  }
})

const loadHistory = async (page: number) => {
  if (!props.user) return
  loading.value = true
  currentPage.value = page
  try {
    const res = await adminAPI.users.getUserBalanceHistory(
      props.user.id,
      page,
	  pageSize,
    )
    history.value = res.items || []
    total.value = res.total || 0
    totalRecharged.value = res.total_recharged || 0
  } catch (error) {
    console.error('Failed to load balance history:', error)
  } finally {
    loading.value = false
  }
}

// Helper: check if admin type
const isAdminType = (type: string) => type === 'admin_balance' || type === 'admin_concurrency'

// Helper: check if balance type (includes admin_balance)
const isBalanceType = (type: string) => !isSubscriptionType(type) && type !== 'concurrency' && type !== 'admin_concurrency'

// Helper: check if subscription type
const isSubscriptionType = (type: string) => type === 'subscription'

// Icon name based on type
const getIconName = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) return 'sparkles'
  if (isSubscriptionType(item.type)) return 'badge'
  return 'bolt' // concurrency
}

// Icon background color
const getIconBg = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'components-admin-user-user-balance-history-modal__state'
      : 'components-admin-user-user-balance-history-modal__state-2'
  }
  if (isSubscriptionType(item.type)) return 'components-admin-user-user-balance-history-modal__state-3'
  return item.value >= 0
    ? 'components-admin-user-user-balance-history-modal__state-4'
    : 'components-admin-user-user-balance-history-modal__state-5'
}

// Icon text color
const getIconColor = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'components-admin-user-user-balance-history-modal__state-6'
      : 'components-admin-user-user-balance-history-modal__state-7'
  }
  if (isSubscriptionType(item.type)) return 'components-admin-user-user-balance-history-modal__state-8'
  return item.value >= 0
    ? 'components-admin-user-user-balance-history-modal__state-9'
    : 'components-admin-user-user-balance-history-modal__state-10'
}

// Value text color
const getValueColor = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'components-admin-user-user-balance-history-modal__state-6'
      : 'components-admin-user-user-balance-history-modal__state-7'
  }
  if (isSubscriptionType(item.type)) return 'components-admin-user-user-balance-history-modal__state-8'
  return item.value >= 0
    ? 'components-admin-user-user-balance-history-modal__state-9'
    : 'components-admin-user-user-balance-history-modal__state-10'
}

// Item title
const getItemTitle = (item: BalanceHistoryItem) => {
  switch (item.type) {
    case 'balance':
      return t('redeem.balanceAddedRedeem')
    case 'affiliate_balance':
      return t('redeem.balanceAddedAffiliate')
    case 'admin_balance':
      return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
    case 'concurrency':
      return t('redeem.concurrencyAddedRedeem')
    case 'admin_concurrency':
      return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
    case 'subscription':
      return t('redeem.subscriptionAssigned')
    default:
	  return t(`admin.users.walletActions.${item.type}`)
  }
}

// Format display value
const formatValue = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${formatPoints(item.value)}`
  }
  if (isSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}d - ${groupName}` : `${days}d`
  }
  // concurrency types
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value}`
}
</script>
