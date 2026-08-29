<template>
  <div class="stat-card">
    <div :class="['stat-icon', iconClass]">
      <component v-if="icon" :is="icon" class="stat-card__icon" aria-hidden="true" />
    </div>
    <div class="stat-card__content">
      <p class="stat-card__title stat-label">{{ title }}</p>
      <div class="stat-card__value-row">
        <p class="stat-value" :title="String(formattedValue)">{{ formattedValue }}</p>
        <span v-if="change !== undefined" :class="['stat-trend', trendClass]">
          <Icon
            v-if="changeType !== 'neutral'"
            name="arrowUp"
            size="xs"
            :class="changeType === 'down' && 'stat-card__trend-icon--down'"
          />
          {{ formattedChange }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import Icon from '@/components/icons/Icon.vue'

type ChangeType = 'up' | 'down' | 'neutral'
type IconVariant = 'primary' | 'success' | 'warning' | 'danger'

interface Props {
  title: string
  value: number | string
  icon?: Component
  iconVariant?: IconVariant
  change?: number
  changeType?: ChangeType
  formatValue?: (value: number | string) => string
}

const props = withDefaults(defineProps<Props>(), {
  changeType: 'neutral',
  iconVariant: 'primary'
})

const formattedValue = computed(() => {
  if (props.formatValue) {
    return props.formatValue(props.value)
  }
  if (typeof props.value === 'number') {
    return props.value.toLocaleString()
  }
  return props.value
})

const formattedChange = computed(() => {
  if (props.change === undefined) return ''
  const absChange = Math.abs(props.change)
  return `${absChange}%`
})

const iconClass = computed(() => {
  const classes: Record<IconVariant, string> = {
    primary: 'stat-icon-primary',
    success: 'stat-icon-success',
    warning: 'stat-icon-warning',
    danger: 'stat-icon-danger'
  }
  return classes[props.iconVariant]
})

const trendClass = computed(() => {
  const classes: Record<ChangeType, string> = {
    up: 'stat-trend-up',
    down: 'stat-trend-down',
    neutral: 'stat-card__trend--neutral'
  }
  return classes[props.changeType]
})
</script>

<style scoped>
.stat-card__icon { width: 1.5rem; height: 1.5rem; }

.stat-card__content {
  min-width: 0;
  flex: 1 1 0%;
}

.stat-card__title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stat-card__value-row {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.stat-card__trend-icon--down { transform: rotate(180deg); }
.stat-card__trend--neutral { color: var(--color-text-secondary); }
</style>
