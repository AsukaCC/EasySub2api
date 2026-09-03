<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.inactiveCleanup.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="inactive-user-cleanup">
      <p class="inactive-user-cleanup__warning">
        {{ t('admin.users.inactiveCleanup.warning') }}
      </p>

      <div class="inactive-user-cleanup__filters">
        <label class="inactive-user-cleanup__field">
          <span class="input-label">{{ t('admin.users.inactiveCleanup.maxBalance') }}</span>
          <input
            v-model="form.maxBalance"
            class="input"
            type="number"
            min="0"
            step="0.00000001"
            data-test="inactive-max-balance"
          />
          <span class="input-hint">{{ t('admin.users.inactiveCleanup.maxBalanceHint') }}</span>
        </label>

        <label class="inactive-user-cleanup__field">
          <span class="input-label">{{ t('admin.users.inactiveCleanup.lastUsedBefore') }}</span>
          <input
            v-model="form.lastUsedBefore"
            class="input"
            type="datetime-local"
            step="1"
            data-test="inactive-last-used-before"
          />
          <span class="input-hint">{{ t('admin.users.inactiveCleanup.lastUsedBeforeHint') }}</span>
        </label>

        <label class="inactive-user-cleanup__field">
          <span class="input-label">{{ t('admin.users.inactiveCleanup.maxUsage7d') }}</span>
          <input
            v-model="form.maxUsage7d"
            class="input"
            type="number"
            min="0"
            step="0.00000001"
            data-test="inactive-max-usage-7d"
          />
          <span class="input-hint">{{ t('admin.users.inactiveCleanup.maxUsage7dHint') }}</span>
        </label>
      </div>

      <p v-if="validationError" class="inactive-user-cleanup__error">
        {{ validationError }}
      </p>

      <div class="inactive-user-cleanup__preview-action">
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="previewing || deleting || !!validationError"
          data-test="inactive-preview"
          @click="loadPreview"
        >
          {{ previewing ? t('admin.users.inactiveCleanup.previewing') : t('admin.users.inactiveCleanup.preview') }}
        </button>
        <span v-if="preview" class="inactive-user-cleanup__generated">
          {{ t('admin.users.inactiveCleanup.generatedAt', { time: formatDateTime(preview.generated_at) }) }}
        </span>
      </div>

      <template v-if="preview">
        <div class="inactive-user-cleanup__summary">
          <div>
            <strong>{{ preview.total }}</strong>
            <span>{{ t('admin.users.inactiveCleanup.matchedUsers') }}</span>
          </div>
          <div>
            <strong>{{ formatPoints(preview.total_balance) }}</strong>
            <span>{{ t('admin.users.inactiveCleanup.totalBalance') }}</span>
          </div>
          <div>
            <strong>{{ formatPoints(preview.total_usage_7d) }}</strong>
            <span>{{ t('admin.users.inactiveCleanup.totalUsage7d') }}</span>
          </div>
        </div>

        <div v-if="preview.items.length" class="inactive-user-cleanup__table-wrap">
          <table class="inactive-user-cleanup__table">
            <thead>
              <tr>
                <th>{{ t('admin.users.email') }}</th>
                <th>{{ t('admin.users.columns.balance') }}</th>
                <th>{{ t('admin.users.inactiveCleanup.usage7d') }}</th>
                <th>{{ t('admin.users.columns.lastUsed') }}</th>
                <th>{{ t('admin.users.columns.created') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in preview.items" :key="item.id">
                <td>{{ item.email }}</td>
                <td>{{ formatPoints(item.balance) }}</td>
                <td>{{ formatPoints(item.usage_7d) }}</td>
                <td>{{ item.last_used_at ? formatDateTime(item.last_used_at) : t('admin.users.inactiveCleanup.neverUsed') }}</td>
                <td>{{ formatDateTime(item.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <p v-if="preview.total > preview.items.length" class="input-hint">
            {{ t('admin.users.inactiveCleanup.sampleLimit', { count: preview.items.length, total: preview.total }) }}
          </p>
        </div>

        <div v-if="preview.total > 0" class="inactive-user-cleanup__confirmation">
          <p>
            {{ t('admin.users.inactiveCleanup.confirmHint', { phrase: confirmationPhrase }) }}
          </p>
          <input
            v-model="confirmation"
            class="input"
            type="text"
            :placeholder="confirmationPhrase"
            autocomplete="off"
            data-test="inactive-confirmation"
          />
        </div>
        <p v-else class="inactive-user-cleanup__empty">
          {{ t('admin.users.inactiveCleanup.noMatches') }}
        </p>
      </template>
    </div>

    <template #footer>
      <div class="inactive-user-cleanup__footer">
        <button type="button" class="btn btn-secondary" :disabled="deleting" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="btn btn-danger"
          :disabled="!canDelete"
          data-test="inactive-delete"
          @click="permanentlyDelete"
        >
          {{ deleting ? t('admin.users.inactiveCleanup.deleting') : t('admin.users.inactiveCleanup.deleteAction') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { InactiveUserDeletePreview, InactiveUserFilterRequest } from '@/api/admin/users'
import { useAppStore } from '@/stores/app'
import { formatDateTime, formatPoints } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  close: []
  success: [deleted: number]
}>()

const { t } = useI18n()
const appStore = useAppStore()
const preview = ref<InactiveUserDeletePreview | null>(null)
const previewing = ref(false)
const deleting = ref(false)
const confirmation = ref('')
const form = reactive({
  maxBalance: '0',
  lastUsedBefore: '',
  maxUsage7d: '0'
})

const toLocalDateTimeValue = (date: Date): string => {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const reset = () => {
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - 30)
  form.maxBalance = '0'
  form.lastUsedBefore = toLocalDateTimeValue(cutoff)
  form.maxUsage7d = '0'
  preview.value = null
  confirmation.value = ''
  previewing.value = false
  deleting.value = false
}

const parsedMaxBalance = computed(() => Number(form.maxBalance))
const parsedMaxUsage7d = computed(() => Number(form.maxUsage7d))
const validationError = computed(() => {
  if (!Number.isFinite(parsedMaxBalance.value) || parsedMaxBalance.value < 0) {
    return t('admin.users.inactiveCleanup.invalidBalance')
  }
  if (!Number.isFinite(parsedMaxUsage7d.value) || parsedMaxUsage7d.value < 0) {
    return t('admin.users.inactiveCleanup.invalidUsage7d')
  }
  if (!form.lastUsedBefore || Number.isNaN(new Date(form.lastUsedBefore).getTime())) {
    return t('admin.users.inactiveCleanup.invalidLastUsedBefore')
  }
  if (new Date(form.lastUsedBefore).getTime() > Date.now()) {
    return t('admin.users.inactiveCleanup.futureLastUsedBefore')
  }
  return ''
})

const buildFilter = (): InactiveUserFilterRequest => ({
  max_balance: parsedMaxBalance.value,
  last_used_before: new Date(form.lastUsedBefore).toISOString(),
  max_usage_7d: parsedMaxUsage7d.value
})

const confirmationPhrase = computed(() => `DELETE ${preview.value?.total ?? 0} USERS`)
const canDelete = computed(() =>
  !!preview.value
  && preview.value.total > 0
  && confirmation.value === confirmationPhrase.value
  && !previewing.value
  && !deleting.value
  && !validationError.value
)

watch(
  () => props.show,
  (show) => {
    if (show) reset()
  },
  { immediate: true }
)

watch(
  () => [form.maxBalance, form.lastUsedBefore, form.maxUsage7d],
  () => {
    preview.value = null
    confirmation.value = ''
  }
)

const loadPreview = async () => {
  if (validationError.value) return
  previewing.value = true
  confirmation.value = ''
  try {
    preview.value = await adminAPI.users.previewInactiveUsers(buildFilter())
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.message
      || error.response?.data?.detail
      || t('admin.users.inactiveCleanup.previewFailed')
    )
  } finally {
    previewing.value = false
  }
}

const permanentlyDelete = async () => {
  if (!canDelete.value || !preview.value) return
  deleting.value = true
  try {
    const result = await adminAPI.users.permanentlyDeleteInactiveUsers({
      ...buildFilter(),
      expected_count: preview.value.total,
      snapshot_token: preview.value.snapshot_token,
      confirmation: confirmation.value
    })
    appStore.showSuccess(t('admin.users.inactiveCleanup.deletedSuccess', { count: result.deleted }))
    emit('success', result.deleted)
    emit('close')
  } catch (error: any) {
    appStore.showError(
      error.response?.data?.message
      || error.response?.data?.detail
      || t('admin.users.inactiveCleanup.deleteFailed')
    )
    preview.value = null
    confirmation.value = ''
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.inactive-user-cleanup {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.inactive-user-cleanup__warning,
.inactive-user-cleanup__confirmation {
  padding: 0.875rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-text-danger) 35%, transparent);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--color-text-danger) 8%, transparent);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.inactive-user-cleanup__filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.inactive-user-cleanup__field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.inactive-user-cleanup__error {
  color: var(--color-text-danger);
}

.inactive-user-cleanup__preview-action,
.inactive-user-cleanup__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

.inactive-user-cleanup__generated {
  color: var(--color-text-tertiary);
  font-size: var(--type-caption-size);
}

.inactive-user-cleanup__summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.inactive-user-cleanup__summary > div {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-inset-bg);
}

.inactive-user-cleanup__summary strong {
  font-size: var(--type-section-title-size);
}

.inactive-user-cleanup__summary span,
.inactive-user-cleanup__empty {
  color: var(--color-text-secondary);
}

.inactive-user-cleanup__table-wrap {
  overflow-x: auto;
}

.inactive-user-cleanup__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--type-control-size);
}

.inactive-user-cleanup__table th,
.inactive-user-cleanup__table td {
  padding: 0.65rem;
  border-bottom: 1px solid var(--glass-border);
  text-align: left;
  white-space: nowrap;
}

.inactive-user-cleanup__confirmation {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

@media (max-width: 900px) {
  .inactive-user-cleanup__filters,
  .inactive-user-cleanup__summary {
    grid-template-columns: 1fr;
  }
}
</style>
