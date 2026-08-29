<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  used: number
  limit: number
  label?: string // 文字前缀，如 "D" / "W"；不传时显示 icon
}>()

const { t } = useI18n()

const badgeClass = computed(() => {
  if (props.used >= props.limit) {
    return 'components-account-quota-badge__state'
  }
  if (props.used >= props.limit * 0.8) {
    return 'components-account-quota-badge__state-2'
  }
  return 'components-account-quota-badge__state-3'
})

const tooltip = computed(() => {
  if (props.used >= props.limit) {
    return t('admin.accounts.capacity.quota.exceeded')
  }
  return t('admin.accounts.capacity.quota.normal')
})

const fmt = (v: number) => v.toFixed(2)
</script>

<template>
  <span
    :class="[
      'components-account-quota-badge__text-4',
      badgeClass
    ]"
    :title="tooltip"
  >
    <span v-if="label" class="components-account-quota-badge__text">{{ label }}</span>
    <svg v-else class="components-account-quota-badge__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
    </svg>
    <span class="components-account-quota-badge__text-2">${{ fmt(used) }}</span>
    <span class="components-account-quota-badge__text-3">/</span>
    <span class="components-account-quota-badge__text-2">${{ fmt(limit) }}</span>
  </span>
</template>
