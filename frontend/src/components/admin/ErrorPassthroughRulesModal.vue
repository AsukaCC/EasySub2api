<template>
  <BaseDialog
    :show="show"
    :title="t('admin.errorPassthrough.title')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div class="components-admin-error-passthrough-rules-modal__panel">
      <!-- Header -->
      <div class="components-admin-error-passthrough-rules-modal__panel-2">
        <p class="components-admin-error-passthrough-rules-modal__description">
          {{ t('admin.errorPassthrough.description') }}
        </p>
        <button @click="showCreateModal = true" class="btn btn-primary btn-sm">
          <Icon name="plus" size="sm" class="components-admin-error-passthrough-rules-modal__icon" />
          {{ t('admin.errorPassthrough.createRule') }}
        </button>
      </div>

      <!-- Rules Table -->
      <div v-if="loading" class="components-admin-error-passthrough-rules-modal__panel-3">
        <Icon name="refresh" size="lg" class="components-admin-error-passthrough-rules-modal__icon-2" />
      </div>

      <div v-else-if="rules.length === 0" class="components-admin-error-passthrough-rules-modal__panel-4">
        <div class="components-admin-error-passthrough-rules-modal__panel-5">
          <Icon name="shield" size="lg" class="components-admin-error-passthrough-rules-modal__icon-3" />
        </div>
        <h4 class="components-admin-error-passthrough-rules-modal__heading">
          {{ t('admin.errorPassthrough.noRules') }}
        </h4>
        <p class="components-admin-error-passthrough-rules-modal__description">
          {{ t('admin.errorPassthrough.createFirstRule') }}
        </p>
      </div>

      <div v-else class="components-admin-error-passthrough-rules-modal__panel-6">
        <table class="components-admin-error-passthrough-rules-modal__table">
          <thead class="components-admin-error-passthrough-rules-modal__header">
            <tr>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.priority') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.name') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.conditions') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.platforms') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.behavior') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.status') }}
              </th>
              <th class="components-admin-error-passthrough-rules-modal__heading-2">
                {{ t('admin.errorPassthrough.columns.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="components-admin-error-passthrough-rules-modal__body">
            <tr v-for="rule in rules" :key="rule.id" class="components-admin-error-passthrough-rules-modal__row">
              <td class="components-admin-error-passthrough-rules-modal__cell">
                <span class="components-admin-error-passthrough-rules-modal__text">
                  {{ rule.priority }}
                </span>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <div class="components-admin-error-passthrough-rules-modal__panel-7">{{ rule.name }}</div>
                <div v-if="rule.description" class="components-admin-error-passthrough-rules-modal__panel-8">
                  {{ rule.description }}
                </div>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <div class="components-admin-error-passthrough-rules-modal__panel-9">
                  <span
                    v-for="code in rule.error_codes.slice(0, 3)"
                    :key="code"
                    class="components-admin-error-passthrough-rules-modal__text-2 badge badge-danger"
                  >
                    {{ code }}
                  </span>
                  <span
                    v-if="rule.error_codes.length > 3"
                    class="components-admin-error-passthrough-rules-modal__text-3"
                  >
                    +{{ rule.error_codes.length - 3 }}
                  </span>
                  <span
                    v-for="keyword in rule.keywords.slice(0, 1)"
                    :key="keyword"
                    class="components-admin-error-passthrough-rules-modal__text-2 badge badge-gray"
                  >
                    "{{ keyword.length > 10 ? keyword.substring(0, 10) + '...' : keyword }}"
                  </span>
                  <span
                    v-if="rule.keywords.length > 1"
                    class="components-admin-error-passthrough-rules-modal__text-3"
                  >
                    +{{ rule.keywords.length - 1 }}
                  </span>
                </div>
                <div class="components-admin-error-passthrough-rules-modal__panel-10">
                  {{ t('admin.errorPassthrough.matchMode.' + rule.match_mode) }}
                </div>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <div v-if="rule.platforms.length === 0" class="components-admin-error-passthrough-rules-modal__panel-11">
                  {{ t('admin.errorPassthrough.allPlatforms') }}
                </div>
                <div v-else class="components-admin-error-passthrough-rules-modal__panel-12">
                  <span
                    v-for="platform in rule.platforms.slice(0, 2)"
                    :key="platform"
                    class="components-admin-error-passthrough-rules-modal__text-2 badge badge-primary"
                  >
                    {{ platform }}
                  </span>
                  <span v-if="rule.platforms.length > 2" class="components-admin-error-passthrough-rules-modal__text-3">
                    +{{ rule.platforms.length - 2 }}
                  </span>
                </div>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <div class="components-admin-error-passthrough-rules-modal__panel-13">
                  <div class="components-admin-error-passthrough-rules-modal__panel-14">
                    <Icon
                      :name="rule.passthrough_code ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="rule.passthrough_code ? 'components-admin-error-passthrough-rules-modal__icon-6' : 'components-admin-error-passthrough-rules-modal__icon-3'"
                    />
                    <span class="components-admin-error-passthrough-rules-modal__text-4">
                      {{ t('admin.errorPassthrough.code') }}:
                      {{ rule.passthrough_code ? t('admin.errorPassthrough.passthrough') : (rule.response_code || '-') }}
                    </span>
                  </div>
                  <div class="components-admin-error-passthrough-rules-modal__panel-14">
                    <Icon
                      :name="rule.passthrough_body ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="rule.passthrough_body ? 'components-admin-error-passthrough-rules-modal__icon-6' : 'components-admin-error-passthrough-rules-modal__icon-3'"
                    />
                    <span class="components-admin-error-passthrough-rules-modal__text-4">
                      {{ t('admin.errorPassthrough.body') }}:
                      {{ rule.passthrough_body ? t('admin.errorPassthrough.passthrough') : t('admin.errorPassthrough.custom') }}
                    </span>
                  </div>
                  <div v-if="rule.skip_monitoring" class="components-admin-error-passthrough-rules-modal__panel-14">
                    <Icon
                      name="checkCircle"
                      size="xs"
                      class="components-admin-error-passthrough-rules-modal__icon-4"
                    />
                    <span class="components-admin-error-passthrough-rules-modal__text-4">
                      {{ t('admin.errorPassthrough.skipMonitoring') }}
                    </span>
                  </div>
                </div>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <button
                  @click="toggleEnabled(rule)"
                  :class="[
                    'components-admin-error-passthrough-rules-modal__action-3',
                    rule.enabled ? 'components-admin-error-passthrough-rules-modal__action-4' : 'components-admin-error-passthrough-rules-modal__action-5'
                  ]"
                >
                  <span
                    :class="[
                      'components-admin-error-passthrough-rules-modal__text-7',
                      rule.enabled ? 'toggle-thumb--on' : 'components-admin-error-passthrough-rules-modal__text-8'
                    ]"
                  />
                </button>
              </td>
              <td class="components-admin-error-passthrough-rules-modal__cell-2">
                <div class="components-admin-error-passthrough-rules-modal__panel-14">
                  <button
                    @click="handleEdit(rule)"
                    class="components-admin-error-passthrough-rules-modal__action"
                    :title="t('common.edit')"
                  >
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    @click="handleDelete(rule)"
                    class="components-admin-error-passthrough-rules-modal__action-2"
                    :title="t('common.delete')"
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

    <template #footer>
      <div class="components-admin-error-passthrough-rules-modal__panel-15">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('admin.errorPassthrough.editRule') : t('admin.errorPassthrough.createRule')"
      width="wide"
      @close="closeFormModal"
    >
      <form @submit.prevent="handleSubmit" class="components-admin-error-passthrough-rules-modal__panel">
        <!-- Basic Info -->
        <div class="components-admin-error-passthrough-rules-modal__panel-16">
          <div>
            <label class="input-label">{{ t('admin.errorPassthrough.form.name') }}</label>
            <input
              v-model="form.name"
              type="text"
              required
              class="input"
              :placeholder="t('admin.errorPassthrough.form.namePlaceholder')"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.errorPassthrough.form.priority') }}</label>
            <input
              v-model.number="form.priority"
              type="number"
              min="0"
              class="input"
            />
            <p class="input-hint">{{ t('admin.errorPassthrough.form.priorityHint') }}</p>
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.errorPassthrough.form.description') }}</label>
          <input
            v-model="form.description"
            type="text"
            class="input"
            :placeholder="t('admin.errorPassthrough.form.descriptionPlaceholder')"
          />
        </div>

        <!-- Match Conditions -->
        <div class="components-admin-error-passthrough-rules-modal__panel-17">
          <h4 class="components-admin-error-passthrough-rules-modal__heading-3">
            {{ t('admin.errorPassthrough.form.matchConditions') }}
          </h4>

          <div class="components-admin-error-passthrough-rules-modal__panel-18">
            <div>
              <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.errorCodes') }}</label>
              <input
                v-model="errorCodesInput"
                type="text"
                class="components-admin-error-passthrough-rules-modal__field input"
                :placeholder="t('admin.errorPassthrough.form.errorCodesPlaceholder')"
              />
              <p class="components-admin-error-passthrough-rules-modal__text-2 input-hint">{{ t('admin.errorPassthrough.form.errorCodesHint') }}</p>
            </div>
            <div>
              <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.keywords') }}</label>
              <textarea
                v-model="keywordsInput"
                rows="2"
                class="components-admin-error-passthrough-rules-modal__field-2 input"
                :placeholder="t('admin.errorPassthrough.form.keywordsPlaceholder')"
              />
              <p class="components-admin-error-passthrough-rules-modal__text-2 input-hint">{{ t('admin.errorPassthrough.form.keywordsHint') }}</p>
            </div>
          </div>

          <div class="components-admin-error-passthrough-rules-modal__panel-19">
            <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.matchMode') }}</label>
            <div class="components-admin-error-passthrough-rules-modal__panel-20">
              <label
                v-for="option in matchModeOptions"
                :key="option.value"
                class="components-admin-error-passthrough-rules-modal__label"
              >
                <input
                  type="radio"
                  :value="option.value"
                  v-model="form.match_mode"
                  class="components-admin-error-passthrough-rules-modal__field-3"
                />
                <div class="components-admin-error-passthrough-rules-modal__panel-21">
                  <span class="components-admin-error-passthrough-rules-modal__text-5">{{ option.label }}</span>
                  <p class="components-admin-error-passthrough-rules-modal__panel-11">{{ option.description }}</p>
                </div>
              </label>
            </div>
          </div>

          <div class="components-admin-error-passthrough-rules-modal__panel-19">
            <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.platforms') }}</label>
            <div class="components-admin-error-passthrough-rules-modal__panel-22">
              <label
                v-for="platform in platformOptions"
                :key="platform.value"
                class="components-admin-error-passthrough-rules-modal__label-2"
              >
                <input
                  type="checkbox"
                  :value="platform.value"
                  v-model="form.platforms"
                  class="components-admin-error-passthrough-rules-modal__field-4"
                />
                <span class="components-admin-error-passthrough-rules-modal__text-6">{{ platform.label }}</span>
              </label>
            </div>
            <p class="components-admin-error-passthrough-rules-modal__description-2 input-hint">{{ t('admin.errorPassthrough.form.platformsHint') }}</p>
          </div>
        </div>

        <!-- Response Behavior -->
        <div class="components-admin-error-passthrough-rules-modal__panel-17">
          <h4 class="components-admin-error-passthrough-rules-modal__heading-3">
            {{ t('admin.errorPassthrough.form.responseBehavior') }}
          </h4>

          <div class="components-admin-error-passthrough-rules-modal__panel-18">
            <div>
              <label class="components-admin-error-passthrough-rules-modal__label-3">
                <input
                  type="checkbox"
                  v-model="form.passthrough_code"
                  class="components-admin-error-passthrough-rules-modal__field-4"
                />
                <span class="components-admin-error-passthrough-rules-modal__text-5">
                  {{ t('admin.errorPassthrough.form.passthroughCode') }}
                </span>
              </label>
              <div v-if="!form.passthrough_code" class="components-admin-error-passthrough-rules-modal__panel-23">
                <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.responseCode') }}</label>
                <input
                  v-model.number="form.response_code"
                  type="number"
                  min="100"
                  max="599"
                  class="components-admin-error-passthrough-rules-modal__field input"
                  placeholder="422"
                />
              </div>
            </div>
            <div>
              <label class="components-admin-error-passthrough-rules-modal__label-3">
                <input
                  type="checkbox"
                  v-model="form.passthrough_body"
                  class="components-admin-error-passthrough-rules-modal__field-4"
                />
                <span class="components-admin-error-passthrough-rules-modal__text-5">
                  {{ t('admin.errorPassthrough.form.passthroughBody') }}
                </span>
              </label>
              <div v-if="!form.passthrough_body" class="components-admin-error-passthrough-rules-modal__panel-23">
                <label class="components-admin-error-passthrough-rules-modal__text-2 input-label">{{ t('admin.errorPassthrough.form.customMessage') }}</label>
                <input
                  v-model="form.custom_message"
                  type="text"
                  class="components-admin-error-passthrough-rules-modal__field input"
                  :placeholder="t('admin.errorPassthrough.form.customMessagePlaceholder')"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Skip Monitoring -->
        <div class="components-admin-error-passthrough-rules-modal__label-3">
          <input
            type="checkbox"
            v-model="form.skip_monitoring"
            class="components-admin-error-passthrough-rules-modal__field-5"
          />
          <span class="components-admin-error-passthrough-rules-modal__text-5">
            {{ t('admin.errorPassthrough.form.skipMonitoring') }}
          </span>
        </div>
        <p class="components-admin-error-passthrough-rules-modal__description-3 input-hint">{{ t('admin.errorPassthrough.form.skipMonitoringHint') }}</p>

        <!-- Enabled -->
        <div class="components-admin-error-passthrough-rules-modal__label-3">
          <input
            type="checkbox"
            v-model="form.enabled"
            class="components-admin-error-passthrough-rules-modal__field-4"
          />
          <span class="components-admin-error-passthrough-rules-modal__text-5">
            {{ t('admin.errorPassthrough.form.enabled') }}
          </span>
        </div>
      </form>

      <template #footer>
        <div class="components-admin-error-passthrough-rules-modal__panel-24">
          <button @click="closeFormModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleSubmit" :disabled="submitting" class="btn btn-primary">
            <Icon v-if="submitting" name="refresh" size="sm" class="components-admin-error-passthrough-rules-modal__icon-5" />
            {{ showEditModal ? t('common.update') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.errorPassthrough.deleteRule')"
      :message="t('admin.errorPassthrough.deleteConfirm', { name: deletingRule?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { ErrorPassthroughRule } from '@/api/admin/errorPassthrough'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { accountPlatformOptions } from '@/utils/accountPlatforms'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

// eslint-disable-next-line @typescript-eslint/no-unused-vars
void emit // suppress unused warning - emit is used via $emit in template

const { t } = useI18n()
const appStore = useAppStore()

const rules = ref<ErrorPassthroughRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const editingRule = ref<ErrorPassthroughRule | null>(null)
const deletingRule = ref<ErrorPassthroughRule | null>(null)

// Form inputs for arrays
const errorCodesInput = ref('')
const keywordsInput = ref('')

const form = reactive({
  name: '',
  enabled: true,
  priority: 0,
  match_mode: 'any' as 'any' | 'all',
  platforms: [] as string[],
  passthrough_code: true,
  response_code: null as number | null,
  passthrough_body: true,
  custom_message: null as string | null,
  skip_monitoring: false,
  description: null as string | null
})

const matchModeOptions = computed(() => [
  { value: 'any', label: t('admin.errorPassthrough.matchMode.any'), description: t('admin.errorPassthrough.matchMode.anyHint') },
  { value: 'all', label: t('admin.errorPassthrough.matchMode.all'), description: t('admin.errorPassthrough.matchMode.allHint') }
])

const platformOptions = computed(() => accountPlatformOptions(t))

// Load rules when dialog opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadRules()
  }
})

