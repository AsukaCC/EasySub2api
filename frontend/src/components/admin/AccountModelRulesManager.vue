<template>
  <div class="account-model-rules-modal account-model-rules-modal--page">
    <div class="account-model-rules-modal__toolbar">
      <div class="account-model-rules-modal__heading">
        <h1 class="account-model-rules-modal__title">
          {{ t('admin.accounts.modelRules.title') }}
        </h1>
        <p class="input-hint account-model-rules-modal__description">
          {{ t('admin.accounts.modelRules.description') }}
        </p>
      </div>
      <div class="account-model-rules-modal__toolbar-actions">
        <button
          type="button"
          class="account-model-rules-modal__refresh btn btn-secondary"
          :disabled="loading"
          :title="t('common.refresh')"
          @click="loadRules"
        >
          <Icon name="refresh" size="sm" :class="loading ? 'account-model-rules-modal__spin' : ''" />
        </button>
        <Select
          v-model="platformFilter"
          :options="platformFilterOptions"
          :placeholder="t('admin.accounts.modelRules.platform')"
          :clearable="true"
          class="account-model-rules-modal__filter"
        />
        <button type="button" class="btn btn-primary" @click="openCreate">
          <Icon name="plus" size="sm" />
          <span>{{ t('admin.accounts.modelRules.createRule') }}</span>
        </button>
      </div>
    </div>

    <div class="account-model-rules-modal__surface">
      <div v-if="loading" class="account-model-rules-modal__state">
        <Icon name="refresh" size="lg" class="account-model-rules-modal__spin" />
      </div>
      <div v-else-if="rules.length === 0" class="account-model-rules-modal__state">
        <Icon name="swap" size="lg" />
        <strong>{{ t('admin.accounts.modelRules.noRules') }}</strong>
        <span>{{ t('admin.accounts.modelRules.createFirst') }}</span>
      </div>
      <div v-else class="account-model-rules-modal__table-wrap">
        <table class="account-model-rules-modal__table">
          <thead>
            <tr>
              <th>{{ t('admin.accounts.modelRules.name') }}</th>
              <th>{{ t('admin.accounts.modelRules.platform') }}</th>
              <th>{{ t('admin.accounts.modelRules.mapping') }}</th>
              <th>{{ t('admin.accounts.modelRules.descriptionLabel') }}</th>
              <th>{{ t('admin.accounts.modelRules.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id">
              <td class="account-model-rules-modal__name">{{ rule.name }}</td>
              <td>
                <span class="account-model-rules-modal__platform">
                  <PlatformIcon :platform="rule.platform" size="xs" />
                  {{ platformLabel(rule.platform) }}
                </span>
              </td>
              <td>
                <span class="account-model-rules-modal__count">
                  {{ t('admin.accounts.modelRules.whitelistCount', { count: (rule.whitelist || []).length }) }}
                  ·
                  {{ t('admin.accounts.modelRules.mappingCount', { count: Object.keys(rule.mapping || {}).length }) }}
                </span>
                <span class="account-model-rules-modal__preview">{{ mappingPreview(rule) }}</span>
              </td>
              <td>{{ rule.description || '—' }}</td>
              <td>
                <div class="account-model-rules-modal__row-actions">
                  <button
                    type="button"
                    class="icon-btn"
                    :title="t('common.edit')"
                    @click="openEdit(rule)"
                  >
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    type="button"
                    class="icon-btn icon-btn--danger"
                    :title="t('common.delete')"
                    @click="openDelete(rule)"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <BaseDialog
    :show="showForm"
    :title="editingRule ? t('admin.accounts.modelRules.editRule') : t('admin.accounts.modelRules.createRule')"
    width="wide"
    @close="closeForm"
  >
    <form class="account-model-rules-modal__form" @submit.prevent="handleSubmit">
      <div class="account-model-rules-modal__form-grid">
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.name') }}</label>
          <input
            v-model="form.name"
            class="input"
            type="text"
            maxlength="100"
            required
            :placeholder="t('admin.accounts.modelRules.namePlaceholder')"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.accounts.modelRules.platform') }}</label>
          <Select v-model="form.platform" :options="platformOptions" />
        </div>
      </div>

      <div>
        <label class="input-label">{{ t('admin.accounts.modelRules.descriptionLabel') }}</label>
        <textarea
          v-model="form.description"
          class="input account-model-rules-modal__textarea"
          rows="2"
          :placeholder="t('admin.accounts.modelRules.descriptionPlaceholder')"
        />
      </div>

      <div class="account-model-rules-modal__mode-toggle" role="tablist">
        <button
          type="button"
          role="tab"
          :aria-selected="modelRestrictionMode === 'whitelist'"
          :class="['btn btn-sm', modelRestrictionMode === 'whitelist' ? 'btn-primary' : 'btn-secondary']"
          @click="modelRestrictionMode = 'whitelist'"
        >
          <Icon name="check" size="sm" />
          <span>{{ t('admin.accounts.modelRules.whitelist') }}</span>
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="modelRestrictionMode === 'mapping'"
          :class="['btn btn-sm', modelRestrictionMode === 'mapping' ? 'btn-primary' : 'btn-secondary']"
          @click="modelRestrictionMode = 'mapping'"
        >
          <Icon name="swap" size="sm" />
          <span>{{ t('admin.accounts.modelRules.mapping') }}</span>
        </button>
      </div>

      <div v-if="modelRestrictionMode === 'whitelist'" class="account-model-rules-modal__restriction-panel">
        <ModelWhitelistSelector
          v-model="form.whitelist"
          :platform="form.platform"
        />
        <p class="input-hint account-model-rules-modal__restriction-hint">
          {{ t('admin.accounts.modelRules.whitelistHint') }}
        </p>
      </div>

      <div v-else class="account-model-rules-modal__restriction-panel">
        <div class="account-model-rules-modal__mapping-header">
          <label class="input-label">{{ t('admin.accounts.modelRules.mapping') }}</label>
          <div class="account-model-rules-modal__mapping-actions">
            <button type="button" class="btn btn-secondary btn-sm" @click="openMappingModelImport">
              <Icon name="download" size="sm" />
              <span>{{ t('admin.accounts.modelRules.importLatestModels') }}</span>
            </button>
            <button type="button" class="btn btn-secondary btn-sm" @click="addMapping">
              <Icon name="plus" size="sm" />
              <span>{{ t('admin.accounts.modelRules.addMapping') }}</span>
            </button>
          </div>
        </div>
        <div class="account-model-rules-modal__mapping-list">
          <div
            v-for="(row, index) in form.mappings"
            :key="row.id"
            :class="[
              'account-model-rules-modal__mapping-row',
              form.platform === 'openai' && 'model-rules__reasoning-row'
            ]"
          >
            <input
              v-model="row.from"
              class="input"
              type="text"
              :placeholder="t('admin.accounts.modelRules.fromModel')"
            />
            <span aria-hidden="true">→</span>
            <input
              v-model="row.to"
              class="input"
              type="text"
              :placeholder="t('admin.accounts.modelRules.toModel')"
            />
            <Select
              v-if="form.platform === 'openai'"
              v-model="row.reasoning_effort"
              :options="reasoningEffortOptions"
              :placeholder="t('admin.accounts.modelRules.reasoningEffortFollowRequest')"
              :aria-label="t('admin.accounts.modelRules.reasoningEffort')"
              :searchable="false"
              clearable
            />
            <button
              type="button"
              class="icon-btn icon-btn--danger"
              :title="t('admin.accounts.modelRules.removeMapping')"
              :disabled="form.mappings.length === 1"
              @click="removeMapping(index)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </div>
      </div>
      <p v-if="formError" class="account-model-rules-modal__error">{{ formError }}</p>
    </form>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="closeForm">{{ t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
        <Icon v-if="submitting" name="refresh" size="sm" class="account-model-rules-modal__spin" />
        <span>{{ t('admin.accounts.modelRules.save') }}</span>
      </button>
    </template>
  </BaseDialog>

  <BaseDialog
    :show="showMappingModelImport"
    :title="t('admin.accounts.modelRules.importLatestModels')"
    width="wide"
    :z-index="60"
    @close="closeMappingModelImport"
  >
    <div class="account-model-rules-modal__mapping-import-dialog">
      <p class="input-hint account-model-rules-modal__mapping-import-hint">
        {{ t('admin.accounts.modelRules.importLatestModelsHint') }}
      </p>
      <div class="account-model-rules-modal__mapping-import-toolbar">
        <input
          v-model="mappingModelSearch"
          class="input"
          type="search"
          :placeholder="t('admin.accounts.modelRules.searchModels')"
          :disabled="mappingModelsLoading || mappingModelOptions.length === 0"
        />
        <div class="account-model-rules-modal__mapping-import-toolbar-actions">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="mappingModelsLoading || mappingModelOptions.length === 0"
            @click="selectAllMappingModels"
          >
            <Icon name="check" size="sm" />
            <span>{{ t('admin.accounts.modelRules.selectAllModels') }}</span>
          </button>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="mappingImportModels.length === 0"
            @click="clearMappingModelSelection"
          >
            <Icon name="x" size="sm" />
            <span>{{ t('admin.accounts.modelRules.clearModelSelection') }}</span>
          </button>
        </div>
      </div>
      <div v-if="mappingModelsLoading" class="account-model-rules-modal__mapping-import-state">
        <Icon name="refresh" size="lg" class="account-model-rules-modal__spin" />
        <span>{{ t('admin.accounts.modelRules.loadingModels') }}</span>
      </div>
      <div v-else-if="mappingModelLoadError" class="account-model-rules-modal__mapping-import-state account-model-rules-modal__mapping-import-state--error">
        <span>{{ mappingModelLoadError }}</span>
        <button type="button" class="btn btn-secondary btn-sm" @click="loadMappingModelOptions">
          <Icon name="refresh" size="sm" />
          <span>{{ t('admin.accounts.retry') }}</span>
        </button>
      </div>
      <div v-else-if="filteredMappingModelOptions.length === 0" class="account-model-rules-modal__mapping-import-state">
        <span>{{ t('admin.accounts.modelRules.noModels') }}</span>
      </div>
      <div v-else class="account-model-rules-modal__mapping-model-grid">
        <label
          v-for="model in filteredMappingModelOptions"
          :key="model"
          class="account-model-rules-modal__mapping-model-option"
        >
          <input v-model="mappingImportModels" type="checkbox" :value="model" />
          <span>{{ model }}</span>
        </label>
      </div>
      <p class="input-hint account-model-rules-modal__mapping-import-count">
        {{ t('admin.accounts.modelRules.selectedModelsCount', { count: mappingImportModels.length }) }}
      </p>
    </div>
    <template #footer>
      <button type="button" class="btn btn-secondary" @click="closeMappingModelImport">
        {{ t('common.cancel') }}
      </button>
      <button
        type="button"
        class="btn btn-primary"
        :disabled="mappingModelsLoading || mappingImportModels.length === 0"
        @click="importSelectedModels"
      >
        <Icon name="download" size="sm" />
        <span>{{ t('admin.accounts.modelRules.importSelectedModels') }}</span>
      </button>
    </template>
  </BaseDialog>

  <ConfirmDialog
    :show="showDeleteConfirm"
    :title="t('admin.accounts.modelRules.deleteRule')"
    :message="t('admin.accounts.modelRules.deleteConfirm', { name: deletingRule?.name || '' })"
    :confirm-text="t('common.delete')"
    :cancel-text="t('common.cancel')"
    :danger="true"
    @confirm="confirmDelete"
    @cancel="cancelDelete"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import ModelWhitelistSelector from '@/components/account/ModelWhitelistSelector.vue'
import { adminAPI } from '@/api/admin'
import type { AccountModelRule } from '@/api/admin/accountModelRules'
import type { AccountPlatform } from '@/types'
import { accountPlatformOptions } from '@/utils/accountPlatforms'
import { isValidWildcardPattern } from '@/composables/useModelWhitelist'
import { useAppStore } from '@/stores/app'

interface MappingRow {
  id: number
  from: string
  to: string
  reasoning_effort: string
}

const { t } = useI18n()
const appStore = useAppStore()

const rules = ref<AccountModelRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const platformFilter = ref<AccountPlatform | ''>('')
const showForm = ref(false)
const editingRule = ref<AccountModelRule | null>(null)
const deletingRule = ref<AccountModelRule | null>(null)
const showDeleteConfirm = ref(false)
const formError = ref('')
const showMappingModelImport = ref(false)
const mappingImportModels = ref<string[]>([])
const mappingModelOptions = ref<string[]>([])
const mappingModelSearch = ref('')
const mappingModelsLoading = ref(false)
const mappingModelLoadError = ref('')
let mappingModelLoadVersion = 0
let nextRowId = 1

const platformOptions = computed(() => accountPlatformOptions(t))
const platformFilterOptions = computed(() => [
  { value: '', label: t('admin.accounts.allPlatforms') },
  ...platformOptions.value
])
const reasoningEffortOptions = computed(() =>
  ['minimal', 'low', 'medium', 'high', 'xhigh', 'max'].map(value => ({ value, label: value }))
)

const form = reactive({
  name: '',
  description: '',
  platform: 'anthropic' as AccountPlatform,
  whitelist: [] as string[],
  mappings: [] as MappingRow[]
})

const modelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const filteredMappingModelOptions = computed(() => {
  const query = mappingModelSearch.value.trim().toLowerCase()
  if (!query) return mappingModelOptions.value
  return mappingModelOptions.value.filter(model => model.toLowerCase().includes(query))
})

const platformLabel = (platform: string) =>
  t(`admin.accounts.platforms.${platform}`, platform)

const mappingPreview = (rule: AccountModelRule) => {
  const entries = Object.entries(rule.mapping || {}).slice(0, 2)
  const preview = entries.map(([from, to]) => `${from} → ${to}`).join(', ')
  return Object.keys(rule.mapping || {}).length > 2 ? `${preview}, …` : preview
}

async function loadRules() {
  loading.value = true
  try {
    rules.value = await adminAPI.accountModelRules.list(platformFilter.value)
  } catch (error) {
    appStore.showError(t('admin.accounts.modelRules.loadFailed'))
    console.error('Failed to load account model rules:', error)
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  form.description = ''
  form.platform = 'anthropic'
  form.whitelist = []
  form.mappings = [{ id: nextRowId++, from: '', to: '', reasoning_effort: '' }]
  modelRestrictionMode.value = 'whitelist'
  showMappingModelImport.value = false
  mappingImportModels.value = []
  mappingModelOptions.value = []
  mappingModelSearch.value = ''
  mappingModelLoadError.value = ''
  formError.value = ''
}

function openCreate() {
  editingRule.value = null
  resetForm()
  showForm.value = true
}

function openEdit(rule: AccountModelRule) {
  editingRule.value = rule
  form.name = rule.name
  form.description = rule.description || ''
  form.platform = rule.platform
  form.whitelist = [...(rule.whitelist || [])]
  form.mappings = Object.entries(rule.mapping || {}).map(([from, to]) => ({
    id: nextRowId++,
    from,
    to,
    reasoning_effort: rule.reasoning_efforts?.[from] || ''
  }))
  if (form.mappings.length === 0) {
    form.mappings.push({ id: nextRowId++, from: '', to: '', reasoning_effort: '' })
  }
  modelRestrictionMode.value = Object.keys(rule.mapping || {}).length > 0 ? 'mapping' : 'whitelist'
  formError.value = ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingRule.value = null
  resetForm()
}

function addMapping() {
  form.mappings.push({ id: nextRowId++, from: '', to: '', reasoning_effort: '' })
}

function closeMappingModelImport() {
  mappingModelLoadVersion += 1
  showMappingModelImport.value = false
  mappingImportModels.value = []
  mappingModelSearch.value = ''
}

async function loadMappingModelOptions() {
  const requestVersion = ++mappingModelLoadVersion
  mappingModelsLoading.value = true
  mappingModelLoadError.value = ''
  mappingModelOptions.value = []
  try {
    const result = await adminAPI.channels.syncPricingModels(form.platform)
    const models = Array.from(new Set(result.models.map(model => model.trim()).filter(Boolean)))
    if (requestVersion !== mappingModelLoadVersion || !showMappingModelImport.value) return
    mappingModelOptions.value = models
    if (models.length === 0) {
      mappingModelLoadError.value = t('admin.accounts.modelRules.noModels')
    }
  } catch (error) {
    if (requestVersion !== mappingModelLoadVersion || !showMappingModelImport.value) return
    mappingModelLoadError.value = t('admin.accounts.modelRules.loadModelsFailed')
    console.error('Failed to load latest supported models for account model rule:', error)
  } finally {
    if (requestVersion === mappingModelLoadVersion) mappingModelsLoading.value = false
  }
}

function openMappingModelImport() {
  showMappingModelImport.value = true
  mappingImportModels.value = []
  mappingModelSearch.value = ''
  void loadMappingModelOptions()
}

function selectAllMappingModels() {
  mappingImportModels.value = [...mappingModelOptions.value]
}

function clearMappingModelSelection() {
  mappingImportModels.value = []
}

function importSelectedModels() {
  const models = Array.from(new Set(mappingImportModels.value.map(model => model.trim()).filter(Boolean)))
  if (models.length === 0) {
    appStore.showInfo(t('admin.accounts.modelRules.importModelsEmpty'))
    return
  }

  const existingSources = new Set(form.mappings.map(row => row.from.trim()).filter(Boolean))
  if (form.mappings.length === 1 && !form.mappings[0].from.trim() && !form.mappings[0].to.trim()) {
    form.mappings.splice(0, 1)
  }

  let addedCount = 0
  for (const model of models) {
    if (existingSources.has(model)) continue
    form.mappings.push({ id: nextRowId++, from: model, to: model, reasoning_effort: '' })
    existingSources.add(model)
    addedCount += 1
  }

  if (addedCount === 0) {
    appStore.showInfo(t('admin.accounts.modelRules.importModelsNoChanges'))
  } else {
    appStore.showSuccess(t('admin.accounts.modelRules.importModelsSuccess', { count: addedCount }))
  }
  closeMappingModelImport()
}

function removeMapping(index: number) {
  if (form.mappings.length <= 1) return
  form.mappings.splice(index, 1)
}

function buildWhitelist(): string[] | null {
  const whitelist: string[] = []
  for (const rawModel of form.whitelist) {
    const model = rawModel.trim()
    if (!model) continue
    if (model.includes('*')) {
      formError.value = t('admin.accounts.modelRules.whitelistRequired')
      return null
    }
    if (!whitelist.includes(model)) whitelist.push(model)
  }
  return whitelist
}

function buildMapping(): Record<string, string> | null {
  const mapping: Record<string, string> = {}
  for (const row of form.mappings) {
    const from = row.from.trim()
    const to = row.to.trim()
    if (!from && !to) continue
    if (!from || !to) {
      formError.value = t('admin.accounts.modelRules.mappingRequired')
      return null
    }
    if (!isValidWildcardPattern(from) || to.includes('*')) {
      formError.value = t('admin.accounts.modelRules.mappingRequired')
      return null
    }
    if (mapping[from]) {
      formError.value = t('admin.accounts.modelRules.duplicateSource')
      return null
    }
    mapping[from] = to
  }
  return mapping
}

function buildReasoningEfforts(mapping: Record<string, string>): Record<string, string> {
  if (form.platform !== 'openai') return {}
  const reasoningEfforts: Record<string, string> = {}
  for (const row of form.mappings) {
    const from = row.from.trim()
    const effort = row.reasoning_effort.trim().toLowerCase()
    if (from && mapping[from] && effort) reasoningEfforts[from] = effort
  }
  return reasoningEfforts
}

async function handleSubmit() {
  formError.value = ''
  if (!form.name.trim()) {
    formError.value = t('admin.accounts.modelRules.nameRequired')
    return
  }
  const mapping = buildMapping()
  if (!mapping) return
  const whitelist = buildWhitelist()
  if (!whitelist) return
  if (Object.keys(mapping).length === 0 && whitelist.length === 0) {
    formError.value = t('admin.accounts.modelRules.restrictionRequired')
    return
  }

  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description.trim() || null,
      platform: form.platform,
      whitelist,
      mapping,
      reasoning_efforts: buildReasoningEfforts(mapping)
    }
    if (editingRule.value) {
      await adminAPI.accountModelRules.update(editingRule.value.id, payload)
      appStore.showSuccess(t('admin.accounts.modelRules.updateSuccess'))
    } else {
      await adminAPI.accountModelRules.create(payload)
      appStore.showSuccess(t('admin.accounts.modelRules.createSuccess'))
    }
    closeForm()
    await loadRules()
  } catch (error: any) {
    const message = error?.response?.data?.detail || error?.message
    appStore.showError(message || t('admin.accounts.modelRules.saveFailed'))
    console.error('Failed to save account model rule:', error)
  } finally {
    submitting.value = false
  }
}

function openDelete(rule: AccountModelRule) {
  deletingRule.value = rule
  showDeleteConfirm.value = true
}

function cancelDelete() {
  showDeleteConfirm.value = false
  deletingRule.value = null
}

async function confirmDelete() {
  if (!deletingRule.value) return
  try {
    await adminAPI.accountModelRules.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.accounts.modelRules.deleteSuccess'))
    cancelDelete()
    await loadRules()
  } catch (error: any) {
    const message = error?.response?.data?.detail || error?.message
    appStore.showError(message || t('admin.accounts.modelRules.deleteFailed'))
    console.error('Failed to delete account model rule:', error)
  }
}

onMounted(() => void loadRules())
watch(platformFilter, () => void loadRules())
watch(
  () => form.platform,
  () => {
    if (showForm.value) {
      modelRestrictionMode.value = 'whitelist'
      form.whitelist = []
      form.mappings = [{ id: nextRowId++, from: '', to: '', reasoning_effort: '' }]
      closeMappingModelImport()
    }
  },
  { flush: 'sync' }
)
</script>

<style scoped>
.account-model-rules-modal {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.account-model-rules-modal--page {
  height: calc(100vh - var(--app-shell-height) - 3.25rem);
  min-height: 0;
}

.account-model-rules-modal__toolbar,
.account-model-rules-modal__toolbar-actions,
.account-model-rules-modal__mapping-header,
.account-model-rules-modal__mode-toggle,
.account-model-rules-modal__row-actions,
.account-model-rules-modal__platform,
.account-model-rules-modal__count {
  display: flex;
  align-items: center;
}

.account-model-rules-modal__toolbar {
  justify-content: space-between;
  gap: 1rem;
}

.account-model-rules-modal__heading {
  min-width: 0;
}

.account-model-rules-modal__title {
  margin: 0 0 0.25rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-xl);
  font-weight: 650;
  letter-spacing: 0;
}

.account-model-rules-modal__description {
  margin: 0;
  max-width: 58rem;
}

.account-model-rules-modal__toolbar-actions {
  gap: 0.5rem;
  flex-shrink: 0;
}

.account-model-rules-modal__refresh {
  width: 2.5rem;
  height: 2.5rem;
  flex: 0 0 2.5rem;
  padding: 0;
}

.account-model-rules-modal__filter {
  min-width: 12rem;
}

.account-model-rules-modal__surface {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  background: var(--glass-bg);
  box-shadow: var(--shadow-sm);
}

.account-model-rules-modal__state {
  min-height: 12rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-secondary);
  text-align: center;
}

.account-model-rules-modal__table-wrap {
  height: 100%;
  overflow-x: auto;
  overflow-y: auto;
}

.account-model-rules-modal__table {
  width: 100%;
  border-collapse: collapse;
  min-width: 720px;
}

.account-model-rules-modal__table th,
.account-model-rules-modal__table td {
  padding: 0.7rem 0.6rem;
  border-bottom: 1px solid var(--color-border);
  text-align: left;
  vertical-align: top;
}

.account-model-rules-modal__table th {
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-weight: 600;
}

.account-model-rules-modal__name {
  font-weight: 600;
}

.account-model-rules-modal__platform {
  gap: 0.35rem;
  white-space: nowrap;
}

.account-model-rules-modal__count {
  gap: 0.35rem;
  white-space: nowrap;
}

.account-model-rules-modal__preview {
  display: block;
  max-width: 20rem;
  overflow: hidden;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-model-rules-modal__row-actions {
  gap: 0.35rem;
}

.icon-btn {
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: 5px;
  background: transparent;
  color: inherit;
  cursor: pointer;
}

.icon-btn:hover:not(:disabled) {
  background: var(--color-surface-hover);
}

.icon-btn--danger {
  color: var(--color-text-danger);
}

.icon-btn:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.account-model-rules-modal__form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.account-model-rules-modal__form-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(12rem, 0.65fr);
  gap: 1rem;
}

.account-model-rules-modal__textarea {
  resize: vertical;
}

.account-model-rules-modal__mode-toggle {
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}

.account-model-rules-modal__restriction-panel {
  min-height: 10rem;
}

.account-model-rules-modal__mapping-header {
  justify-content: space-between;
  gap: 1rem;
}

.account-model-rules-modal__mapping-actions,
.account-model-rules-modal__mapping-import-toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.account-model-rules-modal__mapping-import-dialog {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 18rem;
}

.account-model-rules-modal__mapping-import-hint,
.account-model-rules-modal__mapping-import-count {
  margin: 0;
}

.account-model-rules-modal__mapping-import-toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.account-model-rules-modal__mapping-import-toolbar > .input {
  min-width: 0;
  flex: 1;
}

.account-model-rules-modal__mapping-import-state {
  display: flex;
  flex: 1;
  min-height: 12rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  color: var(--color-text-secondary);
  text-align: center;
}

.account-model-rules-modal__mapping-import-state--error {
  color: var(--color-text-danger);
}

.account-model-rules-modal__mapping-model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(15rem, 1fr));
  gap: 0.5rem;
  max-height: min(26rem, 50vh);
  overflow-y: auto;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
}

