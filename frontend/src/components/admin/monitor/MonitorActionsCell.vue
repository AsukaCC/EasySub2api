<template>
  <div class="components-admin-monitor-monitor-actions-cell__panel">
    <button
      @click="$emit('run', row)"
      :disabled="running"
      class="components-admin-monitor-monitor-actions-cell__action"
    >
      <Icon name="refresh" size="sm" :class="running ? 'components-admin-monitor-monitor-actions-cell__icon' : ''" />
      <span class="components-admin-monitor-monitor-actions-cell__text">{{ t('admin.channelMonitor.runNow') }}</span>
    </button>
    <button
      data-testid="monitor-duplicate"
      :title="duplicateTitle"
      :disabled="duplicating || Boolean(row.api_key_decrypt_failed)"
      @click="$emit('duplicate', row)"
      class="components-admin-monitor-monitor-actions-cell__action-2"
    >
      <Icon name="copy" size="sm" />
      <span class="components-admin-monitor-monitor-actions-cell__text">
        {{ duplicating ? t('admin.channelMonitor.duplicating') : t('admin.channelMonitor.duplicate') }}
      </span>
    </button>
    <button
      @click="$emit('edit', row)"
      class="components-admin-monitor-monitor-actions-cell__action"
    >
      <Icon name="edit" size="sm" />
      <span class="components-admin-monitor-monitor-actions-cell__text">{{ t('common.edit') }}</span>
    </button>
    <button
      @click="$emit('delete', row)"
      class="components-admin-monitor-monitor-actions-cell__action-3"
    >
      <Icon name="trash" size="sm" />
      <span class="components-admin-monitor-monitor-actions-cell__text">{{ t('common.delete') }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChannelMonitor } from '@/api/admin/channelMonitor'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  row: ChannelMonitor
  running: boolean
  duplicating: boolean
}>()

defineEmits<{
  (e: 'run', row: ChannelMonitor): void
  (e: 'duplicate', row: ChannelMonitor): void
  (e: 'edit', row: ChannelMonitor): void
  (e: 'delete', row: ChannelMonitor): void
}>()

const { t } = useI18n()
const duplicateTitle = computed(() => {
  if (props.row.api_key_decrypt_failed) return t('admin.channelMonitor.duplicateKeyUnavailable')
  if (props.duplicating) return t('admin.channelMonitor.duplicating')
  return t('admin.channelMonitor.duplicate')
})
</script>
