<template>
  <section aria-labelledby="prompt-pool-title" class="features-prompt-audit-components-endpoint-pool__section">
    <div class="features-prompt-audit-components-endpoint-pool__panel">
      <div>
        <h2 id="prompt-pool-title" class="features-prompt-audit-components-endpoint-pool__heading">{{ t('admin.promptAudit.pool.title') }}</h2>
        <p class="features-prompt-audit-components-endpoint-pool__description">{{ t('admin.promptAudit.pool.description') }}</p>
      </div>
      <button type="button" class="btn btn-primary btn-sm" data-test="add-endpoint" @click="openCreate">
        {{ t('admin.promptAudit.pool.add') }}
      </button>
    </div>

    <div v-if="endpoints.length === 0" class="features-prompt-audit-components-endpoint-pool__panel-2">
      {{ t('admin.promptAudit.pool.empty') }}
    </div>
    <div v-else class="features-prompt-audit-components-endpoint-pool__panel-3">
      <div class="features-prompt-audit-components-endpoint-pool__panel-4">
        <span>{{ t('admin.promptAudit.pool.node') }}</span>
        <span>{{ t('admin.promptAudit.pool.model') }}</span>
        <span>{{ t('admin.promptAudit.pool.limits') }}</span>
        <span>{{ t('admin.promptAudit.pool.credential') }}</span>
        <span class="features-prompt-audit-components-endpoint-pool__text">{{ t('admin.promptAudit.common.actions') }}</span>
      </div>

      <div class="features-prompt-audit-components-endpoint-pool__panel-5">
        <article
          v-for="endpoint in endpoints"
          :key="endpoint.id"
          :data-test="`endpoint-${endpoint.id}`"
          class="features-prompt-audit-components-endpoint-pool__article"
        >
          <div class="features-prompt-audit-components-endpoint-pool__panel-6">
            <button
              type="button"
              role="switch"
              :aria-checked="endpoint.enabled"
              :aria-label="t('admin.promptAudit.pool.toggleNode', { name: endpoint.name })"
              class="features-prompt-audit-components-endpoint-pool__action"
              :class="endpoint.enabled ? 'features-prompt-audit-components-endpoint-pool__action-3' : 'features-prompt-audit-components-endpoint-pool__action-4'"
              @click="toggleEndpoint(endpoint.id)"
            >
              <span
                class="features-prompt-audit-components-endpoint-pool__text-2"
                :class="endpoint.enabled ? 'toggle-thumb--on' : 'features-prompt-audit-components-endpoint-pool__text-7'"
              />
            </button>
            <div class="features-prompt-audit-components-endpoint-pool__panel-7">
              <div class="features-prompt-audit-components-endpoint-pool__panel-8">
                <p class="features-prompt-audit-components-endpoint-pool__description-2">{{ endpoint.name }}</p>
                <span class="features-prompt-audit-components-endpoint-pool__text-3" :class="endpoint.enabled ? 'features-prompt-audit-components-endpoint-pool__text-8' : 'features-prompt-audit-components-endpoint-pool__text-9'" aria-hidden="true" />
              </div>
              <p class="features-prompt-audit-components-endpoint-pool__description-3" :title="endpoint.base_url">{{ endpoint.base_url }}</p>
            </div>
          </div>

          <div class="features-prompt-audit-components-endpoint-pool__panel-9">
            <p class="features-prompt-audit-components-endpoint-pool__description-4">{{ t('admin.promptAudit.pool.model') }}</p>
            <p class="features-prompt-audit-components-endpoint-pool__description-5" :title="endpoint.model">{{ endpoint.model }}</p>
          </div>

          <div>
            <p class="features-prompt-audit-components-endpoint-pool__description-4">{{ t('admin.promptAudit.pool.limits') }}</p>
            <div class="features-prompt-audit-components-endpoint-pool__panel-10">
              <span class="features-prompt-audit-components-endpoint-pool__text-4">{{ endpoint.timeout_ms }} ms</span>
              <span class="features-prompt-audit-components-endpoint-pool__text-4">{{ endpoint.input_limit }} chars</span>
            </div>
          </div>

          <div class="features-prompt-audit-components-endpoint-pool__panel-7">
            <p class="features-prompt-audit-components-endpoint-pool__description-4">{{ t('admin.promptAudit.pool.credential') }}</p>
            <div class="features-prompt-audit-components-endpoint-pool__panel-11" :class="credentialInvalid(endpoint) ? 'features-prompt-audit-components-endpoint-pool__panel-14' : hasCredential(endpoint) ? 'features-prompt-audit-components-endpoint-pool__panel-15' : 'features-prompt-audit-components-endpoint-pool__panel-16'">
              <span class="features-prompt-audit-components-endpoint-pool__text-5" :class="credentialInvalid(endpoint) ? 'features-prompt-audit-components-endpoint-pool__text-10' : hasCredential(endpoint) ? 'features-prompt-audit-components-endpoint-pool__text-8' : 'features-prompt-audit-components-endpoint-pool__text-9'" aria-hidden="true" />
              {{ credentialInvalid(endpoint) ? t('admin.promptAudit.pool.invalid') : hasCredential(endpoint) ? t('admin.promptAudit.pool.configured') : t('admin.promptAudit.pool.missing') }}
            </div>
            <p v-if="probingIds.includes(endpoint.id)" class="features-prompt-audit-components-endpoint-pool__description-6">
              {{ t('admin.promptAudit.pool.probeProgress') }}
            </p>
            <p v-if="probeResults[endpoint.id]" class="features-prompt-audit-components-endpoint-pool__description-7" :class="probeResults[endpoint.id].ok ? 'features-prompt-audit-components-endpoint-pool__description-8' : 'features-prompt-audit-components-endpoint-pool__panel-14'">
              {{ t('admin.promptAudit.pool.probeResult', { status: probeResults[endpoint.id].status, http: probeResults[endpoint.id].http_status || '—', latency: probeResults[endpoint.id].latency_ms }) }}
              · {{ probeResults[endpoint.id].message }}
            </p>
          </div>

          <div class="features-prompt-audit-components-endpoint-pool__panel-12">
            <button type="button" class="btn btn-secondary btn-sm" :disabled="probingIds.includes(endpoint.id)" @click="$emit('probe', endpoint)">
              {{ probingIds.includes(endpoint.id) ? t('admin.promptAudit.pool.probing') : t('admin.promptAudit.pool.probe') }}
            </button>
            <button type="button" class="btn btn-ghost btn-sm" @click="openEdit(endpoint)">{{ t('common.edit') }}</button>
            <button type="button" class="features-prompt-audit-components-endpoint-pool__action-2 btn btn-ghost btn-sm" @click="removeEndpoint(endpoint)">{{ t('common.delete') }}</button>
          </div>
        </article>
      </div>
    </div>

    <BaseDialog :show="Boolean(editing)" :title="editingIndex < 0 ? t('admin.promptAudit.pool.add') : t('admin.promptAudit.pool.edit')" width="wide" @close="closeEditor">
      <form v-if="editing" class="features-prompt-audit-components-endpoint-pool__form" @submit.prevent="saveEditor">
        <label class="features-prompt-audit-components-endpoint-pool__label">
          <span>{{ t('admin.promptAudit.pool.name') }}</span>
          <input v-model="editing.name" class="features-prompt-audit-components-endpoint-pool__field input" required :aria-label="t('admin.promptAudit.pool.name')" />
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label">
          <span>{{ t('admin.promptAudit.pool.id') }}</span>
          <input v-model="editing.id" class="features-prompt-audit-components-endpoint-pool__field input" required :disabled="editingIndex >= 0" :aria-label="t('admin.promptAudit.pool.id')" />
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label-2">
          <span>{{ t('admin.promptAudit.pool.baseUrl') }}</span>
          <input v-model="editing.base_url" class="features-prompt-audit-components-endpoint-pool__field input" required inputmode="url" :aria-label="t('admin.promptAudit.pool.baseUrl')" />
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label-2">
          <span>{{ t('admin.promptAudit.pool.apiKey') }}</span>
          <input v-model="editing.token" class="features-prompt-audit-components-endpoint-pool__field input" type="password" autocomplete="new-password" :placeholder="editing.has_token ? (editing.token_status === 'invalid' ? t('admin.promptAudit.pool.reenterSecret') : t('admin.promptAudit.pool.keepSecret')) : ''" :aria-label="t('admin.promptAudit.pool.apiKey')" />
          <span class="features-prompt-audit-components-endpoint-pool__text-6">{{ t('admin.promptAudit.pool.secretHint') }}</span>
        </label>
        <label v-if="editing.has_token" class="features-prompt-audit-components-endpoint-pool__label-3">
          <input v-model="editing.clear_token" type="checkbox" :aria-label="t('admin.promptAudit.pool.clearSecret')" />
          {{ t('admin.promptAudit.pool.clearSecret') }}
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label-2">
          <span>{{ t('admin.promptAudit.pool.model') }}</span>
          <input v-model="editing.model" class="features-prompt-audit-components-endpoint-pool__field input" :aria-label="t('admin.promptAudit.pool.model')" />
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label">
          <span>{{ t('admin.promptAudit.pool.timeout') }}</span>
          <input v-model.number="editing.timeout_ms" class="features-prompt-audit-components-endpoint-pool__field input" type="number" min="100" max="30000" required :aria-label="t('admin.promptAudit.pool.timeout')" />
        </label>
        <label class="features-prompt-audit-components-endpoint-pool__label">
          <span>{{ t('admin.promptAudit.pool.inputLimit') }}</span>
          <input v-model.number="editing.input_limit" class="features-prompt-audit-components-endpoint-pool__field input" type="number" min="128" max="100000" required :aria-label="t('admin.promptAudit.pool.inputLimit')" />
        </label>
      </form>
      <template #footer>
        <div class="features-prompt-audit-components-endpoint-pool__panel-13">
          <button type="button" class="btn btn-secondary" @click="closeEditor">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" data-test="save-endpoint" @click="saveEditor">{{ t('common.save') }}</button>
        </div>
      </template>
    </BaseDialog>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PromptAuditEndpointDraft, PromptProbeResult } from '../types'
