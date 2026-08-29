<template>
  <div class="components-user-platform-usage-breakdown__panel">
    <div class="components-user-platform-usage-breakdown__panel-2">
      <span class="components-user-platform-usage-breakdown__text">{{ t('admin.users.today') }}:</span>
      <span class="components-user-platform-usage-breakdown__text-2">{{ formatPoints(today) }}</span>
      <Icon
        v-if="hasBreakdown"
        name="infoCircle"
        size="xs"
        class="components-user-platform-usage-breakdown__icon"
      />
    </div>
    <div class="components-user-platform-usage-breakdown__panel-3">
      <span class="components-user-platform-usage-breakdown__text">{{ t('admin.users.total') }}:</span>
      <span class="components-user-platform-usage-breakdown__text-2">{{ formatPoints(total) }}</span>
    </div>

    <div
      v-if="hasBreakdown"
      class="components-user-platform-usage-breakdown__panel-4"
    >
      <div class="components-user-platform-usage-breakdown__panel-5">
        <span>{{ t('admin.users.platformBreakdown') }}</span>
        <span class="components-user-platform-usage-breakdown__text-3">{{ t('admin.users.today') }} / {{ t('admin.users.total') }}</span>
      </div>
      <div
        v-for="item in sortedBreakdown"
        :key="item.platform"
        class="components-user-platform-usage-breakdown__panel-6"
        :class="{ 'components-user-platform-usage-breakdown__panel-7': item.isOther }"
      >
        <span class="components-user-platform-usage-breakdown__text-4">
          {{ item.isOther ? t('admin.users.platformOther') : platformLabel(item.platform) }}
        </span>
        <span class="components-user-platform-usage-breakdown__text-3">
          {{ formatPointAmount(item.today_actual_cost) }}
          <span class="components-user-platform-usage-breakdown__text-5">/</span>
          {{ formatPoints(item.total_actual_cost) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { PlatformUsage } from '@/api/admin/dashboard'
import { formatPointAmount, formatPoints } from '@/utils/format'

const props = defineProps<{
  today: number
  total: number
  byPlatform?: PlatformUsage[]
}>()

const { t } = useI18n()

// 与 UserDashboardStats 保持一致：把"总值 - 各平台之和"的差作为"其他"行展示，
// 避免 tooltip 内各平台费用加总与列首总值对不上。
const OTHER_THRESHOLD = 0.0001

interface BreakdownRow {
  platform: string
  today_actual_cost: number
  total_actual_cost: number
  isOther?: boolean
}

const sortedBreakdown = computed<BreakdownRow[]>(() => {
  const list = props.byPlatform ?? []
  const rows: BreakdownRow[] = [...list]
    .sort((a, b) => b.total_actual_cost - a.total_actual_cost)
    .map((p) => ({ ...p }))

  const sumTotal = rows.reduce((s, r) => s + r.total_actual_cost, 0)
  const sumToday = rows.reduce((s, r) => s + r.today_actual_cost, 0)
  const diffTotal = Math.max(0, props.total - sumTotal)
  const diffToday = Math.max(0, props.today - sumToday)
  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    rows.push({
      platform: '__other__',
      today_actual_cost: diffToday,
      total_actual_cost: diffTotal,
      isOther: true
    })
  }
  return rows
})

const hasBreakdown = computed(() => sortedBreakdown.value.length > 0)

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI'
}

function platformLabel(platform: string): string {
  return PLATFORM_LABELS[platform] ?? platform
}
</script>
