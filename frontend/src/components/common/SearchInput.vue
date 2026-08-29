<template>
  <div class="search-input">
    <div class="search-input__leading-icon">
      <Icon name="search" size="md" class="search-input__icon" />
    </div>
    <input
      :value="modelValue"
      type="text"
      class="search-input__field input"
      :placeholder="placeholder"
      @input="handleInput"
    />
  </div>
</template>

<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  debounceMs?: number
}>(), {
  placeholder: 'Search...',
  debounceMs: 300
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'search', value: string): void
}>()

const debouncedEmitSearch = useDebounceFn((value: string) => {
  emit('search', value)
}, props.debounceMs)

const handleInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  emit('update:modelValue', value)
  debouncedEmitSearch(value)
}
</script>

<style scoped>
.search-input {
  position: relative;
  width: 100%;
}

.search-input__leading-icon {
  position: absolute;
  inset-block: 0;
  left: 0;
  display: flex;
  align-items: center;
  padding-left: 0.75rem;
  pointer-events: none;
}

.search-input__icon {
  color: var(--color-text-tertiary);
}

.search-input__field {
  padding-left: 2.5rem;
}
</style>
