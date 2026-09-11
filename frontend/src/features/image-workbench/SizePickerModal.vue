<template>
  <BaseDialog
    :show="show"
    :title="t('imageWorkbench.sizePickerTitle')"
    width="normal"
    close-on-click-outside
    :z-index="60"
    @close="emit('close')"
  >
    <p class="input-hint size-picker__current">{{ t('imageWorkbench.sizeCurrent', { size: currentSize || 'auto' }) }}</p>

    <div class="size-picker-tabs" role="tablist">
      <button
        v-for="item in modes"
        :key="item.id"
        type="button"
        role="tab"
        :class="{ active: mode === item.id }"
        :aria-selected="mode === item.id"
        @click="mode = item.id"
      >
        {{ item.label }}
      </button>
    </div>

    <div v-if="mode === 'auto'" class="size-picker-note">
      <strong>{{ t('imageWorkbench.sizeAutoTitle') }}</strong>
      <p>{{ t('imageWorkbench.sizeAutoHint') }}</p>
      <p>{{ t('imageWorkbench.sizeAutoHint2') }}</p>
    </div>

    <div v-else-if="mode === 'ratio'" class="size-picker-form">
      <div>
        <span class="input-label">{{ t('imageWorkbench.sizeBaseResolution') }}</span>
        <div class="size-picker-chips">
          <button
            v-for="item in SIZE_TIERS"
            :key="item"
            type="button"
            class="size-picker-chip"
            :class="{ active: tier === item }"
            @click="tier = item"
          >
            {{ item }}
          </button>
        </div>
      </div>
      <div>
        <span class="input-label">{{ t('imageWorkbench.sizeImageRatio') }}</span>
        <div class="size-picker-chips">
          <button
            v-for="item in RATIO_PRESETS"
            :key="item.value"
            type="button"
            class="size-picker-chip"
            :class="{ active: ratio === item.value }"
            @click="ratio = item.value"
          >
            <span class="ratio-shape" :class="`ratio-shape--${ratioOrientation(item.value)}`" />
            {{ item.label }}
          </button>
          <button
            type="button"
            class="size-picker-chip"
            :class="{ active: ratio === 'custom' }"
            @click="ratio = 'custom'"
          >
            <span class="ratio-shape ratio-shape--custom" />
            {{ t('imageWorkbench.sizeCustomRatio') }}
          </button>
        </div>
      </div>
      <label v-if="ratio === 'custom'">
        <span class="input-label">{{ t('imageWorkbench.sizeCustomRatioLabel') }}</span>
        <input
          v-model="customRatio"
          class="input"
          :class="{ 'input-error': !customRatioValid }"
          :placeholder="t('imageWorkbench.sizeCustomRatioPlaceholder')"
        />
      </label>
    </div>

    <div v-else class="size-picker-form">
      <span class="input-label">{{ t('imageWorkbench.sizeCustomPixels') }}</span>
      <div class="size-picker-xy">
        <label>
          <span class="input-label">{{ t('imageWorkbench.sizeWidth') }}</span>
          <input v-model="customW" class="input" inputmode="numeric" placeholder="1024" />
        </label>
        <span class="size-picker-xy-sep" aria-hidden="true">×</span>
        <label>
          <span class="input-label">{{ t('imageWorkbench.sizeHeight') }}</span>
          <input v-model="customH" class="input" inputmode="numeric" placeholder="1024" />
        </label>
      </div>
      <p class="input-hint">{{ t('imageWorkbench.sizeLimitHint') }}</p>
    </div>

    <div class="size-picker-preview">
      <span class="input-label">{{ t('imageWorkbench.sizeWillUse') }}</span>
      <strong>{{ previewSize || t('imageWorkbench.sizeInvalid') }}</strong>
      <p v-if="isClamped" class="input-hint">{{ t('imageWorkbench.sizeClamped') }}</p>
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="!previewSize" @click="applySize">
        {{ t('common.confirm') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import {
  RATIO_PRESETS,
  SIZE_TIERS,
  calculateImageSize,
  findPresetForSize,
  normalizeImageSize,
  parseRatio,
  parseSize,
  ratioOrientation,
  type SizeMode,
  type SizeTier,
} from './size'

const props = defineProps<{
  show: boolean
  currentSize: string
}>()

const emit = defineEmits<{
  select: [size: string]
  close: []
}>()

const { t } = useI18n()
const modes = computed(() => [
  { id: 'auto' as const, label: t('imageWorkbench.sizeModeAuto') },
  { id: 'ratio' as const, label: t('imageWorkbench.sizeModeRatio') },
  { id: 'resolution' as const, label: t('imageWorkbench.sizeModeCustom') },
])

const mode = ref<SizeMode>('auto')
const tier = ref<SizeTier>('1K')
const ratio = ref('1:1')
const customRatio = ref('16:9')
const customW = ref('1024')
const customH = ref('1024')

const activeRatio = computed(() => (ratio.value === 'custom' ? customRatio.value : ratio.value))
const parsedCustomRatio = computed(() => parseRatio(customRatio.value))
const customRatioValid = computed(() => ratio.value !== 'custom' || Boolean(parsedCustomRatio.value))
const customRatioClamped = computed(() => {
  if (ratio.value !== 'custom' || !parsedCustomRatio.value) return false
  const { width, height } = parsedCustomRatio.value
  return Math.max(width, height) / Math.min(width, height) > 3
})

const previewSize = computed(() => {
  if (mode.value === 'auto') return 'auto'
  if (mode.value === 'ratio') {
    const size = calculateImageSize(tier.value, activeRatio.value)
    return size ? normalizeImageSize(size) : ''
  }
  const width = Number.parseInt(customW.value, 10)
  const height = Number.parseInt(customH.value, 10)
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return ''
  return normalizeImageSize(`${width}x${height}`)
})

const isClamped = computed(() => {
  if (!previewSize.value || previewSize.value === 'auto') return false
  if (mode.value === 'ratio' && ratio.value === 'custom') return customRatioClamped.value
  if (mode.value === 'resolution') {
    const width = Number.parseInt(customW.value, 10)
    const height = Number.parseInt(customH.value, 10)
    if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return false
    return `${width}x${height}` !== previewSize.value
  }
  return false
})

function hydrate() {
  const preset = findPresetForSize(props.currentSize)
  const parsed = parseSize(props.currentSize)
  mode.value = !props.currentSize || props.currentSize === 'auto' ? 'auto' : preset ? 'ratio' : 'resolution'
  tier.value = preset?.tier ?? '1K'
  ratio.value = preset?.ratio ?? '1:1'
  customRatio.value = '16:9'
  customW.value = String(parsed?.width || 1024)
  customH.value = String(parsed?.height || 1024)
}

function applySize() {
  if (!previewSize.value) return
  emit('select', previewSize.value)
  emit('close')
}

watch(() => props.show, (open) => {
  if (open) hydrate()
}, { immediate: true })
</script>

<style scoped>
.size-picker__current {
  margin: 0 0 1rem;
}

.size-picker-tabs {
  display: flex;
  gap: 0.25rem;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.size-picker-tabs button {
  flex: 1;
  border: 0;
  border-bottom: 2px solid transparent;
  padding: 0.6rem;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  font: inherit;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.size-picker-tabs button.active {
  border-color: var(--theme-accent);
  color: var(--theme-accent);
}

.size-picker-note,
.size-picker-form {
  display: grid;
  gap: 0.85rem;
}

.size-picker-note {
  padding: 0.9rem 0.95rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-subtle);
}

.size-picker-note strong {
  display: block;
  margin-bottom: 0.35rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
}

.size-picker-note p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

.size-picker-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.size-picker-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  min-height: 2rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-full);
  background: var(--glass-bg-interactive);
  color: var(--color-text-secondary);
  padding: 0.3rem 0.85rem;
  cursor: pointer;
  font: inherit;
  font-size: var(--font-size-sm);
}

.size-picker-chip:hover,
.size-picker-chip.active {
  color: var(--color-text-brand);
  border-color: var(--color-primary-border);
  background: var(--color-primary-subtle);
}

.ratio-shape {
  display: block;
  width: 0.85rem;
  height: 0.85rem;
  border: 1.5px solid currentColor;
  border-radius: var(--radius-xs);
}

.ratio-shape--landscape {
  width: 1.05rem;
  height: 0.58rem;
}

.ratio-shape--portrait {
  width: 0.58rem;
  height: 1.05rem;
}

.ratio-shape--custom {
  border-style: dashed;
}

.size-picker-xy {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: end;
  gap: 0.65rem;
}

.size-picker-xy-sep {
  padding-bottom: 0.65rem;
  color: var(--color-text-tertiary);
}

.size-picker-preview {
  margin-top: 1.15rem;
  padding-top: 0.9rem;
  border-top: 1px solid var(--color-border-subtle);
}

.size-picker-preview strong {
  display: block;
  color: var(--color-text-primary);
  font-size: var(--font-size-lg);
}
</style>
