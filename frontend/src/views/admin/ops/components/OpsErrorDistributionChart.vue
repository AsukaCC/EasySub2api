<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OpsErrorDistributionResponse } from '@/api/admin/ops'
import type { ChartState } from '../types'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import D3DonutChart from '@/components/charts/d3/D3DonutChart.vue'

interface Props {
  data: OpsErrorDistributionResponse | null
  loading: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'openDetails'): void
}>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const colors = computed(() => ({
  blue: '#3b82f6',
  red: '#ef4444',
  orange: '#f59e0b',
  gray: '#9ca3af',
  text: isDarkMode.value ? '#9ca3af' : '#71717a'
}))

const totalSlaErrors = computed(() =>
  (props.data?.items ?? []).reduce((total, item) => total + Number(item.sla || 0), 0)
)

const hasData = computed(() => totalSlaErrors.value > 0)

const state = computed<ChartState>(() => {
  if (hasData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

interface ErrorCategory {
  label: string
  count: number
  color: string
}

const categories = computed<ErrorCategory[]>(() => {
  if (!props.data) return []

  let upstream = 0 // 502, 503, 504
  let client = 0 // 4xx
  let system = 0 // 500
  let other = 0

  for (const item of props.data.items || []) {
    const code = Number(item.status_code || 0)
    const count = Number(item.sla || 0)
    if (!Number.isFinite(code) || !Number.isFinite(count)) continue

    if ([502, 503, 504].includes(code)) upstream += count
    else if (code >= 400 && code < 500) client += count
    else if (code === 500) system += count
    else other += count
  }

  const out: ErrorCategory[] = []
  if (upstream > 0) out.push({ label: t('admin.ops.upstream'), count: upstream, color: colors.value.orange })
  if (client > 0) out.push({ label: t('admin.ops.client'), count: client, color: colors.value.blue })
  if (system > 0) out.push({ label: t('admin.ops.system'), count: system, color: colors.value.red })
  if (other > 0) out.push({ label: t('admin.ops.other'), count: other, color: colors.value.gray })
  return out
})

const topReason = computed(() => {
  if (categories.value.length === 0) return null
  return categories.value.reduce((prev, cur) => (cur.count > prev.count ? cur : prev))
})

const chartData = computed(() => {
  if (!hasData.value || categories.value.length === 0) return null
  return {
    labels: categories.value.map((c) => c.label),
    datasets: [
      {
        data: categories.value.map((c) => c.count),
        backgroundColor: categories.value.map((c) => c.color),
        borderWidth: 0
      }
    ]
  }
})

const options = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: isDarkMode.value ? '#1b1b1f' : '#ffffff',
      titleColor: isDarkMode.value ? '#f3f4f6' : '#09090b',
      bodyColor: isDarkMode.value ? '#d1d5db' : '#3f3f46'
    }
  }
}))
</script>

<template>
  <div class="views-admin-ops-components-ops-error-distribution-chart__panel card-body">
    <div class="views-admin-ops-components-ops-error-distribution-chart__panel-2">
      <h3 class="views-admin-ops-components-ops-error-distribution-chart__heading">
        <svg class="views-admin-ops-components-ops-error-distribution-chart__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        {{ t('admin.ops.errorDistribution') }}
        <HelpTooltip :content="t('admin.ops.tooltips.errorDistribution')" />
      </h3>
      <button
        type="button"
        class="views-admin-ops-components-ops-error-distribution-chart__action"
        :disabled="state !== 'ready'"
        :title="t('admin.ops.errorTrend')"
        @click="emit('openDetails')"
      >
        {{ t('admin.ops.requestDetails.details') }}
      </button>
    </div>

    <div class="views-admin-ops-components-ops-error-distribution-chart__panel-3">
      <div v-if="state === 'ready' && chartData" class="views-admin-ops-components-ops-error-distribution-chart__panel-4">
        <div class="views-admin-ops-components-ops-error-distribution-chart__panel-5">
          <D3DonutChart :data="chartData" :options="{ ...options, cutout: '65%' }" />
        </div>
        <div class="views-admin-ops-components-ops-error-distribution-chart__panel-6">
          <div v-if="topReason" class="views-admin-ops-components-ops-error-distribution-chart__panel-7">
            {{ t('admin.ops.top') }}: <span :style="{ color: topReason.color }">{{ topReason.label }}</span>
          </div>
          <div class="views-admin-ops-components-ops-error-distribution-chart__panel-8">
            <div v-for="item in categories" :key="item.label" class="views-admin-ops-components-ops-error-distribution-chart__panel-9">
              <span class="views-admin-ops-components-ops-error-distribution-chart__text" :style="{ backgroundColor: item.color }"></span>
              <span class="views-admin-ops-components-ops-error-distribution-chart__text-2">{{ item.label }} {{ item.count }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="views-admin-ops-components-ops-error-distribution-chart__panel-10">
        <div v-if="state === 'loading'" class="views-admin-ops-components-ops-error-distribution-chart__panel-11">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyError')" />
      </div>
    </div>
  </div>
</template>
