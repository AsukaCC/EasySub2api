<template>
  <div v-if="!disabled && isSupportedPlatform" class="account-model-rule-selector">
    <div class="account-model-rule-selector__header">
      <label class="input-label">{{ t('admin.accounts.modelRules.importLabel') }}</label>
      <span v-if="loading" class="input-hint">{{ t('admin.accounts.modelRules.loading') }}</span>
    </div>
    <div class="account-model-rule-selector__controls">
      <Select
        v-model="selectedRuleId"
        :options="ruleOptions"
        :placeholder="t('admin.accounts.modelRules.selectPlaceholder')"
        :disabled="loading || rules.length === 0"
        :searchable="'auto'"
        :clearable="true"
        class="account-model-rule-selector__select"
      />
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="!selectedRule || loading"
        :title="t('admin.accounts.modelRules.importAction')"
        @click="requestApply"
      >
        <Icon name="download" size="sm" />
        <span>{{ t('admin.accounts.modelRules.importAction') }}</span>
      </button>
    </div>
    <p v-if="rules.length === 0 && !loading" class="input-hint">
      {{ t('admin.accounts.modelRules.noRulesForPlatform') }}
    </p>
    <p v-if="loadError" class="account-model-rule-selector__error">{{ loadError }}</p>

    <ConfirmDialog
      :show="showOverwriteConfirm"
      :title="t('admin.accounts.modelRules.overwriteTitle')"
      :message="t('admin.accounts.modelRules.overwriteMessage', { name: pendingRule?.name || '' })"
      :confirm-text="t('admin.accounts.modelRules.overwriteConfirm')"
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
import type { AccountModelRule } from '@/api/admin/accountModelRules'
import type { AccountPlatform } from '@/types'
import { isAccountPlatform } from '@/utils/accountPlatforms'
import type { ModelMappingEntry } from '@/composables/useModelWhitelist'

const props = withDefaults(defineProps<{
  platform: string | null | undefined
  disabled?: boolean
  hasExistingMappings?: boolean
}>(), {
  disabled: false,
  hasExistingMappings: false
})

const emit = defineEmits<{
  (event: 'apply', payload: { name: string; allowedModels: string[]; mappings: ModelMappingEntry[] }): void
}>()

const { t } = useI18n()
const rules = ref<AccountModelRule[]>([])
const selectedRuleId = ref<string | null>(null)
const loading = ref(false)
const loadError = ref('')
const showOverwriteConfirm = ref(false)
const pendingRule = ref<AccountModelRule | null>(null)

const isSupportedPlatform = computed(() => isAccountPlatform(props.platform))
const selectedRule = computed(() => rules.value.find(rule => rule.id === selectedRuleId.value) || null)
const ruleOptions = computed(() => rules.value.map(rule => ({
  value: rule.id,
  label: `${rule.name} (${(rule.whitelist || []).length} + ${Object.keys(rule.mapping || {}).length})`
})))

async function loadRules(platform: string | null | undefined) {
  selectedRuleId.value = null
  rules.value = []
  loadError.value = ''
  if (props.disabled || !isAccountPlatform(platform)) return

  loading.value = true
  try {
    rules.value = await adminAPI.accountModelRules.list(platform as AccountPlatform)
  } catch (error) {
    loadError.value = t('admin.accounts.modelRules.loadFailed')
    console.error('Failed to load account model rules:', error)
  } finally {
    loading.value = false
  }
}

function requestApply() {
  if (!selectedRule.value) return
  if (props.hasExistingMappings) {
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
  const mappings = Object.entries(rule.mapping || {}).map(([from, to]) => ({
    from,
    to,
    ...(rule.reasoning_efforts?.[from] ? { reasoning_effort: rule.reasoning_efforts[from] } : {})
  }))
  emit('apply', { name: rule.name, allowedModels: [...(rule.whitelist || [])], mappings })
}

watch(() => props.platform, loadRules, { immediate: true })
watch(() => props.disabled, (disabled) => {
  if (disabled) {
    selectedRuleId.value = null
    rules.value = []
  } else {
    void loadRules(props.platform)
  }
})
</script>

<style scoped>
.account-model-rule-selector {
  margin: 0.75rem 0;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface-muted);
}

.account-model-rule-selector__header,
.account-model-rule-selector__controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.account-model-rule-selector__header {
  justify-content: space-between;
  margin-bottom: 0.4rem;
}

.account-model-rule-selector__select {
  flex: 1;
  min-width: 0;
}

.account-model-rule-selector__error {
  margin: 0.4rem 0 0;
  color: var(--color-text-danger);
  font-size: var(--font-size-xs);
}

@media (max-width: 640px) {
  .account-model-rule-selector__controls {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
