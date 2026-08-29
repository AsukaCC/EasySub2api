<template>
  <Teleport to="body">
    <div
      class="toast-stack"
      aria-live="polite"
      aria-atomic="true"
    >
      <TransitionGroup
        enter-active-class="toast-motion toast-motion--enter"
        enter-from-class="toast-motion--offscreen"
        enter-to-class="toast-motion--onscreen"
        leave-active-class="toast-motion toast-motion--leave"
        leave-from-class="toast-motion--onscreen"
        leave-to-class="toast-motion--offscreen"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'app-toast',
            getBorderColor(toast.type)
          ]"
        >
          <div class="app-toast__body">
            <div class="app-toast__row">
              <!-- Icon -->
              <div class="app-toast__icon">
                <Icon
                  :name="getToastIconName(toast.type)"
                  size="md"
                  :class="getIconColor(toast.type)"
                  aria-hidden="true"
                />
              </div>

              <!-- Content -->
              <div class="app-toast__content">
                <p v-if="toast.title" class="app-toast__title">
                  {{ toast.title }}
                </p>
                <p
                  :class="[
                    'app-toast__message',
                    toast.title
                      ? 'app-toast__message--with-title'
                      : 'app-toast__message--standalone'
                  ]"
                >
                  {{ toast.message }}
                </p>
              </div>

              <!-- Close button -->
              <button
                @click="removeToast(toast.id)"
                class="app-toast__close"
                aria-label="Close notification"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>

          <!-- Progress bar -->
          <div v-if="toast.duration" class="app-toast__timer">
            <div
              :class="['app-toast__progress', getProgressBarColor(toast.type)]"
              :style="{ animationDuration: `${toast.duration}ms` }"
            ></div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const toasts = computed(() => appStore.toasts)

const getToastIconName = (type: string): 'checkCircle' | 'xCircle' | 'exclamationTriangle' | 'infoCircle' => {
  switch (type) {
    case 'success':
      return 'checkCircle'
    case 'error':
      return 'xCircle'
    case 'warning':
      return 'exclamationTriangle'
    case 'info':
    default:
      return 'infoCircle'
  }
}

const getIconColor = (type: string): string => {
  const colors: Record<string, string> = {
    success: 'status-text--success',
    error: 'status-text--danger',
    warning: 'status-text--warning',
    info: 'status-text--info'
  }
  return colors[type] || colors.info
}

const getBorderColor = (type: string): string => {
  const colors: Record<string, string> = {
    success: 'toast-success',
    error: 'toast-error',
    warning: 'toast-warning',
    info: 'toast-info'
  }
  return colors[type] || colors.info
}

const getProgressBarColor = (type: string): string => {
  const colors: Record<string, string> = {
    success: 'status-fill--success',
    error: 'status-fill--danger',
    warning: 'status-fill--warning',
    info: 'status-fill--info'
  }
  return colors[type] || colors.info
}

const removeToast = (id: string) => {
  appStore.hideToast(id)
}
</script>

<style scoped>
.toast-stack {
  position: fixed;
  top: 1rem;
  right: 1rem;
  z-index: 9999;
  pointer-events: none;
}

.toast-stack > :not([hidden]) ~ :not([hidden]) {
  margin-top: 0.75rem;
}

.toast-motion {
  transition-property: opacity, transform;
}

.toast-motion--enter {
  transition-duration: 300ms;
  transition-timing-function: ease-out;
}

.toast-motion--leave {
  transition-duration: 200ms;
  transition-timing-function: ease-in;
}

.toast-motion--offscreen {
  opacity: 0;
  transform: translateX(100%);
}

.toast-motion--onscreen {
  opacity: 1;
  transform: translateX(0);
}

.app-toast {
  min-width: min(20rem, calc(100vw - 2rem));
  max-width: min(28rem, calc(100vw - 2rem));
  overflow: hidden;
  border: 1px solid var(--glass-border-hover);
  border-left-width: 4px;
  border-radius: var(--radius-lg);
  background-color: var(--glass-layer-floating-bg);
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  pointer-events: auto;
}

.app-toast.toast-success { border-left-color: var(--color-success); }
.app-toast.toast-error { border-left-color: var(--color-danger); }
.app-toast.toast-warning { border-left-color: var(--color-warning); }
.app-toast.toast-info { border-left-color: var(--color-info); }

.app-toast__body {
  padding: 1rem;
}

.app-toast__row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.app-toast__icon {
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.app-toast__content {
  flex: 1 1 0%;
  min-width: 0;
}

.app-toast__title,
.app-toast__message--standalone {
  color: var(--color-text-primary);
}

.app-toast__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  line-height: 1.25rem;
}

.app-toast__message {
  font-size: var(--font-size-sm);
  line-height: 1.625;
  overflow-wrap: anywhere;
}

.app-toast__message--with-title {
  margin-top: 0.25rem;
  color: var(--color-text-secondary);
}

.app-toast__close {
  flex-shrink: 0;
  margin: -0.25rem;
  padding: 0.25rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background-color: transparent;
  color: var(--color-text-tertiary);
  transition: color 150ms ease, border-color 150ms ease, box-shadow 150ms ease;
}

.app-toast__close:hover {
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  color: var(--color-text-secondary);
  box-shadow: 0 1px 0 var(--glass-highlight-hover) inset;
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.app-toast__close:focus-visible {
  border-color: var(--glass-border-active);
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.22);
}

.app-toast__timer {
  height: 0.25rem;
  background-color: var(--glass-bg-interactive);
}

.app-toast__progress {
  width: 100%;
  height: 100%;
  animation-name: toast-progress-shrink;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}

@keyframes toast-progress-shrink {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .toast-motion {
    transition-duration: 1ms;
  }

  .app-toast__progress {
    animation: none;
  }
}
</style>
