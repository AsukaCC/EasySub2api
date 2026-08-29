<script setup lang="ts">
defineProps<{
  colorClass: string
  tooltip?: string
  current: string | number
  max: string | number
  suffix?: string
}>()
</script>

<template>
  <span
    :class="[
      'capacity-badge',
      colorClass
    ]"
    :title="tooltip"
  >
    <slot />
    <span class="capacity-badge__value">{{ current }}</span>
    <span class="capacity-badge__separator">/</span>
    <span class="capacity-badge__value">{{ max }}</span>
    <span v-if="suffix" class="capacity-badge__suffix">{{ suffix }}</span>
  </span>
</template>

<style scoped>
.capacity-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 1px 0.375rem;
  border-radius: var(--radius-md);
  font-size: var(--type-micro-size);
  font-weight: var(--font-weight-medium);
  line-height: var(--line-height-tight);
}

.capacity-badge__value {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.capacity-badge__separator { color: var(--color-text-tertiary); }
.capacity-badge__suffix { font-size: var(--type-micro-size); opacity: 0.6; }
</style>
