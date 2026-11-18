<template>
  <section class="model-forwarding" data-testid="group-model-forwarding">
    <h3 class="input-label">{{ t('admin.groups.modelForwarding') }}</h3>
    <p class="input-hint">{{ t('admin.groups.modelForwardingHint') }}</p>
    <div v-for="(row, index) in modelValue" :key="index" class="model-forwarding__row">
      <input
        :value="row.from"
        class="input"
        required
        pattern="[^\s*]*\*?"
        :aria-label="t('admin.accounts.requestModel')"
        :placeholder="t('admin.accounts.requestModel')"
        data-testid="forwarding-source"
        @input="updateRow(index, 'from', ($event.target as HTMLInputElement).value)"
      />
      <span aria-hidden="true">→</span>
      <input
        :value="row.to"
        class="input"
        required
        :list="listID"
        :aria-label="t('admin.accounts.actualModel')"
        :placeholder="t('admin.accounts.actualModel')"
        data-testid="forwarding-target"
        @input="updateRow(index, 'to', ($event.target as HTMLInputElement).value)"
      />
      <button type="button" class="btn btn-secondary" :aria-label="t('common.delete')" @click="removeRow(index)">
        <Icon name="x" size="sm" />
      </button>
    </div>
    <datalist :id="listID">
      <option v-for="model in targetModels" :key="model" :value="model" />
    </datalist>
    <button type="button" class="btn btn-secondary" data-testid="add-forwarding" @click="emit('update:modelValue', [...modelValue, { from: '', to: '' }])">
      {{ t('admin.groups.addModelForwarding') }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ModelMappingEntry } from '@/composables/useModelWhitelist'

const props = defineProps<{ modelValue: ModelMappingEntry[], candidates: string[] }>()
const emit = defineEmits<{ 'update:modelValue': [rows: ModelMappingEntry[]] }>()
const { t } = useI18n()
const listID = useId()
const targetModels = computed(() => props.candidates.filter(model => /^[^\s*]+$/.test(model)))

function updateRow(index: number, field: 'from' | 'to', value: string) {
  emit('update:modelValue', props.modelValue.map((row, i) => i === index ? { ...row, [field]: value } : row))
}

function removeRow(index: number) {
  emit('update:modelValue', props.modelValue.filter((_, i) => i !== index))
}
</script>

<style scoped>
.model-forwarding { display: grid; gap: 0.75rem; min-width: 0; }
.model-forwarding__row { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) auto; align-items: center; gap: 0.5rem; }
.model-forwarding__row > * { min-width: 0; }
@media (max-width: 480px) {
  .model-forwarding__row { grid-template-columns: minmax(0, 1fr) auto; }
  .model-forwarding__row > input { grid-column: 1 / -1; }
  .model-forwarding__row > span { display: none; }
}
</style>
