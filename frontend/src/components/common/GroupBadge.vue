<template>
  <span
    :class="[
      'components-common-group-badge__text-4',
      badgeClass
    ]"
  >
    <!-- Platform logo -->
    <PlatformIcon v-if="platform" :platform="platform" size="sm" />
    <!-- Group name -->
    <span class="components-common-group-badge__text">{{ name }}</span>
    <!-- Right side label -->
    <span v-if="showLabel" :class="labelClass">
      <GroupRateDisplay
        v-if="showsRateValue"
        :rate-multiplier="rateMultiplier"
        :user-rate-multiplier="userRateMultiplier"
      />
      <template v-else>{{ labelText }}</template>
    </span>
    <span v-if="hasPeakRate" :class="peakRateClass" :title="peakRateTitle">
      {{ peakRateText }}
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionType, GroupPlatform } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import PlatformIcon from './PlatformIcon.vue'
import GroupRateDisplay from './GroupRateDisplay.vue'

interface Props {
  name: string
  platform?: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null // 用户专属倍率
  peakRateEnabled?: boolean
  peakStart?: string
  peakEnd?: string
  peakRateMultiplier?: number
  showRate?: boolean
  daysRemaining?: number | null // 剩余天数（订阅类型时使用）
  /**
   * 订阅分组默认在右侧 label 展示"订阅"或剩余天数；
   * 开启后订阅分组也改为显示倍率（保留订阅主题色 label，配合可用渠道这类
   * 只关心费率、不关心有效期的场景）。
   */
  alwaysShowRate?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  showRate: true,
  daysRemaining: null,
  userRateMultiplier: null,
  peakRateEnabled: false,
  alwaysShowRate: false
})

const { t } = useI18n()

const isSubscription = computed(() => props.subscriptionType === 'subscription')

const showsRateValue = computed(() => !isSubscription.value || props.alwaysShowRate)

const appStore = useAppStore()

const hasPeakRate = computed(() => {
  return Boolean(props.showRate && props.peakRateEnabled && props.peakStart && props.peakEnd)
})

const peakRateText = computed(() => {
  return formatPeakRateWindow(
    {
      peak_rate_enabled: props.peakRateEnabled,
      peak_start: props.peakStart,
      peak_end: props.peakEnd,
      peak_rate_multiplier: props.peakRateMultiplier
    },
    serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  )
})

const peakRateTitle = computed(() => {
  return t('common.peakRateTooltip', { window: peakRateText.value })
})

// 是否显示右侧标签
const showLabel = computed(() => {
  if (!props.showRate) return false
  // 订阅类型：显示天数或"订阅"
  if (isSubscription.value) return true
  // 标准类型：显示倍率（包括专属倍率）
  return props.rateMultiplier !== undefined || props.userRateMultiplier !== null
})

// Label text
const labelText = computed(() => {
  const rateLabel = props.rateMultiplier !== undefined ? `${props.rateMultiplier}x` : ''
  if (isSubscription.value && !props.alwaysShowRate) {
    // 如果有剩余天数，显示天数
    if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
      if (props.daysRemaining <= 0) {
        return t('admin.users.expired')
      }
      return t('admin.users.daysRemaining', { days: props.daysRemaining })
    }
    // 否则显示"订阅"
    return t('groups.subscription')
  }
  return rateLabel
})

// Label style based on type and days remaining
const labelClass = computed(() => {
  const base = 'components-common-group-badge__state'

  if (!isSubscription.value) {
    // Standard: subtle background (不再为专属倍率使用不同的背景色)
    return `${base} group-badge__label--neutral`
  }

  // 订阅类型：根据剩余天数显示不同颜色
  if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
    if (props.daysRemaining <= 0 || props.daysRemaining <= 3) {
      // 已过期或紧急（<=3天）：红色
      return `${base} group-badge__label--danger`
    }
    if (props.daysRemaining <= 7) {
      // 警告（<=7天）：橙色
      return `${base} group-badge__label--warning`
    }
  }

  // 正常状态或无天数：根据平台显示主题色
  if (props.platform === 'anthropic') {
    return `${base} group-badge__label--anthropic`
  }
  if (props.platform === 'openai') {
    return `${base} group-badge__label--openai`
  }
  if (props.platform === 'grok') {
    return `${base} group-badge__label--grok`
  }
  if (props.platform === 'kimi') {
    return `${base} group-badge__label--kimi`
  }
  if (props.platform === 'zhipu') {
    return `${base} group-badge__label--zhipu`
  }
  if (props.platform === 'deepseek') {
    return `${base} group-badge__label--deepseek`
  }
  if (props.platform === 'composite') {
    return `${base} group-badge__label--composite`
  }
  return `${base} group-badge__label--default`
})

const peakRateClass = computed(() => {
  return 'components-common-group-badge__state-2'
})

// Badge color based on platform and subscription type
const badgeClass = computed(() => {
  if (props.platform === 'anthropic') {
    // Claude: orange theme
    return isSubscription.value
      ? 'components-common-group-badge__state-3'
      : 'components-common-group-badge__state-4'
  } else if (props.platform === 'openai') {
    // OpenAI: green theme
    return isSubscription.value
      ? 'components-common-group-badge__state-5'
      : 'components-common-group-badge__state-6'
  }
  if (props.platform === 'grok') {
    return isSubscription.value
      ? 'components-common-group-badge__state-11'
      : 'components-common-group-badge__state-12'
  }
  if (props.platform === 'kimi') {
    return isSubscription.value
      ? 'components-common-group-badge__state-13'
      : 'components-common-group-badge__state-14'
  }
  if (props.platform === 'zhipu') {
    return isSubscription.value
      ? 'components-common-group-badge__state-15'
      : 'components-common-group-badge__state-16'
  }
  if (props.platform === 'deepseek') {
    return isSubscription.value
      ? 'components-common-group-badge__state-17'
      : 'components-common-group-badge__state-18'
  }
  if (props.platform === 'composite') {
    return isSubscription.value
      ? 'components-common-group-badge__state-19'
      : 'components-common-group-badge__state-20'
  }
  // Fallback: original colors
  return isSubscription.value
    ? 'components-common-group-badge__state-21'
    : 'components-common-group-badge__state-5'
})
</script>
