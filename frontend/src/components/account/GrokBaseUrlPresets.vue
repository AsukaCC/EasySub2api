<template>
  <div class="grok-base-url-presets">
    <button
      v-for="preset in GROK_BASE_URL_PRESETS"
      :key="preset.url"
      type="button"
      data-testid="grok-base-url-preset"
      class="grok-base-url-presets__option"
      @click="emit('select', preset.url)"
    >
      {{ presetLabel(preset) }} ({{ displayUrl(preset.url) }})
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { GROK_BASE_URL_PRESETS, type GrokBaseUrlPreset } from './credentialsBuilder'

// Grok 快捷端点：点击把预设地址填入调用方的输入框。
// 仅是快速填充，不限制可填值——输入框仍接受任意第三方转发地址。
const emit = defineEmits<{
  (e: 'select', url: string): void
}>()

const { t } = useI18n()

const presetLabel = (preset: GrokBaseUrlPreset) =>
  preset.label ?? t(`admin.accounts.grokCustomBaseUrl.presets.${preset.labelKey}`)

const displayUrl = (url: string) => url.replace(/^https?:\/\//i, '')
</script>

<style scoped>
.grok-base-url-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.grok-base-url-presets__option {
  padding: 0.25rem 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  color: var(--color-text-primary);
  font-size: var(--type-caption-size);
  line-height: var(--type-caption-line-height);
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  transition: color 150ms ease, border-color 150ms ease, box-shadow 150ms ease, backdrop-filter 150ms ease;
}

.grok-base-url-presets__option:hover {
  border-color: var(--color-primary-border);
  color: var(--color-text-brand);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}
</style>
