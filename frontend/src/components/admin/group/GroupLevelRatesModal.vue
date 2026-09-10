<template>
  <BaseDialog :show="show" :title="t('admin.groups.levelRates.title')" width="full" @close="close">
    <div v-if="group" class="level-rates-modal">
      <div class="level-rates-modal__context">
        <div class="level-rates-modal__context-icon">
          <PlatformIcon :platform="group.platform" size="sm" />
        </div>
        <div class="level-rates-modal__context-copy">
          <strong>{{ group.name }}</strong>
          <span>{{ t('admin.groups.levelRates.hint') }}</span>
        </div>
      </div>

      <section class="level-rates-modal__section">
        <div class="level-rates-modal__section-heading">
          <h4>{{ t('admin.groups.levelRates.dynamicRules') }}</h4>
          <button type="button" class="btn btn-secondary btn-sm" @click="addRule"><Icon name="plus" size="sm" />{{ t('admin.groups.levelRates.addRule') }}</button>
        </div>
        <p class="level-rates-modal__quota-hint">{{ t('admin.groups.levelRates.independentQuotaHint') }}</p>
        <p v-if="hasLegacyRules" class="level-rates-modal__legacy-hint">{{ t('admin.groups.levelRates.legacyHint') }}</p>
        <p v-if="rules.length === 0" class="level-rates-modal__empty">{{ t('admin.groups.levelRates.noRules') }}</p>
        <div v-for="(rule, index) in rules" :key="rule.id" class="level-rates-modal__rule">
          <div class="level-rates-modal__rule-heading">
            <div class="level-rates-modal__rule-title">
              <strong>{{ rule.name || t('admin.groups.levelRates.name') }}</strong>
              <span :class="['level-rates-modal__status', `level-rates-modal__status--${ruleStatus(rule)}`]">{{ statusLabel(ruleStatus(rule)) }}</span>
            </div>
            <button type="button" class="level-rates-modal__delete" :title="t('common.delete')" @click="rules.splice(index, 1)">
              <Icon name="trash" size="sm" />
            </button>
          </div>
          <div class="level-rates-modal__rule-grid">
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--name"><span>{{ t('admin.groups.levelRates.name') }}</span><input v-model="rule.name" class="input" maxlength="100" /></label>
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--datetime"><span>{{ t('admin.groups.levelRates.startAt') }}</span><input :value="toLocalDateTimeInput(rule.start_at)" class="input" type="datetime-local" step="1" @input="updateDateTime(rule, 'start_at', ($event.target as HTMLInputElement).value)" /></label>
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--datetime"><span>{{ t('admin.groups.levelRates.endAt') }}</span><input :value="toLocalDateTimeInput(rule.end_at)" class="input" type="datetime-local" step="1" @input="updateDateTime(rule, 'end_at', ($event.target as HTMLInputElement).value)" /></label>
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--metric"><span>{{ t('admin.groups.levelRates.discountCoefficient') }}</span><input v-model.number="rule.discount_coefficient" class="input" type="number" min="0.01" max="1" step="0.01" /></label>
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--metric"><span>{{ t('admin.groups.levelRates.activationSpend') }}</span><input v-model.number="rule.activation_spend" class="input" type="number" min="0" step="0.01" /></label>
            <label class="level-rates-modal__rule-field level-rates-modal__rule-field--metric"><span>{{ t('admin.groups.levelRates.personalQuotaAmount') }}</span><input v-model.number="rule.personal_quota_amount" class="input" type="number" min="0" step="0.01" /></label>
          </div>
          <div class="level-rates-modal__rule-footer">
            <label class="level-rates-modal__check level-rates-modal__check--enabled"><input v-model="rule.enabled" type="checkbox" />{{ t('admin.groups.levelRates.enabled') }}</label>
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
const rules = ref<GroupDynamicRateRule[]>([])
const saving = ref(false)
const nowTick = ref(Date.now())
let statusTimer: number | undefined

const hasLegacyRules = computed(() => rules.value.some((rule) => isLegacyDynamicRateRule(rule)))
function cloneRule(rule: GroupDynamicRateRule): GroupDynamicRateRule {
  return {
    ...rule,
    start_at: rule.start_at || '',
    end_at: rule.end_at || '',
    // Old shared quota values are displayed as the new per-user quota until
    // migration 263 has rewritten the persisted group configuration.
    shared_quota_amount: 0,
    personal_quota_amount: Number(rule.personal_quota_amount ?? rule.shared_quota_amount ?? 0),
    discount_coefficient: Number(rule.discount_coefficient ?? rule.multiplier ?? 1),
    multiplier: undefined
  }
}

