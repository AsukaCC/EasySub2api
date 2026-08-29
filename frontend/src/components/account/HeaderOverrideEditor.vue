<template>
  <div v-if="rows.length > 0" class="components-account-header-override-editor__panel">
    <div
      v-for="(row, index) in rows"
      :key="getHeaderOverrideRowKey(row)"
      class="components-account-header-override-editor__panel-2"
    >
      <input
        v-model="row.name"
        type="text"
        class="components-account-header-override-editor__field input"
        :placeholder="t('admin.accounts.headerOverride.namePlaceholder')"
      />
      <input
        v-model="row.value"
        type="text"
        class="components-account-header-override-editor__field input"
        :placeholder="t('admin.accounts.headerOverride.valuePlaceholder')"
      />
      <button
        type="button"
        class="components-account-header-override-editor__action"
        @click="removeRow(index)"
      >
        <svg class="components-account-header-override-editor__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          />
        </svg>
      </button>
    </div>
  </div>

  <button
    type="button"
    class="components-account-header-override-editor__action-2"
    @click="addRow"
  >
    <svg class="components-account-header-override-editor__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
    </svg>
    {{ t('admin.accounts.headerOverride.addRow') }}
  </button>

  <div class="components-account-header-override-editor__panel-3">
    <HeaderOverrideJsonTools :rows="rows" @update:rows="emit('update:rows', $event)" />
  </div>

  <p class="components-account-header-override-editor__description">
    {{ t('admin.accounts.headerOverride.emptyValueHint') }}
  </p>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import HeaderOverrideJsonTools from './HeaderOverrideJsonTools.vue'
import type { HeaderOverrideRow } from './credentialsBuilder'

const props = defineProps<{
  rows: HeaderOverrideRow[]
}>()

const emit = defineEmits<{
  (e: 'update:rows', rows: HeaderOverrideRow[]): void
}>()

const { t } = useI18n()

const getHeaderOverrideRowKey = createStableObjectKeyResolver<HeaderOverrideRow>(
  'header-override-row'
)

const addRow = () => {
  emit('update:rows', [...props.rows, { name: '', value: '' }])
}

const removeRow = (index: number) => {
  emit('update:rows', props.rows.filter((_, i) => i !== index))
}
</script>
