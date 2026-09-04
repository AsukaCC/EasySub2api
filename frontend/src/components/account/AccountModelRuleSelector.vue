<template>
  <div v-if="!disabled && isSupportedPlatform" class="model-rule-selector">
    <div class="model-rule-selector__header">
      <label class="input-label">{{ t('admin.accounts.modelRules.bindingLabel') }}</label>
      <span v-if="loading" class="input-hint">{{ t('admin.accounts.modelRules.loading') }}</span>
    </div>

    <div v-if="currentBindingLabel" class="model-rule-selector__current">
      <span>{{ t('admin.accounts.modelRules.currentBinding') }}</span>
      <strong>{{ currentBindingLabel }}</strong>
    </div>

    <div class="model-rule-selector__controls">
      <Select
        v-model="selectedRuleId"
        :options="ruleOptions"
        :placeholder="t('admin.accounts.modelRules.selectPlaceholder')"
        :disabled="loading || rules.length === 0"
        :searchable="'auto'"
        :clearable="true"
        class="model-rule-selector__select"
      />
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="!selectedRule || loading"
        :title="t('admin.accounts.modelRules.bindAction')"
        @click="requestApply"
      >
        <Icon name="link" size="sm" />
        <span>{{ t('admin.accounts.modelRules.bindAction') }}</span>
      </button>
      <button
        v-if="modelValue || allowUnbind"
        type="button"
        class="btn btn-ghost btn-sm"
        :disabled="loading"
        :title="t('admin.accounts.modelRules.unbindAction')"
        @click="unbindRule"
      >
        <Icon name="x" size="sm" />
        <span>{{ t('admin.accounts.modelRules.unbindAction') }}</span>
      </button>
    </div>
    <p v-if="rules.length === 0 && !loading" class="input-hint">
      {{ t('admin.accounts.modelRules.noRulesForPlatform') }}
    </p>
    <p v-if="loadError" class="model-rule-selector__error">{{ loadError }}</p>

    <ConfirmDialog
      :show="showOverwriteConfirm"
      :title="t('admin.accounts.modelRules.overwriteTitle')"
      :message="t('admin.accounts.modelRules.bindConfirmMessage', { name: pendingRule?.name || '' })"
      :confirm-text="t('admin.accounts.modelRules.bindConfirm')"
      :cancel-text="t('common.cancel')"
      @confirm="confirmApply"
      @cancel="cancelApply"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { AccountModelRoute, AccountModelRule } from '@/api/admin/accountModelRules'
import type { AccountPlatform } from '@/types'
import { isAccountPlatform } from '@/utils/accountPlatforms'
import type { ModelMappingEntry } from '@/composables/useModelWhitelist'

interface AccountModelRuleBindingPayload {
  id: string | null
  name: string
  routes: AccountModelRoute[]
  allowedModels: string[]
  mappings: ModelMappingEntry[]
}

const props = withDefaults(defineProps<{
  platform: string | null | undefined
  subscriptionTier?: string | null
  modelValue?: string | null
  currentRuleName?: string | null
  allowUnbind?: boolean
  disabled?: boolean
  hasExistingMappings?: boolean
}>(), {
  subscriptionTier: null,
  modelValue: null,
  currentRuleName: null,
  allowUnbind: false,
  disabled: false,
  hasExistingMappings: false
})

const emit = defineEmits<{
  (event: 'apply', payload: AccountModelRuleBindingPayload): void
}>()

const { t } = useI18n()
const rules = ref<AccountModelRule[]>([])
const selectedRuleId = ref<string | null>(props.modelValue)
const loading = ref(false)
const loadError = ref('')
const showOverwriteConfirm = ref(false)
const pendingRule = ref<AccountModelRule | null>(null)

const isSupportedPlatform = computed(() => isAccountPlatform(props.platform))
const selectedRule = computed(() => rules.value.find(rule => rule.id === selectedRuleId.value) || null)
const currentBindingLabel = computed(() => {
  if (!props.modelValue) return ''
  return props.currentRuleName || rules.value.find(rule => rule.id === props.modelValue)?.name || props.modelValue
})
const ruleOptions = computed(() => rules.value.map(rule => ({
  value: rule.id,
  label: `${rule.name} · ${rule.subscription_tier || t('admin.accounts.modelRules.allTiers')} · ${t('admin.accounts.modelRules.routeCount', { count: (rule.model_routes || []).length })}`
})))

async function loadRules() {
  rules.value = []
  loadError.value = ''
  if (props.disabled || !isAccountPlatform(props.platform)) return
  if (!adminAPI.accountModelRules?.list) return

  loading.value = true
  try {
    rules.value = await adminAPI.accountModelRules.list(
      props.platform as AccountPlatform,
      props.subscriptionTier || undefined
    )
    selectedRuleId.value = props.modelValue
  } catch (error) {
    loadError.value = t('admin.accounts.modelRules.loadFailed')
    console.error('Failed to load account model rules:', error)
  } finally {
    loading.value = false
  }
}

function requestApply() {
  if (!selectedRule.value) return
  if (props.hasExistingMappings || (props.modelValue && props.modelValue !== selectedRule.value.id)) {
    pendingRule.value = selectedRule.value
    showOverwriteConfirm.value = true
    return
  }
  emitApply(selectedRule.value)
}

function confirmApply() {
  if (pendingRule.value) emitApply(pendingRule.value)
  cancelApply()
}

function cancelApply() {
  showOverwriteConfirm.value = false
  pendingRule.value = null
}

function emitApply(rule: AccountModelRule) {
  const routes = (rule.model_routes || []).map(route => ({ ...route }))
  const mappings = routes.map(route => ({
    from: route.request_model,
    to: route.upstream_model,
    ...(route.reasoning_effort ? { reasoning_effort: route.reasoning_effort } : {})
  }))
  emit('apply', {
    id: rule.id,
    name: rule.name,
    routes,
    allowedModels: [],
    mappings
  })
}

function unbindRule() {
  selectedRuleId.value = null
  emit('apply', {
    id: null,
    name: props.currentRuleName || '',
    routes: [],
    allowedModels: [],
    mappings: []
  })
}

watch(
  [() => props.platform, () => props.subscriptionTier, () => props.disabled],
  () => { void loadRules() },
  { immediate: true }
)
watch(() => props.modelValue, (value) => {
  selectedRuleId.value = value
})
</script>

<style scoped>
.model-rule-selector {
  margin: 0.75rem 0;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
}

.model-rule-selector__header,
.model-rule-selector__controls,
.model-rule-selector__current {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.model-rule-selector__header {
  justify-content: space-between;
  margin-bottom: 0.4rem;
}

.model-rule-selector__current {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

.model-rule-selector__current strong {
  color: var(--color-text-primary);
}

.model-rule-selector__select {
  flex: 1;
  min-width: 0;
}

.model-rule-selector__error {
  margin: 0.4rem 0 0;
  color: var(--color-text-danger);
  font-size: var(--font-size-xs);
}

@media (max-width: 640px) {
  .model-rule-selector__controls {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
