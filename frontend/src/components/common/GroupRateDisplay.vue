<template>
  <span
    v-if="rate.effective !== null"
    class="group-rate"
    :class="{ 'group-rate--compared': rate.hasComparison }"
    :title="comparisonTitle"
  >
    <template v-if="rate.hasComparison">
      <span class="group-rate__original">{{ formatMultiplier(rate.original!) }}x</span>
      <Icon name="arrowRight" size="xs" class="group-rate__arrow" />
    </template>
    <span class="group-rate__minimum">{{ formatMultiplier(rate.effective) }}x</span>
    <span v-if="showLabel" class="group-rate__label">{{ t('admin.groups.rateLabel') }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatMultiplier, lowestAvailableGroupRate } from '@/utils/formatters'

const props = withDefaults(defineProps<{
  rateMultiplier?: number
  userRateMultiplier?: number | null
  showLabel?: boolean
}>(), {
  userRateMultiplier: null,
  showLabel: false,
})

const { t } = useI18n()

const rate = computed(() => {
  const original = lowestAvailableGroupRate(props.rateMultiplier, null)
  const effective = lowestAvailableGroupRate(props.rateMultiplier, props.userRateMultiplier)

  return {
    original,
    effective,
    hasComparison: original !== null && effective !== null && effective < original,
  }
})

const comparisonTitle = computed(() => {
  if (!rate.value.hasComparison || rate.value.original === null || rate.value.effective === null) return undefined
  return t('common.groupRateComparison', {
    original: formatMultiplier(rate.value.original),
    lowest: formatMultiplier(rate.value.effective),
  })
})
</script>

<style scoped>
.group-rate {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  white-space: nowrap;
}

.group-rate__original {
  color: var(--color-text-quaternary);
  font-size: var(--type-micro-size);
  text-decoration: line-through;
  text-decoration-thickness: 1px;
}

.group-rate__arrow {
  color: var(--color-text-quaternary);
}

.group-rate__minimum {
  font-weight: var(--font-weight-semibold);
}

.group-rate--compared .group-rate__minimum {
  padding: 0.0625rem 0.25rem;
  border: 1px solid color-mix(in srgb, var(--color-success) 32%, transparent);
  border-radius: var(--radius-sm);
  background: var(--glass-tint-success);
  color: var(--color-text-success);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.group-rate__label {
  color: currentColor;
  font-weight: var(--font-weight-medium);
}
</style>
