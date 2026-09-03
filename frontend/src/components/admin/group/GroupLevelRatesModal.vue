<template>
  <BaseDialog :show="show" :title="t('admin.groups.levelRates.title')" width="extra-wide" @close="close">
    <div v-if="group" class="level-rates-modal">
      <div class="level-rates-modal__context">
        <PlatformIcon :platform="group.platform" size="sm" />
        <strong>{{ group.name }}</strong>
        <span>{{ t('admin.groups.levelRates.hint') }}</span>
      </div>

      <section class="level-rates-modal__section">
        <h4>{{ t('admin.groups.levelRates.levelOverrides') }}</h4>
        <div class="level-rates-modal__levels">
          <label v-for="level in levels" :key="level" class="level-rates-modal__field">
            <span>{{ t('admin.groups.levelRates.level', { level }) }}</span>
            <input v-model="levelRates[String(level)]" class="input" type="number" min="0.01" max="100" step="0.01" :placeholder="t('admin.groups.levelRates.inherit')" />
          </label>
        </div>
      </section>

      <section class="level-rates-modal__section">
        <div class="level-rates-modal__section-heading">
          <h4>{{ t('admin.groups.levelRates.dynamicRules') }}</h4>
          <button type="button" class="btn btn-secondary btn-sm" @click="addRule"><Icon name="plus" size="sm" />{{ t('admin.groups.levelRates.addRule') }}</button>
        </div>
        <p v-if="hasLegacyRules" class="level-rates-modal__legacy-hint">{{ t('admin.groups.levelRates.legacyHint') }}</p>
        <p v-if="rules.length === 0" class="level-rates-modal__empty">{{ t('admin.groups.levelRates.noRules') }}</p>
        <div v-for="(rule, index) in rules" :key="rule.id" class="level-rates-modal__rule">
          <div class="level-rates-modal__rule-heading">
            <strong>{{ rule.name || t('admin.groups.levelRates.name') }}</strong>
            <span :class="['level-rates-modal__status', `level-rates-modal__status--${ruleStatus(rule)}`]">{{ statusLabel(ruleStatus(rule)) }}</span>
          </div>
          <div class="level-rates-modal__rule-grid">
            <label><span>{{ t('admin.groups.levelRates.name') }}</span><input v-model="rule.name" class="input" maxlength="100" /></label>
            <label><span>{{ t('admin.groups.levelRates.startAt') }}</span><input :value="toLocalDateTimeInput(rule.start_at)" class="input" type="datetime-local" step="1" @input="updateDateTime(rule, 'start_at', ($event.target as HTMLInputElement).value)" /></label>
            <label><span>{{ t('admin.groups.levelRates.endAt') }}</span><input :value="toLocalDateTimeInput(rule.end_at)" class="input" type="datetime-local" step="1" @input="updateDateTime(rule, 'end_at', ($event.target as HTMLInputElement).value)" /></label>
            <label><span>{{ t('admin.groups.levelRates.multiplier') }}</span><input v-model.number="rule.multiplier" class="input" type="number" min="0.01" max="100" step="0.01" /></label>
            <label><span>{{ t('admin.groups.levelRates.activationSpend') }}</span><input v-model.number="rule.activation_spend" class="input" type="number" min="0" step="0.01" /></label>
            <label><span>{{ t('admin.groups.levelRates.sharedQuotaAmount') }}</span><input v-model.number="rule.shared_quota_amount" class="input" type="number" min="0" step="0.01" /></label>
            <label><span>{{ t('admin.groups.levelRates.personalQuotaAmount') }}</span><input v-model.number="rule.personal_quota_amount" class="input" type="number" min="0" step="0.01" /></label>
          </div>
          <div class="level-rates-modal__usage">
            <div class="level-rates-modal__usage-label">
              <span>{{ t('admin.groups.levelRates.sharedUsage') }}</span>
              <span>{{ formatAmount(sharedQuotaUsed(rule)) }} / {{ sharedQuotaTotal(rule) > 0 ? formatAmount(sharedQuotaTotal(rule)) : t('admin.groups.levelRates.unlimited') }}</span>
            </div>
            <div v-if="sharedQuotaTotal(rule) > 0" class="level-rates-modal__progress" role="progressbar" :aria-valuenow="sharedQuotaUsed(rule)" :aria-valuemax="sharedQuotaTotal(rule)">
              <span :style="{ width: `${sharedUsagePercent(rule)}%` }" />
            </div>
          </div>
          <div class="level-rates-modal__rule-footer">
            <label class="level-rates-modal__check"><input v-model="rule.enabled" type="checkbox" />{{ t('admin.groups.levelRates.enabled') }}</label>
            <span>{{ t('admin.groups.levelRates.levels') }}</span>
            <label class="level-rates-modal__check"><input type="checkbox" :checked="rule.levels.length === 0" @change="toggleAll(rule, ($event.target as HTMLInputElement).checked)" />{{ t('admin.groups.levelRates.allLevels') }}</label>
            <label v-for="level in levels" :key="level" class="level-rates-modal__check"><input type="checkbox" :checked="rule.levels.includes(level)" :disabled="rule.levels.length === 0" @change="toggleLevel(rule, level, ($event.target as HTMLInputElement).checked)" />L{{ level }}</label>
            <button type="button" class="icon-button danger" :title="t('common.delete')" @click="rules.splice(index, 1)"><Icon name="trash" size="sm" /></button>
          </div>
        </div>
      </section>
    </div>
    <template #footer>
      <div class="level-rates-modal__footer">
        <button type="button" class="btn btn-secondary" @click="close">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving" @click="save"><Icon v-if="saving" name="refresh" size="sm" />{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { DynamicRateUsageSummary } from '@/api/admin/groups'
