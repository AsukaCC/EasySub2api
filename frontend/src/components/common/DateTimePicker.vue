<template>
  <div ref="rootRef" class="date-time-picker" :class="$attrs.class" :style="$attrs.style as StyleValue">
    <input
      v-bind="inputAttrs"
      ref="inputRef"
      type="text"
      class="date-time-picker__input"
      :value="text"
      :placeholder="placeholder || displayFormat"
      :disabled="disabled"
      :readonly="readonly"
      :required="required"
      :aria-label="$attrs['aria-label'] as string || t(type === 'time' ? 'dates.selectTime' : 'dates.selectDate')"
      :aria-invalid="invalid || undefined"
      autocomplete="off"
      @input="onTextInput"
      @change="onTextChange"
      @keydown.down.prevent="openPanel"
      @keydown.esc.stop.prevent="closePanel"
    />
    <div class="date-time-picker__tools">
      <button v-if="clearable && text && !disabled && !readonly" type="button" :title="t('dates.clear')" :aria-label="t('dates.clear')" @click="clear">
        <Icon name="x" size="xs" />
      </button>
      <button ref="triggerRef" type="button" :disabled="disabled || readonly" :title="t(type === 'time' ? 'dates.selectTime' : 'dates.selectDate')" :aria-label="t(type === 'time' ? 'dates.selectTime' : 'dates.selectDate')" aria-haspopup="dialog" :aria-expanded="open" :aria-controls="panelId" @click="open ? closePanel() : openPanel()">
        <Icon :name="type === 'time' ? 'clock' : 'calendar'" size="sm" />
      </button>
    </div>
    <Teleport to="body">
      <div v-if="open" :id="panelId" ref="panelRef" class="date-time-panel" data-date-picker-panel :style="panelStyle" role="dialog" :aria-label="t(type === 'time' ? 'dates.selectTime' : 'dates.selectDate')" @click.stop @keydown.esc.stop.prevent="closePanel">
        <template v-if="type !== 'time'">
          <header class="date-time-panel__header">
            <button type="button" :title="t('dates.previousMonth')" :aria-label="t('dates.previousMonth')" :disabled="cursor.year() === 1 && cursor.month() === 0" @click="moveMonth(-1)"><Icon name="chevronLeft" size="sm" /></button>
            <input type="number" min="1" max="9999" :value="cursor.year()" :aria-label="t('dates.year')" @change="setYear(($event.target as HTMLInputElement).value)" />
            <select :value="cursor.month()" :aria-label="t('dates.month')" @change="cursor = cursor.month(Number(($event.target as HTMLSelectElement).value))">
              <option v-for="(month, index) in months" :key="index" :value="index">{{ month }}</option>
            </select>
            <button type="button" :title="t('dates.nextMonth')" :aria-label="t('dates.nextMonth')" :disabled="cursor.year() === 9999 && cursor.month() === 11" @click="moveMonth(1)"><Icon name="chevronRight" size="sm" /></button>
          </header>
          <div class="date-time-panel__week" aria-hidden="true"><span v-for="(weekday, index) in weekdays" :key="index">{{ weekday }}</span></div>
          <div class="date-time-panel__calendar" role="grid" :aria-label="cursor.format('YYYY-MM')">
            <div v-for="week in 6" :key="week" role="row">
              <button v-for="day in days.slice((week - 1) * 7, week * 7)" :key="day.format('YYYY-MM-DD')" type="button" role="gridcell"
                :data-date="day.format('YYYY-MM-DD')" :aria-label="day.format('YYYY-MM-DD')"
                :aria-selected="day.isSame(selected, 'day')" :aria-current="day.isSame(dayjs(), 'day') ? 'date' : undefined"
                :tabindex="day.isSame(focusDay, 'day') ? 0 : -1" :disabled="!dayAllowed(day)"
                :class="{ 'is-outside': day.month() !== cursor.month(), 'is-selected': day.isSame(selected, 'day'), 'is-today': day.isSame(dayjs(), 'day') }"
                @click="selectDay(day)" @keydown="onDayKey($event, day)">{{ day.date() }}</button>
            </div>
          </div>
        </template>
        <div v-if="type !== 'date'" class="date-time-panel__time">
          <Icon name="clock" size="sm" />
          <label><span>{{ t('dates.hour') }}</span><input v-model="hour" type="number" min="0" max="23" :aria-label="t('dates.hour')" /></label>
          <span>:</span>
          <label><span>{{ t('dates.minute') }}</span><input v-model="minute" type="number" min="0" max="59" :aria-label="t('dates.minute')" /></label>
          <template v-if="seconds"><span>:</span><label><span>{{ t('dates.second') }}</span><input v-model="second" type="number" min="0" max="59" :aria-label="t('dates.second')" /></label></template>
        </div>
        <footer class="date-time-panel__footer">
          <button type="button" @click="selectNow">{{ t(type === 'date' ? 'dates.today' : 'dates.now') }}</button>
          <button type="button" class="date-time-panel__apply" :disabled="!draftValid" @click="apply">{{ t('dates.apply') }}</button>
        </footer>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, onBeforeUnmount, onMounted, ref, useAttrs, watch, type StyleValue } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { dayjs, dateLocale, parsePickerValue, pickerFormat, type Dayjs, type PickerType } from '@/utils/datetime'

