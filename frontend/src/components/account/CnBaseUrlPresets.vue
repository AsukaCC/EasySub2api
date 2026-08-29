<template>
  <div class="cn-base-url-presets">
    <button
      v-for="preset in presets"
      :key="preset.mode + ':' + preset.protocol + ':' + preset.url"
      type="button"
      data-testid="cn-base-url-preset"
      :class="[
        'cn-base-url-presets__option',
        isActive(preset)
          ? 'cn-base-url-presets__option--active'
          : 'cn-base-url-presets__option--idle'
      ]"
      @click="emit('select', preset)"
    >
      {{ preset.label }} ({{ displayUrl(preset.url) }})
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CN_BASE_URL_PRESETS, type CnBaseUrlPreset } from './credentialsBuilder'

// 国产供应商快捷端点：点击把预设地址（及对应账号类型/协议）回填到调用方。
// 与 Grok 预设一致，仅作快速填充，输入框仍接受任意第三方转发地址。
// 传入 protocol 时只显示该协议档的预设（协议 × 账号类型正交分档）。
const props = defineProps<{
  platform: 'kimi' | 'zhipu' | 'deepseek'
  /** 当前已选账号类型，用于过滤和高亮匹配的预设 */
  mode?: 'payg' | 'coding'
  /** 当前已选 API 协议，用于过滤和高亮匹配的预设 */
  protocol?: 'chat_completions' | 'anthropic' | 'responses'
  /** 当前输入框中的 base url，用于高亮完全匹配项 */
  currentUrl?: string
}>()

const emit = defineEmits<{
  (e: 'select', preset: CnBaseUrlPreset): void
}>()

const presets = computed(() => {
  const all = CN_BASE_URL_PRESETS[props.platform] ?? []
  // 只按协议过滤：同协议下 payg/coding 两档都展示，点击即同时切换账号类型。
  if (props.protocol == null) return all
  return all.filter(p => p.protocol === props.protocol)
})

const isActive = (preset: CnBaseUrlPreset) =>
  (props.mode != null && preset.mode === props.mode) || preset.url === props.currentUrl

const displayUrl = (url: string) => url.replace(/^https?:\/\//i, '')
</script>

<style scoped>
.cn-base-url-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.cn-base-url-presets__option {
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

.cn-base-url-presets__option--active {
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.cn-base-url-presets__option--idle:hover {
  border-color: var(--color-primary-border);
  color: var(--color-text-brand);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}
</style>