import { useAppStore } from '@/stores/app'
import type { AdminGroup, GroupDynamicRateRule } from '@/types'
import {
  getDynamicRateRuleStatus,
  isLegacyDynamicRateRule,
  localDateTimeToUTC,
  parseAbsoluteWindow,
  toLocalDateTimeInput,
  type DynamicRateRuleStatus
} from './dynamicRateWindow'

const props = defineProps<{ show: boolean; group: AdminGroup | null }>()
const emit = defineEmits<{ (event: 'close'): void; (event: 'success'): void }>()
const { t } = useI18n()
const appStore = useAppStore()
const levels = [1, 2, 3]
const levelRates = ref<Record<string, number | undefined>>({})
const rules = ref<GroupDynamicRateRule[]>([])
const usageByRule = ref<Record<string, DynamicRateUsageSummary>>({})
const saving = ref(false)
const nowTick = ref(Date.now())
let usageTimer: number | undefined
let statusTimer: number | undefined

const hasLegacyRules = computed(() => rules.value.some(isLegacyDynamicRateRule))

function cloneRule(rule: GroupDynamicRateRule): GroupDynamicRateRule {
  return {
    ...rule,
    levels: [...(rule.levels || [])],
    start_at: rule.start_at || '',
    end_at: rule.end_at || '',
    shared_quota_amount: Number(rule.shared_quota_amount ?? 0),
    personal_quota_amount: Number(rule.personal_quota_amount ?? 0)
  }
}

function stopRefresh() {
  if (usageTimer !== undefined) window.clearInterval(usageTimer)
  if (statusTimer !== undefined) window.clearInterval(statusTimer)
  usageTimer = undefined
  statusTimer = undefined
}

async function loadUsage(groupID: string) {
  try {
    const summaries = await adminAPI.groups.getDynamicRateUsage(groupID)
    if (props.show && props.group?.id === groupID) {
      usageByRule.value = Object.fromEntries(summaries.map(summary => [summary.rule_id, summary]))
    }
  } catch {
    // Background refreshes should not interrupt editing with a toast.
  }
}

function startRefresh(groupID: string) {
  stopRefresh()
  usageByRule.value = {}
  void loadUsage(groupID)
  usageTimer = window.setInterval(() => void loadUsage(groupID), 10_000)
  statusTimer = window.setInterval(() => { nowTick.value = Date.now() }, 1_000)
}