defineOptions({ inheritAttrs: false })
const props = withDefaults(defineProps<{
  modelValue?: string | number | null
  type?: PickerType
  min?: string
  max?: string
  step?: string | number
  disabled?: boolean
  readonly?: boolean
  required?: boolean
  clearable?: boolean
  placeholder?: string
}>(), { modelValue: '', type: 'datetime-local', step: 60, clearable: true })
const emit = defineEmits<{ 'update:modelValue': [value: string]; change: [value: string] }>()
const { t, locale } = useI18n()
const attrs = useAttrs()
const inputAttrs = computed(() => Object.fromEntries(Object.entries(attrs).filter(([key]) => key !== 'class' && key !== 'style')))
const rootRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const open = ref(false)
const panelId = `date-time-panel-${getCurrentInstance()?.uid}`
const { panelRef, style: panelStyle } = useFloatingPanel(rootRef, open, { maxWidth: 320, align: 'start', minComfortableHeight: 380, zIndex: 1100 })
const seconds = computed(() => Number(props.step) < 60 || props.step === 'any')
const valueFormat = computed(() => pickerFormat(props.type, seconds.value))
const displayFormat = computed(() => valueFormat.value.replace('T', ' '))
const text = ref('')
const invalid = ref(false)
let pendingInputValue: string | undefined
const selected = ref(dayjs().startOf('day'))
const cursor = ref(dayjs().startOf('month'))
const hour = ref<string | number>(0)
const minute = ref<string | number>(0)
const second = ref<string | number>(0)
const months = computed(() => Array.from({ length: 12 }, (_, month) => dayjs().month(month).locale(dateLocale(locale.value)).format('MMM')))
const weekdays = computed(() => Array.from({ length: 7 }, (_, index) => dayjs().day(index + 1).locale(dateLocale(locale.value)).format('dd')))
const days = computed(() => {
  const first = cursor.value.startOf('month')
  const start = first.subtract((first.day() + 6) % 7, 'day')
  return Array.from({ length: 42 }, (_, index) => start.add(index, 'day'))
})
const focusDay = computed(() => selected.value.isSame(cursor.value, 'month') && dayAllowed(selected.value)
  ? selected.value : days.value.find(day => day.month() === cursor.value.month() && dayAllowed(day)))
const draft = computed(() => selected.value.hour(Number(hour.value)).minute(Number(minute.value)).second(seconds.value ? Number(second.value) : 0).millisecond(0))
const draftValid = computed(() => {
  const parts = props.type === 'date' ? [] : [[hour.value, 23], [minute.value, 59], ...(seconds.value ? [[second.value, 59]] : [])]
  return parts.every(([value, max]) => value !== '' && Number.isInteger(Number(value)) && Number(value) >= 0 && Number(value) <= Number(max)) && valueAllowed(draft.value.format(valueFormat.value))
})

