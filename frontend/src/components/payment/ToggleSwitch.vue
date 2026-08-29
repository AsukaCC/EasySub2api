<template>
  <label class="components-payment-toggle-switch__label">
    <span class="components-payment-toggle-switch__text">{{ label }}</span>
    <button
      type="button"
      role="switch"
      :aria-checked="checked"
      @click="emit('toggle')"
      :class="[
        'components-payment-toggle-switch__action',
        checked ? 'components-payment-toggle-switch__action-2' : 'components-payment-toggle-switch__action-3',
      ]"
    >
      <span :class="[
        'components-payment-toggle-switch__text-2',
        checked ? 'toggle-thumb--on' : 'components-payment-toggle-switch__text-3',
      ]" />
    </button>
  </label>
</template>

<script setup lang="ts">
defineProps<{ label: string; checked: boolean }>()
const emit = defineEmits<{ toggle: [] }>()
</script>
