<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef, nextTick } from 'vue'

const props = withDefaults(defineProps<{
  content?: string
  trigger?: 'hover' | 'click'
  widthClass?: string
}>(), {
  trigger: 'hover',
  widthClass: 'w-64',
})

const show = ref(false)
const triggerRef = useTemplateRef<HTMLElement>('trigger')
const tooltipRef = useTemplateRef<HTMLElement>('tooltip')
const tooltipStyle = ref({ top: '0px', left: '0px' })

function openTooltip() {
  show.value = true
  nextTick(updatePosition)
}

function closeTooltip() {
  show.value = false
}

function onEnter() {
  if (props.trigger !== 'hover') return
  openTooltip()
}

function onLeave() {
  if (props.trigger !== 'hover') return
  closeTooltip()
}

function onFocusIn() {
  if (props.trigger !== 'hover') return
  openTooltip()
}

function onFocusOut() {
  if (props.trigger !== 'hover') return
  closeTooltip()
}

function onClick(event: MouseEvent) {
  if (props.trigger !== 'click') return
  event.stopPropagation()
  if (show.value) {
    closeTooltip()
    return
  }
  openTooltip()
}

function onDocumentClick(event: MouseEvent) {
  if (props.trigger !== 'click' || !show.value) return
  const target = event.target as Node | null
  if (!target) return
  if (triggerRef.value?.contains(target) || tooltipRef.value?.contains(target)) return
  closeTooltip()
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (props.trigger !== 'click') return
  if (event.key === 'Escape') {
    closeTooltip()
  }
}

function onViewportChange() {
  if (!show.value) return
  updatePosition()
}

function updatePosition() {
  const el = triggerRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  tooltipStyle.value = {
    top: `${rect.top}px`,
    left: `${rect.left + rect.width / 2}px`,
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick, true)
  document.addEventListener('keydown', onDocumentKeydown)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick, true)
  document.removeEventListener('keydown', onDocumentKeydown)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<template>
  <div
    ref="trigger"
    class="help-tooltip"
    tabindex="0"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @focusin="onFocusIn"
    @focusout="onFocusOut"
    @click="onClick"
  >
    <!-- Trigger Icon -->
    <slot name="trigger">
      <svg
        class="help-tooltip__icon"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    </slot>

    <!-- Teleport to body to escape modal overflow clipping -->
    <Teleport to="body">
      <div
        ref="tooltip"
        v-show="show"
        role="tooltip"
        :class="[
          'help-tooltip-popover glass-popover',
          props.widthClass,
        ]"
        :style="{ top: `calc(${tooltipStyle.top} - 8px)`, left: tooltipStyle.left }"
      >
        <button
          v-if="props.trigger === 'click'"
          type="button"
          class="help-tooltip-popover__close"
          aria-label="Close"
          @click.stop="closeTooltip"
        >
          <svg class="help-tooltip-popover__close-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <slot>{{ content }}</slot>
        <div class="help-tooltip-popover__arrow"></div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.help-tooltip {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin-left: 0.25rem;
  vertical-align: middle;
}

.help-tooltip__icon {
  width: 1rem;
  height: 1rem;
  color: var(--color-text-tertiary);
  cursor: help;
  transition: color 150ms ease, backdrop-filter 150ms ease;
}

.help-tooltip__icon:hover {
  color: var(--color-text-brand);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}
</style>

<style>
.help-tooltip-popover {
  position: fixed;
  z-index: 99999;
  padding: 0.75rem;
  border: 1px solid var(--glass-border-hover);
  border-radius: var(--radius-md);
  background: var(--glass-layer-floating-bg);
  color: var(--color-text-primary);
  font-size: var(--type-caption-size);
  line-height: var(--line-height-relaxed);
  transform: translate(-50%, -100%);
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
}

.help-tooltip-popover__close {
  position: absolute;
  top: 0.375rem;
  right: 0.375rem;
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  transition: color 150ms ease, backdrop-filter 150ms ease;
}

.help-tooltip-popover__close:hover {
  color: var(--color-text-primary);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.help-tooltip-popover__close-icon { width: 0.875rem; height: 0.875rem; }

.help-tooltip-popover__arrow {
  position: absolute;
  bottom: -0.25rem;
  left: 50%;
  width: 0.5rem;
  height: 0.5rem;
  border-right: 1px solid var(--glass-border-hover);
  border-bottom: 1px solid var(--glass-border-hover);
  background: var(--glass-layer-floating-bg);
  transform: translateX(-50%) rotate(45deg);
}
</style>
