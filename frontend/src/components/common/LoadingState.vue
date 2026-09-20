<template>
  <div
    :class="['loading-state', `loading-state--${variant}`]"
    role="status"
    aria-live="polite"
    aria-busy="true"
  >
    <template v-if="variant === 'inline'">
      <LoadingSpinner :size="resolvedSize" :color="color" decorative />
      <span class="loading-state__label">{{ resolvedLabel }}</span>
    </template>
    <template v-else>
      <span class="sr-only">{{ resolvedLabel }}</span>
      <div class="loading-state__skeleton" aria-hidden="true">
        <Skeleton width="34%" height="1.25rem" />
        <div v-if="variant === 'page'" class="loading-state__summary">
          <Skeleton v-for="index in 3" :key="index" height="5rem" />
        </div>
        <Skeleton v-for="index in 4" :key="index" :width="index === 4 ? '68%' : '100%'" height="1rem" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from './LoadingSpinner.vue'
import Skeleton from './Skeleton.vue'
import { usePageLoading } from '@/composables/usePageLoading'

type LoadingStateVariant = 'page' | 'section' | 'inline'
type LoadingStateColor = 'primary' | 'secondary' | 'white' | 'gray' | 'inherit'
type SpinnerSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

interface Props {
  variant?: LoadingStateVariant
  label?: string
  size?: SpinnerSize
  color?: LoadingStateColor
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'section',
  label: '',
  size: undefined,
  color: 'primary'
})

const { t } = useI18n()
usePageLoading(() => props.variant !== 'inline')

const resolvedLabel = computed(() => props.label || t('common.loading'))
const resolvedSize = computed<SpinnerSize>(() => {
  if (props.size) return props.size
  if (props.variant === 'page') return 'lg'
  if (props.variant === 'inline') return 'xs'
  return 'md'
})
</script>

<style scoped>
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
}

.loading-state--page {
  min-height: min(28rem, 60vh);
  flex-direction: column;
  gap: 0.875rem;
  padding: 3rem 1.5rem;
}

.loading-state__skeleton {
  display: grid;
  width: 100%;
  min-width: 0;
  gap: 1rem;
}

.loading-state__summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
  margin-block: 0.5rem;
}

.loading-state--section {
  min-height: 8rem;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.5rem;
}

.loading-state--inline {
  display: inline-flex;
  min-height: 0;
  gap: 0.5rem;
  padding: 0;
}

.loading-state__label {
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}
</style>