const loadRules = async () => {
  loading.value = true
  try {
    rules.value = await adminAPI.errorPassthrough.list()
  } catch (error) {
    appStore.showError(t('admin.errorPassthrough.failedToLoad'))
    console.error('Error loading rules:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.enabled = true
  form.priority = 0
  form.match_mode = 'any'
  form.platforms = []
  form.passthrough_code = true
  form.response_code = null
  form.passthrough_body = true
  form.custom_message = null
  form.skip_monitoring = false
  form.description = null
  errorCodesInput.value = ''
  keywordsInput.value = ''
}

const closeFormModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingRule.value = null
  resetForm()
}

const handleEdit = (rule: ErrorPassthroughRule) => {
  editingRule.value = rule
  form.name = rule.name
  form.enabled = rule.enabled
  form.priority = rule.priority
  form.match_mode = rule.match_mode
  form.platforms = [...rule.platforms]
  form.passthrough_code = rule.passthrough_code
  form.response_code = rule.response_code
  form.passthrough_body = rule.passthrough_body
  form.custom_message = rule.custom_message
  form.skip_monitoring = rule.skip_monitoring
  form.description = rule.description
  errorCodesInput.value = rule.error_codes.join(', ')
  keywordsInput.value = rule.keywords.join('\n')
  showEditModal.value = true
}

const handleDelete = (rule: ErrorPassthroughRule) => {
  deletingRule.value = rule
  showDeleteDialog.value = true
}

const parseErrorCodes = (): number[] => {
  if (!errorCodesInput.value.trim()) return []
  return errorCodesInput.value
    .split(/[,\s]+/)
    .map(s => parseInt(s.trim(), 10))
    .filter(n => !isNaN(n) && n > 0)
}

const parseKeywords = (): string[] => {
  if (!keywordsInput.value.trim()) return []
  return keywordsInput.value
    .split('\n')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

const handleSubmit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('admin.errorPassthrough.nameRequired'))
    return
  }

  const errorCodes = parseErrorCodes()
  const keywords = parseKeywords()

  if (errorCodes.length === 0 && keywords.length === 0) {
    appStore.showError(t('admin.errorPassthrough.conditionsRequired'))
    return
  }

  submitting.value = true
  try {
    const data = {
      name: form.name.trim(),
      enabled: form.enabled,
      priority: form.priority,
      error_codes: errorCodes,
      keywords: keywords,
      match_mode: form.match_mode,
      platforms: form.platforms,
      passthrough_code: form.passthrough_code,
      response_code: form.passthrough_code ? null : form.response_code,
      passthrough_body: form.passthrough_body,
      custom_message: form.passthrough_body ? null : form.custom_message,
      skip_monitoring: form.skip_monitoring,
      description: form.description?.trim() || null
    }

    if (showEditModal.value && editingRule.value) {
      await adminAPI.errorPassthrough.update(editingRule.value.id, data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleUpdated'))
    } else {
      await adminAPI.errorPassthrough.create(data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleCreated'))
    }

    closeFormModal()
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToSave'))
    console.error('Error saving rule:', error)
  } finally {
    submitting.value = false
  }
}

const toggleEnabled = async (rule: ErrorPassthroughRule) => {
  try {
    await adminAPI.errorPassthrough.toggleEnabled(rule.id, !rule.enabled)
    rule.enabled = !rule.enabled
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToToggle'))
    console.error('Error toggling rule:', error)
  }
}

const confirmDelete = async () => {
  if (!deletingRule.value) return

  try {
    await adminAPI.errorPassthrough.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.errorPassthrough.ruleDeleted'))
    showDeleteDialog.value = false
    deletingRule.value = null
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToDelete'))
    console.error('Error deleting rule:', error)
  }
}
</script>
