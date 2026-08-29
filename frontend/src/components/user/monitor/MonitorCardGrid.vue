<template>
  <div>
    <div
      v-if="loading && items.length === 0"
      class="components-user-monitor-monitor-card-grid__panel"
    >
      <div
        v-for="i in 6"
        :key="i"
        class="components-user-monitor-monitor-card-grid__panel-2"
      >
        <div class="components-user-monitor-monitor-card-grid__panel-3">
          <div class="components-user-monitor-monitor-card-grid__panel-4"></div>
          <div class="components-user-monitor-monitor-card-grid__panel-5">
            <div class="components-user-monitor-monitor-card-grid__panel-6"></div>
            <div class="components-user-monitor-monitor-card-grid__panel-7"></div>
          </div>
          <div class="components-user-monitor-monitor-card-grid__panel-8"></div>
        </div>
        <div class="components-user-monitor-monitor-card-grid__panel-9">
          <div class="components-user-monitor-monitor-card-grid__panel-10"></div>
          <div class="components-user-monitor-monitor-card-grid__panel-10"></div>
        </div>
        <div class="components-user-monitor-monitor-card-grid__panel-11"></div>
      </div>
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      :title="t('channelStatus.empty.title')"
      :description="t('channelStatus.empty.description')"
    />

    <div
      v-else
      class="components-user-monitor-monitor-card-grid__panel"
    >
      <MonitorCard
        v-for="item in items"
        :key="item.id"
        :item="item"
        :window="window"
        :availability-value="resolveAvailability(item)"
        :countdown-seconds="countdownSeconds"
        @click="emit('cardClick', item)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { UserMonitorView, UserMonitorDetail } from '@/api/channelMonitor'
import EmptyState from '@/components/common/EmptyState.vue'
import MonitorCard from './MonitorCard.vue'

const props = defineProps<{
  items: UserMonitorView[]
  window: '7d' | '15d' | '30d'
  countdownSeconds: number
  loading: boolean
  detailCache: Record<string, UserMonitorDetail>
}>()

const emit = defineEmits<{
  (e: 'cardClick', item: UserMonitorView): void
}>()

const { t } = useI18n()

function resolveAvailability(item: UserMonitorView): number | null {
  if (props.window === '7d') {
    return item.availability_7d ?? null
  }
  const detail = props.detailCache[item.id]
  if (!detail) return null
  const primary = detail.models.find(m => m.model === item.primary_model)
  if (!primary) return null
  return props.window === '15d' ? primary.availability_15d ?? null : primary.availability_30d ?? null
}
</script>
