<template>
  <BaseDialog :show="show" :title="t('admin.accounts.fingerprint.title')" width="normal" @close="emit('close')">
    <div class="fingerprint-dialog">
      <div class="fingerprint-account">{{ account?.name }}</div>
      <label class="fingerprint-label">{{ t('admin.accounts.selectTestModel') }}</label>
      <Select v-model="selectedModel" :options="models" value-key="id" label-key="display_name"
        searchable creatable
        :disabled="loadingModels || busy || starting"
        :placeholder="loadingModels ? t('common.loading') : t('admin.accounts.selectTestModel')" />
      <p v-if="loadError" class="fingerprint-error" role="alert">{{ t('admin.accounts.fingerprint.loadFailed') }}</p>
      <p v-else-if="!loadingModels && !models.length" class="fingerprint-note">{{ t('admin.accounts.fingerprint.noModels') }}</p>
      <div class="fingerprint-meta"><span>{{ t('admin.accounts.fingerprint.samples') }}</span><span>{{ t('admin.accounts.fingerprint.serial') }}</span></div>
      <div v-if="snapshot" class="fingerprint-output"><ModelFingerprintResult :snapshot="snapshot" /></div>
      <p v-if="pollingError" class="fingerprint-error" role="status">{{ t('admin.accounts.fingerprint.pollFailed') }}</p>
      <p class="fingerprint-note">{{ t('admin.accounts.fingerprint.scope') }}</p>
      <p v-if="submitError" class="fingerprint-error" role="alert">{{ submitError }}</p>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="emit('close')">{{ t('common.close') }}</button>
      <button class="btn btn-primary" :disabled="!selectedModel || loadingModels || busy || starting || pollingError" @click="start">
        <Icon :name="busy || starting ? 'clock' : 'play'" size="sm" />
        {{ busy ? t('admin.accounts.fingerprint.inBackground') : t('admin.accounts.fingerprint.start') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import { Icon } from '@/components/icons'
import { adminAPI } from '@/api/admin'
import { isFingerprintTextModel, startModelFingerprint, type ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'
import { useModelFingerprint } from '@/composables/useModelFingerprint'
import type { Account, ClaudeModel } from '@/types'
import ModelFingerprintResult from './ModelFingerprintResult.vue'

const props = defineProps<{ show: boolean; account: Account | null }>()
const emit = defineEmits<{ close: []; update: [accountId: string, snapshot: ModelFingerprintSnapshot | null] }>()
const { t } = useI18n()
const accountId = computed(() => props.show ? props.account?.id || null : null)
const { snapshot, pollingError, track } = useModelFingerprint(accountId, value => {
  if (accountId.value) emit('update', accountId.value, value)
})
const models = ref<ClaudeModel[]>([])
const selectedModel = ref('')
const loadingModels = ref(false)
const loadError = ref(false)
const starting = ref(false)
const submitError = ref('')
const busy = computed(() => snapshot.value?.status === 'running')

watch(accountId, async (id, _, onCleanup) => {
  let stale = false
  onCleanup(() => { stale = true })
  models.value = []
  selectedModel.value = ''
  loadError.value = false
  submitError.value = ''
  loadingModels.value = Boolean(id)
  if (!id) return
  try {
    const items = await adminAPI.accounts.getAvailableModels(id, true)
    if (stale) return
    models.value = items.filter(item => isFingerprintTextModel(item.id))
    const previous = (props.account?.extra?.model_fingerprint as ModelFingerprintSnapshot | undefined)?.model
    selectedModel.value = models.value.find(item => item.id === previous)?.id || models.value[0]?.id || ''
  } catch { if (!stale) loadError.value = true }
  finally { if (!stale) loadingModels.value = false }
}, { immediate: true })

async function start() {
  const id = accountId.value
  if (!id || busy.value || starting.value || !selectedModel.value) return
  starting.value = true
  submitError.value = ''
  try {
    const value = await startModelFingerprint(id, selectedModel.value)
    if (accountId.value === id) track(value)
    else emit('update', id, value)
  } catch {
    if (accountId.value === id) submitError.value = t('admin.accounts.fingerprint.startFailed')
  } finally { starting.value = false }
}
</script>

<style scoped>
.fingerprint-dialog { display: grid; gap: 12px; min-width: 0; }
.fingerprint-account { font-weight: 600; overflow-wrap: anywhere; }
.fingerprint-label { font-size: var(--font-size-sm); margin-bottom: -6px; }
.fingerprint-meta { display: flex; flex-wrap: wrap; gap: 8px 18px; font-size: var(--font-size-xs); color: var(--color-text-tertiary); }
.fingerprint-output { border-top: 1px solid var(--color-border); padding-top: 16px; margin-top: 4px; }
.fingerprint-note { margin: 0; font-size: var(--font-size-xs); color: var(--color-text-tertiary); line-height: 1.6; }
.fingerprint-error { margin: 0; font-size: var(--font-size-sm); color: var(--color-text-danger); }
</style>
