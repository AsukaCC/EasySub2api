<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import QuotaNotifyToggle from './QuotaNotifyToggle.vue'
import Select from '@/components/common/Select.vue'
import type { QuotaThresholdType, QuotaResetMode } from '@/constants/account'

const { t } = useI18n()

const props = defineProps<{
  dim: 'daily' | 'weekly' | 'total'
  label: string
  limit: number | null
  quotaNotifyGlobalEnabled: boolean
  notifyEnabled: boolean | null
  notifyThreshold: number | null
  notifyThresholdType: QuotaThresholdType | null
  // Reset mode (only for daily/weekly, null for total)
  resetMode: QuotaResetMode | null
  resetHour: number | null
  resetDay: number | null  // weekly only
  resetTimezone: string | null
  hintRolling: string
  hintFixed: string
  // Shared options passed from parent
  hourOptions: number[]
  dayOptions: { value: number; key: string }[]
  timezoneOptions?: string[]
}>()

const emit = defineEmits<{
  'update:limit': [value: number | null]
  'update:notifyEnabled': [value: boolean | null]
  'update:notifyThreshold': [value: number | null]
  'update:notifyThresholdType': [value: QuotaThresholdType | null]
  'update:resetMode': [value: QuotaResetMode | null]
  'update:resetHour': [value: number | null]
  'update:resetDay': [value: number | null]
  'update:resetTimezone': [value: string | null]
}>()

const hasResetMode = props.dim !== 'total'

const onLimitInput = (e: Event) => {
  const raw = (e.target as HTMLInputElement).valueAsNumber
  emit('update:limit', Number.isNaN(raw) ? null : raw)
}

const onModeChange = (value: string | number | boolean | null) => {
  const val = value as QuotaResetMode
  emit('update:resetMode', val)
  if (val === 'fixed') {
    if (props.resetHour == null) emit('update:resetHour', 0)
    if (props.dim === 'weekly' && props.resetDay == null) emit('update:resetDay', 1)
    if (!props.resetTimezone) emit('update:resetTimezone', 'UTC')
  }
}

function getTimezoneOffsetLabel(tz: string): string {
  try {
    const dtf = new Intl.DateTimeFormat('en-US', { timeZone: tz, timeZoneName: 'shortOffset' })
    const parts = dtf.formatToParts(new Date())
    const tzPart = parts.find(p => p.type === 'timeZoneName')
    return tzPart ? (tzPart.value === 'GMT' ? 'GMT+0' : tzPart.value) : ''
  } catch {
    return ''
  }
}
</script>

<template>
  <div>
    <!-- Title row (only when global notify is enabled) -->
    <div v-if="quotaNotifyGlobalEnabled" class="components-account-quota-dimension-row__panel">
      <span class="components-account-quota-dimension-row__text">{{ label }}</span>
      <span v-if="limit && limit > 0" class="components-account-quota-dimension-row__text">{{ t('admin.accounts.quotaNotify.alert') }}</span>
    </div>
    <label v-else class="components-account-quota-dimension-row__label">{{ label }}</label>

    <!-- Input row -->
    <div class="components-account-quota-dimension-row__panel-2">
      <div :class="['components-account-quota-dimension-row__panel-4', quotaNotifyGlobalEnabled ? 'components-account-quota-dimension-row__quota-notify-toggle' : 'components-account-quota-dimension-row__panel-5']">
        <span class="components-account-quota-dimension-row__text-2">$</span>
        <input :value="limit" @input="onLimitInput" type="number" min="0" step="0.01" class="components-account-quota-dimension-row__field input" :placeholder="t('admin.accounts.quotaLimitPlaceholder')" />
      </div>
      <QuotaNotifyToggle
        v-if="quotaNotifyGlobalEnabled && limit && limit > 0"
        class="components-account-quota-dimension-row__quota-notify-toggle"
        :enabled="notifyEnabled" :threshold="notifyThreshold" :threshold-type="notifyThresholdType"
        @update:enabled="emit('update:notifyEnabled', $event)" @update:threshold="emit('update:notifyThreshold', $event)" @update:threshold-type="emit('update:notifyThresholdType', $event)"
      />
    </div>

    <!-- Reset mode row (daily/weekly only) -->
    <div v-if="hasResetMode" class="components-account-quota-dimension-row__panel-3">
      <label class="components-account-quota-dimension-row__label-2">{{ t('admin.accounts.quotaResetMode') }}</label>
      <Select :model-value="resetMode || 'rolling'" @change="onModeChange" class="components-account-quota-dimension-row__field-2" :options="[
        { value: 'rolling', label: t('admin.accounts.quotaResetModeRolling') },
        { value: 'fixed', label: t('admin.accounts.quotaResetModeFixed') }
      ]" />
      <template v-if="resetMode === 'fixed'">
        <!-- Weekly: day of week selector -->
        <template v-if="dim === 'weekly'">
          <label class="components-account-quota-dimension-row__label-2">{{ t('admin.accounts.quotaWeeklyResetDay') }}</label>
          <Select :model-value="resetDay ?? 1" @change="emit('update:resetDay', Number($event))" class="components-account-quota-dimension-row__field-3" :options="dayOptions.map(d => ({ value: d.value, label: t('admin.accounts.dayOfWeek.' + d.key) }))" />
        </template>
        <label class="components-account-quota-dimension-row__label-2">{{ t('admin.accounts.quotaResetHour') }}</label>
        <Select :model-value="resetHour ?? 0" @change="emit('update:resetHour', Number($event))" class="components-account-quota-dimension-row__field-4" :options="hourOptions.map(h => ({ value: h, label: `${String(h).padStart(2, '0')}:00` }))" />
        <template v-if="timezoneOptions && timezoneOptions.length > 0">
          <Select :model-value="resetTimezone || 'UTC'" @change="emit('update:resetTimezone', String($event))" class="components-account-quota-dimension-row__field-2" :options="timezoneOptions.map(tz => ({ value: tz, label: `${tz} (${getTimezoneOffsetLabel(tz)})` }))" searchable />
        </template>
      </template>
      <span class="components-account-quota-dimension-row__text-3">
        <template v-if="resetMode === 'fixed'">{{ hintFixed }}</template>
        <template v-else>{{ hintRolling }}</template>
      </span>
    </div>

    <!-- Total dimension hint (no reset mode) -->
    <p v-if="!hasResetMode" class="components-account-quota-dimension-row__description input-hint">{{ hintRolling }}</p>
  </div>
</template>
