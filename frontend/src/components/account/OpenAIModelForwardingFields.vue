<template>
  <section class="model-forwarding" data-testid="openai-model-forwarding">
    <label class="input-label">{{ t('admin.accounts.openai.modelForwarding') }}</label>
    <p class="input-hint">{{ t('admin.accounts.openai.modelForwardingHint') }}</p>
    <div v-for="(row, index) in modelValue" :key="index" class="model-forwarding__row">
      <input
        :value="row.from"
        class="input"
        required
        pattern="[^\s*]+\*?"
        :aria-label="t('admin.accounts.requestModel')"
        placeholder="gpt-5.6-luna"
        data-testid="forwarding-source"
        @input="updateRow(index, 'from', ($event.target as HTMLInputElement).value)"
      />
      <span aria-hidden="true">→</span>
      <input
        :value="row.to"
        class="input"
        required
        pattern="[^\s*]+"
        :aria-label="t('admin.accounts.actualModel')"
        placeholder="gpt-5.6-terra"
        data-testid="forwarding-target"
        @input="updateRow(index, 'to', ($event.target as HTMLInputElement).value)"
      />
      <button type="button" class="btn btn-secondary" :aria-label="t('common.delete')" @click="removeRow(index)">
        <Icon name="x" size="sm" />
      </button>
    </div>
    <button type="button" class="btn btn-secondary" data-testid="add-forwarding" @click="emit('update:modelValue', [...modelValue, { from: '', to: '' }])">
      {{ t('admin.accounts.openai.addModelForwarding') }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ModelMappingEntry } from '@/composables/useModelWhitelist'

const props = defineProps<{ modelValue: ModelMappingEntry[] }>()
const emit = defineEmits<{ 'update:modelValue': [rows: ModelMappingEntry[]] }>()
const { t } = useI18n()

function updateRow(index: number, field: 'from' | 'to', value: string) {
  emit('update:modelValue', props.modelValue.map((row, i) => i === index ? { ...row, [field]: value } : row))
}

function removeRow(index: number) {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== index))
}
</script>

<style scoped>
.model-forwarding { display: grid; gap: 0.75rem; }
.model-forwarding__row { display: flex; align-items: center; gap: 0.5rem; }
.model-forwarding__row input { min-width: 0; flex: 1; }
</style>