import { cloneData, createDefaultEndpoint } from '../viewModel'

const props = defineProps<{
  endpoints: PromptAuditEndpointDraft[]
  probeResults: Record<string, PromptProbeResult>
  probingIds: string[]
}>()
const emit = defineEmits<{
  (event: 'update:endpoints', value: PromptAuditEndpointDraft[]): void
  (event: 'probe', endpoint: PromptAuditEndpointDraft): void
}>()
const { t } = useI18n()
const editing = ref<PromptAuditEndpointDraft | null>(null)
const editingIndex = ref(-1)

function openCreate() {
  editingIndex.value = -1
  editing.value = createDefaultEndpoint(props.endpoints.length + 1)
}
function openEdit(endpoint: PromptAuditEndpointDraft) {
  editingIndex.value = props.endpoints.findIndex((item) => item.id === endpoint.id)
  editing.value = cloneData(endpoint)
}
function closeEditor() {
  editing.value = null
  editingIndex.value = -1
}
function saveEditor() {
  if (!editing.value?.id.trim() || !editing.value.name.trim() || !editing.value.base_url.trim()) return
  const next = props.endpoints.map((item) => cloneData(item))
  const value = cloneData(editing.value)
  if (value.token.trim()) value.clear_token = false
  if (editingIndex.value < 0) next.push(value)
  else next.splice(editingIndex.value, 1, value)
  emit('update:endpoints', next)
  closeEditor()
}
function toggleEndpoint(id: string) {
  emit('update:endpoints', props.endpoints.map((item) => item.id === id ? { ...item, enabled: !item.enabled } : cloneData(item)))
}
function removeEndpoint(endpoint: PromptAuditEndpointDraft) {
  if (!window.confirm(t('admin.promptAudit.pool.deleteConfirm', { name: endpoint.name }))) return
  emit('update:endpoints', props.endpoints.filter((item) => item.id !== endpoint.id).map((item) => cloneData(item)))
}
function hasCredential(endpoint: PromptAuditEndpointDraft): boolean {
  return Boolean(endpoint.token.trim() || (endpoint.has_token && !endpoint.clear_token))
}
function credentialInvalid(endpoint: PromptAuditEndpointDraft): boolean {
  return endpoint.token_status === 'invalid' && !endpoint.token.trim() && !endpoint.clear_token
}
</script>
