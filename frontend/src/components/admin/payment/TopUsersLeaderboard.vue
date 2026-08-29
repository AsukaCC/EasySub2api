<template>
  <div class="components-admin-payment-top-users-leaderboard__panel card">
    <h3 class="components-admin-payment-top-users-leaderboard__heading">
      {{ t('payment.admin.topUsers') }}
    </h3>
    <div
      v-if="!hasUsers(props.users)"
      class="components-admin-payment-top-users-leaderboard__panel-2"
    >
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else class="components-admin-payment-top-users-leaderboard__panel-3">
      <div v-for="[currency, currencyUsers] in sortedUsers(props.users)" :key="currency" class="components-admin-payment-top-users-leaderboard__panel-3">
        <p class="components-admin-payment-top-users-leaderboard__description">{{ currency }}</p>
        <div
          v-for="(user, idx) in currencyUsers"
          :key="user.user_id"
          class="components-admin-payment-top-users-leaderboard__panel-4"
        >
          <div class="components-admin-payment-top-users-leaderboard__panel-5">
            <span
              :class="[
                'components-admin-payment-top-users-leaderboard__text-3',
                rankClass(idx),
              ]"
            >
              {{ idx + 1 }}
            </span>
            <span class="components-admin-payment-top-users-leaderboard__text">{{ user.email }}</span>
          </div>
          <span class="components-admin-payment-top-users-leaderboard__text-2">
            {{ formatMoney(currency, user.amount) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { TopUserPaymentStats } from '@/types/payment'

const { t } = useI18n()

const props = defineProps<{
  users: Record<string, TopUserPaymentStats[]>
}>()

function rankClass(idx: number): string {
  if (idx === 0) return 'components-admin-payment-top-users-leaderboard__state'
  if (idx === 1) return 'components-admin-payment-top-users-leaderboard__state-2'
  if (idx === 2) return 'components-admin-payment-top-users-leaderboard__state-3'
  return 'components-admin-payment-top-users-leaderboard__state-4'
}

function hasUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): boolean {
  return Object.values(usersByCurrency).some(users => users.length > 0)
}

function sortedUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): [string, TopUserPaymentStats[]][] {
  return Object.entries(usersByCurrency).sort(([left], [right]) => left.localeCompare(right))
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}
</script>
