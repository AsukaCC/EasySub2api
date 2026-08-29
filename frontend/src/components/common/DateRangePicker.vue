<template>
  <div class="date-range-picker" ref="containerRef">
    <button
      ref="triggerRef"
      type="button"
      @click="toggle"
      :class="['date-range-picker__trigger', isOpen && 'date-range-picker__trigger--open']"
    >
      <span class="date-range-picker__calendar">
        <Icon name="calendar" size="sm" />
      </span>
      <span class="date-range-picker__value">
        {{ displayValue }}
      </span>
      <span class="date-range-picker__chevron">
        <Icon
          name="chevronDown"
          size="sm"
          :class="['date-range-picker__chevron-icon', isOpen && 'date-range-picker__chevron-icon--open']"
        />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="date-picker-dropdown">
        <div
          v-if="isOpen"
          ref="panelRef"
          class="date-range-dropdown"
          :style="panelStyle"
          @click.stop
        >
        <!-- Quick presets -->
        <div class="date-range-dropdown__presets">
          <button
            v-for="preset in presets"
            :key="preset.value"
            @click="selectPreset(preset)"
            :class="['date-range-dropdown__preset', isPresetActive(preset) && 'date-range-dropdown__preset--active']"
          >
            {{ t(preset.labelKey) }}
          </button>
        </div>

        <div class="date-range-dropdown__divider"></div>

        <!-- Custom date range inputs -->
        <div class="date-range-dropdown__custom">
          <div class="date-range-dropdown__field">
            <label class="date-range-dropdown__label">{{ t('dates.startDate') }}</label>
            <input
              type="date"
              v-model="localStartDate"
              :max="localEndDate || tomorrow"
              class="date-range-dropdown__input"
              @change="onDateChange"
            />
          </div>
          <div class="date-range-dropdown__separator">
            <Icon name="arrowRight" size="sm" class="date-range-dropdown__arrow" />
          </div>
          <div class="date-range-dropdown__field">
            <label class="date-range-dropdown__label">{{ t('dates.endDate') }}</label>
            <input
              type="date"
              v-model="localEndDate"
              :min="localStartDate"
              :max="tomorrow"
              class="date-range-dropdown__input"
              @change="onDateChange"
            />
          </div>
        </div>

        <!-- Apply button -->
        <div class="date-range-dropdown__actions">
          <button @click="apply" class="date-range-dropdown__apply">
            {{ t('dates.apply') }}
          </button>
        </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useFloatingPanel } from '@/composables/useFloatingPanel'

interface DatePreset {
  labelKey: string
  value: string
  getRange: () => { start: string; end: string }
}

interface Props {
  startDate: string
  endDate: string
}

interface Emits {
  (e: 'update:startDate', value: string): void
  (e: 'update:endDate', value: string): void
  (e: 'change', range: { startDate: string; endDate: string; preset: string | null }): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t, locale } = useI18n()

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, isOpen, {
  maxWidth: 360,
  align: 'start',
  minComfortableHeight: 280
})
const localStartDate = ref(props.startDate)
const localEndDate = ref(props.endDate)
const activePreset = ref<string | null>('last24Hours')

const today = computed(() => {
  // Use local timezone to avoid UTC timezone issues
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
})

// Tomorrow's date - used for max date to handle timezone differences
// When user is in a timezone behind the server, "today" on server might be "tomorrow" locally
const tomorrow = computed(() => {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return formatDateToString(d)
})

// Helper function to format date to YYYY-MM-DD using local timezone
const formatDateToString = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const presets: DatePreset[] = [
  {
    labelKey: 'dates.today',
    value: 'today',
    getRange: () => {
      const t = today.value
      return { start: t, end: t }
    }
  },
  {
    labelKey: 'dates.yesterday',
    value: 'yesterday',
    getRange: () => {
      const d = new Date()
      d.setDate(d.getDate() - 1)
      const yesterday = formatDateToString(d)
      return { start: yesterday, end: yesterday }
    }
  },
  {
    labelKey: 'dates.last24Hours',
    value: 'last24Hours',
    getRange: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      return {
        start: formatDateToString(start),
        end: formatDateToString(end)
      }
    }
  },
  {
    labelKey: 'dates.last7Days',
    value: '7days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 6)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last14Days',
    value: '14days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 13)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last30Days',
    value: '30days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 29)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.thisMonth',
    value: 'thisMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 1))
      return { start, end: today.value }
    }
  },
  {
    labelKey: 'dates.lastMonth',
    value: 'lastMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth() - 1, 1))
      const end = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 0))
      return { start, end }
    }
  }
]

const displayValue = computed(() => {
  if (activePreset.value) {
    const preset = presets.find((p) => p.value === activePreset.value)
    if (preset) return t(preset.labelKey)
  }

  if (localStartDate.value && localEndDate.value) {
    if (localStartDate.value === localEndDate.value) {
      return formatDate(localStartDate.value)
    }
    return `${formatDate(localStartDate.value)} - ${formatDate(localEndDate.value)}`
  }

  return t('dates.selectDateRange')
})

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr + 'T00:00:00')
  const dateLocale = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  return date.toLocaleDateString(dateLocale, { month: 'short', day: 'numeric' })
}

