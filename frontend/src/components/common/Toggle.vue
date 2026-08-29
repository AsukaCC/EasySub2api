<template>
  <button
    type="button"
    @click="toggle"
    class="app-toggle"
    :class="modelValue ? 'app-toggle--checked' : 'app-toggle--unchecked'"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
  >
    <span
      class="app-toggle__thumb"
      :class="modelValue ? 'app-toggle__thumb--checked' : 'app-toggle__thumb--unchecked'"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>

<style scoped>
.app-toggle {
  position: relative;
  display: inline-flex;
  width: 2.75rem;
  height: 1.5rem;
  flex-shrink: 0;
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: 9999px;
  transition: border-color 200ms ease, box-shadow 200ms ease, backdrop-filter 200ms ease;
  backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
}

.app-toggle--checked {
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
}

.app-toggle--unchecked {
  border-color: var(--glass-border);
  background: var(--glass-layer-inset-bg);
}

.app-toggle:hover {
  border-color: var(--glass-border-hover);
  backdrop-filter: blur(var(--glass-blur-thin-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.app-toggle:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.app-toggle__thumb {
  display: inline-block;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 9999px;
  background: var(--glass-layer-floating-bg);
  box-shadow: var(--shadow-sm);
  pointer-events: none;
  transition: transform 200ms ease, backdrop-filter 200ms ease;
  backdrop-filter: blur(var(--glass-blur-thin)) saturate(var(--glass-saturate));
}

.app-toggle__thumb--checked {
  transform: translateX(1.25rem);
}

.app-toggle__thumb--unchecked {
  transform: translateX(0);
}

.app-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
