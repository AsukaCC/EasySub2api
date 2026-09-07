<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.bulkLimits.title')"
    width="normal"
    @close="emit('close')"
  >
    <form id="bulk-edit-user-limits-form" class="components-admin-user-bulk-edit-user-modal__form" @submit.prevent="handleSubmit">
      <p class="components-admin-user-bulk-edit-user-modal__description">
        {{ t('admin.users.bulkLimits.selectedCount', { count: selectedIds.length }) }}
      </p>

      <div class="components-admin-user-bulk-edit-user-modal__panel">
        <div class="components-admin-user-bulk-edit-user-modal__panel-2">
          <div class="components-admin-user-bulk-edit-user-modal__panel-3">
            <label for="bulk-concurrency" class="components-admin-user-bulk-edit-user-modal__label input-label">
              {{ t('admin.users.columns.concurrency') }}
            </label>
            <Toggle
              v-model="enableConcurrency"
              :aria-label="t('admin.users.bulkLimits.enableConcurrency')"
              data-test="enable-concurrency"
            />
          </div>
          <input
            v-if="enableConcurrency"
            id="bulk-concurrency"
            v-model="concurrencyValue"
            type="number"
            min="0"
            step="1"
            class="input"
            data-test="concurrency-input"
          />
        </div>

        <div class="components-admin-user-bulk-edit-user-modal__panel-2">
          <div class="components-admin-user-bulk-edit-user-modal__panel-3">
            <label for="bulk-rpm-limit" class="components-admin-user-bulk-edit-user-modal__label input-label">
              {{ t('admin.users.form.rpmLimit') }}
            </label>
            <Toggle
              v-model="enableRPMLimit"
              :aria-label="t('admin.users.bulkLimits.enableRPMLimit')"
              data-test="enable-rpm-limit"
            />
          </div>
          <div v-if="enableRPMLimit">
            <input
              id="bulk-rpm-limit"
              v-model="rpmLimitValue"
              type="number"
              min="0"
              step="1"
              class="input"
              data-test="rpm-limit-input"
            />
            <p v-if="parsedRPMLimit === 0" class="input-hint">
              {{ t('admin.users.bulkLimits.unlimited') }}
            </p>
          </div>
        </div>
      </div>

      <div class="components-admin-user-bulk-edit-user-modal__panel components-admin-user-bulk-edit-user-modal__level-rules">
        <div class="components-admin-user-bulk-edit-user-modal__panel-2">
          <div class="components-admin-user-bulk-edit-user-modal__panel-3">
            <label for="bulk-level-rules" class="components-admin-user-bulk-edit-user-modal__label input-label">
              {{ t('admin.users.levels.assignedRules') }}
            </label>
            <Toggle
              v-model="enableLevelRules"
              :aria-label="t('admin.users.levels.enableBulk')"
              data-test="enable-level-rules"
            />
          </div>
          <div v-if="enableLevelRules" class="components-admin-user-bulk-edit-user-modal__level-rule-fields">
            <select v-model="levelRuleOperation" class="input">
              <option value="add">{{ t('admin.users.levels.addOperation') }}</option>
              <option value="remove">{{ t('admin.users.levels.removeOperation') }}</option>
              <option value="replace">{{ t('admin.users.levels.replaceOperation') }}</option>
            </select>
            <select id="bulk-level-rules" v-model="selectedLevelRuleIDs" class="input" multiple size="4" :disabled="levelRulesLoading">
              <option v-for="rule in levelRules" :key="rule.id" :value="rule.id">
                {{ rule.name }} ({{ rule.window_days }}d)
              </option>
            </select>
          </div>
        </div>
      </div>

      <p v-if="hasInvalidValue" class="components-admin-user-bulk-edit-user-modal__description-2">
        {{ t('admin.users.bulkLimits.nonNegativeInteger') }}
      </p>
      <p v-if="selectionTooLarge" class="components-admin-user-bulk-edit-user-modal__description-2">
        {{ t('admin.users.bulkLimits.selectionLimit', { max: MAX_BATCH_USER_IDS }) }}
      </p>
    </form>

    <template #footer>
      <div class="components-admin-user-bulk-edit-user-modal__panel-4">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="bulk-edit-user-limits-form"
          class="btn btn-primary"
          :disabled="!canSubmit"
          data-test="submit"
        >
          {{ submitting ? t('admin.users.bulkLimits.applying') : t('admin.users.bulkLimits.apply') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { BatchUpdateUserLimitsRequest } from '@/api/admin/users'
import type { UserLevelRule } from '@/api/admin/users'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Toggle from '@/components/common/Toggle.vue'

const props = defineProps<{
  show: boolean
  selectedIds: string[]
}>()

const emit = defineEmits<{
  close: []
  success: [affected: number]
}>()

const { t } = useI18n()
const appStore = useAppStore()
const enableConcurrency = ref(false)
const enableRPMLimit = ref(false)
const concurrencyValue = ref<string | number>('')
const rpmLimitValue = ref<string | number>('')
const submitting = ref(false)
const MAX_BATCH_USER_IDS = 500
const enableLevelRules = ref(false)
const levelRuleOperation = ref<'add' | 'remove' | 'replace'>('replace')
const selectedLevelRuleIDs = ref<string[]>([])
const levelRules = ref<UserLevelRule[]>([])
const levelRulesLoading = ref(false)

const parseLimit = (value: string | number): number | null | undefined => {
  const trimmed = String(value).trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  if (!Number.isInteger(parsed) || parsed < 0) return null
  return parsed
}

const parsedConcurrency = computed(() =>
  enableConcurrency.value ? parseLimit(concurrencyValue.value) : undefined
)
const parsedRPMLimit = computed(() =>
  enableRPMLimit.value ? parseLimit(rpmLimitValue.value) : undefined
)
const hasInvalidValue = computed(() =>
  parsedConcurrency.value === null || parsedRPMLimit.value === null
)
const hasUpdate = computed(() =>
  (parsedConcurrency.value !== undefined && parsedConcurrency.value !== null)
  || (parsedRPMLimit.value !== undefined && parsedRPMLimit.value !== null)
  || enableLevelRules.value
)
const levelRulesSelectionValid = computed(() =>
  !enableLevelRules.value
  || levelRuleOperation.value === 'replace'
  || selectedLevelRuleIDs.value.length > 0
)
const selectionTooLarge = computed(() => props.selectedIds.length > MAX_BATCH_USER_IDS)
const canSubmit = computed(() =>
  props.selectedIds.length > 0
  && !selectionTooLarge.value
  && hasUpdate.value
  && levelRulesSelectionValid.value
  && !hasInvalidValue.value
  && !submitting.value
)

const reset = () => {
  enableConcurrency.value = false
  enableRPMLimit.value = false
  concurrencyValue.value = ''
  rpmLimitValue.value = ''
  enableLevelRules.value = false
  levelRuleOperation.value = 'replace'
  selectedLevelRuleIDs.value = []
  submitting.value = false
}

async function loadLevelRules() {
  levelRulesLoading.value = true
  try {
    levelRules.value = await adminAPI.users.listLevelRules()
  } catch {
    levelRules.value = []
  } finally {
    levelRulesLoading.value = false
  }
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      reset()
      void loadLevelRules()
    }
  }
)

