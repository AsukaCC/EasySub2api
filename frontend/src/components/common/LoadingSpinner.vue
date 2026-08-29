<template>
  <div
    :class="['loading-spinner', sizeClasses, colorClass]"
    role="status"
    :aria-label="t('common.loading')"
  >
    <span class="loading-spinner__label">{{ t('common.loading') }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

type SpinnerSize = 'sm' | 'md' | 'lg' | 'xl'
type SpinnerColor = 'primary' | 'secondary' | 'white' | 'gray'

interface Props {
  size?: SpinnerSize
  color?: SpinnerColor
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  color: 'primary'
})

const sizeClasses = computed(() => {
  const sizes: Record<SpinnerSize, string> = {
    sm: 'loading-spinner--sm',
    md: 'loading-spinner--md',
    lg: 'loading-spinner--lg',
    xl: 'loading-spinner--xl'
  }
  return sizes[props.size]
})

const colorClass = computed(() => {
  const colors: Record<SpinnerColor, string> = {
    primary: 'loading-spinner--primary',
    secondary: 'loading-spinner--secondary',
    white: 'loading-spinner--white',
    gray: 'loading-spinner--muted'
  }
  return colors[props.color]
})
</script>

<style scoped>
.loading-spinner {
  display: inline-block;
  flex: 0 0 auto;
  border-style: solid;
  border-color: color-mix(in srgb, currentColor 24%, transparent);
  border-top-color: currentColor;
  border-radius: var(--radius-full);
  animation: spin 0.75s linear infinite;
}

.loading-spinner__label {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.loading-spinner--sm { width: 1rem; height: 1rem; border-width: 2px; }
.loading-spinner--md { width: 2rem; height: 2rem; border-width: 2px; }
.loading-spinner--lg { width: 3rem; height: 3rem; border-width: 3px; }
.loading-spinner--xl { width: 4rem; height: 4rem; border-width: 4px; }
.loading-spinner--primary { color: var(--color-primary); }
.loading-spinner--secondary { color: var(--color-text-secondary); }
.loading-spinner--white { color: var(--color-text-inverse); }
.loading-spinner--muted { color: var(--color-text-muted); }

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
