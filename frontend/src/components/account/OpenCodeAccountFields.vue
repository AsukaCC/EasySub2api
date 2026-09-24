<template>
  <div class="opencode-fields">
    <label>
      <span class="input-label">{{ t('admin.accounts.opencode.mode') }}</span>
      <Select
        :model-value="modelValue.account_mode"
        :options="modeOptions"
        :searchable="false"
        @update:model-value="changeMode"
      />
    </label>
    <label>
      <span class="input-label">{{ t('admin.accounts.opencode.protocol') }}</span>
      <Select
        :model-value="modelValue.api_protocol"
        :options="protocolOptions"
        :searchable="false"
        @update:model-value="changeProtocol"
      />
    </label>
    <div v-if="modelValue.api_protocol === 'adaptive'" class="opencode-rules">
      <div class="opencode-rule-heading">
        <span class="input-label">{{ t('admin.accounts.opencode.rules') }}</span>
        <button type="button" class="btn btn-secondary" @click="resetRules">{{ t('admin.accounts.opencode.defaults') }}</button>
      </div>
      <div v-for="(rule, index) in rules" :key="index" class="opencode-rule">
        <input class="input" :value="rule.pattern" :aria-label="t('admin.accounts.opencode.pattern')" maxlength="128" @input="updateRule(index, 'pattern', ($event.target as HTMLInputElement).value)" />
        <Select
          :model-value="rule.protocol"
          :options="ruleProtocolOptions"
          :searchable="false"
          :aria-label="t('admin.accounts.opencode.protocol')"
          @update:model-value="(value) => updateRule(index, 'protocol', String(value ?? ''))"
        />
        <button type="button" class="btn btn-secondary" :title="t('common.delete')" :aria-label="t('common.delete')" @click="removeRule(index)"><Icon name="trash" size="sm" /></button>
      </div>
      <button type="button" class="btn btn-secondary" :disabled="rules.length >= 64" @click="addRule"><Icon name="plus" size="sm" />{{ t('admin.accounts.opencode.addRule') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import { defaultOpenCodeRules, openCodeBaseUrl, type OpenCodeSettings, type OpenCodeProtocol, type OpenCodeRule } from './openCodeCredentials'

const props = defineProps<{ modelValue: OpenCodeSettings; baseUrl: string }>()
const emit = defineEmits<{ 'update:modelValue': [OpenCodeSettings]; 'update:baseUrl': [string] }>()
const { t } = useI18n()
const modeOptions = [
  { value: 'zen', label: 'Zen' },
  { value: 'go', label: 'GO' },
]
const protocolOptions = computed(() => [
  { value: 'adaptive', label: t('admin.accounts.opencode.adaptive') },
  { value: 'chat_completions', label: 'Chat Completions' },
  { value: 'responses', label: 'Responses' },
  { value: 'anthropic', label: 'Anthropic Messages' },
])
const ruleProtocolOptions = [
  { value: 'chat_completions', label: 'Chat Completions' },
  { value: 'responses', label: 'Responses' },
  { value: 'anthropic', label: 'Anthropic Messages' },
]
const rules = computed(() => props.modelValue.protocol_rules ?? defaultOpenCodeRules(props.modelValue.account_mode))
function changeMode(mode: string | number | boolean | null) {
  const next = mode === 'zen' ? 'zen' : 'go'
  if (!props.baseUrl || props.baseUrl === openCodeBaseUrl(props.modelValue.account_mode)) emit('update:baseUrl', openCodeBaseUrl(next))
  emit('update:modelValue', { ...props.modelValue, account_mode: next })
}
function changeProtocol(protocol: string | number | boolean | null) {
  emit('update:modelValue', { ...props.modelValue, api_protocol: String(protocol ?? 'adaptive') as OpenCodeProtocol })
}
function setRules(protocol_rules?: OpenCodeRule[]) { emit('update:modelValue', { ...props.modelValue, protocol_rules }) }
function resetRules() { setRules(undefined) }
function addRule() { setRules([...rules.value, { pattern: '', protocol: 'chat_completions' }]) }
function removeRule(index: number) { setRules(rules.value.filter((_, i) => i !== index)) }
function updateRule(index: number, field: keyof OpenCodeRule, value: string) {
  setRules(rules.value.map((rule, i) => i === index ? { ...rule, [field]: value } : { ...rule }))
}
</script>

<style scoped>
.opencode-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.opencode-rules { grid-column: 1 / -1; display: grid; gap: 8px; }
.opencode-rule-heading { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.opencode-rule { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 36px; gap: 8px; }
.opencode-rule button { width: 36px; height: 36px; padding: 8px; }
@media (max-width: 480px) { .opencode-fields { grid-template-columns: minmax(0, 1fr); } .opencode-rule { grid-template-columns: minmax(0, 1fr) 36px; } .opencode-rule input { grid-column: 1 / -1; } }
</style>
