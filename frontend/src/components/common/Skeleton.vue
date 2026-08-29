<template>
  <div
    :class="[
      'app-skeleton',
      variant === 'circle' ? 'app-skeleton--circle' : 'app-skeleton--rounded',
      customClass
    ]"
    :style="style"
  ></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'rect' | 'circle' | 'text'
  width?: string | number
  height?: string | number
  class?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'rect',
  width: '100%'
})

const customClass = computed(() => props.class || '')

const style = computed(() => {
  const s: Record<string, string> = {}
  
  if (props.width) {
    s.width = typeof props.width === 'number' ? `${props.width}px` : props.width
  }
  
  if (props.height) {
    s.height = typeof props.height === 'number' ? `${props.height}px` : props.height
  } else if (props.variant === 'text') {
    s.height = '1em'
    s.marginTop = '0.25em'
    s.marginBottom = '0.25em'
  }
  
  return s
})
</script>

<style scoped>
.app-skeleton {
  animation: skeleton-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  background: var(--glass-layer-inset-bg);
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.app-skeleton--circle { border-radius: 9999px; }
.app-skeleton--rounded { border-radius: var(--radius-md); }

@keyframes skeleton-pulse {
  50% { opacity: 0.5; }
}
</style>