function stopRefresh() {
  if (statusTimer !== undefined) window.clearInterval(statusTimer)
  statusTimer = undefined
}

function startStatusRefresh() {
  stopRefresh()
  statusTimer = window.setInterval(() => { nowTick.value = Date.now() }, 1_000)
}

watch(() => [props.show, props.group] as const, ([show, group]) => {
  stopRefresh()
  if (!show || !group) return
  rules.value = (group.dynamic_rate_rules || []).map(cloneRule)
  nowTick.value = Date.now()
  startStatusRefresh()
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
    discount_coefficient: 1,
    activation_spend: 0,
    shared_quota_amount: 0,
    personal_quota_amount: 0
  })
}

function updateDateTime(rule: GroupDynamicRateRule, field: 'start_at' | 'end_at', value: string) {
  rule[field] = localDateTimeToUTC(value)
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

function validateRules(): string | null {
  if (hasLegacyRules.value) return 'legacyRules'
  for (const rule of rules.value) {
    if (!rule.start_at || !rule.end_at) return 'timeRequired'
    if (!parseAbsoluteWindow(rule)) return 'invalidRange'
    const personal = Number(rule.personal_quota_amount ?? 0)
    if (!Number.isFinite(personal) || personal < 0) {
      return 'invalidQuota'
    }
    const coefficient = Number(rule.discount_coefficient ?? 1)
    if (!Number.isFinite(coefficient) || coefficient < 0.01 || coefficient > 1) return 'invalidCoefficient'
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
    shared_quota_amount: 0,
    personal_quota_amount: Number(rule.personal_quota_amount ?? 0),
    discount_coefficient: Number(rule.discount_coefficient ?? 1)
  }
  delete output.timezone
  delete output.start_time
  delete output.end_time
  delete output.quota_amount
  delete output.level_tier_ids
  delete output.levels
  delete output.multiplier
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
    const dynamicRateRules = rules.value.map(absoluteRuleForSave)
    await adminAPI.groups.update(props.group.id, { dynamic_rate_rules: dynamicRateRules })
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
.level-rates-modal {
  display: grid;
  min-width: 0;
  gap: 1.5rem;
}

.level-rates-modal__context,
.level-rates-modal__section-heading,
.level-rates-modal__rule-heading,
.level-rates-modal__rule-title,
.level-rates-modal__rule-footer,
.level-rates-modal__scope,
.level-rates-modal__footer {
  display: flex;
  align-items: center;
}

.level-rates-modal__context {
  min-width: 0;
  gap: .875rem;
  padding: .875rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface-muted);
}

.level-rates-modal__context-icon {
  display: grid;
  flex: 0 0 auto;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface-elevated);
}

.level-rates-modal__context-copy {
  display: grid;
  min-width: 0;
  gap: .2rem;
}

.level-rates-modal__context-copy strong {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.level-rates-modal__context-copy span {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  line-height: 1.5;
}

.level-rates-modal__section {
  display: grid;
  min-width: 0;
  gap: .875rem;
}

.level-rates-modal__section-heading {
  justify-content: space-between;
  gap: .75rem;
}

.level-rates-modal__section h4 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
}

.level-rates-modal__levels {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: .875rem;
}

.level-rates-modal__field,
.level-rates-modal__rule-field {
  display: grid;
  min-width: 0;
  gap: .45rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.level-rates-modal__field {
  padding: .75rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  background: var(--color-surface-muted);
}

.level-rates-modal__field .input,
.level-rates-modal__rule-field .input {
  width: 100%;
  min-width: 0;
}

.level-rates-modal__legacy-hint,
.level-rates-modal__quota-hint,
.level-rates-modal__empty {
  margin: 0;
  border-radius: var(--radius-lg);
  font-size: var(--font-size-xs);
  line-height: 1.55;
}

.level-rates-modal__legacy-hint {
  padding: .75rem .875rem;
  border: 1px solid color-mix(in srgb, var(--color-text-danger) 32%, transparent);
  color: var(--color-text-danger);
  background: color-mix(in srgb, var(--color-text-danger) 8%, transparent);
}

.level-rates-modal__quota-hint {
  padding: .7rem .875rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, var(--color-border-subtle));
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-primary) 6%, var(--color-surface-muted));
}

.level-rates-modal__empty {
  padding: 1.5rem;
  border: 1px dashed var(--color-border-strong);
  color: var(--color-text-muted);
  text-align: center;
  background: var(--color-surface-muted);
}