function valueAllowed(value: string): boolean {
  const date = parsePickerValue(value, props.type)
  if (!date) return false
  const min = props.min && parsePickerValue(props.min, props.type)
  const max = props.max && parsePickerValue(props.max, props.type)
  if ((min && date.isBefore(min)) || (max && date.isAfter(max))) return false
  if (props.type !== 'date' && props.step !== 'any') {
    const step = Number(props.step)
    const base = min || date.startOf('day')
    if (step > 0 && date.diff(base, 'second') % step !== 0) return false
  }
  return true
}
function dayAllowed(day: Dayjs): boolean {
  if (day.year() < 1 || day.year() > 9999) return false
  const min = props.min && parsePickerValue(props.min, props.type)
  const max = props.max && parsePickerValue(props.max, props.type)
  return !(min && day.isBefore(min, 'day')) && !(max && day.isAfter(max, 'day'))
}
function syncDraft(date: Dayjs) {
  selected.value = date.startOf('day')
  cursor.value = date.startOf('month')
  hour.value = date.format('HH')
  minute.value = date.format('mm')
  second.value = date.format('ss')
}
watch(() => props.modelValue, value => {
  if (pendingInputValue !== undefined && value === pendingInputValue) {
    pendingInputValue = undefined
    return
  }
  pendingInputValue = undefined
  text.value = String(value ?? '').replace('T', ' ')
  validateText()
}, { immediate: true })
watch([() => props.min, () => props.max, () => props.step], validateText)
watch([() => props.disabled, () => props.readonly], ([disabled, readonly]) => { if (disabled || readonly) open.value = false })
function normalizedText(): string {
  const value = text.value.trim().replace(/：/g, ':')
  if (props.type === 'time' && /^24:00(?::00)?$/.test(value)) return value.replace('24:', '00:')
  return props.type === 'datetime-local' ? value.replace(' ', 'T') : value
}
function validateText() {
  invalid.value = !!text.value && !valueAllowed(normalizedText())
  inputRef.value?.setCustomValidity(invalid.value ? t('dates.invalidDate') : '')
}
function onTextInput(event: Event) {
  text.value = (event.target as HTMLInputElement).value
  validateText()
  pendingInputValue = invalid.value ? '' : normalizedText()
  emit('update:modelValue', pendingInputValue)
}
function onTextChange() {
  validateText()
  if (!invalid.value) emit('change', normalizedText())
}
async function openPanel() {
  if (props.disabled || props.readonly) return
  syncDraft(parsePickerValue(normalizedText(), props.type) || dayjs().second(0))
  open.value = true
  await nextTick()
  const focus = panelRef.value?.querySelector<HTMLElement>('[role="gridcell"][tabindex="0"]:not(:disabled)')
    || panelRef.value?.querySelector<HTMLElement>('input:not(:disabled)')
  focus?.focus()
}
function closePanel() { open.value = false; triggerRef.value?.focus() }
function commit(value: string) {
  text.value = value.replace('T', ' ')
  validateText()
  emit('update:modelValue', value)
  emit('change', value)
}
function clear() { commit(''); open.value = false; inputRef.value?.focus() }
function apply() { if (draftValid.value) { commit(draft.value.format(valueFormat.value)); closePanel() } }
function selectDay(day: Dayjs) {
  selected.value = day
  cursor.value = day.startOf('month')
  if (props.type === 'date') apply()
}
function selectNow() { syncDraft(dayjs().second(0)); if (props.type === 'date') apply() }
function moveMonth(amount: number) { cursor.value = cursor.value.add(amount, 'month') }
function setYear(value: string) { const year = Number(value); if (Number.isInteger(year) && year >= 1 && year <= 9999) cursor.value = cursor.value.year(year) }
async function onDayKey(event: KeyboardEvent, day: Dayjs) {
  const offset: Record<string, number> = { ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7, Home: -((day.day() + 6) % 7), End: 6 - ((day.day() + 6) % 7) }
  let next: Dayjs
  if (event.key in offset) next = day.add(offset[event.key], 'day')
  else if (event.key === 'PageUp' || event.key === 'PageDown') next = day.add(event.key === 'PageUp' ? -1 : 1, 'month')
  else return
  event.preventDefault()
  if (!dayAllowed(next)) return
  selected.value = next
  cursor.value = next.startOf('month')
  await nextTick()
  panelRef.value?.querySelector<HTMLElement>(`[data-date="${next.format('YYYY-MM-DD')}"]`)?.focus()
}
function onOutside(event: Event) {
  if (rootRef.value?.contains(event.target as Node) || panelRef.value?.contains(event.target as Node)) return
  open.value = false
}
onMounted(() => { validateText(); document.addEventListener('pointerdown', onOutside); document.addEventListener('focusin', onOutside) })
onBeforeUnmount(() => { document.removeEventListener('pointerdown', onOutside); document.removeEventListener('focusin', onOutside) })
defineExpose({ focus: () => inputRef.value?.focus(), checkValidity: () => inputRef.value?.checkValidity() ?? false })
</script>

