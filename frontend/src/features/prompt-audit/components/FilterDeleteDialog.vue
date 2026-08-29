<template>
  <BaseDialog :show="show" :title="t('admin.promptAudit.events.filterDeleteDialogTitle')" width="wide" @close="$emit('close')">
    <div class="features-prompt-audit-components-filter-delete-dialog__panel">
      <p class="features-prompt-audit-components-filter-delete-dialog__description">{{ t('admin.promptAudit.events.filterDeleteDialogDesc') }}</p>

      <fieldset>
        <legend class="features-prompt-audit-components-filter-delete-dialog__legend">{{ t('admin.promptAudit.events.filterTimeRange') }}</legend>
        <div class="features-prompt-audit-components-filter-delete-dialog__panel-2" role="radiogroup" :aria-label="t('admin.promptAudit.events.filterTimeRange')">
          <label
            v-for="option in DELETE_RANGE_PRESETS"
            :key="option.id"
            class="features-prompt-audit-components-filter-delete-dialog__label"
          >
            <input v-model="preset" type="radio" name="prompt-delete-range" :value="option.id" class="features-prompt-audit-components-filter-delete-dialog__field" :data-test="`range-preset-${option.id}`" @change="criteriaChanged" />
            <span class="features-prompt-audit-components-filter-delete-dialog__text">
              {{ t(`admin.promptAudit.events.timePresets.${option.id}`) }}
            </span>
          </label>
        </div>
        <p class="features-prompt-audit-components-filter-delete-dialog__description-2">{{ t('admin.promptAudit.events.filterTimeRangeHint') }}</p>
        <div v-if="preset === 'custom'" class="features-prompt-audit-components-filter-delete-dialog__panel-3" data-test="custom-range">
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.startAt') }}</span>
            <input v-model="local.start_at" type="datetime-local" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.startAt')" @change="criteriaChanged" />
          </label>
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.endAt') }}</span>
            <input v-model="local.end_at" type="datetime-local" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.endAt')" @change="criteriaChanged" />
          </label>
          <p v-if="!canPreview" class="features-prompt-audit-components-filter-delete-dialog__description-3">{{ t('admin.promptAudit.events.customRangeInvalid') }}</p>
        </div>
      </fieldset>

      <div class="features-prompt-audit-components-filter-delete-dialog__panel-4">
        <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
          <span>{{ t('admin.promptAudit.events.decision') }}</span>
          <Select v-model="local.decision" class="features-prompt-audit-components-filter-delete-dialog__field-2" :aria-label="t('admin.promptAudit.events.decision')" data-test="delete-decision" :options="[
            { value: '', label: t('common.all') },
            { value: 'pass', label: t('admin.promptAudit.decisions.pass') },
            { value: 'flag', label: t('admin.promptAudit.decisions.flag') },
            { value: 'critical', label: t('admin.promptAudit.decisions.critical') }
          ]" @change="criteriaChanged" />
        </label>
        <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
          <span>{{ t('admin.promptAudit.events.risk') }}</span>
          <Select v-model="local.risk_level" class="features-prompt-audit-components-filter-delete-dialog__field-2" :aria-label="t('admin.promptAudit.events.risk')" data-test="delete-risk" :options="[
            { value: '', label: t('common.all') },
            { value: 'low', label: t('admin.promptAudit.riskLevels.low') },
            { value: 'medium', label: t('admin.promptAudit.riskLevels.medium') },
            { value: 'high', label: t('admin.promptAudit.riskLevels.high') },
            { value: 'critical', label: t('admin.promptAudit.riskLevels.critical') }
          ]" @change="criteriaChanged" />
        </label>
      </div>

      <details class="features-prompt-audit-components-filter-delete-dialog__details" data-test="more-conditions">
        <summary class="features-prompt-audit-components-filter-delete-dialog__summary">{{ t('admin.promptAudit.events.moreConditions') }}</summary>
        <div class="features-prompt-audit-components-filter-delete-dialog__panel-3">
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.endpoint') }}</span>
            <input v-model="local.endpoint" type="text" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.endpoint')" @input="criteriaChanged" />
          </label>
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.keyword') }}</span>
            <input v-model="local.keyword" type="text" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.keyword')" @input="criteriaChanged" />
          </label>
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.groupId') }}</span>
            <input v-model="local.group_id" type="text" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.groupId')" @input="criteriaChanged" />
          </label>
          <label class="features-prompt-audit-components-filter-delete-dialog__label-2">
            <span>{{ t('admin.promptAudit.events.userId') }}</span>
            <input v-model="local.user_id" type="text" class="features-prompt-audit-components-filter-delete-dialog__field-2 input" :aria-label="t('admin.promptAudit.events.userId')" @input="criteriaChanged" />
          </label>
        </div>
      </details>

      <div v-if="preview" class="features-prompt-audit-components-filter-delete-dialog__panel-5" data-test="delete-preview-result">
        <p class="features-prompt-audit-components-filter-delete-dialog__description-4">{{ t('admin.promptAudit.events.filterDeleteCount', { count: preview.matched_count }) }}</p>
        <dl class="features-prompt-audit-components-filter-delete-dialog__dl">
          <dt>{{ t('admin.promptAudit.events.snapshotMax') }}</dt>
          <dd>{{ preview.snapshot_max_id }}</dd>
          <dt>Filter SHA-256</dt>
          <dd class="features-prompt-audit-components-filter-delete-dialog__dd">{{ preview.filter_hash }}</dd>
          <dt>{{ t('admin.promptAudit.events.expiresAt') }}</dt>
          <dd>{{ formatDate(preview.expires_at) }}</dd>
        </dl>
        <p class="features-prompt-audit-components-filter-delete-dialog__description-5">{{ t('admin.promptAudit.events.filterDeleteWarning') }}</p>
      </div>
      <p v-else class="features-prompt-audit-components-filter-delete-dialog__description-6" data-test="delete-preview-empty">
        {{ t('admin.promptAudit.events.filterDeleteNeedPreview') }}
      </p>
    </div>

    <template #footer>
      <div class="features-prompt-audit-components-filter-delete-dialog__panel-6">
        <p v-if="confirmDisabledReason" class="features-prompt-audit-components-filter-delete-dialog__description-7" data-test="confirm-disabled-reason">
          {{ t(confirmDisabledReason) }}
        </p>
        <button type="button" class="btn btn-secondary" @click="$emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-secondary" :disabled="!canPreview || previewing || deleting" data-test="run-delete-preview" @click="requestPreview">
          {{ previewing ? t('admin.promptAudit.events.filterDeletePreviewing') : t('admin.promptAudit.events.filterDeletePreviewAction') }}
        </button>
        <button
          type="button"
          class="btn btn-danger"
          :disabled="confirmDisabled"
          :title="confirmDisabledReason ? t(confirmDisabledReason) : undefined"
          data-test="confirm-filter-delete"
          @click="requestConfirm"
        >
          {{ deleting ? t('common.submitting') : t('admin.promptAudit.events.confirmFilterDelete') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import type { PromptDeletePreview, PromptEventFilters } from '../types'
import {
  DELETE_RANGE_PRESETS,
  cloneData,
  emptyEventFilters,
  hasExplicitDeleteRange,
  resolveDeleteRangeFilters,
  type DeleteRangePreset,
} from '../viewModel'

const props = defineProps<{
  show: boolean
  initialFilters: PromptEventFilters
  preview: PromptDeletePreview | null
  previewing: boolean
  deleting: boolean
}>()
const emit = defineEmits<{
  (event: 'close'): void
  (event: 'preview', value: PromptEventFilters): void
  (event: 'confirm', value: PromptEventFilters): void
  (event: 'criteria-change'): void
}>()
const { t, locale } = useI18n()

const preset = ref<DeleteRangePreset>('7d')
const local = reactive<PromptEventFilters>(emptyEventFilters())

watch(
  () => props.show,
  (visible) => {
    if (!visible) return
    const initial = cloneData(props.initialFilters)
    // Only inherit an explicit list-filter range; otherwise default to the
    // seven-day preset so a careless click can never target everything.
    preset.value = hasExplicitDeleteRange(initial) ? 'custom' : '7d'
    Object.assign(local, initial)
  },
  { immediate: true },
)

const canPreview = computed(() => preset.value !== 'custom' || hasExplicitDeleteRange(local))

// One-click flow: a valid criteria selection is enough to confirm — the parent
// mints the server-side confirmation token on the fly. The button stays
// disabled only when the range is invalid, work is in flight, or a fresh
// preview already proved there is nothing to delete.
const confirmDisabled = computed(
  () => !canPreview.value || props.previewing || props.deleting || (props.preview !== null && props.preview.matched_count === 0),
)
const confirmDisabledReason = computed(() => {
  if (props.previewing || props.deleting) return ''
  if (!canPreview.value) return 'admin.promptAudit.events.filterDeleteConfirmInvalidRange'
  if (props.preview && props.preview.matched_count === 0) return 'admin.promptAudit.events.filterDeleteConfirmNoMatches'
  return ''
})

function criteriaChanged() {
  emit('criteria-change')
}
function requestPreview() {
  if (!canPreview.value) return
  emit('preview', resolveDeleteRangeFilters(local, preset.value))
}
function requestConfirm() {
  if (confirmDisabled.value) return
  emit('confirm', resolveDeleteRangeFilters(local, preset.value))
}
function formatDate(value: string): string {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
}
</script>
