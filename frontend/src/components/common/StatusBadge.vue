<template>
  <div class="status-badge">
    <span
      :class="[
        'status-badge__indicator',
        variantClass
      ]"
    ></span>
    <span class="status-badge__label">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  label: string
}>()

const variantClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'status-fill--success'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'status-fill--warning'
    case 'error':
    case 'danger':
      return 'status-fill--danger'
    default:
      return 'status-fill--neutral'
  }
})
</script>

<style scoped>
.status-badge {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.status-badge__indicator {
  display: inline-block;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 9999px;
}

.status-badge__label {
  color: var(--color-text-primary);
  font-size: var(--type-control-size);
  line-height: var(--type-control-line-height);
}
</style>
