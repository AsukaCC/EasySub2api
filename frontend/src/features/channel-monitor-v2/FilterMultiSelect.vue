<template>
  <div
    ref="containerRef"
    class="features-channel-monitor-v2-filter-multi-select__panel filter-menu"
    :class="compact ? 'features-channel-monitor-v2-filter-multi-select__panel-2' : 'features-channel-monitor-v2-filter-multi-select__panel-3'"
  >
    <button
      ref="triggerRef"
      type="button"
      class="features-channel-monitor-v2-filter-multi-select__action select-trigger"
      :class="[
        isOpen ? 'select-trigger-open' : '',
        compact ? 'features-channel-monitor-v2-filter-multi-select__action-4' : 'features-channel-monitor-v2-filter-multi-select__action-5',
      ]"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
      :aria-label="label"
      @click="toggleOpen"
    >
      <span
        class="features-channel-monitor-v2-filter-multi-select__text select-value"
        :class="compact ? 'features-channel-monitor-v2-filter-multi-select__text-6' : 'filter-multi-select__value--wide'"
      >
        {{ t('channelMonitorV2.filters.labelValue', { label, value: selectionLabel }) }}
      </span>
      <span class="features-channel-monitor-v2-filter-multi-select__text-2 select-icon" :class="isOpen ? 'features-channel-monitor-v2-filter-multi-select__text-7' : ''">
        <Icon name="chevronDown" size="sm" />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="select-dropdown">
        <div
          v-if="isOpen"
          ref="dropdownRef"
          class="select-dropdown-portal dropdown filter-dropdown"
          :class="[instanceId]"
          :style="dropdownStyle"
          role="listbox"
          aria-multiselectable="true"
          @click.stop
          @mousedown.stop
        >
          <button
            type="button"
            class="features-channel-monitor-v2-filter-multi-select__action-2 dropdown-item select-option select-option-group"
            @click="clear"
          >
            <span>{{ allLabel }}</span>
            <Icon v-if="modelValue.length === 0" name="check" size="sm" class="features-channel-monitor-v2-filter-multi-select__icon" />
          </button>

          <button
            v-for="option in options"
            :key="option.value"
            type="button"
            role="option"
            class="features-channel-monitor-v2-filter-multi-select__action-3 dropdown-item select-option"
            :class="modelValue.includes(option.value) ? 'select-option-selected' : ''"
            :aria-selected="modelValue.includes(option.value)"
            @click="toggle(option.value)"
          >
            <span class="features-channel-monitor-v2-filter-multi-select__text-3">
              <span
                class="features-channel-monitor-v2-filter-multi-select__text-4 checkbox"
                :class="modelValue.includes(option.value) ? 'features-channel-monitor-v2-filter-multi-select__text-8' : ''"
              >
                <Icon v-if="modelValue.includes(option.value)" name="check" size="sm" class="features-channel-monitor-v2-filter-multi-select__icon" />
              </span>
              <span class="features-channel-monitor-v2-filter-multi-select__text-5">{{ option.label }}</span>
            </span>
            <small v-if="option.count != null" class="features-channel-monitor-v2-filter-multi-select__small">{{ formatCount(option.count) }}</small>
          </button>
          <p v-if="options.length === 0" class="features-channel-monitor-v2-filter-multi-select__description">{{ t('channelMonitorV2.filters.empty') }}</p>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { CSSProperties } from 'vue'
import { useI18n } from 'vue-i18n'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { monitorIntlLocale } from '@/features/channel-monitor-v2/monitorFormat'

interface FilterOption {
  value: string
  label: string
  count?: number
}

const props = withDefaults(
  defineProps<{
    label: string
    allLabel: string
    modelValue: string[]
    /** Options for this picker (parent may cascade by platform). */
    options: FilterOption[]
    /** Compact trigger for single-row toolbars. */
    compact?: boolean
  }>(),
  { compact: false },
)
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()
const { t, locale } = useI18n()

const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const instanceId = `filter-select-${Math.random().toString(36).slice(2, 9)}`

const selectionLabel = computed(() => {
  if (props.modelValue.length === 0) return props.allLabel
  if (props.modelValue.length === 1) {
    return props.options.find((item) => item.value === props.modelValue[0])?.label || props.modelValue[0]
  }
  return t('channelMonitorV2.filters.selectedCount', { count: props.modelValue.length })
})

