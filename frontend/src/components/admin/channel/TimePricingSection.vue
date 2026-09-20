<template>
  <section class="components-admin-channel-time-pricing-section__section">
    <div class="components-admin-channel-time-pricing-section__panel">
      <div class="components-admin-channel-time-pricing-section__panel-2">
        <label class="components-admin-channel-time-pricing-section__label">
          {{ t('admin.channels.form.timePricing') }}
        </label>
        <label class="components-admin-channel-time-pricing-section__label-2">
          {{ t('admin.channels.form.timezone') }}
        </label>
        <Select
          :model-value="modelValue.timezone"
          :options="timezoneOptions"
          :aria-label="t('admin.channels.form.timezone')"
          searchable
          creatable
          class="components-admin-channel-time-pricing-section__field"
          @update:model-value="updateTimezone"
        />
      </div>
      <button
        type="button"
        class="components-admin-channel-time-pricing-section__action"
        data-testid="add-time-period"
        @click="addPeriod"
      >
        + {{ t('admin.channels.form.addTimePeriod') }}
      </button>
    </div>

    <div v-if="modelValue.periods.length > 0" class="components-admin-channel-time-pricing-section__panel-3">
      <div
        v-for="(period, index) in modelValue.periods"
        :key="index"
        class="components-admin-channel-time-pricing-section__panel-4"
      >
        <div class="components-admin-channel-time-pricing-section__panel-5">
          <label :for="`${inputIdPrefix}-start-${index}`" class="components-admin-channel-time-pricing-section__label-3">
            {{ t('admin.channels.form.startTime') }}
          </label>
          <DateTimePicker
            :id="`${inputIdPrefix}-start-${index}`"
            :model-value="period.start_time"
            type="time"
            :step="1"
            class="components-admin-channel-time-pricing-section__field-2 input"
            @update:model-value="updatePeriod(index, 'start_time', $event)"
          />
        </div>
        <div class="components-admin-channel-time-pricing-section__panel-5">
          <label :for="`${inputIdPrefix}-end-${index}`" class="components-admin-channel-time-pricing-section__label-3">
            {{ t('admin.channels.form.endTime') }}
          </label>
          <DateTimePicker
            :id="`${inputIdPrefix}-end-${index}`"
            :model-value="period.end_time"
            type="time"
            :step="1"
            class="components-admin-channel-time-pricing-section__field-2 input"
            @update:model-value="updatePeriod(index, 'end_time', $event)"
          />
        </div>
        <div class="components-admin-channel-time-pricing-section__panel-5">
          <label :for="`${inputIdPrefix}-multiplier-${index}`" class="components-admin-channel-time-pricing-section__label-3">
            {{ t('admin.channels.form.multiplier') }}
          </label>
          <input
            :id="`${inputIdPrefix}-multiplier-${index}`"
            :value="period.multiplier"
            type="number"
            min="0.01"
            step="0.01"
            class="components-admin-channel-time-pricing-section__field-2 input"
            @input="updatePeriod(index, 'multiplier', ($event.target as HTMLInputElement).value)"
            @blur="formatMultiplier(index, ($event.target as HTMLInputElement).value)"
          />
        </div>
        <button
          type="button"
          class="components-admin-channel-time-pricing-section__action-2"
          :title="t('admin.channels.form.removeTimePeriod')"
          :aria-label="t('admin.channels.form.removeTimePeriod')"
          :data-testid="`remove-time-period-${index}`"
          @click="removePeriod(index)"
        >
          <Icon name="trash" size="sm" />
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import DateTimePicker from '@/components/common/DateTimePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  COMMON_TIMEZONES,
  formatTimezoneOffset,
  isValidTimePricingMultiplier,
  type TimePricingFormEntry,
  type TimePricingPeriodFormEntry,
} from './types'

const { t } = useI18n()

const props = defineProps<{ modelValue: TimePricingFormEntry }>()
const emit = defineEmits<{ 'update:modelValue': [value: TimePricingFormEntry] }>()
const inputIdPrefix = `time-pricing-${getCurrentInstance()?.uid}`

const timezoneOptions = COMMON_TIMEZONES.map(value => {
  const offset = formatTimezoneOffset(value)
  return { value, label: offset ? `${value} (${offset})` : value }
})

function updateTimezone(value: string | number | boolean | null) {
  emit('update:modelValue', { ...props.modelValue, timezone: String(value ?? '') })
}

function addPeriod() {
  emit('update:modelValue', {
    ...props.modelValue,
    periods: [
      ...props.modelValue.periods,
      { start_time: '', end_time: '', multiplier: '1.00' },
    ],
  })
}

function updatePeriod(index: number, field: keyof TimePricingPeriodFormEntry, value: string) {
  const periods = props.modelValue.periods.map((period, current) =>
    current === index ? { ...period, [field]: value } : period)
  emit('update:modelValue', { ...props.modelValue, periods })
}

function formatMultiplier(index: number, value: string) {
  if (!isValidTimePricingMultiplier(value)) return
  updatePeriod(index, 'multiplier', Number(value).toFixed(2))
}

function removePeriod(index: number) {
  emit('update:modelValue', {
    ...props.modelValue,
    periods: props.modelValue.periods.filter((_period, current) => current !== index),
  })
}
</script>
