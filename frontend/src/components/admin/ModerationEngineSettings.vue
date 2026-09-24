<template>
  <section class="engine-settings">
    <h3>{{ t('upstreamUpdate.engine') }}</h3>
    <p>{{ t('upstreamUpdate.savedEngine') }}: {{ activeEngine }}</p>
    <div class="engine-settings__tabs" role="tablist">
      <button v-for="engine in engines" :key="engine" type="button" role="tab" :aria-selected="selected === engine" :disabled="busy" class="btn btn-secondary" @click="selected = engine">{{ engine === 'typesafe' ? 'TypeSafe' : 'OpenAI' }}</button>
    </div>
    <div v-if="draft" class="engine-settings__fields">
      <label>{{ t('upstreamUpdate.endpoint') }}<input v-model="draft.base_url" type="url" class="input" /></label>
      <label>{{ t('upstreamUpdate.model') }}<input v-model="draft.model" class="input" /></label>
      <label>{{ t('upstreamUpdate.proxy') }}<Select v-model="draft.proxy_id" :options="proxyOptions" :searchable="false" /></label>
      <label>{{ t('upstreamUpdate.timeout') }}<input v-model.number="draft.timeout_ms" type="number" min="100" max="60000" class="input" /></label>
      <label>{{ t('upstreamUpdate.retries') }}<input v-model.number="draft.retry_count" type="number" min="0" max="5" class="input" /></label>
      <label>{{ t('upstreamUpdate.key') }}<input v-model="keys[selected]" type="password" autocomplete="new-password" class="input" :placeholder="t('upstreamUpdate.keyPlaceholder')" /></label>
      <p>{{ t('upstreamUpdate.configuredKeys') }}: {{ draft.api_key_count }}</p>
      <label><input v-model="clearKeys[selected]" type="checkbox" />{{ t('upstreamUpdate.clearKeys') }}</label>
      <details class="engine-settings__thresholds"><summary>{{ t('upstreamUpdate.thresholds') }}</summary><div class="engine-settings__fields"><label v-for="(_, category) in draft.thresholds" :key="category">{{ category }}<input v-model.number="draft.thresholds[category]" class="input" type="number" min="0" max="1" step="0.01" /></label></div></details>
      <label>{{ t('upstreamUpdate.testPrompt') }}<input v-model="testPrompt" class="input" /></label>
      <div class="engine-settings__actions"><button type="button" class="btn btn-secondary" :disabled="busy || !testPrompt.trim()" @click="test">{{ t('upstreamUpdate.test') }}</button><button type="button" class="btn btn-primary" :disabled="busy" @click="save">{{ t('common.save') }}</button></div>
    </div>
    <button v-else class="btn btn-secondary" :disabled="busy" @click="load"><Icon name="refresh" size="sm" />{{ t('upstreamUpdate.refresh') }}</button>
    <p v-if="message" role="status">{{ message }}</p>
  </section>
</template>
<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import { getConfig, updateConfig, testAPIKeys, type ContentModerationConfig, type ModerationEngine } from '@/api/admin/riskControl'
import type { Proxy } from '@/types'
const emit = defineEmits<{ saved: [] }>()
const { t } = useI18n()
const engines: ModerationEngine[] = ['openai', 'typesafe']
const selected = ref<ModerationEngine>('openai'), activeEngine = ref('openai'), busy = ref(false), message = ref(''), testPrompt = ref('')
const profiles = ref<Partial<Record<ModerationEngine, ContentModerationConfig>>>({}), proxies = ref<Proxy[]>([])
const keys = reactive({ openai: '', typesafe: '' }), clearKeys = reactive({ openai: false, typesafe: false })
const draft = computed(() => profiles.value[selected.value])
const proxyOptions = computed(() => [
  { value: null, label: t('upstreamUpdate.direct') },
  ...proxies.value.map(proxy => ({ value: proxy.id, label: proxy.name })),
])
function apply(config: ContentModerationConfig) { activeEngine.value = config.engine || 'openai'; profiles.value = config.engine_configs || { openai: config }; keys.openai = ''; keys.typesafe = ''; clearKeys.openai = false; clearKeys.typesafe = false }
async function load() { busy.value = true; message.value = ''; try { const config = await getConfig(); apply(config); selected.value = config.engine || 'openai'; proxies.value = await adminAPI.proxies.getAll() } catch { message.value = t('upstreamUpdate.loadFailed') } finally { busy.value = false } }
async function save() {
  if (!draft.value) return
  busy.value = true; message.value = ''
  const engine = selected.value, value = draft.value
  try { apply(await updateConfig({ engine, engine_configs: { [engine]: { base_url: value.base_url, model: value.model, proxy_id: value.proxy_id, timeout_ms: value.timeout_ms, retry_count: value.retry_count, thresholds: value.thresholds, ...(keys[engine] ? { api_key: keys[engine] } : {}), clear_api_key: clearKeys[engine] } } })); emit('saved') }
  catch { message.value = t('upstreamUpdate.saveFailed') } finally { busy.value = false }
}
async function test() { if (!draft.value) return; busy.value = true; message.value = ''; try { const result = await testAPIKeys({ engine: selected.value, base_url: draft.value.base_url, model: draft.value.model, proxy_id: draft.value.proxy_id, thresholds: draft.value.thresholds, timeout_ms: draft.value.timeout_ms, prompt: testPrompt.value, ...(keys[selected.value] ? { api_keys: [keys[selected.value]] } : {}) }); message.value = result.items.every(item => item.status === 'ok') ? t('upstreamUpdate.tested') : t('upstreamUpdate.queryFailed') } catch { message.value = t('upstreamUpdate.queryFailed') } finally { busy.value = false } }
onMounted(load)
</script>
<style scoped>
.engine-settings { padding: 16px 0; border-bottom: 1px solid var(--border-color); }
.engine-settings h3 { font-size: 14px; }
.engine-settings__tabs, .engine-settings__actions { display: flex; gap: 8px; margin: 12px 0; }
.engine-settings [aria-selected=true] { border-color: var(--color-primary); }
.engine-settings__fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px; }
.engine-settings label { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; font-size: 12px; }
.engine-settings .input,
.engine-settings :deep(.app-select) { width: 100%; min-width: 0; }
.engine-settings__thresholds { grid-column: 1 / -1; font-size: 12px; }
</style>
