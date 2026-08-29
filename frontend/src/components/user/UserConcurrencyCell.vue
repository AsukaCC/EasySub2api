<template>
  <div class="components-user-user-concurrency-cell__panel">
    <span
      :class="[
        'components-user-user-concurrency-cell__text-3',
        statusClass
      ]"
    >
      <!-- Four-square grid icon -->
      <svg class="components-user-user-concurrency-cell__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
      </svg>
      <span class="components-user-user-concurrency-cell__text">{{ current }}</span>
      <span class="components-user-user-concurrency-cell__text-2">/</span>
      <span class="components-user-user-concurrency-cell__text">{{ max }}</span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  current: number
  max: number
}>()

// Status color based on usage
const statusClass = computed(() => {
  const { current, max } = props

  // Full: red
  if (current >= max && max > 0) {
    return 'components-user-user-concurrency-cell__state'
  }
  // In use: yellow
  if (current > 0) {
    return 'components-user-user-concurrency-cell__state-2'
  }
  // Idle: gray
  return 'components-user-user-concurrency-cell__state-3'
})
</script>
