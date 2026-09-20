<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useNavigationLoadingState } from '@/composables/useNavigationLoading'

const { isLoading } = useNavigationLoadingState()
const { t } = useI18n()
const visible = ref(false)
const progress = ref(0)
let trickleTimer: ReturnType<typeof setInterval> | undefined
let finishTimer: ReturnType<typeof setTimeout> | undefined
let hideTimer: ReturnType<typeof setTimeout> | undefined

function clearTimers() {
  clearInterval(trickleTimer)
  clearTimeout(finishTimer)
  clearTimeout(hideTimer)
}

watch(isLoading, (loading) => {
  clearTimers()
  if (loading) {
    if (!visible.value || progress.value === 1) progress.value = 0.12
    visible.value = true
    trickleTimer = setInterval(() => {
      progress.value += (0.9 - progress.value) * 0.08
    }, 200)
  } else if (visible.value) {
    // Allow route resolution to hand off to the new page's data loading.
    finishTimer = setTimeout(() => {
      progress.value = 1
      hideTimer = setTimeout(() => { visible.value = false }, 220)
    }, 150)
  }
}, { immediate: true })

onBeforeUnmount(clearTimers)
</script>

<template>
  <Transition name="progress-fade">
    <div
      v-show="visible"
      class="navigation-progress"
      role="progressbar"
      :aria-label="t('common.loading')"
    >
      <div class="navigation-progress-bar" :style="{ transform: `scaleX(${progress})` }" />
    </div>
  </Transition>
</template>

<style scoped>
.navigation-progress {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  z-index: var(--z-toast);
  pointer-events: none;
  overflow: hidden;
  background: transparent;
}

.navigation-progress-bar {
  height: 100%;
  width: 100%;
  background: var(--color-primary);
  box-shadow: 0 0 8px var(--color-primary-border);
  transform-origin: left;
  transition: transform 200ms ease-out;
}

/* 淡入淡出过渡 */
.progress-fade-enter-active {
  transition: opacity 0.15s ease-out;
}

.progress-fade-leave-active {
  transition: opacity 0.3s ease-out;
}

.progress-fade-enter-from,
.progress-fade-leave-to {
  opacity: 0;
}

/* 减少动画模式 */
@media (prefers-reduced-motion: reduce) {
  .navigation-progress-bar,
  .progress-fade-enter-active,
  .progress-fade-leave-active {
    transition: none;
  }
}
</style>
