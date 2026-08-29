<template>
  <div class="components-admin-monitor-monitor-primary-model-cell__panel">
    <div class="components-admin-monitor-monitor-primary-model-cell__panel-2">
      <span class="components-admin-monitor-monitor-primary-model-cell__text">{{ row.primary_model }}</span>
      <HelpTooltip>
      <template #trigger>
        <span
          class="components-admin-monitor-monitor-primary-model-cell__text-2"
          :class="statusBadgeClass(row.primary_status)"
        >
          {{ statusLabel(row.primary_status) }}
        </span>
      </template>
      <div class="components-admin-monitor-monitor-primary-model-cell__panel-3">
        <div class="components-admin-monitor-monitor-primary-model-cell__panel-4">
          {{ row.primary_model }}
          <span
            class="components-admin-monitor-monitor-primary-model-cell__text-3"
            :class="statusBadgeClass(row.primary_status)"
          >
            {{ statusLabel(row.primary_status) }}
          </span>
        </div>
        <div v-if="(row.extra_models?.length ?? 0) === 0" class="components-admin-monitor-monitor-primary-model-cell__panel-5">
          {{ t('monitorCommon.extraModelsEmpty') }}
        </div>
        <div v-else class="components-admin-monitor-monitor-primary-model-cell__panel-6">
          <div class="components-admin-monitor-monitor-primary-model-cell__panel-7">
            {{ t('monitorCommon.extraModelsHeader') }}
          </div>
          <table class="components-admin-monitor-monitor-primary-model-cell__table">
            <thead>
              <tr class="components-admin-monitor-monitor-primary-model-cell__row">
                <th class="components-admin-monitor-monitor-primary-model-cell__heading">{{ t('admin.channelMonitor.columns.primaryModel') }}</th>
                <th class="components-admin-monitor-monitor-primary-model-cell__heading">{{ t('admin.channelMonitor.columns.actions') }}</th>
                <th class="components-admin-monitor-monitor-primary-model-cell__heading-2">{{ t('admin.channelMonitor.columns.latency') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in (row.extra_models_status || [])" :key="m.model">
                <td class="components-admin-monitor-monitor-primary-model-cell__cell">{{ m.model }}</td>
                <td class="components-admin-monitor-monitor-primary-model-cell__cell-2">
                  <span
                    class="components-admin-monitor-monitor-primary-model-cell__text-4"
                    :class="statusBadgeClass(m.status)"
                  >
                    {{ statusLabel(m.status) }}
                  </span>
                </td>
                <td class="components-admin-monitor-monitor-primary-model-cell__cell-3">{{ formatLatency(m.latency_ms) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      </HelpTooltip>
    </div>
    <!-- 配额模式监控：主模型行内联展示最新用量/余额快照（管理端不受用户端开关限制） -->
    <MonitorQuotaView :snapshot="row.latest_quota" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { ChannelMonitor } from '@/api/admin/channelMonitor'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import MonitorQuotaView from '@/components/common/MonitorQuotaView.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

defineProps<{
  row: ChannelMonitor
}>()

const { t } = useI18n()
const { statusLabel, statusBadgeClass, formatLatency } = useChannelMonitorFormat()
</script>