const isPresetActive = (preset: DatePreset): boolean => {
  return activePreset.value === preset.value
}

const selectPreset = (preset: DatePreset) => {
  const range = preset.getRange()
  localStartDate.value = range.start
  localEndDate.value = range.end
  activePreset.value = preset.value
}

const onDateChange = () => {
  // Check if current dates match any preset
  activePreset.value = null
  for (const preset of presets) {
    const range = preset.getRange()
    if (range.start === localStartDate.value && range.end === localEndDate.value) {
      activePreset.value = preset.value
      break
    }
  }
}

const toggle = () => {
  isOpen.value = !isOpen.value
}

const apply = () => {
  emit('update:startDate', localStartDate.value)
  emit('update:endDate', localEndDate.value)
  emit('change', {
    startDate: localStartDate.value,
    endDate: localEndDate.value,
    preset: activePreset.value
  })
  isOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node
  if (containerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  isOpen.value = false
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isOpen.value) {
    isOpen.value = false
  }
}

// Sync local state with props
watch(
  () => props.startDate,
  (val) => {
    localStartDate.value = val
    onDateChange()
  }
)

watch(
  () => props.endDate,
  (val) => {
    localEndDate.value = val
    onDateChange()
  }
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
  // Initialize active preset detection
  onDateChange()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.date-range-picker {
  position: relative;
}

.date-range-picker__trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--glass-field-bg);
  color: var(--color-text-secondary);
  font-size: var(--type-control-size);
  cursor: pointer;
  transition: border-color 200ms ease, box-shadow 200ms ease;
}

.date-range-picker__trigger:hover {
  border-color: var(--glass-border-hover);
  background: var(--glass-field-bg-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
}

.date-range-picker__trigger:focus-visible {
  border-color: var(--color-primary);
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
}

.date-range-picker__trigger--open {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
}

.date-range-picker__calendar {
  color: var(--color-text-tertiary);
}

.date-range-picker__value {
  font-weight: 500;
}

.date-range-picker__chevron {
  color: var(--color-text-tertiary);
}

.date-range-picker__chevron-icon {
  transition: transform 200ms ease;
}

.date-range-picker__chevron-icon--open {
  transform: rotate(180deg);
}

.date-range-dropdown {
  position: absolute;
  left: 0;
  z-index: 100;
  min-width: 320px;
  margin-top: 0.5rem;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-floating-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
}

.date-range-dropdown__presets {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.25rem;
  padding: 0.5rem;
}

.date-range-dropdown__preset {
  padding: 0.375rem 0.75rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease;
}

.date-range-dropdown__preset:hover {
  color: var(--color-text-primary);
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.date-range-dropdown__preset--active {
  border-color: var(--color-primary-border);
  background-color: var(--glass-tint-brand);
  color: var(--color-text-brand);
  font-weight: var(--font-weight-semibold);
}

:global(.dark) .date-range-dropdown__preset--active {
  border-color: var(--color-primary-border);
  background-color: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.date-range-dropdown__divider {
  border-top: 1px solid var(--color-border-subtle);
}

.date-range-dropdown__custom {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  padding: 0.75rem;
}

.date-range-dropdown__field {
  flex: 1;
}

.date-range-dropdown__label {
  display: block;
  margin-bottom: 0.25rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.date-range-dropdown__input {
  width: 100%;
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-sm);
  background: var(--glass-field-bg);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  transition: border-color 150ms ease, box-shadow 150ms ease;
}

.date-range-dropdown__input:focus {
  border-color: var(--color-primary);
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.22), 0 1px 0 var(--glass-highlight) inset;
}

.date-range-dropdown__input::-webkit-calendar-picker-indicator {
  filter: invert(0.5);
  cursor: pointer;
  opacity: 0.6;
}

.date-range-dropdown__input::-webkit-calendar-picker-indicator:hover {
  opacity: 1;
}

:global(.dark) .date-range-dropdown__input::-webkit-calendar-picker-indicator {
  filter: none;
}

.date-range-dropdown__separator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-bottom: 0.25rem;
}

.date-range-dropdown__actions {
  display: flex;
  justify-content: flex-end;
  padding: 0 0.5rem 0.5rem;
}

.date-range-dropdown__apply {
  padding: 0.375rem 1rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-primary-border);
  background: var(--glass-tint-brand);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
  color: var(--color-text-brand);
  font-size: var(--type-control-size);
  font-weight: 500;
  transition: background-color 150ms ease;
}

.date-range-dropdown__apply:hover {
  background: var(--glass-tint-brand);
  -webkit-backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate));
}

/* Dropdown animation */
.date-picker-dropdown-enter-active,
.date-picker-dropdown-leave-active {
  transition: all 0.2s ease;
}

.date-picker-dropdown-enter-from,
.date-picker-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
