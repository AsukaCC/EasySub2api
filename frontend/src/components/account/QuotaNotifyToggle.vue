<script setup lang="ts">
import { QUOTA_THRESHOLD_TYPE_FIXED, QUOTA_THRESHOLD_TYPE_PERCENTAGE, type QuotaThresholdType } from '@/constants/account'
import Select from '@/components/common/Select.vue'

defineProps<{
  enabled: boolean | null
  threshold: number | null
  thresholdType: QuotaThresholdType | null
}>()

const emit = defineEmits<{
  'update:enabled': [value: boolean | null]
  'update:threshold': [value: number | null]
  'update:thresholdType': [value: QuotaThresholdType | null]
}>()
</script>

<template>
  <div class="components-account-quota-notify-toggle__panel">
    <button
      type="button"
      @click="emit('update:enabled', !enabled)"
      :class="[
        'components-account-quota-notify-toggle__action',
        enabled ? 'components-account-quota-notify-toggle__action-2' : 'components-account-quota-notify-toggle__action-3'
      ]"
    >
      <span
        :class="[
          'components-account-quota-notify-toggle__text',
          enabled ? 'toggle-thumb--on' : 'components-account-quota-notify-toggle__text-2'
        ]"
      />
    </button>
    <template v-if="enabled">
      <input
        :value="threshold"
        @input="emit('update:threshold', parseFloat(($event.target as HTMLInputElement).value) || null)"
        type="number"
        min="0"
        :max="thresholdType === QUOTA_THRESHOLD_TYPE_PERCENTAGE ? 100 : undefined"
        :step="thresholdType === QUOTA_THRESHOLD_TYPE_PERCENTAGE ? 1 : 0.01"
        class="components-account-quota-notify-toggle__field input"
      />
      <Select
        :model-value="thresholdType || QUOTA_THRESHOLD_TYPE_FIXED"
        :options="[
          { value: QUOTA_THRESHOLD_TYPE_FIXED, label: '$' },
          { value: QUOTA_THRESHOLD_TYPE_PERCENTAGE, label: '%' }
        ]"
        @change="emit('update:thresholdType', $event as QuotaThresholdType)"
        class="components-account-quota-notify-toggle__field-2 input"
      />
    </template>
  </div>
</template>