watch(() => [props.show, props.group] as const, ([show, group]) => {
  stopRefresh()
  if (!show || !group) return
  levelRates.value = { ...(group.level_rate_multipliers || {}) }
  rules.value = (group.dynamic_rate_rules || []).map(cloneRule)
  nowTick.value = Date.now()
  startRefresh(group.id)
}, { immediate: true })

onUnmounted(stopRefresh)

function addRule() {
  const id = globalThis.crypto?.randomUUID?.() || `rule-${Date.now()}-${Math.random().toString(36).slice(2)}`
  rules.value.push({
    id,
    name: '',
    enabled: true,
    start_at: '',
    end_at: '',
    levels: [],
    multiplier: 1,
    activation_spend: 0,
    shared_quota_amount: 0,
    personal_quota_amount: 0
  })
}

function updateDateTime(rule: GroupDynamicRateRule, field: 'start_at' | 'end_at', value: string) {
  rule[field] = localDateTimeToUTC(value)
}

function toggleAll(rule: GroupDynamicRateRule, all: boolean) { rule.levels = all ? [] : [1] }

function toggleLevel(rule: GroupDynamicRateRule, level: number, checked: boolean) {
  const next = new Set(rule.levels)
  checked ? next.add(level) : next.delete(level)
  if (!checked && next.size === 0) return
  rule.levels = [...next].sort()
}

function ruleStatus(rule: GroupDynamicRateRule): DynamicRateRuleStatus {
  void nowTick.value
  return getDynamicRateRuleStatus(rule)
}

function statusLabel(status: DynamicRateRuleStatus): string {
  switch (status) {
    case 'legacy': return t('admin.groups.levelRates.legacy')
    case 'not_started': return t('admin.groups.levelRates.notStarted')
    case 'active': return t('admin.groups.levelRates.active')
    case 'expired': return t('admin.groups.levelRates.expired')
    default: return t('admin.groups.levelRates.invalid')
  }
}

function sharedQuotaTotal(rule: GroupDynamicRateRule): number {
  return Math.max(0, Number(rule.shared_quota_amount ?? 0))
}

function sharedQuotaUsed(rule: GroupDynamicRateRule): number {
  const summary = usageByRule.value[rule.id]
  const window = parseAbsoluteWindow(rule)
  const summaryStart = summary?.start_at ? new Date(summary.start_at) : null
  if (!summary || !window || !summaryStart || Number.isNaN(summaryStart.getTime()) || summaryStart.getTime() !== window.start.getTime()) {
    return 0
  }
  return Math.max(0, summary.shared_used_amount)
}

function sharedUsagePercent(rule: GroupDynamicRateRule): number {
  const total = sharedQuotaTotal(rule)
  if (total <= 0) return 0
  return Math.min(100, Math.max(0, sharedQuotaUsed(rule) / total * 100))
}

function formatAmount(value: number): string {
  if (!Number.isFinite(value)) return '0'
  return value.toFixed(8).replace(/0+$/, '').replace(/\.$/, '') || '0'
}

function validateRules(): string | null {
  if (hasLegacyRules.value) return 'legacyRules'
  for (const rule of rules.value) {
    if (!rule.start_at || !rule.end_at) return 'timeRequired'
    if (!parseAbsoluteWindow(rule)) return 'invalidRange'
    const shared = Number(rule.shared_quota_amount ?? 0)
    const personal = Number(rule.personal_quota_amount ?? 0)
    if (!Number.isFinite(shared) || shared < 0 || !Number.isFinite(personal) || personal < 0) {
      return 'invalidQuota'
    }
  }
  return null
}

function absoluteRuleForSave(rule: GroupDynamicRateRule): GroupDynamicRateRule {
  const window = parseAbsoluteWindow(rule)
  const output = {
    ...rule,
    name: rule.name.trim(),
    start_at: window?.start.toISOString() || '',
    end_at: window?.end.toISOString() || '',
    shared_quota_amount: Number(rule.shared_quota_amount ?? 0),
    personal_quota_amount: Number(rule.personal_quota_amount ?? 0)
  }
  delete output.timezone
  delete output.start_time
  delete output.end_time
  delete output.quota_amount
  return output
}

