<template>
  <BaseDialog :show="show" :title="t('usage.exporting')" width="narrow" @close="handleCancel">
    <div class="components-common-export-progress-dialog__panel">
      <div class="components-common-export-progress-dialog__panel-2">
        {{ t('usage.exportingProgress') }}
      </div>
      <div class="components-common-export-progress-dialog__panel-3">
        <span>{{ t('usage.exportedCount', { current, total }) }}</span>
        <span class="components-common-export-progress-dialog__text">{{ normalizedProgress }}%</span>
      </div>
      <div class="components-common-export-progress-dialog__panel-4">
        <div
          role="progressbar"
          :aria-valuenow="normalizedProgress"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-label="`${t('usage.exportingProgress')}: ${normalizedProgress}%`"
          class="components-common-export-progress-dialog__panel-5"
          :style="{ width: `${normalizedProgress}%` }"
        ></div>
      </div>
      <div v-if="estimatedTime" class="components-common-export-progress-dialog__panel-6" aria-live="polite" aria-atomic="true">
        {{ t('usage.estimatedTime', { time: estimatedTime }) }}
      </div>
    </div>

    <template #footer>
      <button
        @click="handleCancel"
        type="button"
        class="components-common-export-progress-dialog__action"
      >
        {{ t('usage.cancelExport') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

interface Props {
  show: boolean
  progress: number
  current: number
  total: number
  estimatedTime: string
}

interface Emits {
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const { t } = useI18n()

const normalizedProgress = computed(() => {
  const value = Number.isFinite(props.progress) ? props.progress : 0
  return Math.min(100, Math.max(0, Math.round(value)))
})

const handleCancel = () => {
  emit('cancel')
}
</script>
