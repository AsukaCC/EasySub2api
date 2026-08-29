<template>
  <section aria-labelledby="prompt-events-title" class="features-prompt-audit-components-event-workspace__section">
    <div class="features-prompt-audit-components-event-workspace__panel">
      <div>
        <h2 id="prompt-events-title" class="features-prompt-audit-components-event-workspace__heading">{{ t('admin.promptAudit.events.title') }}</h2>
        <p class="features-prompt-audit-components-event-workspace__description">{{ t('admin.promptAudit.events.description') }}</p>
      </div>
      <div class="features-prompt-audit-components-event-workspace__panel-2">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="selectedIds.length === 0" @click="$emit('batch-delete')">
          {{ t('admin.promptAudit.events.deleteSelected', { count: selectedIds.length }) }}
        </button>
        <button type="button" class="btn btn-danger btn-sm" data-test="filter-delete" @click="$emit('preview-delete')">
          {{ t('admin.promptAudit.events.deleteByFilter') }}
        </button>
      </div>
    </div>

    <form class="features-prompt-audit-components-event-workspace__form" @submit.prevent="applyFilters">
      <label class="features-prompt-audit-components-event-workspace__label">
        <span>{{ t('admin.promptAudit.events.decision') }}</span>
        <Select v-model="localFilters.decision" class="features-prompt-audit-components-event-workspace__field" :aria-label="t('admin.promptAudit.events.decision')" :options="[
          { value: '', label: t('common.all') },
          { value: 'pass', label: t('admin.promptAudit.decisions.pass') },
          { value: 'flag', label: t('admin.promptAudit.decisions.flag') },
          { value: 'critical', label: t('admin.promptAudit.decisions.critical') }
        ]" @change="filtersChanged" />
      </label>
      <label class="features-prompt-audit-components-event-workspace__label">
        <span>{{ t('admin.promptAudit.events.risk') }}</span>
        <Select v-model="localFilters.risk_level" class="features-prompt-audit-components-event-workspace__field" :aria-label="t('admin.promptAudit.events.risk')" :options="[
          { value: '', label: t('common.all') },
          { value: 'low', label: t('admin.promptAudit.riskLevels.low') },
          { value: 'medium', label: t('admin.promptAudit.riskLevels.medium') },
          { value: 'high', label: t('admin.promptAudit.riskLevels.high') },
          { value: 'critical', label: t('admin.promptAudit.riskLevels.critical') }
        ]" @change="filtersChanged" />
      </label>
      <FilterInput v-model="localFilters.endpoint" :label="t('admin.promptAudit.events.endpoint')" @change="filtersChanged" />
      <FilterInput v-model="localFilters.group_id" :label="t('admin.promptAudit.events.groupId')" type="text" @change="filtersChanged" />
      <FilterInput v-model="localFilters.user_id" :label="t('admin.promptAudit.events.userId')" type="text" @change="filtersChanged" />
      <FilterInput v-model="localFilters.api_key_id" :label="t('admin.promptAudit.events.apiKeyId')" type="text" @change="filtersChanged" />
      <FilterInput v-model="localFilters.request_id" :label="t('admin.promptAudit.events.requestId')" @change="filtersChanged" />
      <FilterInput v-model="localFilters.prompt_hash" :label="t('admin.promptAudit.events.promptHash')" @change="filtersChanged" />
      <FilterInput v-model="localFilters.keyword" :label="t('admin.promptAudit.events.keyword')" @change="filtersChanged" />
      <label class="features-prompt-audit-components-event-workspace__label">
        <span>{{ t('admin.promptAudit.events.startAt') }}</span>
        <input v-model="localFilters.start_at" type="datetime-local" class="features-prompt-audit-components-event-workspace__field input" :aria-label="t('admin.promptAudit.events.startAt')" @change="filtersChanged" />
      </label>
      <label class="features-prompt-audit-components-event-workspace__label">
        <span>{{ t('admin.promptAudit.events.endAt') }}</span>
        <input v-model="localFilters.end_at" type="datetime-local" class="features-prompt-audit-components-event-workspace__field input" :aria-label="t('admin.promptAudit.events.endAt')" @change="filtersChanged" />
      </label>
      <div class="features-prompt-audit-components-event-workspace__panel-3">
        <button type="submit" class="btn btn-primary btn-sm">{{ t('common.search') }}</button>
        <button type="button" class="btn btn-ghost btn-sm" @click="resetFilters">{{ t('common.reset') }}</button>
      </div>
    </form>
    <div v-if="error" role="alert" class="features-prompt-audit-components-event-workspace__panel-4">{{ error }}</div>
    <div class="features-prompt-audit-components-event-workspace__panel-5">
      <table class="features-prompt-audit-components-event-workspace__table">
        <thead class="features-prompt-audit-components-event-workspace__header">
          <tr>
            <th class="features-prompt-audit-components-event-workspace__heading-2"><input type="checkbox" :checked="allSelected" :aria-label="t('admin.promptAudit.events.selectAll')" @change="toggleAll" /></th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.time') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.identity') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.group') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.route') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.result') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-3">{{ t('admin.promptAudit.events.preview') }}</th>
            <th class="features-prompt-audit-components-event-workspace__heading-4">{{ t('admin.promptAudit.common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="features-prompt-audit-components-event-workspace__body">
          <tr v-if="loading"><td colspan="8" class="features-prompt-audit-components-event-workspace__cell" aria-busy="true">{{ t('common.loading') }}</td></tr>
          <tr v-else-if="events.length === 0"><td colspan="8" class="features-prompt-audit-components-event-workspace__cell">{{ t('admin.promptAudit.events.empty') }}</td></tr>
          <tr v-for="event in events" v-else :key="event.id" :data-test="`event-${event.id}`" class="features-prompt-audit-components-event-workspace__row">
            <td class="features-prompt-audit-components-event-workspace__cell-2"><input type="checkbox" :checked="selectedIds.includes(event.id)" :aria-label="t('admin.promptAudit.events.selectEvent', { id: event.id })" @change="toggleOne(event.id)" /></td>
            <td class="features-prompt-audit-components-event-workspace__cell-3">{{ formatDate(event.created_at) }}</td>
            <td class="features-prompt-audit-components-event-workspace__cell-2">
              <CopyLine :label="t('admin.promptAudit.events.user')" :value="event.snapshot.username" />
              <CopyLine :label="t('admin.promptAudit.events.email')" :value="event.snapshot.user_email" />
              <CopyLine :label="t('admin.promptAudit.events.apiKey')" :value="event.snapshot.api_key_name" />
            </td>
            <td class="features-prompt-audit-components-event-workspace__cell-4">{{ event.snapshot.group_name || '—' }}</td>
            <td class="features-prompt-audit-components-event-workspace__cell-2">
              <p class="features-prompt-audit-components-event-workspace__description-2">{{ event.snapshot.endpoint }}</p>
              <p class="features-prompt-audit-components-event-workspace__description-3">{{ event.snapshot.model }} · {{ event.snapshot.protocol }} · {{ event.snapshot.stage || 'http' }}</p>
            </td>
            <td class="features-prompt-audit-components-event-workspace__cell-2">
              <span class="features-prompt-audit-components-event-workspace__text" :class="decisionClass(event.decision)">{{ formatDecisionRisk(event.decision, event.risk_level) }}</span>
              <p class="features-prompt-audit-components-event-workspace__description-4" :title="formatCategories(event.categories)">{{ formatCategories(event.categories) }}</p>
            </td>
            <td class="features-prompt-audit-components-event-workspace__cell-5"><p class="features-prompt-audit-components-event-workspace__description-5">{{ event.snapshot.redacted_preview || '—' }}</p></td>
            <td class="features-prompt-audit-components-event-workspace__cell-6">
              <button type="button" class="btn btn-ghost btn-sm" @click="$emit('view', event.id)">{{ t('common.view') }}</button>
              <button type="button" class="features-prompt-audit-components-event-workspace__action btn btn-ghost btn-sm" @click="$emit('delete', event.id)">{{ t('common.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
      <Pagination :total="total" :page="page" :page-size="pageSize" @update:page="$emit('page', $event)" @update:page-size="$emit('page-size', $event)" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import type { PromptAuditEvent, PromptEventFilters } from '../types'
import { cloneData, emptyEventFilters, SCANNER_CATALOG } from '../viewModel'

const props = defineProps<{
  events: PromptAuditEvent[]; total: number; page: number; pageSize: number
  filters: PromptEventFilters; selectedIds: string[]; loading: boolean; error: string
}>()
const emit = defineEmits<{
  (event: 'filters-change', value: PromptEventFilters): void
  (event: 'search', value: PromptEventFilters): void
  (event: 'selection', value: string[]): void
  (event: 'page', value: number): void
  (event: 'page-size', value: number): void
  (event: 'view', id: string): void
  (event: 'delete', id: string): void
  (event: 'batch-delete'): void
  (event: 'preview-delete'): void
}>()
const { t, locale } = useI18n()
const localFilters = reactive<PromptEventFilters>(cloneData(props.filters))
watch(() => props.filters, (value) => Object.assign(localFilters, cloneData(value)), { deep: true })
const allSelected = computed(() => props.events.length > 0 && props.events.every((event) => props.selectedIds.includes(event.id)))

const FilterInput = defineComponent({
  props: { modelValue: { type: String, required: true }, label: { type: String, required: true }, type: { type: String, default: 'text' } },
  emits: ['update:modelValue', 'change'],
  setup(componentProps, { emit: componentEmit }) {
    return () => h('label', { class: 'features-prompt-audit-components-event-workspace__label' }, [
      h('span', componentProps.label),
      h('input', {
        value: componentProps.modelValue, type: componentProps.type, class: 'features-prompt-audit-components-event-workspace__field input', 'aria-label': componentProps.label,
        onInput: (event: Event) => componentEmit('update:modelValue', (event.target as HTMLInputElement).value),
        onChange: () => componentEmit('change'),
      }),
    ])
  },
})

const CopyLine = defineComponent({
  props: { label: { type: String, required: true }, value: { type: String, default: '' } },
  setup(componentProps) {
    return () => h('div', { class: 'features-prompt-audit-components-event-workspace__render' }, [
      h('span', { class: 'features-prompt-audit-components-event-workspace__render-2' }, componentProps.label),
      h('span', { class: 'features-prompt-audit-components-event-workspace__render-3' }, componentProps.value || '—'),
      componentProps.value ? h('button', {
        type: 'button', class: 'features-prompt-audit-components-event-workspace__render-4', 'aria-label': `${t('common.copy')} ${componentProps.label}`,
        onClick: () => navigator.clipboard?.writeText(componentProps.value),
      }, t('common.copy')) : null,
    ])
  },
})

function filtersChanged() {
  emit('filters-change', cloneData(localFilters))
}
function applyFilters() {
  const value = cloneData(localFilters)
  emit('filters-change', value)
  emit('search', value)
}
function resetFilters() {
  Object.assign(localFilters, emptyEventFilters())
  applyFilters()
}
function toggleOne(id: string) {
  const selected = new Set(props.selectedIds)
  if (selected.has(id)) selected.delete(id)
  else selected.add(id)
  emit('selection', [...selected])
}
function toggleAll() {
  emit('selection', allSelected.value ? [] : props.events.map((event) => event.id))
}
function formatDate(value: string): string {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'short', timeStyle: 'medium' }).format(new Date(value))
}
function decisionClass(decision: string): string {
  if (decision === 'critical') return 'features-prompt-audit-components-event-workspace__state'
  if (decision === 'flag') return 'features-prompt-audit-components-event-workspace__state-2'
  return 'features-prompt-audit-components-event-workspace__state-3'
}
const DECISIONS = new Set(['pass', 'flag', 'critical'])
const RISK_LEVELS = new Set(['low', 'medium', 'high', 'critical'])

function translateDecision(decision: string): string {
  return DECISIONS.has(decision) ? t(`admin.promptAudit.decisions.${decision}`) : decision
}
function translateRiskLevel(riskLevel: string): string {
  return RISK_LEVELS.has(riskLevel) ? t(`admin.promptAudit.riskLevels.${riskLevel}`) : riskLevel
}
function translateCategory(category: string): string {
  return SCANNER_CATALOG.some((scanner) => scanner.id === category)
    ? t(`admin.promptAudit.scanners.${category}`)
    : category
}
function formatDecisionRisk(decision: string, riskLevel: string): string {
  return `${translateDecision(decision)} · ${translateRiskLevel(riskLevel)}`
}
function formatCategories(categories: string[]): string {
  if (!categories.length) return '—'
  return categories.map(translateCategory).join(', ')
}
</script>
