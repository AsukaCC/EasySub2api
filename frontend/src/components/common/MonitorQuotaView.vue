<template>
  <div v-if="snapshot" class="components-common-monitor-quota-view__panel" data-testid="monitor-quota-view">
    <!-- 套餐等级徽章（如智谱 plan level / Claude 订阅档） -->
    <div v-if="snapshot.plan_level" class="components-common-monitor-quota-view__panel-2">
      <span class="components-common-monitor-quota-view__text">
        {{ snapshot.plan_level }}
      </span>
    </div>

    <!-- 用量窗口条形图（样式/阈值对齐账号页 CNProviderQuotaCell） -->
    <div v-if="snapshot.success && tierRows.length" class="components-common-monitor-quota-view__panel">
      <div v-for="row in tierRows" :key="row.key" class="components-common-monitor-quota-view__panel-3">
        <span class="components-common-monitor-quota-view__text-2" :title="row.title">
          {{ row.label }}
        </span>
        <div class="components-common-monitor-quota-view__panel-4">
          <div
            class="components-common-monitor-quota-view__panel-5"
            :class="utilizationColor(row.tier.used_percent)"
            :style="{ width: `${Math.min(100, Math.max(0, row.tier.used_percent))}%` }"
          />
        </div>
        <span :class="['components-common-monitor-quota-view__text-4', utilizationTextColor(row.tier.used_percent)]">
          {{ Math.round(row.tier.used_percent) }}%
        </span>
        <span v-if="row.tier.reset_at" class="components-common-monitor-quota-view__text-3" :title="row.tier.reset_at">
          · {{ formatReset(row.tier.reset_at) }}
        </span>
      </div>
    </div>

    <!-- 余额（国产 payg；支持多币种） -->
    <div v-if="snapshot.success && balanceRows.length" class="components-common-monitor-quota-view__panel-6">
      <span
        v-for="b in balanceRows"
        :key="b.currency"
        :class="['components-common-monitor-quota-view__text-5', b.balance <= 0 ? 'components-common-monitor-quota-view__text-6' : 'components-common-monitor-quota-view__text-7']"
      >
        {{ b.balance.toFixed(2) }} {{ b.currency }}
      </span>
    </div>

    <div v-if="!snapshot.success" class="components-common-monitor-quota-view__panel-7" :title="snapshot.error" data-testid="monitor-quota-error">
      {{ truncatedError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorQuotaSnapshot, MonitorQuotaTier } from '@/api/admin/channelMonitor'

/**
 * 配额快照渲染（管理端监控列表/运行结果 + 用户端监控卡片共用）。
 * 展示形态对齐账号管理侧的用量视图（CNProviderQuotaCell：同阈值配色、
 * 同倒计时格式）；tier 的 Window/Label 是后端约定的机器 token，
 * 已知 token 走 i18n，未知 token 原样展示（前向兼容）。
 */
const props = defineProps<{
  snapshot?: MonitorQuotaSnapshot | null
}>()

const { t, te } = useI18n()

interface QuotaTierRow {
  key: string
  label: string
  title: string
  tier: MonitorQuotaTier
}

// 已知的 window/label 机器 token → i18n key（monitorCommon.quota.*）。
const windowI18nKeys: Record<string, string> = {
  '5h': 'monitorCommon.quota.windows.5h',
  '7d': 'monitorCommon.quota.windows.7d',
  '7d-sonnet': 'monitorCommon.quota.windows.7dSonnet',
  '7d-fable': 'monitorCommon.quota.windows.7dFable',
  weekly: 'monitorCommon.quota.windows.weekly',
  daily: 'monitorCommon.quota.windows.daily',
  '30d': 'monitorCommon.quota.windows.30d',
  total: 'monitorCommon.quota.windows.total',
}

const labelI18nKeys: Record<string, string> = {
  requests: 'monitorCommon.quota.labels.requests',
  tokens: 'monitorCommon.quota.labels.tokens',
  shared: 'monitorCommon.quota.labels.shared',
  pro: 'monitorCommon.quota.labels.pro',
  flash: 'monitorCommon.quota.labels.flash',
}

function windowLabel(window: string): string {
  const key = windowI18nKeys[window]
  return key && te(key) ? t(key) : window
}

function tierLabel(tier: MonitorQuotaTier): string {
  const window = windowLabel(tier.window)
  if (!tier.label) return window
  const labelKey = labelI18nKeys[tier.label]
  const label = labelKey && te(labelKey) ? t(labelKey) : tier.label
  return `${label}/${window}`
}

const tierRows = computed<QuotaTierRow[]>(() =>
  (props.snapshot?.tiers || []).map((tier, idx) => ({
    key: `${tier.window}-${tier.label || ''}-${idx}`,
    label: tierLabel(tier),
    title: tierLabel(tier),
    tier,
  })),
)

const balanceRows = computed(() => {
  const snapshot = props.snapshot
  if (!snapshot) return []
  if (snapshot.balances?.length) return snapshot.balances
  if (snapshot.balance != null) {
    return [{ currency: snapshot.currency || '?', balance: snapshot.balance }]
  }
  return []
})

const truncatedError = computed(() => {
  const error = props.snapshot?.error || t('monitorCommon.quota.unavailable')
  return error.length > 48 ? `${error.slice(0, 48)}…` : error
})

const utilizationColor = (pct: number) => {
  if (pct >= 90) return 'status-fill--danger'
  if (pct >= 75) return 'status-fill--warning'
  return 'status-fill--success'
}

const utilizationTextColor = (pct: number) => {
  if (pct >= 90) return 'components-common-monitor-quota-view__text-6'
  if (pct >= 75) return 'components-common-monitor-quota-view__state'
  return 'components-common-monitor-quota-view__state-2'
}

// 重置时间相对/绝对简短显示（与账号页一致）。
const formatReset = (iso: string) => {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const now = Date.now()
  const diffMs = d.getTime() - now
  if (diffMs <= 0) return t('monitorCommon.quota.resetSoon')
  if (diffMs < 3_600_000) return `${Math.max(1, Math.round(diffMs / 60_000))}m`
  const hours = Math.round(diffMs / 3_600_000)
  if (hours < 48) return `${hours}h`
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${mm}-${dd}`
}
</script>
