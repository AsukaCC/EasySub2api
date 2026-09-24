<template>
  <fieldset class="protection-panel" :disabled="busy">
    <div class="protection-heading">
      <label class="protection-toggle">
        <input type="checkbox" :checked="current.anti_degradation === true" @change="requestToggle" />
        <span>{{ t('admin.accounts.protection.title') }}</span>
      </label>
      <span v-if="current.anti_degradation">{{ modeName(current.protection_mode || 'legacy') }}</span>
    </div>
    <template v-if="!compact">
      <div class="protection-controls">
        <label>
          <span class="input-label">{{ t('admin.accounts.protection.strategy') }}</span>
          <Select v-model="selected" :options="strategyOptions" :searchable="false" />
        </label>
        <button type="button" class="btn btn-secondary" @click="previewStrategy">
          <Icon name="shield" size="sm" />{{ t('admin.accounts.protection.preview') }}
        </button>
        <label v-if="current.platform === 'openai'">
          <span class="input-label">{{ t('admin.accounts.protection.integrity') }}</span>
          <Select :model-value="integrity" :options="integrityOptions" :searchable="false" @update:model-value="saveIntegrity" />
        </label>
      </div>
      <label class="protection-toggle">
        <input v-model="diagnostics" type="checkbox" />{{ t('admin.accounts.protection.diagnostics') }}
      </label>
      <dl v-if="runtime" class="protection-runtime">
        <dt>{{ t('admin.accounts.protection.configuredTLS') }}</dt><dd>{{ runtime.configured_tls }}</dd>
        <dt>{{ t('admin.accounts.protection.effectiveTLS') }}</dt><dd>{{ runtime.effective_tls }}</dd>
        <dt>{{ t('admin.accounts.protection.observation') }}</dt><dd>{{ t('admin.accounts.protection.' + (runtime.observed ? 'observed' : 'unverified')) }}</dd>
        <dt>{{ t('admin.accounts.protection.concurrency') }}</dt><dd>{{ runtime.concurrency }}</dd>
      </dl>
      <p v-if="runtime?.tls_reason" class="text-sm text-amber-600">{{ runtime.tls_reason }}</p>
    </template>
    <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
  </fieldset>
  <ConfirmDialog :show="confirm !== null" :title="t('admin.accounts.protection.title')"
    :message="t('admin.accounts.protection.' + (confirm === 'disable' ? 'disableConfirm' : 'applyConfirm'))"
    :danger="confirm === 'disable'" @cancel="confirm = null" @confirm="commit">
    <div v-if="preview" class="protection-preview">
      <p v-if="preview.reason">{{ preview.reason }}</p>
      <p v-for="issue in preview.issues || []" :key="issue" class="text-amber-600">{{ issue }}</p>
      <div v-for="change in preview.changes" :key="change.key">
        <strong>{{ change.key }}</strong>: {{ formatValue(change.from) }} → {{ formatValue(change.to) }}
      </div>
    </div>
  </ConfirmDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'
import { accountProtection, type IntegrityMode, type ProtectionPreview, type ProtectionStrategy } from '@/api/admin/accountProtection'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ account: Account; compact?: boolean }>()
const emit = defineEmits<{ updated: [account: Account]; busy: [value: boolean] }>()
const { t } = useI18n()
const current = ref(props.account)
const selected = ref('legacy')
const strategies = ref<ProtectionStrategy[]>([])
const diagnostics = ref(false)
const busy = ref(false)
const error = ref('')
const confirm = ref<'apply' | 'enable' | 'disable' | null>(null)
const preview = ref<ProtectionPreview | null>(null)
const runtime = ref<ProtectionPreview['runtime']>()
const integrityModes: IntegrityMode[] = ['off', 'observe', 'enforce']
const integrity = computed((): IntegrityMode => {
  const mode = current.value.extra?.request_integrity_mode ?? runtime.value?.integrity_mode
  return mode === 'observe' || mode === 'enforce' ? mode : 'off'
})
const eligibleStrategies = computed(() => strategies.value.filter(s =>
  s.apply_supported && (!s.diagnostic_only || diagnostics.value) &&
  (!s.requires_openai || (current.value.platform === 'openai' && ['oauth', 'setup-token'].includes(current.value.type)))
))
const strategyOptions = computed(() => eligibleStrategies.value.map(strategy => ({
  value: strategy.id,
  label: modeName(strategy.id),
})))
const integrityOptions = computed(() => integrityModes.map(mode => ({
  value: mode,
  label: t('admin.accounts.protection.' + mode),
})))
const modeName = (mode: string) => t('admin.accounts.protection.modes.' + mode)
const formatValue = (value: unknown) => value == null ? '-' : typeof value === 'object' ? JSON.stringify(value) : String(value)

async function run(action: () => Promise<void>) {
  if (busy.value) return
  busy.value = true
  emit('busy', true)
  error.value = ''
  try { await action() } catch (e) { error.value = (e as { message?: string }).message || t('common.error') }
  finally { busy.value = false; emit('busy', false) }
}
async function refresh() {
  const id = current.value.id
  const result = await accountProtection.preview(id)
  if (current.value.id === id) runtime.value = result.runtime
}
watch(() => props.account, account => { current.value = account }, { immediate: true })
watch(() => props.account.id, () => {
  confirm.value = null
  preview.value = null
  if (!props.compact) void run(async () => { strategies.value = await accountProtection.strategies(); await refresh() })
}, { immediate: true })
function requestToggle(event: Event) {
  (event.target as HTMLInputElement).checked = current.value.anti_degradation === true
  preview.value = null
  confirm.value = current.value.anti_degradation ? 'disable' : 'enable'
}
async function previewStrategy() {
  await run(async () => {
    preview.value = await accountProtection.preview(current.value.id, selected.value)
    if (preview.value.eligible) confirm.value = 'apply'
    else error.value = preview.value.reason || ''
  })
}
async function commit() {
  const action = confirm.value
  if (!action) return
  await run(async () => {
    const id = current.value.id
    const updated = action === 'apply'
      ? await accountProtection.apply(id, selected.value)
      : await accountProtection.set(id, action === 'enable')
    if (current.value.id === id) current.value = updated
    confirm.value = null
    emit('updated', updated)
    if (!props.compact) await refresh()
  })
}
async function saveIntegrity(mode: string | number | boolean | null) {
  const next = String(mode ?? 'off') as IntegrityMode
  await run(async () => {
    const id = current.value.id
    const updated = await accountProtection.integrity(id, next)
    if (current.value.id === id) current.value = updated
    emit('updated', updated)
    await refresh()
  })
}
</script>

<style scoped>
.protection-panel { min-width: 0; border: 0; padding: 0; margin: 0; }
.protection-heading, .protection-toggle { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.protection-heading { justify-content: space-between; }
.protection-controls { display: flex; gap: 12px; align-items: end; flex-wrap: wrap; margin: 12px 0; }
.protection-controls label { flex: 1 1 180px; min-width: 0; }
.protection-runtime { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 6px 12px; margin-top: 12px; font-size: 13px; }
.protection-runtime dd { margin: 0; overflow-wrap: anywhere; }
.protection-preview { overflow-wrap: anywhere; font-size: 13px; }
</style>