const handleSubmit = async () => {
  if (!canSubmit.value) return

  const request: BatchUpdateUserLimitsRequest = {
    user_ids: [...props.selectedIds],
    all: false
  }
  const fields: string[] = []
  if (parsedConcurrency.value !== undefined && parsedConcurrency.value !== null) {
    request.concurrency = parsedConcurrency.value
    fields.push(
      t('admin.users.bulkLimits.concurrencyValue', { value: parsedConcurrency.value })
    )
  }
  if (parsedRPMLimit.value !== undefined && parsedRPMLimit.value !== null) {
    request.rpm_limit = parsedRPMLimit.value
    fields.push(
      parsedRPMLimit.value === 0
        ? t('admin.users.bulkLimits.rpmUnlimitedValue')
        : t('admin.users.bulkLimits.rpmValue', { value: parsedRPMLimit.value })
    )
  }
  if (enableLevelRules.value) {
    fields.push(t(`admin.users.levels.${levelRuleOperation.value}Operation`))
  }

  const confirmed = window.confirm(
    t('admin.users.bulkLimits.confirm', {
      count: props.selectedIds.length,
      fields: fields.join(', ')
    })
  )
  if (!confirmed) return

  submitting.value = true
  try {
    let affected = 0
    if (parsedConcurrency.value !== undefined || parsedRPMLimit.value !== undefined) {
      const result = await adminAPI.users.batchUpdateLimits(request)
      affected = result.affected
    }
    if (enableLevelRules.value) {
      const ruleResult = await adminAPI.users.batchAssignLevelRules({
        user_ids: [...props.selectedIds],
        rule_ids: [...selectedLevelRuleIDs.value],
        operation: levelRuleOperation.value
      })
      affected = Math.max(affected, ruleResult.affected)
    }
    appStore.showSuccess(
      t('admin.users.bulkLimits.success', { count: affected })
    )
    emit('success', affected)
    emit('close')
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.message
      || error.response?.data?.detail
      || t('admin.users.bulkLimits.failed')
    )
  } finally {
    submitting.value = false
  }
}
</script>
