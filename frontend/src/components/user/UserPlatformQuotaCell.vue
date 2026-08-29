<template>
  <span v-if="props.quotas === undefined" class="components-user-user-platform-quota-cell__text">…</span>
  <span v-else-if="configured.length === 0" class="components-user-user-platform-quota-cell__text">
    {{ t('admin.users.platformQuota.cellNotConfigured') }}
  </span>
  <div v-else class="components-user-user-platform-quota-cell__panel">
    <div
      v-for="row in configured"
      :key="row.platform"
      class="components-user-user-platform-quota-cell__panel-2"
    >
      <span class="components-user-user-platform-quota-cell__text-2">{{ row.platform }}</span>
      <span class="components-user-user-platform-quota-cell__text-3">
        {{ t('admin.users.platformQuota.windowDaily') }}
        <span class="components-user-user-platform-quota-cell__text-4">{{ fmtRange(usagePoints(row, 'daily'), limitPoints(row, 'daily')) }}</span>
      </span>
      <span class="components-user-user-platform-quota-cell__text-3">
        {{ t('admin.users.platformQuota.windowWeekly') }}
        <span class="components-user-user-platform-quota-cell__text-4">{{ fmtRange(usagePoints(row, 'weekly'), limitPoints(row, 'weekly')) }}</span>
      </span>
      <span class="components-user-user-platform-quota-cell__text-3">
        {{ t('admin.users.platformQuota.windowMonthly') }}
        <span class="components-user-user-platform-quota-cell__text-4">{{ fmtRange(usagePoints(row, 'monthly'), limitPoints(row, 'monthly')) }}</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PlatformQuotaItem, PlatformQuotaPlatform } from '@/api/admin/users'
import { formatPointAmount, formatPoints } from '@/utils/format'

const props = defineProps<{ quotas?: PlatformQuotaItem[] }>()
const { t } = useI18n()

const PLATFORM_ORDER: PlatformQuotaPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']

// 仅展示「至少一档限额非空」的平台（配额列，非用量列）
const configured = computed(() => {
  if (!props.quotas) return []
  return props.quotas
    .filter(
      (q) =>
        limitPoints(q, 'daily') != null ||
        limitPoints(q, 'weekly') != null ||
        limitPoints(q, 'monthly') != null
    )
    .slice()
    .sort((a, b) => PLATFORM_ORDER.indexOf(a.platform) - PLATFORM_ORDER.indexOf(b.platform))
})

type QuotaWindow = 'daily' | 'weekly' | 'monthly'

function usagePoints(row: PlatformQuotaItem, window: QuotaWindow): number {
  return row[`${window}_usage_points`] ?? row[`${window}_usage_usd`] ?? 0
}

function limitPoints(row: PlatformQuotaItem, window: QuotaWindow): number | null {
  return row[`${window}_limit_points`] ?? row[`${window}_limit_usd`] ?? null
}

function fmtRange(used: number, limit: number | null): string {
  return limit == null ? `${formatPointAmount(used)} / —` : `${formatPointAmount(used)} / ${formatPoints(limit)}`
}
</script>