<style scoped>
.date-time-picker { position: relative; display: flex; min-width: 0; width: 100%; padding: 0 !important; border: 0 !important; background: transparent !important; box-shadow: none !important; }
.date-time-picker__input { width: 100%; min-width: 0; height: 2.5rem; padding: 0.5rem 4.25rem 0.5rem 0.75rem; border: 1px solid var(--color-border); border-radius: var(--radius-md); background: var(--glass-field-bg); color: var(--color-text-primary); font-size: var(--type-control-size); font-variant-numeric: tabular-nums; }
.date-time-picker__input:focus { outline: 2px solid var(--color-primary-border); outline-offset: 1px; border-color: var(--color-primary); }
.date-time-picker__input[aria-invalid='true'] { border-color: var(--color-text-danger); }
.date-time-picker__input:disabled { opacity: 0.5; cursor: not-allowed; }
.date-time-picker__tools { position: absolute; right: 0.375rem; inset-block: 0; display: flex; align-items: center; }
.date-time-picker button, .date-time-panel button { display: inline-flex; align-items: center; justify-content: center; color: var(--color-text-secondary); border-radius: var(--radius-sm); cursor: pointer; }
.date-time-picker__tools button { width: 1.75rem; height: 1.75rem; }
.date-time-picker button:hover, .date-time-panel button:hover:not(:disabled) { background: var(--glass-bg-interactive-hover); color: var(--color-text-primary); }
.date-time-picker button:focus-visible, .date-time-panel button:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }
.date-time-panel { overflow: auto; padding: 0.75rem; border: 1px solid var(--color-border); border-radius: 8px; background: var(--glass-layer-floating-bg); backdrop-filter: blur(var(--glass-layer-floating-blur)); box-shadow: var(--glass-shadow-hover); color: var(--color-text-primary); font-size: var(--font-size-sm); }
.date-time-panel__header { display: grid; grid-template-columns: 2rem minmax(0, 1fr) minmax(0, 1fr) 2rem; align-items: center; gap: 0.5rem; margin-bottom: 0.75rem; }
.date-time-panel__header button { width: 2rem; height: 2rem; }
.date-time-panel input, .date-time-panel select { min-width: 0; width: 100%; height: 2rem; padding: 0.25rem; border: 1px solid var(--color-border); border-radius: var(--radius-sm); color: var(--color-text-primary); background: var(--glass-field-bg); font-size: var(--font-size-sm); }
.date-time-panel input:focus, .date-time-panel select:focus { outline: 2px solid var(--color-primary-border); outline-offset: 1px; border-color: var(--color-primary); }
.date-time-panel__week, .date-time-panel__calendar > div { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 0.125rem; }
.date-time-panel__week { text-align: center; color: var(--color-text-tertiary); font-size: var(--font-size-xs); margin-bottom: 0.25rem; }
.date-time-panel__calendar button { width: 100%; height: 2rem; font-variant-numeric: tabular-nums; }
.date-time-panel__calendar .is-outside { color: var(--color-text-tertiary); }
.date-time-panel__calendar .is-today { box-shadow: inset 0 0 0 1px var(--color-primary-border); }
.date-time-panel__calendar .is-selected { color: var(--color-text-brand); background: var(--glass-tint-brand); font-weight: 600; }
.date-time-panel button:disabled { opacity: 0.35; cursor: not-allowed; }
.date-time-panel__time { display: flex; align-items: flex-end; justify-content: center; gap: 0.5rem; padding-top: 0.75rem; }
.date-time-panel__time > .app-icon { align-self: center; }
.date-time-panel__time label { width: 3.5rem; }
.date-time-panel__time label span { display: block; margin-bottom: 0.25rem; color: var(--color-text-tertiary); font-size: var(--font-size-xs); }
.date-time-panel__time input { text-align: center; }
.date-time-panel__footer { display: flex; justify-content: space-between; margin-top: 0.75rem; padding-top: 0.5rem; border-top: 1px solid var(--color-border-subtle); }
.date-time-panel__footer button { min-height: 2rem; padding: 0.25rem 0.75rem; }
.date-time-panel__footer .date-time-panel__apply { background: var(--glass-tint-brand); color: var(--color-text-brand); }
</style>
