<template>
  <BaseDialog :show="show" :title="t('admin.accounts.fingerprint.title')" width="wide" @close="emit('close')">
    <div class="fingerprint-dialog">
      <div class="fingerprint-account">{{ account?.name }}</div>
      <label class="fingerprint-label">{{ t('admin.accounts.fingerprint.apiKey') }}</label>
      <Select v-model="selectedKeyId" :options="keys" value-key="id" label-key="name" searchable
        :disabled="busy || starting || savingSchedule || loadingKeys" />
      <p v-if="keyError" class="fingerprint-error" role="alert">{{ t('admin.accounts.fingerprint.keysFailed') }}</p>
      <p v-else-if="!loadingKeys && !keys.length" class="fingerprint-note">{{ t('admin.accounts.fingerprint.noKeys') }}</p>
      <label class="fingerprint-label">{{ t('admin.accounts.fingerprint.protocol') }}</label>
      <Select v-model="protocol" :options="protocolOptions" :disabled="busy || starting" />
      <label class="fingerprint-label">{{ t('admin.accounts.selectTestModel') }}</label>
      <Select v-model="selectedModel" :options="models" value-key="id" label-key="display_name"
        searchable
        :disabled="loadingModels || busy || starting"
        :placeholder="loadingModels ? t('common.loading') : t('admin.accounts.selectTestModel')" />
      <p v-if="loadError" class="fingerprint-error" role="alert">{{ t('admin.accounts.fingerprint.loadFailed') }}</p>
      <p v-else-if="!loadingModels && !models.length" class="fingerprint-note">{{ t('admin.accounts.fingerprint.noModels') }}</p>
      <label class="fingerprint-label">{{ t('admin.accounts.fingerprint.effort') }}</label>
      <Select v-model="effort" :options="effortOptions" :disabled="busy || starting || loadingModels" />
      <div class="fingerprint-schedule">
        <label><input v-model="scheduleEnabled" type="checkbox" :disabled="savingSchedule || loadingSchedule" /> {{ t('admin.accounts.fingerprint.schedule') }}</label>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="savingSchedule || loadingSchedule || !accountId || (scheduleEnabled && (!selectedKeyId || !selectedModel || loadError || loadingModels || keyError))" @click="saveSchedule">
          <Icon name="check" size="sm" />{{ t('common.save') }}
        </button>
      </div>
      <p v-if="nextRunAt" class="fingerprint-note">{{ t('admin.accounts.fingerprint.nextRun') }} {{ new Date(nextRunAt).toLocaleString() }}</p>
      <p v-if="scheduleMessage" :class="scheduleFailed ? 'fingerprint-error' : 'fingerprint-note'" role="status">{{ scheduleMessage }}</p>
      <div class="fingerprint-meta"><span>{{ t('admin.accounts.fingerprint.samples') }}</span><span>{{ t('admin.accounts.fingerprint.serial') }}</span></div>
      <div v-if="viewedSnapshot || snapshot" ref="resultElement" class="fingerprint-output"><ModelFingerprintResult :snapshot="viewedSnapshot || snapshot" /></div>
      <p v-if="pollingError" class="fingerprint-error" role="status">{{ t('admin.accounts.fingerprint.pollFailed') }}</p>
      <p class="fingerprint-note">{{ t('admin.accounts.fingerprint.scope') }}</p>
      <p v-if="submitError" class="fingerprint-error" role="alert">{{ submitError }}</p>
      <div class="fingerprint-history-heading">
        <strong>{{ t('admin.accounts.fingerprint.history') }}</strong>
        <RouterLink v-if="accountId" :to="{ path: '/admin/usage/admin', query: { account_id: accountId, request_type: 'test' } }" @click="emit('close')">{{ t('admin.accounts.fingerprint.usageRecords') }}</RouterLink>
      </div>
      <p v-if="historyError" class="fingerprint-error" role="alert">{{ t('admin.accounts.fingerprint.historyFailed') }}</p>
      <DataTable :columns="historyColumns" :data="history.items" :loading="loadingHistory" @row-click="viewHistory">
        <template #cell-started_at="{ row }">{{ new Date(row.started_at).toLocaleString() }}</template>
        <template #cell-source="{ row }">{{ t(`admin.accounts.fingerprint.${row.source === 'scheduled' ? 'scheduled' : 'manual'}`) }}</template>
        <template #cell-protocol="{ row }">{{ protocolLabel(row.resolved_protocol || row.protocol) }} / {{ row.reasoning_effort || t('admin.accounts.fingerprint.defaultEffort') }}</template>
        <template #cell-result="{ row }"><ModelFingerprintResult :snapshot="row" compact /></template>
      </DataTable>
      <Pagination v-if="history.total > 0" :page="history.page" :page-size="10" :total="history.total" :show-page-size-selector="false" @update:page="loadHistory" />
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="emit('close')">{{ t('common.close') }}</button>
      <button class="btn btn-primary" :disabled="!selectedKeyId || keyError || !selectedModel || loadingModels || busy || starting || pollingError" @click="start">
        <Icon :name="busy || starting ? 'clock' : 'play'" size="sm" />
        {{ busy ? t('admin.accounts.fingerprint.inBackground') : t('admin.accounts.fingerprint.start') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import { Icon } from '@/components/icons'
import { getFingerprintKeys, getFingerprintModels, getFingerprintHistory, getFingerprintSchedule, setFingerprintSchedule, startModelFingerprint, type FingerprintKey, type FingerprintOptions, type FingerprintProtocol, type FingerprintModel, type FingerprintHistory, type ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'
import { useModelFingerprint } from '@/composables/useModelFingerprint'
import type { Account } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'
import ModelFingerprintResult from './ModelFingerprintResult.vue'

const props = defineProps<{ show: boolean; account: Account | null }>()
const emit = defineEmits<{ close: []; update: [accountId: string, snapshot: ModelFingerprintSnapshot | null] }>()
const { t } = useI18n()
const selectedKeyId = ref('')
const accountId = computed(() => props.show ? props.account?.id || null : null)
const { snapshot, pollingError, track, refresh } = useModelFingerprint(accountId, value => {
  if (accountId.value) emit('update', accountId.value, value)
})
const models = ref<FingerprintModel[]>([])
const keys = ref<FingerprintKey[]>([])
const keyError = ref(false)
const loadingKeys = ref(false)
const savedOptions = ref<FingerprintOptions | null>(null)
const protocol = ref<FingerprintProtocol>('auto')
const effort = ref('')
const protocolOptions = computed(() => [
  { value: 'auto', label: t('admin.accounts.fingerprint.autoProtocol') },
  { value: 'chat', label: 'OpenAI Chat Completions' },
  { value: 'anthropic', label: 'Anthropic Messages' }
])
const protocolLabel = (value?: string) => protocolOptions.value.find(item => item.value === value)?.label || '-'
const effortOptions = computed(() => [
  { value: '', label: t('admin.accounts.fingerprint.defaultEffort') },
  ...(models.value.find(item => item.id === selectedModel.value)?.reasoning_levels || []).map(value => ({ value, label: value }))
])
const scheduleEnabled = ref(false)
const loadingSchedule = ref(false)
const savingSchedule = ref(false)
const nextRunAt = ref('')
const scheduleMessage = ref('')
const scheduleFailed = ref(false)
const loadingHistory = ref(false)
const historyError = ref(false)
const history = ref<FingerprintHistory>({ items: [], total: 0, page: 1, page_size: 10 })
const viewedSnapshot = ref<ModelFingerprintSnapshot | null>(null)
const resultElement = ref<HTMLElement | null>(null)
function viewHistory(value: ModelFingerprintSnapshot) {
  viewedSnapshot.value = value
  resultElement.value?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
}
const historyColumns = computed(() => [
  { key: 'started_at', label: t('admin.accounts.fingerprint.time') },
  { key: 'model', label: t('admin.accounts.selectTestModel') },
  { key: 'protocol', label: t('admin.accounts.fingerprint.protocolEffort') },
  { key: 'source', label: t('admin.accounts.fingerprint.source') },
  { key: 'result', label: t('admin.accounts.fingerprint.attribution') }
])
const selectedModel = ref('')
const loadingModels = ref(false)
const loadError = ref(false)
const starting = ref(false)
const submitError = ref('')
const busy = computed(() => snapshot.value?.status === 'running')
watch(selectedModel, () => { effort.value = '' }, { flush: 'sync' })

let historyGeneration = 0
async function loadHistory(page = 1) {
  const id = accountId.value
  const generation = ++historyGeneration
  if (!id) return
  loadingHistory.value = true
  historyError.value = false
  try {
    const result = await getFingerprintHistory(id, page)
    if (generation === historyGeneration && id === accountId.value) history.value = result
    if (viewedSnapshot.value && Date.now() - Date.parse(viewedSnapshot.value.started_at) >= 24 * 60 * 60 * 1000) viewedSnapshot.value = null
  } catch { if (generation === historyGeneration && id === accountId.value) historyError.value = true }
  finally { if (generation === historyGeneration) loadingHistory.value = false }
}

async function saveSchedule() {
  const id = accountId.value
  if (!id || savingSchedule.value || (scheduleEnabled.value && (!selectedKeyId.value || !selectedModel.value || loadingModels.value || loadError.value || keyError.value))) return
  savingSchedule.value = true
  scheduleMessage.value = ''
  try {
    const result = await setFingerprintSchedule(id, { enabled: scheduleEnabled.value, options: { api_key_id: selectedKeyId.value, model_id: selectedModel.value, protocol: protocol.value, reasoning_effort: effort.value } })
    if (id !== accountId.value) return
    nextRunAt.value = result.enabled ? result.next_run_at || '' : ''
    scheduleFailed.value = false
    scheduleMessage.value = t('admin.accounts.fingerprint.scheduleSaved')
  } catch (error) {
    if (id !== accountId.value) return
    scheduleFailed.value = true
    scheduleMessage.value = extractApiErrorMessage(error, t('admin.accounts.fingerprint.scheduleFailed'))
  } finally { savingSchedule.value = false }
}

watch(accountId, async (id, _, onCleanup) => {
  let stale = false
  onCleanup(() => { stale = true })
  models.value = []
  keys.value = []
  selectedKeyId.value = ''
  savedOptions.value = null
  keyError.value = false
  viewedSnapshot.value = null
  selectedModel.value = ''
  loadError.value = false
  submitError.value = ''
  protocol.value = 'auto'
  effort.value = ''
  scheduleEnabled.value = false
  nextRunAt.value = ''
  scheduleMessage.value = ''
  historyGeneration++
  history.value = { items: [], total: 0, page: 1, page_size: 10 }
  loadingKeys.value = Boolean(id)
  loadingSchedule.value = Boolean(id)
  if (!id) return
  loadingSchedule.value = true
  const schedulePromise = getFingerprintSchedule(id).then(value => {
    if (stale) return
    scheduleEnabled.value = value.enabled
    nextRunAt.value = value.enabled ? value.next_run_at || '' : ''
    savedOptions.value = value.options
    return value
  }).catch(() => {
    if (!stale) { scheduleFailed.value = true; scheduleMessage.value = t('admin.accounts.fingerprint.scheduleFailed') }
  }).finally(() => { if (!stale) loadingSchedule.value = false })
  void loadHistory()
  try {
    const [items, schedule] = await Promise.all([getFingerprintKeys(id), schedulePromise])
    if (stale) return
    keys.value = items
    const savedKey = schedule?.options?.api_key_id
    selectedKeyId.value = savedKey ? items.find(item => item.id === savedKey)?.id || '' : schedule?.enabled ? '' : items[0]?.id || ''
    if (schedule?.enabled && !selectedKeyId.value) {
      scheduleFailed.value = true
      scheduleMessage.value = t('admin.accounts.fingerprint.scheduleKeyRequired')
    }
  } catch { if (!stale) keyError.value = true }
  finally { if (!stale) loadingKeys.value = false }
}, { immediate: true })

watch([accountId, selectedKeyId], async ([id, keyID], _, onCleanup) => {
  let stale = false
  onCleanup(() => { stale = true })
  models.value = []
  selectedModel.value = ''
  loadError.value = false
  loadingModels.value = Boolean(id && keyID)
  if (!id || !keyID) return
  try {
    const items = await getFingerprintModels(id, keyID)
    if (stale) return
    models.value = items
    const previous = (props.account?.extra?.model_fingerprint as ModelFingerprintSnapshot | undefined)?.model
    selectedModel.value = models.value.find(item => item.id === previous)?.id || models.value[0]?.id || ''
    const options = savedOptions.value
    if (options?.api_key_id === keyID && options.model_id) {
      selectedModel.value = models.value.find(item => item.id === options.model_id)?.id || selectedModel.value
      protocol.value = options.protocol || 'auto'
      effort.value = effortOptions.value.some(item => item.value === options.reasoning_effort) ? options.reasoning_effort : ''
    }
  } catch { if (!stale) loadError.value = true }
  finally { if (!stale) loadingModels.value = false }
}, { immediate: true })

async function start() {
  const id = accountId.value
  if (!id || !selectedKeyId.value || keyError.value || loadingModels.value || busy.value || starting.value || !selectedModel.value) return
  starting.value = true
  viewedSnapshot.value = null
  submitError.value = ''
  try {
    const value = await startModelFingerprint(id, selectedModel.value, { api_key_id: selectedKeyId.value, protocol: protocol.value, reasoning_effort: effort.value })
    if (accountId.value === id) track(value)
    else emit('update', id, value)
    if (accountId.value === id) void loadHistory()
  } catch (error) {
    if (accountId.value === id) submitError.value = extractApiErrorMessage(error, t('admin.accounts.fingerprint.startFailed'))
  } finally { starting.value = false }
}
watch(() => snapshot.value?.status, (status, previous) => {
  if (previous === 'running' && status !== 'running') void loadHistory()
})
const historyTimer = setInterval(() => {
  if (accountId.value && !document.hidden) { void loadHistory(history.value.page); if (scheduleEnabled.value && !busy.value) void refresh() }
}, 30000)
onUnmounted(() => { clearInterval(historyTimer); historyGeneration++ })
</script>

<style scoped>
.fingerprint-dialog { display: grid; gap: 12px; min-width: 0; }
.fingerprint-account { font-weight: 600; overflow-wrap: anywhere; }
.fingerprint-label { font-size: var(--font-size-sm); margin-bottom: -6px; }
.fingerprint-meta { display: flex; flex-wrap: wrap; gap: 8px 18px; font-size: var(--font-size-xs); color: var(--color-text-tertiary); }
.fingerprint-output { border-top: 1px solid var(--color-border); padding-top: 16px; margin-top: 4px; }
.fingerprint-note { margin: 0; font-size: var(--font-size-xs); color: var(--color-text-tertiary); line-height: 1.6; }
.fingerprint-error { margin: 0; font-size: var(--font-size-sm); color: var(--color-text-danger); }
.fingerprint-schedule, .fingerprint-history-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 12px; }
.fingerprint-schedule label { display: flex; align-items: center; gap: 8px; }
.fingerprint-history-heading { border-top: 1px solid var(--color-border); padding-top: 16px; }
</style>