const dropdownStyle = computed<CSSProperties>(() => {
  const trigger = triggerRef.value
  if (!trigger) return {}
  const rect = trigger.getBoundingClientRect()
  const padding = 8
  const viewportRight = Math.max(padding, window.innerWidth - padding)
  const left = Math.min(Math.max(padding, rect.left), viewportRight)
  const availableWidth = Math.max(0, viewportRight - left)
  const preferredMinWidth = Math.max(200, rect.width)
  const minWidth = Math.min(preferredMinWidth, availableWidth)
  return {
    position: 'fixed' as const,
    left: `${left}px`,
    top: `${rect.bottom + 4}px`,
    minWidth: `${minWidth}px`,
    maxWidth: `${availableWidth}px`,
    zIndex: '100000020',
  }
})

function clear() {
  emit('update:modelValue', [])
  // Stay open so users can re-select without reopening.
}

function toggle(value: string) {
  const selected = new Set(props.modelValue)
  if (selected.has(value)) selected.delete(value)
  else selected.add(value)
  emit('update:modelValue', [...selected])
  // Stay open on toggle (multi-select).
}

function toggleOpen() {
  isOpen.value ? close() : open()
}

function open() {
  isOpen.value = true
  void nextTick(() => positionDropdown())
}

function close() {
  isOpen.value = false
}

function positionDropdown() {
  const trigger = triggerRef.value
  const dropdown = dropdownRef.value
  if (!trigger || !dropdown) return
  const rect = trigger.getBoundingClientRect()
  const padding = 8
  const viewportRight = Math.max(padding, window.innerWidth - padding)
  const left = Math.min(Math.max(padding, rect.left), viewportRight)
  const availableWidth = Math.max(0, viewportRight - left)
  const preferredMinWidth = Math.max(200, rect.width)
  const minWidth = Math.min(preferredMinWidth, availableWidth)
  dropdown.style.left = `${left}px`
  dropdown.style.top = `${rect.bottom + 4}px`
  dropdown.style.minWidth = `${minWidth}px`
  dropdown.style.maxWidth = `${availableWidth}px`
}

function formatCount(value: number) {
  return Intl.NumberFormat(locale.value || monitorIntlLocale(), {
    notation: value >= 10000 ? 'compact' : 'standard',
  }).format(value)
}

function onDocumentMouseDown(event: MouseEvent) {
  if (!isOpen.value) return
  const target = event.target as Node | null
  if (!target) return
  if (containerRef.value?.contains(target)) return
  if (dropdownRef.value?.contains(target)) return
  close()
}

function onDocumentKeyDown(event: KeyboardEvent) {
  if (event.key === 'Escape') close()
}

function onWindowChange() {
  if (isOpen.value) positionDropdown()
}

watch(isOpen, async (open) => {
  if (open) {
    await nextTick()
    positionDropdown()
    document.addEventListener('mousedown', onDocumentMouseDown)
    document.addEventListener('keydown', onDocumentKeyDown)
    window.addEventListener('resize', onWindowChange)
    window.addEventListener('scroll', onWindowChange, true)
  } else {
    document.removeEventListener('mousedown', onDocumentMouseDown)
    document.removeEventListener('keydown', onDocumentKeyDown)
    window.removeEventListener('resize', onWindowChange)
    window.removeEventListener('scroll', onWindowChange, true)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentMouseDown)
  document.removeEventListener('keydown', onDocumentKeyDown)
  window.removeEventListener('resize', onWindowChange)
  window.removeEventListener('scroll', onWindowChange, true)
})
</script>

<style scoped>
.select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  width: 100%;
  padding: 0.625rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--glass-field-bg);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: border-color 200ms ease, box-shadow 200ms ease;
}

.select-trigger:hover {
  border-color: var(--color-border-strong);
}

.select-trigger-open {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
}

.filter-menu summary::-webkit-details-marker {
  display: none;
}

.filter-dropdown {
  width: max-content;
  min-width: 200px;
  max-height: min(50vh, 360px);
  overflow-y: auto;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-floating-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
}

.dropdown-item {
  cursor: pointer;
}

.checkbox {
  flex: none;
}

@media (max-width: 640px) {
  .filter-menu {
    min-width: 100%;
  }
}
</style>
