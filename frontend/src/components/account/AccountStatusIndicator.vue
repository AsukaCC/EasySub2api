<template>
  <div class="components-account-account-status-indicator__panel">
    <!-- Rate Limit Display (429) - Two-line layout -->
    <div v-if="isRateLimited" class="components-account-account-status-indicator__panel-2">
      <span class="components-account-account-status-indicator__text badge badge-warning">{{ t('admin.accounts.status.rateLimited') }}</span>
      <span class="components-account-account-status-indicator__text-2">{{ rateLimitResumeText }}</span>
    </div>

    <!-- Overload Display (529) - Two-line layout -->
    <div v-else-if="isOverloaded" class="components-account-account-status-indicator__panel-2">
      <span class="components-account-account-status-indicator__text badge badge-danger">{{ t('admin.accounts.status.overloaded') }}</span>
      <span class="components-account-account-status-indicator__text-2">{{ overloadCountdown }}</span>
    </div>

    <!-- Main Status Badge (shown when not rate limited/overloaded) -->
    <template v-else>
      <div v-if="isTempUnschedulable" class="components-account-account-status-indicator__panel-2">
        <button
          type="button"
          :class="['components-account-account-status-indicator__text badge', statusClass, 'components-account-account-status-indicator__action']"
          :title="t('admin.accounts.status.viewTempUnschedDetails')"
          @click="handleTempUnschedClick"
        >
          {{ statusText }}
        </button>
        <span class="components-account-account-status-indicator__text-3">
          {{ tempUnschedRecoveryText }}
        </span>
      </div>
      <span v-else :class="['components-account-account-status-indicator__text badge', statusClass]">
        {{ statusText }}
      </span>
    </template>

    <!-- Error Info Indicator -->
    <div v-if="hasError && account.error_message" class="components-account-account-status-indicator__panel-3">
      <HelpTooltip width-class="w-72">
        <template #trigger>
          <svg
            class="components-account-account-status-indicator__icon"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9.879 7.519c1.171-1.025 3.071-1.025 4.242 0 1.172 1.025 1.172 2.687 0 3.712-.203.179-.43.326-.67.442-.745.361-1.45.999-1.45 1.827v.75M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9 5.25h.008v.008H12v-.008z"
            />
          </svg>
        </template>
        <div class="components-account-account-status-indicator__panel-5">
          {{ account.error_message }}
        </div>
      </HelpTooltip>
    </div>

    <!-- Rate Limit Indicator (429) -->
    <div v-if="isRateLimited" class="components-account-account-status-indicator__panel-7">
      <span
        class="components-account-account-status-indicator__text-4"
      >
        <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
        429
      </span>
      <!-- Tooltip -->
      <div
        class="components-account-account-status-indicator__panel-8"
      >
        {{ t('admin.accounts.status.rateLimitedUntil', { time: formatDateTime(account.rate_limit_reset_at) }) }}
        <div
          class="components-account-account-status-indicator__panel-9"
        ></div>
      </div>
    </div>

    <!-- Model Status Indicators (普通限流 / 超量请求中) -->
    <div
      v-if="activeModelStatuses.length > 0"
      :class="[
        activeModelStatuses.length <= 4
          ? 'components-account-account-status-indicator__panel-12'
          : activeModelStatuses.length <= 8
            ? 'components-account-account-status-indicator__panel-13'
            : 'components-account-account-status-indicator__panel-14'
      ]"
    >
      <div v-for="item in activeModelStatuses" :key="`${item.kind}-${item.model}`" class="components-account-account-status-indicator__panel-10">
        <!-- 积分已用尽 -->
        <span
          v-if="item.kind === 'credits_exhausted'"
          class="components-account-account-status-indicator__text-5"
        >
          <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
          {{ t('admin.accounts.status.creditsExhausted') }}
          <span class="components-account-account-status-indicator__text-6">{{ formatCountdown(item.reset_at) }}</span>
        </span>
        <!-- 正在走积分（模型限流但积分可用）-->
        <span
          v-else-if="item.kind === 'credits_active'"
          class="components-account-account-status-indicator__text-4"
        >
          <span>⚡</span>
          {{ formatScopeName(item.model) }}
          <span class="components-account-account-status-indicator__text-6">{{ formatCountdown(item.reset_at) }}</span>
        </span>
        <!-- 普通模型限流 -->
        <span
          v-else
          class="components-account-account-status-indicator__text-7"
        >
          <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
          {{ formatScopeName(item.model) }}
          <span class="components-account-account-status-indicator__text-6">{{ formatCountdown(item.reset_at) }}</span>
        </span>
        <!-- Tooltip -->
        <div
          class="components-account-account-status-indicator__panel-11"
        >
          {{
            item.kind === 'credits_exhausted'
              ? t('admin.accounts.status.creditsExhaustedUntil', { time: formatDateTimeToMinute(item.reset_at) })
              : item.kind === 'credits_active'
                ? t('admin.accounts.status.modelCreditOveragesUntil', { model: formatScopeName(item.model), time: formatDateTimeToMinute(item.reset_at) })
                : t('admin.accounts.status.modelRateLimitedUntil', { model: formatScopeName(item.model), time: formatDateTimeToMinute(item.reset_at) })
          }}
          <div
            class="components-account-account-status-indicator__panel-9"
          ></div>
        </div>
      </div>
    </div>

    <!-- Overload Indicator (529) -->
    <div v-if="isOverloaded" class="components-account-account-status-indicator__panel-7">
      <span
        class="components-account-account-status-indicator__text-5"
      >
        <Icon name="exclamationTriangle" size="xs" :stroke-width="2" />
        529
      </span>
      <!-- Tooltip -->
      <div
        class="components-account-account-status-indicator__panel-8"
      >
        {{ t('admin.accounts.status.overloadedUntil', { time: formatTime(account.overload_until) }) }}
        <div
          class="components-account-account-status-indicator__panel-9"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import type { Account } from '@/types'
