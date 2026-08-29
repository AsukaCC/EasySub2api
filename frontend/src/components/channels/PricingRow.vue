<template>
  <div class="pricing-row">
    <span class="pricing-row__label">{{ label }}</span>
    <span class="pricing-row__value">{{ display }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatScaled } from '@/utils/pricing'

const props = withDefaults(
  defineProps<{
    label: string
    value: number | null
    unit: string
    scale: number
  }>(),
  { value: null }
)

const display = computed(() =>
  props.value == null ? '-' : `${formatScaled(props.value, props.scale)} ${props.unit}`
)
</script>

<style scoped>
.pricing-row {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}

.pricing-row__label {
  color: var(--color-text-secondary);
}

.pricing-row__value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
</style>
