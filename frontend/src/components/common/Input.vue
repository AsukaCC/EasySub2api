<template>
  <div class="components-common-input__panel">
    <label v-if="label" :for="id" class="components-common-input__label input-label">
      {{ label }}
      <span v-if="required" class="components-common-input__text">*</span>
    </label>
    <div class="components-common-input__panel-2">
      <!-- Prefix Icon Slot -->
      <div
        v-if="$slots.prefix"
        class="components-common-input__panel-3"
      >
        <slot name="prefix"></slot>
      </div>

      <input
        :id="id"
        ref="inputRef"
        :type="type"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        :placeholder="placeholderText"
        :autocomplete="autocomplete"
        :readonly="readonly"
        :class="[
          'components-common-input__field input',
          $slots.prefix ? 'components-common-input__field-2' : '',
          $slots.suffix ? 'components-common-input__field-3' : '',
          error ? 'components-common-input__field-4 input-error' : '',
          disabled ? 'components-common-input__field-5' : ''
        ]"
        @input="onInput"
        @change="$emit('change', ($event.target as HTMLInputElement).value)"
        @blur="$emit('blur', $event)"
        @focus="$emit('focus', $event)"
        @keyup.enter="$emit('enter', $event)"
      />

      <!-- Suffix Slot (e.g. Password Toggle or Clear Button) -->
      <div
        v-if="$slots.suffix"
        class="components-common-input__panel-4"
      >
        <slot name="suffix"></slot>
      </div>
    </div>
    <!-- Hint / Error Text -->
    <p v-if="error" class="components-common-input__description input-error-text">
      {{ error }}
    </p>
    <p v-else-if="hint" class="components-common-input__description input-hint">
      {{ hint }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  modelValue: string | number | null | undefined
  type?: string
  label?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
  readonly?: boolean
  error?: string
  hint?: string
  id?: string
  autocomplete?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  required: false,
  readonly: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'blur', event: FocusEvent): void
  (e: 'focus', event: FocusEvent): void
  (e: 'enter', event: KeyboardEvent): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const placeholderText = computed(() => props.placeholder || '')

const onInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  emit('update:modelValue', value)
}

// Expose focus method
defineExpose({
  focus: () => inputRef.value?.focus(),
  select: () => inputRef.value?.select()
})
</script>