import { formatCountdown, formatDateTime, formatDateTimeToMinute, formatCountdownWithSuffix, formatTime } from '@/utils/format'

const { t } = useI18n()

const props = defineProps<{
  account: Account
}>()

const emit = defineEmits<{
  (e: 'show-temp-unsched', account: Account): void
}>()

// Computed: is rate limited (429)
const isRateLimited = computed(() => {
  if (!props.account.rate_limit_reset_at) return false
  return new Date(props.account.rate_limit_reset_at) > new Date()
})

type AccountModelStatusItem = {
  kind: 'rate_limit' | 'credits_exhausted' | 'credits_active'
  model: string
  reset_at: string
}

// Computed: active model statuses (普通模型限流 + 积分耗尽 + 走积分中)
const activeModelStatuses = computed<AccountModelStatusItem[]>(() => {
  const extra = props.account.extra as Record<string, unknown> | undefined
  const modelLimits = extra?.model_rate_limits as
    | Record<string, { rate_limited_at: string; rate_limit_reset_at: string }>
    | undefined
  const now = new Date()
  const items: AccountModelStatusItem[] = []

  if (!modelLimits) return items

  // 检查 AICredits key 是否生效（积分是否耗尽）
  const aiCreditsEntry = modelLimits['AICredits']
  const hasActiveAICredits = aiCreditsEntry && new Date(aiCreditsEntry.rate_limit_reset_at) > now
  const allowOverages = !!(extra?.allow_overages)

  for (const [model, info] of Object.entries(modelLimits)) {
    if (new Date(info.rate_limit_reset_at) <= now) continue

    if (model === 'AICredits') {
      // AICredits key → 积分已用尽
      items.push({ kind: 'credits_exhausted', model, reset_at: info.rate_limit_reset_at })
    } else if (allowOverages && !hasActiveAICredits) {
      // 普通模型限流 + overages 启用 + 积分可用 → 正在走积分
      items.push({ kind: 'credits_active', model, reset_at: info.rate_limit_reset_at })
    } else {
      // 普通模型限流
      items.push({ kind: 'rate_limit', model, reset_at: info.rate_limit_reset_at })
    }
  }

  return items
})

