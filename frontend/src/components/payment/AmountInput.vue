<template>
  <div class="components-payment-amount-input__panel amount-input-root">
    <!-- Quick Amount Buttons -->
    <div>
      <label class="components-payment-amount-input__label amount-input-label">
        {{ t('payment.quickAmounts') }}
      </label>
      <div class="components-payment-amount-input__panel-2 amount-input-grid">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :class="[
            'components-payment-amount-input__action amount-btn',
            modelValue === amt
              ? 'components-payment-amount-input__action-2 amount-btn--active'
              : 'components-payment-amount-input__action-3',
          ]"
          @click="selectAmount(amt)"
        >
          ¥{{ amt }}
        </button>
      </div>
    </div>

    <!-- Custom Amount Input -->
    <div>
      <label class="components-payment-amount-input__label amount-input-label">
        {{ t('payment.customAmount') }}
      </label>
      <div class="components-payment-amount-input__panel-3 amount-custom-wrap">
        <span class="components-payment-amount-input__text amount-custom-symbol">
          ¥
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="components-payment-amount-input__field amount-custom-field input"
          @input="handleInput"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
  amounts?: number[]
  modelValue: number | null
  min?: number
  max?: number
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500, 1000, 2000, 5000],
  min: 0,
  max: 0,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

// 0 = no limit
const filteredAmounts = computed(() =>
  props.amounts.filter((a) => (props.min <= 0 || a >= props.min) && (props.max <= 0 || a <= props.max))
)

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return `≥ ${props.min}`
  if (props.max > 0) return `≤ ${props.max}`
  return t('payment.enterAmount')
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function handleInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (!AMOUNT_PATTERN.test(val)) return
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>

<style scoped>
.amount-input-root {
  display: grid;
  gap: 0.75rem;
}

.amount-input-label {
  display: block;
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.375rem;
}

.amount-input-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
}

.amount-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 0.5rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease-in-out;
}

.amount-btn:hover {
  background: var(--color-bg-tertiary);
  border-color: var(--color-border-hover, var(--color-primary));
}

.amount-btn--active {
  border-color: var(--color-primary) !important;
  background: var(--color-primary-subtle, rgba(99, 102, 241, 0.1)) !important;
  color: var(--color-primary) !important;
  font-weight: 600;
  box-shadow: 0 0 0 1px var(--color-primary);
}

.amount-custom-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.amount-custom-symbol {
  position: absolute;
  left: 0.875rem;
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-secondary);
  pointer-events: none;
}

.amount-custom-field {
  width: 100%;
  padding-left: 2rem !important;
}
</style>
