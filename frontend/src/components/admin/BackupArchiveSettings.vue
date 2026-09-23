<template>
  <section class="archive-settings">
    <div class="archive-settings__heading">
      <h4>{{ t('admin.backup.archive.title') }}</h4>
      <label><input type="checkbox" :checked="modelValue.enabled" @change="update({ enabled: ($event.target as HTMLInputElement).checked })" />{{ t('admin.backup.archive.enabled') }}</label>
    </div>
    <div v-if="modelValue.enabled" class="archive-settings__body">
      <fieldset>
        <legend>{{ t('admin.backup.archive.dates') }}</legend>
        <div class="archive-settings__days">
          <label v-for="day in 31" :key="day"><input type="checkbox" :checked="modelValue.days.includes(day)" :aria-label="t('admin.backup.archive.day', { day })" @change="toggleDay(day)" />{{ day }}</label>
        </div>
        <label><input type="checkbox" :checked="modelValue.include_month_end" @change="update({ include_month_end: ($event.target as HTMLInputElement).checked })" />{{ t('admin.backup.archive.monthEnd') }}</label>
      </fieldset>
      <div class="archive-settings__retention">
        <label><input type="checkbox" :checked="modelValue.retain_count === 0" @change="update({ retain_count: ($event.target as HTMLInputElement).checked ? 0 : 12 })" />{{ t('admin.backup.archive.forever') }}</label>
        <label v-if="modelValue.retain_count !== 0">{{ t('admin.backup.archive.count') }}<input :value="modelValue.retain_count" type="number" min="1" step="1" class="input" @input="update({ retain_count: ($event.target as HTMLInputElement).valueAsNumber || Number.NaN })" /></label>
      </div>
    </div>
  </section>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { BackupMonthlyArchiveConfig } from '@/api/admin/backup'
const props = defineProps<{ modelValue: BackupMonthlyArchiveConfig }>()
const emit = defineEmits<{ 'update:modelValue': [value: BackupMonthlyArchiveConfig] }>()
const { t } = useI18n()
function update(patch: Partial<BackupMonthlyArchiveConfig>) { emit('update:modelValue', { ...props.modelValue, ...patch }) }
function toggleDay(day: number) { update({ days: props.modelValue.days.includes(day) ? props.modelValue.days.filter(value => value !== day) : [...props.modelValue.days, day].sort((a, b) => a-b) }) }
</script>
<style scoped>
.archive-settings { border-top: 1px solid var(--border-color); padding-top: 16px; margin-top: 16px; }
.archive-settings__heading { display: flex; gap: 12px; flex-wrap: wrap; align-items: center; justify-content: space-between; }
.archive-settings h4 { font-size: 14px; margin: 0; }
.archive-settings label { display: flex; align-items: center; gap: 6px; font-size: 12px; }
.archive-settings__body { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 16px; margin-top: 12px; }
.archive-settings fieldset { border: 0; padding: 0; min-width: 0; }
.archive-settings legend { font-size: 12px; margin-bottom: 8px; }
.archive-settings__days { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 4px; margin-bottom: 12px; }
.archive-settings__days label { min-height: 32px; }
.archive-settings__retention { display: grid; align-content: start; gap: 12px; }
.archive-settings__retention label { flex-wrap: wrap; }
.archive-settings__retention .input { width: 120px; }
</style>
