<template>
  <div class="codex-export-options">
    <label class="form-label">
      {{ t('keys.useKeyModal.codexOptions.model') }}
      <input class="input" data-testid="codex-main-model" :value="modelValue.model" @input="update('model', ($event.target as HTMLInputElement).value)" />
    </label>
    <label class="codex-export-options__toggle">
      <input type="checkbox" data-testid="codex-subagent-follow" :checked="modelValue.subagentModel === null" @change="update('subagentModel', ($event.target as HTMLInputElement).checked ? null : modelValue.model)" />
      {{ t('keys.useKeyModal.codexOptions.followMain') }}
    </label>
    <label v-if="modelValue.subagentModel !== null" class="form-label">
      {{ t('keys.useKeyModal.codexOptions.subagent') }}
      <input class="input" data-testid="codex-subagent-model" :value="modelValue.subagentModel" @input="update('subagentModel', ($event.target as HTMLInputElement).value)" />
    </label>
    <label class="codex-export-options__toggle">
      <input type="checkbox" data-testid="codex-websocket" :checked="modelValue.websocketEnabled" @change="update('websocketEnabled', ($event.target as HTMLInputElement).checked)" />
      {{ t('keys.useKeyModal.codexOptions.websocket') }}
    </label>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CodexExportSettings } from '@/utils/ccswitchImport'

const props = defineProps<{ modelValue: CodexExportSettings }>()
const emit = defineEmits<{ 'update:modelValue': [value: CodexExportSettings] }>()
const { t } = useI18n()
function update<K extends keyof CodexExportSettings>(key: K, value: CodexExportSettings[K]) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<style scoped>
.codex-export-options { display: grid; gap: 12px; min-width: 0; }
.codex-export-options .input { display: block; width: 100%; margin-top: 6px; }
.codex-export-options__toggle { display: flex; align-items: center; gap: 8px; font-size: 14px; }
</style>
