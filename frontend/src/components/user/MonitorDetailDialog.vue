<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="components-user-monitor-detail-dialog__panel">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="components-user-monitor-detail-dialog__panel">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="components-user-monitor-detail-dialog__panel-2">
      <table class="components-user-monitor-detail-dialog__table">
        <thead class="components-user-monitor-detail-dialog__header">
          <tr class="components-user-monitor-detail-dialog__row">
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.model') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.latestStatus') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.latestLatency') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.availability7d') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.availability15d') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.availability30d') }}</th>
            <th class="components-user-monitor-detail-dialog__heading">{{ t('channelStatus.detailColumns.avgLatency7d') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="m in detail.models"
            :key="m.model"
            class="components-user-monitor-detail-dialog__row-2"
          >
            <td class="components-user-monitor-detail-dialog__cell">{{ m.model }}</td>
            <td class="components-user-monitor-detail-dialog__heading">
              <span
                class="components-user-monitor-detail-dialog__text"
                :class="statusBadgeClass(m.latest_status)"
              >
                {{ statusLabel(m.latest_status) }}
              </span>
            </td>
            <td class="components-user-monitor-detail-dialog__cell-2">{{ formatLatency(m.latest_latency_ms) }}</td>
            <td class="components-user-monitor-detail-dialog__cell-2">{{ formatPercent(m.availability_7d) }}</td>
            <td class="components-user-monitor-detail-dialog__cell-2">{{ formatPercent(m.availability_15d) }}</td>
            <td class="components-user-monitor-detail-dialog__cell-2">{{ formatPercent(m.availability_30d) }}</td>
            <td class="components-user-monitor-detail-dialog__cell-2">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <div class="components-user-monitor-detail-dialog__panel-3">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('channelStatus.closeDetail') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  status as fetchChannelMonitorDetail,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  show: boolean
  monitorId: string | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { statusLabel, statusBadgeClass, formatLatency, formatPercent } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const loading = ref(false)

async function load(id: string) {
  detail.value = null
  loading.value = true
  try {
    detail.value = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>