.account-model-rules-modal__mapping-model-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  padding: 0.55rem 0.65rem;
  border: 1px solid transparent;
  border-radius: 5px;
  cursor: pointer;
  overflow-wrap: anywhere;
}

.account-model-rules-modal__mapping-model-option:hover {
  border-color: var(--color-border);
  background: var(--color-surface-hover);
}

.account-model-rules-modal__mapping-header .input-label {
  margin: 0;
}

.account-model-rules-modal__mapping-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.account-model-rules-modal__mapping-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
}

.model-rules__reasoning-row {
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) minmax(10rem, 0.65fr) auto;
}

.account-model-rules-modal__error {
  margin: 0;
  color: var(--color-text-danger);
  font-size: var(--font-size-sm);
}

.account-model-rules-modal__spin {
  animation: account-model-rules-spin 0.9s linear infinite;
}

@keyframes account-model-rules-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 720px) {
  .account-model-rules-modal--page {
    height: auto;
    min-height: calc(100vh - var(--app-shell-height) - 3.25rem);
  }

  .account-model-rules-modal__toolbar,
  .account-model-rules-modal__toolbar-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .account-model-rules-modal__filter {
    min-width: 0;
  }

  .account-model-rules-modal__form-grid {
    grid-template-columns: 1fr;
  }

  .account-model-rules-modal__mapping-row {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .account-model-rules-modal__mapping-row > :nth-child(2) {
    display: none;
  }

  .account-model-rules-modal__mapping-header,
  .account-model-rules-modal__mapping-import-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .account-model-rules-modal__mapping-actions,
  .account-model-rules-modal__mapping-import-toolbar-actions {
    width: 100%;
  }

  .account-model-rules-modal__mapping-actions > .btn,
  .account-model-rules-modal__mapping-import-toolbar-actions > .btn {
    flex: 1;
  }
}
</style>