const formatScopeName = (scope: string): string => {
  const aliases: Record<string, string> = {
    // Claude 系列
    'claude-fable-5-1': 'CFable51',
    'claude-fable-5': 'CFable5',
    'claude-opus-4-6': 'COpus46',
    'claude-opus-4-6-thinking': 'COpus46T',
    'claude-opus-4-7': 'COpus47',
    'claude-opus-4-8': 'COpus48',
    'claude-opus-5': 'COpus5',
    'claude-sonnet-4-6': 'CSon46',
    'claude-sonnet-4-5': 'CSon45',
    'claude-sonnet-4-5-thinking': 'CSon45T',
    'claude-sonnet-5': 'CSon5',
    // 其他
    'gpt-oss-120b-medium': 'GPT120',
    'tab_flash_lite_preview': 'TabFL',
    // 旧版 scope 别名（兼容）
    claude: 'Claude',
    claude_sonnet: 'CSon',
    claude_opus: 'COpus',
    claude_haiku: 'CHaiku',
  }
  return aliases[scope] || scope
}

// Computed: is overloaded (529)
const isOverloaded = computed(() => {
  if (!props.account.overload_until) return false
  return new Date(props.account.overload_until) > new Date()
})

// Computed: is temp unschedulable
const isTempUnschedulable = computed(() => {
  if (!props.account.temp_unschedulable_until) return false
  return new Date(props.account.temp_unschedulable_until) > new Date()
})

// Computed: has error status
const hasError = computed(() => {
  return props.account.status === 'error'
})

const isQuotaExceeded = computed(() => {
  const exceeded = (used?: number | null, limit?: number | null) =>
    typeof limit === 'number' && limit > 0 && typeof used === 'number' && used >= limit
  return (
    exceeded(props.account.quota_used, props.account.quota_limit) ||
    exceeded(props.account.quota_daily_used, props.account.quota_daily_limit) ||
    exceeded(props.account.quota_weekly_used, props.account.quota_weekly_limit)
  )
})

// Computed: countdown text for rate limit (429)
const rateLimitCountdown = computed(() => {
  return formatCountdown(props.account.rate_limit_reset_at)
})

const rateLimitResumeText = computed(() => {
  if (!rateLimitCountdown.value) return ''
  return t('admin.accounts.status.rateLimitedAutoResume', { time: rateLimitCountdown.value })
})

// Computed: countdown text for overload (529)
const overloadCountdown = computed(() => {
  return formatCountdownWithSuffix(props.account.overload_until)
})

const tempUnschedRecoveryText = computed(() => {
  if (!isTempUnschedulable.value || !props.account.temp_unschedulable_until) return ''
  return t('admin.accounts.status.tempUnschedulableUntil', {
    time: formatDateTime(props.account.temp_unschedulable_until)
  })
})

// Computed: status badge class
const statusClass = computed(() => {
  if (hasError.value) {
    return 'badge-danger'
  }
  if (isTempUnschedulable.value) {
    return 'badge-warning'
  }
  if (props.account.status !== 'active') {
    return props.account.status === 'error' ? 'badge-danger' : 'badge-gray'
  }
  if (isQuotaExceeded.value) {
    return 'badge-warning'
  }
  if (!props.account.schedulable) {
    return 'badge-gray'
  }
  return 'badge-success'
})

// Computed: status text
const statusText = computed(() => {
  if (hasError.value) {
    return t('admin.accounts.status.error')
  }
  if (isTempUnschedulable.value) {
    return t('admin.accounts.status.tempUnschedulable')
  }
  if (props.account.status !== 'active') {
    return t(`admin.accounts.status.${props.account.status}`)
  }
  if (isQuotaExceeded.value) {
    return t('admin.accounts.status.quotaExceeded')
  }
  if (!props.account.schedulable) {
    return t('admin.accounts.status.paused')
  }
  return t(`admin.accounts.status.${props.account.status}`)
})

const handleTempUnschedClick = () => {
  if (!isTempUnschedulable.value) return
  emit('show-temp-unsched', props.account)
}
</script>
