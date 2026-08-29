<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.template.applyPickerTitle', { name: templateName })"
    @close="$emit('close')"
  >
    <p class="components-admin-monitor-monitor-template-apply-picker-dialog__description">
      {{ t('admin.channelMonitor.template.applyPickerHint') }}
    </p>

    <div v-if="loading" class="components-admin-monitor-monitor-template-apply-picker-dialog__panel">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="monitors.length === 0" class="components-admin-monitor-monitor-template-apply-picker-dialog__panel">
      {{ t('admin.channelMonitor.template.applyPickerEmpty') }}
    </div>

    <div v-else>
      <!-- 全选/全不选 -->
      <div class="components-admin-monitor-monitor-template-apply-picker-dialog__panel-2">
        <button
          type="button"
          class="components-admin-monitor-monitor-template-apply-picker-dialog__action"
          @click="selectAll"
        >
          {{ t('common.selectAll') }}
        </button>
        <button
          type="button"
          class="components-admin-monitor-monitor-template-apply-picker-dialog__action-2"
          @click="selectNone"
        >
          {{ t('admin.channelMonitor.template.selectNone') }}
        </button>
        <span class="components-admin-monitor-monitor-template-apply-picker-dialog__text">
          {{ t('admin.channelMonitor.template.selectedCount', {
            n: selectedIds.length,
            total: monitors.length,
          }) }}
        </span>
      </div>

      <ul class="components-admin-monitor-monitor-template-apply-picker-dialog__list">
        <li
          v-for="m in monitors"
          :key="m.id"
          class="components-admin-monitor-monitor-template-apply-picker-dialog__item"
          @click="toggle(m.id)"
        >
          <input
            type="checkbox"
            :checked="selectedSet.has(m.id)"
            class="components-admin-monitor-monitor-template-apply-picker-dialog__field"
            @click.stop="toggle(m.id)"
          />
          <span class="components-admin-monitor-monitor-template-apply-picker-dialog__text-2">{{ m.name }}</span>
          <span class="components-admin-monitor-monitor-template-apply-picker-dialog__text-3">{{ m.provider }}</span>
          <span v-if="m.provider === 'openai'" class="components-admin-monitor-monitor-template-apply-picker-dialog__text-3">{{ m.api_mode }}</span>
          <span
            v-if="!m.enabled"
            class="components-admin-monitor-monitor-template-apply-picker-dialog__text-4"
          >
            {{ t('admin.channelMonitor.onlyDisabled').replace(/^仅|^Only /, '') }}
          </span>
        </li>
      </ul>
    </div>

    <template #footer>
      <div class="components-admin-monitor-monitor-template-apply-picker-dialog__panel-3">
        <button class="btn btn-secondary" @click="$emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          :disabled="submitting || selectedIds.length === 0"
          @click="handleApply"
        >
          {{ submitting
            ? t('common.submitting')
            : t('admin.channelMonitor.template.applyPickerConfirm', { n: selectedIds.length }) }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { adminAPI } from '@/api/admin'
import type { AssociatedMonitorBrief } from '@/api/admin/channelMonitorTemplate'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{
  show: boolean
  templateId: string | null
  templateName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'applied', affected: number): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const submitting = ref(false)
const monitors = ref<AssociatedMonitorBrief[]>([])
const selectedIds = ref<string[]>([])

const selectedSet = computed(() => new Set(selectedIds.value))

watch(
  () => [props.show, props.templateId] as const,
  ([show, id]) => {
    if (!show || id == null) return
    void fetchMonitors(id)
  },
  { immediate: true },
)

async function fetchMonitors(id: string) {
  loading.value = true
  monitors.value = []
  selectedIds.value = []
  try {
    const { items } = await adminAPI.channelMonitorTemplate.listAssociatedMonitors(id)
    monitors.value = items
    // 默认全选
    selectedIds.value = items.map((m) => m.id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

function toggle(id: string) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function selectAll() {
  selectedIds.value = monitors.value.map((m) => m.id)
}

function selectNone() {
  selectedIds.value = []
}

async function handleApply() {
  if (props.templateId == null || selectedIds.value.length === 0 || submitting.value) return
  submitting.value = true
  try {
    const { affected } = await adminAPI.channelMonitorTemplate.apply(
      props.templateId,
      [...selectedIds.value],
    )
    appStore.showSuccess(t('admin.channelMonitor.template.applySuccess', { n: affected }))
    emit('applied', affected)
    emit('close')
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    submitting.value = false
  }
}
</script>