.level-rates-modal__rule {
  display: grid;
  min-width: 0;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  background: var(--color-surface-muted);
  box-shadow: var(--shadow-sm);
}

.level-rates-modal__rule-heading {
  min-width: 0;
  justify-content: space-between;
  gap: .75rem;
  padding-bottom: .875rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.level-rates-modal__rule-title {
  min-width: 0;
  flex-wrap: wrap;
  gap: .625rem;
}

.level-rates-modal__rule-title strong {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.level-rates-modal__status {
  flex: 0 0 auto;
  padding: .2rem .55rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  color: var(--color-text-secondary);
  background: var(--color-surface-elevated);
  font-size: var(--font-size-xs);
}

.level-rates-modal__status--active {
  border-color: color-mix(in srgb, var(--color-text-success) 32%, transparent);
  color: var(--color-text-success);
  background: color-mix(in srgb, var(--color-text-success) 8%, transparent);
}

.level-rates-modal__status--expired,
.level-rates-modal__status--legacy,
.level-rates-modal__status--invalid {
  border-color: color-mix(in srgb, var(--color-text-danger) 32%, transparent);
  color: var(--color-text-danger);
  background: color-mix(in srgb, var(--color-text-danger) 8%, transparent);
}

.level-rates-modal__status--not_started {
  border-color: color-mix(in srgb, var(--color-text-warning) 32%, transparent);
  color: var(--color-text-warning);
  background: color-mix(in srgb, var(--color-text-warning) 8%, transparent);
}

.level-rates-modal__delete {
  display: inline-grid;
  flex: 0 0 auto;
  width: 2rem;
  height: 2rem;
  padding: 0;
  place-items: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-danger);
  background: var(--color-surface-elevated);
  cursor: pointer;
  transition: border-color .15s ease, background-color .15s ease, transform .15s ease;
}

.level-rates-modal__delete:hover {
  border-color: color-mix(in srgb, var(--color-text-danger) 45%, transparent);
  background: color-mix(in srgb, var(--color-text-danger) 10%, var(--color-surface-elevated));
  transform: translateY(-1px);
}

.level-rates-modal__rule-grid {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: .875rem 1rem;
}

.level-rates-modal__rule-field--name,
.level-rates-modal__rule-field--datetime {
  grid-column: span 4;
}

.level-rates-modal__rule-field--metric {
  grid-column: span 3;
}

.level-rates-modal__rule-footer {
  min-width: 0;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: .75rem 1rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

.level-rates-modal__scope {
  min-width: 0;
  flex-wrap: wrap;
  gap: .625rem;
}

.level-rates-modal__scope-label {
  color: var(--color-text-tertiary);
  font-weight: 600;
}

.level-rates-modal__check {
  display: inline-flex;
  align-items: center;
  gap: .35rem;
  min-height: 1.5rem;
  white-space: nowrap;
}

.level-rates-modal__check--enabled {
  padding: .35rem .625rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-surface-elevated);
  font-weight: 600;
}

.level-rates-modal__footer {
  justify-content: flex-end;
  gap: .75rem;
}

@media (max-width: 960px) {
  .level-rates-modal__rule-field--name {
    grid-column: span 12;
  }

  .level-rates-modal__rule-field--datetime,
  .level-rates-modal__rule-field--metric {
    grid-column: span 6;
  }
}

@media (max-width: 720px) {
  .level-rates-modal {
    gap: 1.25rem;
  }

  .level-rates-modal__levels {
    grid-template-columns: 1fr;
  }

  .level-rates-modal__rule {
    padding: .875rem;
    border-radius: var(--radius-lg);
  }

  .level-rates-modal__rule-field--name,
  .level-rates-modal__rule-field--datetime,
  .level-rates-modal__rule-field--metric {
    grid-column: span 12;
  }

  .level-rates-modal__rule-footer {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 520px) {
  .level-rates-modal__context {
    align-items: flex-start;
    padding: .75rem;
  }

  .level-rates-modal__section-heading {
    align-items: stretch;
    flex-direction: column;
  }

  .level-rates-modal__section-heading .btn {
    justify-content: center;
    width: 100%;
  }

  .level-rates-modal__scope {
    gap: .5rem .75rem;
  }

  .level-rates-modal__scope-label {
    width: 100%;
  }

  .level-rates-modal__footer {
    width: 100%;
  }

  .level-rates-modal__footer .btn {
    flex: 1;
  }
}
</style>
