<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.runResultTitle')"
    width="normal"
    @close="$emit('close')"
  >
    <div class="components-admin-monitor-monitor-run-result-dialog__panel">
      <div
        v-for="r in results"
        :key="r.model"
        class="components-admin-monitor-monitor-run-result-dialog__panel-2"
      >
        <div class="components-admin-monitor-monitor-run-result-dialog__panel-3">
          <span class="components-admin-monitor-monitor-run-result-dialog__text">{{ r.model }}</span>
          <span v-if="r.message" class="components-admin-monitor-monitor-run-result-dialog__text-2">{{ r.message }}</span>
          <MonitorQuotaView :snapshot="r.quota" class="components-admin-monitor-monitor-run-result-dialog__monitor-quota-view" />
        </div>
        <div class="components-admin-monitor-monitor-run-result-dialog__panel-4">
          <span
            class="components-admin-monitor-monitor-run-result-dialog__text-3"
            :class="statusBadgeClass(r.status)"
          >
            {{ statusLabel(r.status) }}
          </span>
          <span class="components-admin-monitor-monitor-run-result-dialog__text-2">{{ formatLatency(r.latency_ms) }} ms</span>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="components-admin-monitor-monitor-run-result-dialog__panel-5">
        <button @click="$emit('close')" class="btn btn-primary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CheckResult } from '@/api/admin/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import MonitorQuotaView from '@/components/common/MonitorQuotaView.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

defineProps<{
  show: boolean
  results: CheckResult[]
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const { statusLabel, statusBadgeClass, formatLatency } = useChannelMonitorFormat()
</script>
