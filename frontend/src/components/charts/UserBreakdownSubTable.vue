<template>
  <div class="components-charts-user-breakdown-sub-table__panel">
    <div v-if="loading" class="components-charts-user-breakdown-sub-table__panel-2">
      <LoadingSpinner />
    </div>
    <div v-else-if="items.length === 0" class="components-charts-user-breakdown-sub-table__panel-3">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
    <table v-else class="components-charts-user-breakdown-sub-table__table">
      <tbody>
        <tr
          v-for="user in items"
          :key="user.user_id"
          class="components-charts-user-breakdown-sub-table__row"
        >
          <td class="components-charts-user-breakdown-sub-table__cell" :title="user.email">
            {{ user.email || `User #${user.user_id}` }}
          </td>
          <td class="components-charts-user-breakdown-sub-table__cell-2">
            {{ user.requests.toLocaleString() }}
          </td>
          <td class="components-charts-user-breakdown-sub-table__cell-2">
            {{ formatTokens(user.total_tokens) }}
          </td>
          <td class="components-charts-user-breakdown-sub-table__cell-3">
            {{ formatPoints(user.actual_cost) }}
          </td>
          <td v-if="showAccountCost" class="components-charts-user-breakdown-sub-table__cell-4">
            ${{ formatCost(user.account_cost) }}
          </td>
          <td class="components-charts-user-breakdown-sub-table__cell-5">
            ${{ formatCost(user.cost) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { UserBreakdownItem } from '@/types'
import { formatPoints } from '@/utils/format'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  items: UserBreakdownItem[]
  loading?: boolean
  showAccountCost?: boolean
}>(), {
  loading: false,
  showAccountCost: true,
})

const showAccountCost = computed(() => props.showAccountCost)

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

const formatCost = (value: number | undefined | null): string => {
  if (value == null) return '0.0000'
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>
