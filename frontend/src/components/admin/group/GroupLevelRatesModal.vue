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
        <p v-if="rules.length === 0" class="level-rates-modal__empty">{{ t('admin.groups.levelRates.noRules') }}</p>
        <div v-for="(rule, index) in rules" :key="rule.id" class="level-rates-modal__rule">
          <div class="level-rates-modal__rule-grid">
            <label><span>{{ t('admin.groups.levelRates.name') }}</span><input v-model="rule.name" class="input" maxlength="100" /></label>
            <label><span>{{ t('admin.groups.levelRates.start') }}</span><input v-model="rule.start_time" class="input" type="time" /></label>
            <label><span>{{ t('admin.groups.levelRates.end') }}</span><input v-model="rule.end_time" class="input" type="time" /></label>
            <label><span>{{ t('admin.groups.levelRates.multiplier') }}</span><input v-model.number="rule.multiplier" class="input" type="number" min="0.01" max="100" step="0.01" /></label>
            <label><span>{{ t('admin.groups.levelRates.activationSpend') }}</span><input v-model.number="rule.activation_spend" class="input" type="number" min="0" step="0.01" /></label>
            <label><span>{{ t('admin.groups.levelRates.quotaAmount') }}</span><input v-model.number="rule.quota_amount" class="input" type="number" min="0" step="0.01" /></label>
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
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import type { AdminGroup, GroupDynamicRateRule } from '@/types'

const props = defineProps<{ show: boolean; group: AdminGroup | null }>()
const emit = defineEmits<{ (event: 'close'): void; (event: 'success'): void }>()
const { t } = useI18n()
const appStore = useAppStore()
const levels = [1, 2, 3]
const levelRates = ref<Record<string, number | undefined>>({})
const rules = ref<GroupDynamicRateRule[]>([])
const saving = ref(false)

function cloneRule(rule: GroupDynamicRateRule): GroupDynamicRateRule {
  return { ...rule, levels: [...(rule.levels || [])] }
}

watch(() => [props.show, props.group] as const, ([show, group]) => {
  if (!show || !group) return
  levelRates.value = { ...(group.level_rate_multipliers || {}) }
  rules.value = (group.dynamic_rate_rules || []).map(cloneRule)
}, { immediate: true })

function addRule() {
  const id = globalThis.crypto?.randomUUID?.() || `rule-${Date.now()}-${Math.random().toString(36).slice(2)}`
  rules.value.push({ id, name: '', enabled: true, timezone: 'Asia/Shanghai', start_time: '00:00', end_time: '23:59', levels: [], multiplier: 1, activation_spend: 0, quota_amount: 0 })
}
function toggleAll(rule: GroupDynamicRateRule, all: boolean) { rule.levels = all ? [] : [1] }
function toggleLevel(rule: GroupDynamicRateRule, level: number, checked: boolean) {
  const next = new Set(rule.levels)
  checked ? next.add(level) : next.delete(level)
  if (!checked && next.size === 0) return
  rule.levels = [...next].sort()
}
function close() { if (!saving.value) emit('close') }
async function save() {
  if (!props.group) return
  saving.value = true
  try {
    const normalizedRates: Record<string, number> = {}
    for (const level of levels) {
      const value = levelRates.value[String(level)]
      if (value != null && Number.isFinite(Number(value))) normalizedRates[String(level)] = Number(value)
    }
    await adminAPI.groups.update(props.group.id, { level_rate_multipliers: normalizedRates, dynamic_rate_rules: rules.value })
    emit('success')
    emit('close')
  } catch (error) {
    appStore.showError(t('admin.groups.levelRates.saveFailed'))
  } finally { saving.value = false }
}
</script>

<style scoped>
.level-rates-modal { display: grid; gap: 1.25rem; }
.level-rates-modal__context, .level-rates-modal__section-heading, .level-rates-modal__rule-footer, .level-rates-modal__footer { display: flex; align-items: center; gap: .75rem; }
.level-rates-modal__context { color: var(--color-text-secondary); border-bottom: 1px solid var(--color-border); padding-bottom: .75rem; }
.level-rates-modal__context span { margin-left: auto; font-size: var(--font-size-xs); }
.level-rates-modal__section { display: grid; gap: .75rem; }
.level-rates-modal__section h4 { margin: 0; font-size: var(--font-size-sm); }
.level-rates-modal__levels { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; }
.level-rates-modal__field, .level-rates-modal__rule-grid label { display: grid; gap: .35rem; font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.level-rates-modal__rule { border: 1px solid var(--color-border); padding: .75rem; display: grid; gap: .75rem; }
.level-rates-modal__rule-grid { display: grid; grid-template-columns: 2fr repeat(5, minmax(0, 1fr)); gap: .6rem; }
.level-rates-modal__rule-footer { flex-wrap: wrap; font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.level-rates-modal__check { display: inline-flex; align-items: center; gap: .3rem; }
.level-rates-modal__empty { color: var(--color-text-muted); margin: 0; }
.level-rates-modal__footer { justify-content: flex-end; }
.icon-button { border: 0; background: transparent; cursor: pointer; padding: .35rem; }
.danger { color: var(--color-danger); }
@media (max-width: 900px) { .level-rates-modal__rule-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .level-rates-modal__levels { grid-template-columns: 1fr; } .level-rates-modal__context span { margin-left: 0; } }
</style>
