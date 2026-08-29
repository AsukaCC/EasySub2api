<template>
  <div class="components-user-dashboard-user-dashboard-charts__panel">
    <!-- Date Range Filter -->
    <div class="components-user-dashboard-user-dashboard-charts__panel-2 card">
      <div class="components-user-dashboard-user-dashboard-charts__panel-3">
        <div class="components-user-dashboard-user-dashboard-charts__panel-4">
          <span class="components-user-dashboard-user-dashboard-charts__text">{{ t('dashboard.timeRange') }}:</span>
          <DateRangePicker :start-date="startDate" :end-date="endDate" @update:startDate="$emit('update:startDate', $event)" @update:endDate="$emit('update:endDate', $event)" @change="$emit('dateRangeChange', $event)" />
        </div>
        <button @click="$emit('refresh')" :disabled="loading" class="btn btn-secondary">
          {{ t('common.refresh') }}
        </button>
        <div class="components-user-dashboard-user-dashboard-charts__panel-5">
          <span class="components-user-dashboard-user-dashboard-charts__text">{{ t('dashboard.granularity') }}:</span>
          <div class="components-user-dashboard-user-dashboard-charts__panel-6">
            <Select :model-value="granularity" :options="[{value:'day', label:t('dashboard.day')}, {value:'hour', label:t('dashboard.hour')}]" @update:model-value="$emit('update:granularity', $event)" @change="$emit('granularityChange')" />
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="components-user-dashboard-user-dashboard-charts__panel-7">
      <!-- Model Distribution Chart -->
      <div class="components-user-dashboard-user-dashboard-charts__panel-8 card">
        <div v-if="loading" class="components-user-dashboard-user-dashboard-charts__panel-9">
          <LoadingSpinner size="md" />
        </div>
        <h3 class="components-user-dashboard-user-dashboard-charts__heading">{{ t('dashboard.modelDistribution') }}</h3>
        <div class="components-user-dashboard-user-dashboard-charts__panel-10">
          <div class="components-user-dashboard-user-dashboard-charts__panel-11">
            <D3DonutChart v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="components-user-dashboard-user-dashboard-charts__panel-12">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="components-user-dashboard-user-dashboard-charts__panel-13">
            <table class="components-user-dashboard-user-dashboard-charts__table">
              <thead>
                <tr class="components-user-dashboard-user-dashboard-charts__row">
                  <th class="components-user-dashboard-user-dashboard-charts__heading-2">{{ t('dashboard.model') }}</th>
                  <th class="components-user-dashboard-user-dashboard-charts__heading-3">{{ t('dashboard.requests') }}</th>
                  <th class="components-user-dashboard-user-dashboard-charts__heading-3">{{ t('dashboard.tokens') }}</th>
                  <th class="components-user-dashboard-user-dashboard-charts__heading-3">{{ t('dashboard.actual') }}</th>
                  <th class="components-user-dashboard-user-dashboard-charts__heading-3">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="components-user-dashboard-user-dashboard-charts__row-2">
                  <td class="components-user-dashboard-user-dashboard-charts__cell" :title="model.model">{{ model.model }}</td>
                  <td class="components-user-dashboard-user-dashboard-charts__cell-2">{{ formatNumber(model.requests) }}</td>
                  <td class="components-user-dashboard-user-dashboard-charts__cell-2">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="components-user-dashboard-user-dashboard-charts__cell-3">{{ formatPoints(model.actual_cost) }}</td>
                  <td class="components-user-dashboard-user-dashboard-charts__cell-4">{{ formatUSD(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Token Usage Trend Chart -->
      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import D3DonutChart from '@/components/charts/d3/D3DonutChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatNumberLocaleString as formatNumber, formatPoints, formatTokensK as formatTokens, formatUSD } from '@/utils/format'
const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']
  }]
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}
</script>
