<template>
  <div class="opencode-fields">
    <label>
      <span class="input-label">{{ t('admin.accounts.opencode.mode') }}</span>
      <select class="input" :value="modelValue.account_mode" @change="changeMode">
        <option value="zen">Zen</option><option value="go">GO</option>
      </select>
    </label>
    <label>
      <span class="input-label">{{ t('admin.accounts.opencode.protocol') }}</span>
      <select class="input" :value="modelValue.api_protocol" @change="changeProtocol">
        <option value="adaptive">{{ t('admin.accounts.opencode.adaptive') }}</option>
        <option value="chat_completions">Chat Completions</option>
        <option value="responses">Responses</option>
        <option value="anthropic">Anthropic Messages</option>
      </select>
    </label>
    <div v-if="modelValue.api_protocol === 'adaptive'" class="opencode-rules">
      <div class="opencode-rule-heading">
        <span class="input-label">{{ t('admin.accounts.opencode.rules') }}</span>
        <button type="button" class="btn btn-secondary" @click="resetRules">{{ t('admin.accounts.opencode.defaults') }}</button>
      </div>
      <div v-for="(rule, index) in rules" :key="index" class="opencode-rule">
        <input class="input" :value="rule.pattern" :aria-label="t('admin.accounts.opencode.pattern')" maxlength="128" @input="updateRule(index, 'pattern', ($event.target as HTMLInputElement).value)" />
        <select class="input" :value="rule.protocol" :aria-label="t('admin.accounts.opencode.protocol')" @change="updateRule(index, 'protocol', ($event.target as HTMLSelectElement).value)">
          <option value="chat_completions">Chat Completions</option>
          <option value="responses">Responses</option>
          <option value="anthropic">Anthropic Messages</option>
        </select>
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
import { defaultOpenCodeRules, openCodeBaseUrl, type OpenCodeSettings, type OpenCodeProtocol, type OpenCodeRule } from './openCodeCredentials'

const props = defineProps<{ modelValue: OpenCodeSettings; baseUrl: string }>()
const emit = defineEmits<{ 'update:modelValue': [OpenCodeSettings]; 'update:baseUrl': [string] }>()
const { t } = useI18n()
const rules = computed(() => props.modelValue.protocol_rules ?? defaultOpenCodeRules(props.modelValue.account_mode))
function changeMode(event: Event) {
  const mode = (event.target as HTMLSelectElement).value === 'zen' ? 'zen' : 'go'
  if (!props.baseUrl || props.baseUrl === openCodeBaseUrl(props.modelValue.account_mode)) emit('update:baseUrl', openCodeBaseUrl(mode))
  emit('update:modelValue', { ...props.modelValue, account_mode: mode })
}
function changeProtocol(event: Event) {
  emit('update:modelValue', { ...props.modelValue, api_protocol: (event.target as HTMLSelectElement).value as OpenCodeProtocol })
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
