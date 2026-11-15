<template>
  <section class="model-forwarding" data-testid="openai-model-forwarding">
    <label v-if="!hideLabel" class="input-label">{{ t('admin.accounts.openai.modelForwarding') }}</label>
    <p class="input-hint">{{ t('admin.accounts.openai.modelForwardingHint') }}</p>
    <p v-if="batch" class="input-hint">{{ t('admin.accounts.openai.modelForwardingBatchHint') }}</p>
    <div class="model-forwarding__actions">
      <button v-if="canLoadModels" type="button" class="btn btn-secondary" :disabled="loading" data-testid="load-forwarding-models" @click="loadModels">
        {{ loading ? t('admin.accounts.syncUpstreamModelsLoading') : t('admin.accounts.syncUpstreamModels') }}
      </button>
      <button v-if="modelValue.length" type="button" class="btn btn-secondary" data-testid="clear-forwarding" @click="emit('update:modelValue', [])">
        {{ t('admin.accounts.openai.clearModelForwarding') }}
      </button>
    </div>
    <p v-if="loadError" class="input-hint" role="alert">{{ loadError }}</p>
    <p v-else-if="!canLoadModels" class="input-hint">{{ t('admin.accounts.openai.modelForwardingCredentialsHint') }}</p>
    <div v-for="(row, index) in modelValue" :key="index" class="model-forwarding__row">
      <input
        :value="row.from"
        class="input"
        required
        pattern="[^\s*]+\*?"
        :aria-label="t('admin.accounts.requestModel')"
        placeholder="gpt-5.6-luna"
        data-testid="forwarding-source"
        @input="updateRow(index, 'from', ($event.target as HTMLInputElement).value)"
      />
      <span aria-hidden="true">→</span>
      <Select
        :model-value="row.to"
        :options="targetOptions"
        searchable
        :loading="loading"
        :placeholder="t('admin.accounts.openai.selectForwardingTarget')"
        :aria-label="t('admin.accounts.actualModel')"
        data-testid="forwarding-target"
        @update:model-value="updateRow(index, 'to', String($event ?? ''))"
      />
      <button type="button" class="btn btn-secondary" :aria-label="t('common.delete')" @click="removeRow(index)">
        <Icon name="x" size="sm" />
      </button>
    </div>
    <button type="button" class="btn btn-secondary" data-testid="add-forwarding" @click="emit('update:modelValue', [...modelValue, { from: '', to: '' }])">
      {{ t('admin.accounts.openai.addModelForwarding') }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import { accountsAPI, type SyncUpstreamPreviewParams } from '@/api/admin/accounts'
import type { ModelMappingEntry } from '@/composables/useModelWhitelist'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{
  modelValue: ModelMappingEntry[]
  accountId?: string
  syncCredentials?: SyncUpstreamPreviewParams
  batch?: boolean
  hideLabel?: boolean
}>()
const emit = defineEmits<{ 'update:modelValue': [rows: ModelMappingEntry[]] }>()
const { t } = useI18n()
const loading = ref(false)
const loadError = ref('')
const upstreamModels = ref<string[]>([])
const loaded = ref(false)
let loadVersion = 0
const canLoadModels = computed(() => Boolean(props.accountId || props.syncCredentials?.api_key))
const targetOptions = computed(() => {
  const supported = new Set(upstreamModels.value)
  // Keep saved targets visible while loading, without offering them as discovered models.
  return Array.from(new Set([...upstreamModels.value, ...props.modelValue.map(row => row.to).filter(Boolean)]), value => ({
    value, label: value, disabled: !supported.has(value)
  }))
})

async function loadModels() {
  if (!canLoadModels.value || loading.value) return
  const version = loadVersion
  loading.value = true
  loadError.value = ''
  try {
    const result = props.accountId
      ? await accountsAPI.syncUpstreamModels(props.accountId)
      : await accountsAPI.syncUpstreamModelsPreview(props.syncCredentials!)
    if (version !== loadVersion) return
    upstreamModels.value = Array.from(new Set(result.models.map(model => model.trim()).filter(model => model && !/[\s*]/.test(model))))
    loaded.value = true
    if (!upstreamModels.value.length) loadError.value = t('admin.accounts.syncUpstreamModelsEmpty')
  } catch (error) {
    if (version === loadVersion) loadError.value = extractApiErrorMessage(error, t('admin.accounts.syncUpstreamModelsFailed'))
  } finally {
    if (version === loadVersion) loading.value = false
  }
}

watch([
  () => props.accountId,
  () => props.syncCredentials?.platform,
  () => props.syncCredentials?.type,
  () => props.syncCredentials?.base_url,
  () => props.syncCredentials?.api_key
], () => {
  loadVersion += 1
  loading.value = false
  loaded.value = false
  upstreamModels.value = []
  loadError.value = ''
})
// Fetch when a rule needs a target; preview credentials are only sent on demand.
watch([() => props.accountId, () => props.modelValue.length > 0], () => {
  if (props.modelValue.length && !loaded.value) void loadModels()
}, { immediate: true })
onBeforeUnmount(() => { loadVersion += 1 })

function updateRow(index: number, field: 'from' | 'to', value: string) {
  emit('update:modelValue', props.modelValue.map((row, i) => i === index ? { ...row, [field]: value } : row))
}

function removeRow(index: number) {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== index))
}
</script>

<style scoped>
.model-forwarding { display: grid; gap: 0.75rem; min-width: 0; }
.model-forwarding__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.model-forwarding__row { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) auto; align-items: center; gap: 0.5rem; }
.model-forwarding__row > * { min-width: 0; }
@media (max-width: 480px) {
  .model-forwarding__row { grid-template-columns: minmax(0, 1fr) auto; }
  .model-forwarding__row > input { grid-column: 1 / -1; }
  .model-forwarding__row > span { display: none; }
}
</style>
