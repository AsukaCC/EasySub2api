<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { adminAPI } from '@/api'
import { opsAPI } from '@/api/admin/ops'
import type { AlertRule, MetricType, Operator } from '../types'
import type { OpsSeverity } from '@/api/admin/ops'
import { formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()
const appStore = useAppStore()

// 与 DataTable 一致：< 768px 切换为卡片视图，避免宽表在移动端被截断。
const isDesktopViewport = useMediaQuery('(min-width: 768px)')

const loading = ref(false)
const rules = ref<AlertRule[]>([])

async function load() {
  loading.value = true
  try {
    rules.value = await opsAPI.listAlertRules()
  } catch (err: any) {
    console.error('[OpsAlertRulesCard] Failed to load rules', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertRules.loadFailed'))
    rules.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  load()
  loadGroups()
})

const sortedRules = computed(() => {
  return [...rules.value].sort((a, b) =>
    String(b.created_at || '').localeCompare(String(a.created_at || ''))
  )
})

const showEditor = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const draft = ref<AlertRule | null>(null)

type MetricGroup = 'system' | 'group' | 'account'

interface MetricDefinition {
  type: MetricType
  group: MetricGroup
  label: string
  description: string
  recommendedOperator: Operator
  recommendedThreshold: number
  unit?: string
}

const groupMetricTypes = new Set<MetricType>([
  'group_available_accounts',
  'group_available_ratio',
  'group_rate_limit_ratio'
])

function parseGroupId(value: unknown): string | null {
  if (typeof value !== 'string') return null
  const id = value.trim()
  return id || null
}

const groupOptionsBase = ref<SelectOption[]>([])

async function loadGroups() {
  try {
    const list = await adminAPI.groups.getAll()
    groupOptionsBase.value = list.map((g) => ({ value: g.id, label: g.name }))
  } catch (err) {
    console.error('[OpsAlertRulesCard] Failed to load groups', err)
    groupOptionsBase.value = []
  }
}

const isGroupMetricSelected = computed(() => {
  const metricType = draft.value?.metric_type
  return metricType ? groupMetricTypes.has(metricType) : false
})

const draftGroupId = computed<string | null>({
  get() {
    return parseGroupId(draft.value?.filters?.group_id)
  },
  set(value) {
    if (!draft.value) return
    if (value == null) {
      if (!draft.value.filters) return
      delete draft.value.filters.group_id
      if (Object.keys(draft.value.filters).length === 0) {
        delete draft.value.filters
      }
      return
    }
    if (!draft.value.filters) draft.value.filters = {}
    draft.value.filters.group_id = value
  }
})

const groupOptions = computed<SelectOption[]>(() => {
  if (isGroupMetricSelected.value) return groupOptionsBase.value
  return [{ value: null, label: t('admin.ops.alertRules.form.allGroups') }, ...groupOptionsBase.value]
})

const metricDefinitions = computed(() => {
  return [
    // System-level metrics
    {
      type: 'success_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.successRate'),
      description: t('admin.ops.alertRules.metricDescriptions.successRate'),
      recommendedOperator: '<',
      recommendedThreshold: 99,
      unit: '%'
    },
    {
      type: 'error_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.errorRate'),
      description: t('admin.ops.alertRules.metricDescriptions.errorRate'),
      recommendedOperator: '>',
      recommendedThreshold: 1,
      unit: '%'
    },
    {
      type: 'upstream_error_rate',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.upstreamErrorRate'),
      description: t('admin.ops.alertRules.metricDescriptions.upstreamErrorRate'),
      recommendedOperator: '>',
      recommendedThreshold: 1,
      unit: '%'
    },
    {
      type: 'cpu_usage_percent',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.cpu'),
      description: t('admin.ops.alertRules.metricDescriptions.cpu'),
      recommendedOperator: '>',
      recommendedThreshold: 80,
      unit: '%'
    },
    {
      type: 'memory_usage_percent',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.memory'),
      description: t('admin.ops.alertRules.metricDescriptions.memory'),
      recommendedOperator: '>',
      recommendedThreshold: 80,
      unit: '%'
    },
    {
      type: 'concurrency_queue_depth',
      group: 'system',
      label: t('admin.ops.alertRules.metrics.queueDepth'),
      description: t('admin.ops.alertRules.metricDescriptions.queueDepth'),
      recommendedOperator: '>',
      recommendedThreshold: 10
    },

    // Group-level metrics (requires group_id filter)
    {
      type: 'group_available_accounts',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupAvailableAccounts'),
      description: t('admin.ops.alertRules.metricDescriptions.groupAvailableAccounts'),
      recommendedOperator: '<',
      recommendedThreshold: 1
    },
    {
      type: 'group_available_ratio',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupAvailableRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.groupAvailableRatio'),
      recommendedOperator: '<',
      recommendedThreshold: 50,
      unit: '%'
    },
    {
      type: 'group_rate_limit_ratio',
      group: 'group',
      label: t('admin.ops.alertRules.metrics.groupRateLimitRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.groupRateLimitRatio'),
      recommendedOperator: '>',
      recommendedThreshold: 10,
      unit: '%'
    },

    // Account-level metrics
    {
      type: 'account_rate_limited_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountRateLimitedCount'),
      description: t('admin.ops.alertRules.metricDescriptions.accountRateLimitedCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    },
    {
      type: 'account_error_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountErrorCount'),
      description: t('admin.ops.alertRules.metricDescriptions.accountErrorCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    },
    {
      type: 'account_error_ratio',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountErrorRatio'),
      description: t('admin.ops.alertRules.metricDescriptions.accountErrorRatio'),
      recommendedOperator: '>',
      recommendedThreshold: 5,
      unit: '%'
    },
    {
      type: 'account_temp_unscheduled_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.accountTempUnscheduledCount'),
      description: t('admin.ops.alertRules.metricDescriptions.accountTempUnscheduledCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    },
    {
      type: 'overload_account_count',
      group: 'account',
      label: t('admin.ops.alertRules.metrics.overloadAccountCount'),
      description: t('admin.ops.alertRules.metricDescriptions.overloadAccountCount'),
      recommendedOperator: '>',
      recommendedThreshold: 0
    }
  ] satisfies MetricDefinition[]
})

const selectedMetricDefinition = computed(() => {
  const metricType = draft.value?.metric_type
  if (!metricType) return null
  return metricDefinitions.value.find((m) => m.type === metricType) ?? null
})

const metricOptions = computed(() => {
  const buildGroup = (group: MetricGroup): SelectOption[] => {
    const items = metricDefinitions.value.filter((m) => m.group === group)
    if (items.length === 0) return []
    const headerValue = `__group__${group}`
    return [
      {
        value: headerValue,
        label: t(`admin.ops.alertRules.metricGroups.${group}`),
        disabled: true,
        kind: 'group'
      },
      ...items.map((m) => ({ value: m.type, label: m.label }))
    ]
  }

  return [...buildGroup('system'), ...buildGroup('group'), ...buildGroup('account')]
})

const operatorOptions = computed(() => {
  const ops: Operator[] = ['>', '>=', '<', '<=', '==', '!=']
  return ops.map((o) => ({ value: o, label: o }))
})

const severityOptions = computed(() => {
  const sev: OpsSeverity[] = ['P0', 'P1', 'P2', 'P3']
  return sev.map((s) => ({ value: s, label: s }))
})

const windowOptions = computed(() => {
  const windows = [1, 5, 60]
  return windows.map((m) => ({ value: m, label: `${m}m` }))
})

function newRuleDraft(): AlertRule {
  return {
    name: '',
    description: '',
    enabled: true,
    metric_type: 'error_rate',
    operator: '>',
    threshold: 1,
    window_minutes: 1,
    sustained_minutes: 2,
    severity: 'P1',
    cooldown_minutes: 10,
    notify_email: true
  }
}

function openCreate() {
  editingId.value = null
  draft.value = newRuleDraft()
  showEditor.value = true
}

function openEdit(rule: AlertRule) {
  editingId.value = rule.id ?? null
  draft.value = JSON.parse(JSON.stringify(rule))
  showEditor.value = true
}

const editorValidation = computed(() => {
  const errors: string[] = []
  const r = draft.value
  if (!r) return { valid: true, errors }
  if (!r.name || !r.name.trim()) errors.push(t('admin.ops.alertRules.validation.nameRequired'))
  if (!r.metric_type) errors.push(t('admin.ops.alertRules.validation.metricRequired'))
  if (groupMetricTypes.has(r.metric_type) && !parseGroupId(r.filters?.group_id)) {
    errors.push(t('admin.ops.alertRules.validation.groupIdRequired'))
  }
  if (!r.operator) errors.push(t('admin.ops.alertRules.validation.operatorRequired'))
  if (!(typeof r.threshold === 'number' && Number.isFinite(r.threshold)))
    errors.push(t('admin.ops.alertRules.validation.thresholdRequired'))
  if (!(typeof r.window_minutes === 'number' && Number.isFinite(r.window_minutes) && [1, 5, 60].includes(r.window_minutes))) {
    errors.push(t('admin.ops.alertRules.validation.windowRange'))
  }
  if (!(typeof r.sustained_minutes === 'number' && Number.isFinite(r.sustained_minutes) && r.sustained_minutes >= 1 && r.sustained_minutes <= 1440)) {
    errors.push(t('admin.ops.alertRules.validation.sustainedRange'))
  }
  if (!(typeof r.cooldown_minutes === 'number' && Number.isFinite(r.cooldown_minutes) && r.cooldown_minutes >= 0 && r.cooldown_minutes <= 1440)) {
    errors.push(t('admin.ops.alertRules.validation.cooldownRange'))
  }
  return { valid: errors.length === 0, errors }
})

async function save() {
  if (!draft.value) return
  if (!editorValidation.value.valid) {
    appStore.showError(editorValidation.value.errors[0] || t('admin.ops.alertRules.validation.invalid'))
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await opsAPI.updateAlertRule(editingId.value, draft.value)
    } else {
      await opsAPI.createAlertRule(draft.value)
    }
    showEditor.value = false
    draft.value = null
    editingId.value = null
    await load()
    appStore.showSuccess(t('admin.ops.alertRules.saveSuccess'))
  } catch (err: any) {
    console.error('[OpsAlertRulesCard] Failed to save rule', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertRules.saveFailed'))
  } finally {
    saving.value = false
  }
}

const showDeleteConfirm = ref(false)
const pendingDelete = ref<AlertRule | null>(null)

function requestDelete(rule: AlertRule) {
  pendingDelete.value = rule
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (!pendingDelete.value?.id) return
  try {
    await opsAPI.deleteAlertRule(pendingDelete.value.id)
    showDeleteConfirm.value = false
    pendingDelete.value = null
    await load()
    appStore.showSuccess(t('admin.ops.alertRules.deleteSuccess'))
  } catch (err: any) {
    console.error('[OpsAlertRulesCard] Failed to delete rule', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertRules.deleteFailed'))
  }
}

function cancelDelete() {
  showDeleteConfirm.value = false
  pendingDelete.value = null
}
</script>

<template>
  <div class="views-admin-ops-components-ops-alert-rules-card__panel">
    <div class="views-admin-ops-components-ops-alert-rules-card__panel-2">
      <div>
        <h3 class="views-admin-ops-components-ops-alert-rules-card__heading">{{ t('admin.ops.alertRules.title') }}</h3>
        <p class="views-admin-ops-components-ops-alert-rules-card__description">{{ t('admin.ops.alertRules.description') }}</p>
      </div>

      <div class="views-admin-ops-components-ops-alert-rules-card__panel-3">
        <button class="btn btn-sm btn-primary" :disabled="loading" @click="openCreate">
          {{ t('admin.ops.alertRules.create') }}
        </button>
        <button
          class="views-admin-ops-components-ops-alert-rules-card__action"
          :disabled="loading"
          @click="load"
        >
          <svg class="views-admin-ops-components-ops-alert-rules-card__icon" :class="{ 'views-admin-ops-components-ops-alert-rules-card__icon-2': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="views-admin-ops-components-ops-alert-rules-card__panel-4">
      {{ t('admin.ops.alertRules.loading') }}
    </div>

    <div v-else-if="sortedRules.length === 0" class="views-admin-ops-components-ops-alert-rules-card__panel-5">
      {{ t('admin.ops.alertRules.empty') }}
    </div>

    <div v-else class="views-admin-ops-components-ops-alert-rules-card__panel-6">
      <div class="views-admin-ops-components-ops-alert-rules-card__panel-7">
        <div v-if="!isDesktopViewport" class="views-admin-ops-components-ops-alert-rules-card__panel-8">
          <div v-for="row in sortedRules" :key="row.id" class="views-admin-ops-components-ops-alert-rules-card__panel-9">
            <div class="views-admin-ops-components-ops-alert-rules-card__panel-10">
              <div class="views-admin-ops-components-ops-alert-rules-card__panel-11">
                <div class="views-admin-ops-components-ops-alert-rules-card__panel-12">{{ row.name }}</div>
                <div v-if="row.description" class="views-admin-ops-components-ops-alert-rules-card__panel-13">
                  {{ row.description }}
                </div>
              </div>
              <span class="views-admin-ops-components-ops-alert-rules-card__text">{{ row.severity }}</span>
            </div>
            <div class="views-admin-ops-components-ops-alert-rules-card__panel-14">
              <span class="views-admin-ops-components-ops-alert-rules-card__text-2">{{ row.metric_type }}</span>
              <span class="views-admin-ops-components-ops-alert-rules-card__text-3">{{ row.operator }}</span>
              <span class="views-admin-ops-components-ops-alert-rules-card__text-2">{{ row.threshold }}</span>
            </div>
            <div class="views-admin-ops-components-ops-alert-rules-card__panel-15">
              <span class="views-admin-ops-components-ops-alert-rules-card__panel-14">
                {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
              </span>
              <div class="views-admin-ops-components-ops-alert-rules-card__panel-3">
                <button class="btn btn-sm btn-secondary" @click="openEdit(row)">{{ t('common.edit') }}</button>
                <button class="btn btn-sm btn-danger" @click="requestDelete(row)">{{ t('common.delete') }}</button>
              </div>
            </div>
            <div v-if="row.updated_at" class="views-admin-ops-components-ops-alert-rules-card__panel-16">
              {{ formatDateTime(row.updated_at) }}
            </div>
          </div>
        </div>
        <table v-else class="views-admin-ops-components-ops-alert-rules-card__table">
          <thead class="views-admin-ops-components-ops-alert-rules-card__header">
            <tr>
              <th class="views-admin-ops-components-ops-alert-rules-card__heading-2">
                {{ t('admin.ops.alertRules.table.name') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-rules-card__heading-2">
                {{ t('admin.ops.alertRules.table.metric') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-rules-card__heading-2">
                {{ t('admin.ops.alertRules.table.severity') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-rules-card__heading-2">
                {{ t('admin.ops.alertRules.table.enabled') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-rules-card__heading-3">
                {{ t('admin.ops.alertRules.table.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="views-admin-ops-components-ops-alert-rules-card__body">
            <tr v-for="row in sortedRules" :key="row.id" class="views-admin-ops-components-ops-alert-rules-card__row">
              <td class="views-admin-ops-components-ops-alert-rules-card__cell">
                <div class="views-admin-ops-components-ops-alert-rules-card__panel-12">{{ row.name }}</div>
                <div v-if="row.description" class="views-admin-ops-components-ops-alert-rules-card__panel-13">
                  {{ row.description }}
                </div>
                <div v-if="row.updated_at" class="views-admin-ops-components-ops-alert-rules-card__panel-17">
                  {{ formatDateTime(row.updated_at) }}
                </div>
              </td>
              <td class="views-admin-ops-components-ops-alert-rules-card__cell-2">
                <span class="views-admin-ops-components-ops-alert-rules-card__text-2">{{ row.metric_type }}</span>
                <span class="views-admin-ops-components-ops-alert-rules-card__text-3">{{ row.operator }}</span>
                <span class="views-admin-ops-components-ops-alert-rules-card__text-2">{{ row.threshold }}</span>
              </td>
              <td class="views-admin-ops-components-ops-alert-rules-card__cell-3">
                {{ row.severity }}
              </td>
              <td class="views-admin-ops-components-ops-alert-rules-card__cell-2">
                {{ row.enabled ? t('common.enabled') : t('common.disabled') }}
              </td>
              <td class="views-admin-ops-components-ops-alert-rules-card__cell-4">
                <button class="btn btn-sm btn-secondary" @click="openEdit(row)">{{ t('common.edit') }}</button>
                <button class="views-admin-ops-components-ops-alert-rules-card__action-2 btn btn-sm btn-danger" @click="requestDelete(row)">{{ t('common.delete') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <BaseDialog
      :show="showEditor"
      :title="editingId ? t('admin.ops.alertRules.editTitle') : t('admin.ops.alertRules.createTitle')"
      width="wide"
      @close="showEditor = false"
    >
      <div class="views-admin-ops-components-ops-alert-rules-card__panel-18">
        <div v-if="!editorValidation.valid" class="views-admin-ops-components-ops-alert-rules-card__panel-19">
          <div class="views-admin-ops-components-ops-alert-rules-card__panel-20">{{ t('admin.ops.alertRules.validation.title') }}</div>
          <ul class="views-admin-ops-components-ops-alert-rules-card__list">
            <li v-for="e in editorValidation.errors" :key="e">{{ e }}</li>
          </ul>
        </div>

        <div class="views-admin-ops-components-ops-alert-rules-card__panel-21">
          <div class="views-admin-ops-components-ops-alert-rules-card__panel-22">
            <label class="input-label">{{ t('admin.ops.alertRules.form.name') }}</label>
            <input v-model="draft!.name" class="input" type="text" />
          </div>

          <div class="views-admin-ops-components-ops-alert-rules-card__panel-22">
            <label class="input-label">{{ t('admin.ops.alertRules.form.description') }}</label>
            <input v-model="draft!.description" class="input" type="text" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.metric') }}</label>
            <Select v-model="draft!.metric_type" :options="metricOptions" />
            <div v-if="selectedMetricDefinition" class="views-admin-ops-components-ops-alert-rules-card__panel-23">
              <p>{{ selectedMetricDefinition.description }}</p>
              <p>
                {{
                  t('admin.ops.alertRules.hints.recommended', {
                    operator: selectedMetricDefinition.recommendedOperator,
                    threshold: selectedMetricDefinition.recommendedThreshold,
                    unit: selectedMetricDefinition.unit || ''
                  })
                }}
              </p>
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.operator') }}</label>
            <Select v-model="draft!.operator" :options="operatorOptions" />
          </div>

          <div class="views-admin-ops-components-ops-alert-rules-card__panel-22">
            <label class="input-label">
              {{ t('admin.ops.alertRules.form.groupId') }}
              <span v-if="isGroupMetricSelected" class="views-admin-ops-components-ops-alert-rules-card__text-4">*</span>
            </label>
            <Select
              v-model="draftGroupId"
              :options="groupOptions"
              searchable
              :placeholder="t('admin.ops.alertRules.form.groupPlaceholder')"
              :error="isGroupMetricSelected && !draftGroupId"
            />
            <p class="views-admin-ops-components-ops-alert-rules-card__description">
              {{ isGroupMetricSelected ? t('admin.ops.alertRules.hints.groupRequired') : t('admin.ops.alertRules.hints.groupOptional') }}
            </p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.threshold') }}</label>
            <input v-model.number="draft!.threshold" class="input" type="number" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.severity') }}</label>
            <Select v-model="draft!.severity" :options="severityOptions" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.window') }}</label>
            <Select v-model="draft!.window_minutes" :options="windowOptions" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.sustained') }}</label>
            <input v-model.number="draft!.sustained_minutes" class="input" type="number" min="1" max="1440" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.alertRules.form.cooldown') }}</label>
            <input v-model.number="draft!.cooldown_minutes" class="input" type="number" min="0" max="1440" />
          </div>

          <div class="views-admin-ops-components-ops-alert-rules-card__panel-24">
            <span class="views-admin-ops-components-ops-alert-rules-card__text-5">{{ t('admin.ops.alertRules.form.enabled') }}</span>
            <input v-model="draft!.enabled" type="checkbox" class="views-admin-ops-components-ops-alert-rules-card__field" />
          </div>

          <div class="views-admin-ops-components-ops-alert-rules-card__panel-24">
            <span class="views-admin-ops-components-ops-alert-rules-card__text-5">{{ t('admin.ops.alertRules.form.notifyEmail') }}</span>
            <input v-model="draft!.notify_email" type="checkbox" class="views-admin-ops-components-ops-alert-rules-card__field" />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="views-admin-ops-components-ops-alert-rules-card__panel-25">
          <button class="btn btn-secondary" :disabled="saving" @click="showEditor = false">
            {{ t('common.cancel') }}
          </button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('admin.ops.alertRules.deleteConfirmTitle')"
      :message="t('admin.ops.alertRules.deleteConfirmMessage')"
      :confirmText="t('common.delete')"
      :cancelText="t('common.cancel')"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>
