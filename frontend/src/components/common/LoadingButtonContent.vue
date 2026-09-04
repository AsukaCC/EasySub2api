<template>
  <span class="loading-button-content">
    <span
      class="loading-button-content__layer loading-button-content__normal"
      :class="{ 'loading-button-content__layer--hidden': loading }"
      :aria-hidden="loading ? 'true' : undefined"
    >
      <slot />
    </span>
    <span
      class="loading-button-content__layer loading-button-content__loading"
      :class="{ 'loading-button-content__layer--hidden': !loading }"
      :aria-hidden="loading ? undefined : 'true'"
    >
      <LoadingSpinner size="xs" color="inherit" decorative />
      <span>{{ loadingText }}</span>
    </span>
  </span>
</template>

<script setup lang="ts">
import LoadingSpinner from './LoadingSpinner.vue'

interface Props {
  loading: boolean
  loadingText: string
}

defineProps<Props>()
</script>

<style scoped>
.loading-button-content {
  display: inline-grid;
  align-items: center;
  justify-items: center;
}

.loading-button-content__layer {
  grid-area: 1 / 1;
}

.loading-button-content__normal,
.loading-button-content__loading {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  white-space: nowrap;
}

.loading-button-content__layer--hidden {
  visibility: hidden;
}
</style>