function close() { if (!saving.value) emit('close') }

async function save() {
  if (!props.group) return
  const validationKey = validateRules()
  if (validationKey) {
    appStore.showError(t(`admin.groups.levelRates.${validationKey}`))
    return
  }
  saving.value = true
  try {
    const normalizedRates: Record<string, number> = {}
    for (const level of levels) {
      const value = levelRates.value[String(level)]
      if (value != null && Number.isFinite(Number(value))) normalizedRates[String(level)] = Number(value)
    }
    const dynamicRateRules = rules.value.map(absoluteRuleForSave)
    await adminAPI.groups.update(props.group.id, { level_rate_multipliers: normalizedRates, dynamic_rate_rules: dynamicRateRules })
    emit('success')
    emit('close')
  } catch {
    appStore.showError(t('admin.groups.levelRates.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.level-rates-modal { display: grid; gap: 1.25rem; }
.level-rates-modal__context, .level-rates-modal__section-heading, .level-rates-modal__rule-heading, .level-rates-modal__rule-footer, .level-rates-modal__footer { display: flex; align-items: center; gap: .75rem; }
.level-rates-modal__context { color: var(--color-text-secondary); border-bottom: 1px solid var(--color-border); padding-bottom: .75rem; }
.level-rates-modal__context span { margin-left: auto; font-size: var(--font-size-xs); }
.level-rates-modal__section { display: grid; gap: .75rem; }
.level-rates-modal__section h4 { margin: 0; font-size: var(--font-size-sm); }
.level-rates-modal__levels { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; }
.level-rates-modal__field, .level-rates-modal__rule-grid label { display: grid; gap: .35rem; font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.level-rates-modal__rule { border: 1px solid var(--color-border); padding: .75rem; display: grid; gap: .75rem; }
.level-rates-modal__rule-heading { justify-content: space-between; font-size: var(--font-size-xs); }
.level-rates-modal__status { border-radius: 999px; padding: .2rem .55rem; background: var(--color-surface-secondary); color: var(--color-text-secondary); }
.level-rates-modal__status--active { color: var(--color-text-success); }
.level-rates-modal__status--expired, .level-rates-modal__status--legacy, .level-rates-modal__status--invalid { color: var(--color-text-danger); }
.level-rates-modal__status--not_started { color: var(--color-text-warning); }
.level-rates-modal__rule-grid { display: grid; grid-template-columns: minmax(12rem, 2fr) repeat(6, minmax(8rem, 1fr)); gap: .6rem; }
.level-rates-modal__usage { display: grid; gap: .35rem; }
.level-rates-modal__usage-label { display: flex; justify-content: space-between; font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.level-rates-modal__progress { height: .4rem; overflow: hidden; border-radius: 999px; background: var(--color-surface-secondary); }
.level-rates-modal__progress span { display: block; height: 100%; border-radius: inherit; background: var(--color-primary); transition: width .2s ease; }
.level-rates-modal__rule-footer { flex-wrap: wrap; font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.level-rates-modal__check { display: inline-flex; align-items: center; gap: .3rem; }
.level-rates-modal__empty { color: var(--color-text-muted); margin: 0; }
.level-rates-modal__legacy-hint { color: var(--color-text-danger); margin: 0; font-size: var(--font-size-xs); }
.level-rates-modal__footer { justify-content: flex-end; }
.icon-button { border: 0; background: transparent; cursor: pointer; padding: .35rem; }
.danger { color: var(--color-text-danger); }
@media (max-width: 1200px) { .level-rates-modal__rule-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (max-width: 900px) { .level-rates-modal__rule-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .level-rates-modal__levels { grid-template-columns: 1fr; } .level-rates-modal__context span { margin-left: 0; } }
</style>
