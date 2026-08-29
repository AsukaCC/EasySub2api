<template>
  <div>
    <!-- Window stats row (above progress bar) -->
    <div
      v-if="windowStats && (windowStats.requests > 0 || windowStats.tokens > 0)"
      class="components-account-usage-progress-bar__panel"
    >
      <div class="components-account-usage-progress-bar__panel-2">
        <span class="components-account-usage-progress-bar__text">
          {{ formatRequests }} req
        </span>
        <span class="components-account-usage-progress-bar__text">
          {{ formatTokens }}
        </span>
        <span class="components-account-usage-progress-bar__text" :title="t('usage.accountBilled')">
          A ${{ formatAccountCost }}
        </span>
        <span
          v-if="windowStats?.user_cost != null"
          class="components-account-usage-progress-bar__text"
          :title="t('usage.userBilled')"
        >
          U {{ formatUserCost }}
        </span>
      </div>
    </div>

    <!-- Progress bar row -->
    <div class="components-account-usage-progress-bar__panel-3">
      <!-- Label badge (fixed width for alignment) -->
      <span
        :class="['components-account-usage-progress-bar__text-3', labelClass]"
      >
        {{ label }}
      </span>

      <!-- Progress bar container -->
      <div class="components-account-usage-progress-bar__panel-4">
        <div
          :class="['components-account-usage-progress-bar__panel-5', barClass]"
          :style="{ width: barWidth }"
        ></div>
      </div>

      <!-- Percentage -->
      <span :class="['components-account-usage-progress-bar__text-4', textClass]">
        {{ displayPercent }}
      </span>

      <!-- Reset time -->
      <span v-if="shouldShowResetTime" class="components-account-usage-progress-bar__text-2">
        {{ formatResetTime }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import type { WindowStats } from '@/types'
import { formatCompactNumber, formatPoints } from '@/utils/format'

const props = defineProps<{
  label: string
  utilization: number // Percentage (0-100+)
  resetsAt?: string | null
  color: 'indigo' | 'emerald' | 'purple' | 'amber'
  windowStats?: WindowStats | null
  showNowWhenIdle?: boolean
  remainingCapacity?: boolean
}>()

const { t } = useI18n()

// Reactive clock for countdown — only runs when a reset time is shown,
// to avoid creating many idle timers across large account lists.
const now = ref(new Date())
const { pause: pauseClock, resume: resumeClock } = useIntervalFn(
  () => {
    now.value = new Date()
  },
  60_000,
  { immediate: false },
)
if (props.resetsAt) resumeClock()
watch(
  () => props.resetsAt,
  (val) => {
    if (val) {
      now.value = new Date()
      resumeClock()
    } else {
      pauseClock()
    }
  },
)

// Label background colors
const labelClass = computed(() => {
  const colors = {
    indigo: 'components-account-usage-progress-bar__state',
    emerald: 'components-account-usage-progress-bar__state-2',
    purple: 'components-account-usage-progress-bar__state-3',
    amber: 'components-account-usage-progress-bar__state-4'
  }
  return colors[props.color]
})

// Progress bar color based on utilization
const barClass = computed(() => {
  if (props.remainingCapacity) {
    if (props.utilization <= 20) {
      return 'status-fill--danger'
    } else if (props.utilization <= 50) {
      return 'status-fill--warning'
    }
    return 'status-fill--success'
  }
  if (props.utilization >= 100) {
    return 'status-fill--danger'
  } else if (props.utilization >= 80) {
    return 'status-fill--warning'
  } else {
    return 'status-fill--success'
  }
})

// Text color based on utilization
const textClass = computed(() => {
  if (props.remainingCapacity) {
    if (props.utilization <= 20) {
      return 'components-account-usage-progress-bar__state-5'
    } else if (props.utilization <= 50) {
      return 'components-account-usage-progress-bar__state-6'
    }
    return 'components-account-usage-progress-bar__state-7'
  }
  if (props.utilization >= 100) {
    return 'components-account-usage-progress-bar__state-5'
  } else if (props.utilization >= 80) {
    return 'components-account-usage-progress-bar__state-6'
  } else {
    return 'components-account-usage-progress-bar__state-7'
  }
})

// Bar width (capped at 100%)
const barWidth = computed(() => {
  return `${Math.min(Math.max(props.utilization, 0), 100)}%`
})

// Display percentage (cap at 999% for readability)
const displayPercent = computed(() => {
  const percent = Math.round(
    props.remainingCapacity
      ? Math.min(Math.max(props.utilization, 0), 100)
      : props.utilization
  )
  return percent > 999 ? '>999%' : `${percent}%`
})

const shouldShowResetTime = computed(() => {
  if (props.resetsAt) return true
  return Boolean(props.showNowWhenIdle && props.utilization <= 0)
})

// Format reset time
const formatResetTime = computed(() => {
  // For rolling windows, when utilization is 0%, treat as immediately available.
  if (props.showNowWhenIdle && props.utilization <= 0) {
    return t('usage.resetNow')
  }

  if (!props.resetsAt) return '-'

  const date = new Date(props.resetsAt)
  const diffMs = date.getTime() - now.value.getTime()

  // resetsAt 已过期：utilization>0 说明后端窗口数据还没刷新（active poll 没回写），
  // 显示「待刷新」以区别于真正可用的「现在」。
  if (diffMs <= 0) {
    return props.utilization > 0 ? t('usage.resetPending') : t('usage.resetNow')
  }

  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))

  if (diffHours >= 24) {
    const days = Math.floor(diffHours / 24)
    return `${days}d ${diffHours % 24}h`
  } else if (diffHours > 0) {
    return `${diffHours}h ${diffMins}m`
  } else {
    return `${diffMins}m`
  }
})

// Window stats formatters
const formatRequests = computed(() => {
  if (!props.windowStats) return ''
  return formatCompactNumber(props.windowStats.requests, { allowBillions: false })
})

const formatTokens = computed(() => {
  if (!props.windowStats) return ''
  return formatCompactNumber(props.windowStats.tokens)
})

const formatAccountCost = computed(() => {
  if (!props.windowStats) return '0.00'
  return props.windowStats.cost.toFixed(2)
})

const formatUserCost = computed(() => {
  return formatPoints(props.windowStats?.user_cost)
})

</script>
