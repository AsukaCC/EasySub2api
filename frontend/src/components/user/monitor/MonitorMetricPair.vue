<template>
  <div class="components-user-monitor-monitor-metric-pair__panel">
    <div
      class="components-user-monitor-monitor-metric-pair__panel-2"
    >
      <div
        class="components-user-monitor-monitor-metric-pair__panel-3"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span>{{ primaryLabel }}</span>
      </div>
      <div class="components-user-monitor-monitor-metric-pair__panel-4">
        {{ primaryValue }}<span class="components-user-monitor-monitor-metric-pair__text">{{ primaryUnit }}</span>
      </div>
    </div>
    <div
      class="components-user-monitor-monitor-metric-pair__panel-2"
    >
      <div
        class="components-user-monitor-monitor-metric-pair__panel-3"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span>{{ secondaryLabel }}</span>
      </div>
      <div class="components-user-monitor-monitor-metric-pair__panel-4">
        {{ secondaryValue }}<span class="components-user-monitor-monitor-metric-pair__text">{{ secondaryUnit }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  primaryLabel: string
  primaryValue: string
  primaryUnit: string
  primaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
  secondaryLabel: string
  secondaryValue: string
  secondaryUnit: string
  secondaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
}>()
</script>
