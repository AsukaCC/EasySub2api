<template>
  <BaseDialog
    :show="show"
    :title="t('imageWorkbench.saveToCollection')"
    width="normal"
    close-on-click-outside
    :z-index="60"
    @close="emit('close')"
  >
    <p class="input-hint">{{ t('imageWorkbench.saveToCollectionHint') }}</p>
    <ul class="collection-list">
      <li v-for="collection in collections" :key="collection.id">
        <label class="collection-check">
          <input v-model="checkedIds" type="checkbox" :value="collection.id" />
          <span>{{ collection.name }}</span>
        </label>
      </li>
    </ul>
    <div class="collection-create">
      <input v-model="draft" class="input" :placeholder="t('imageWorkbench.newCollectionPlaceholder')" @keydown.enter.prevent="create" />
      <button type="button" class="btn btn-secondary" :disabled="!draft.trim()" @click="create">{{ t('common.create') }}</button>
    </div>
    <template #footer>
      <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" @click="confirm">{{ t('common.confirm') }}</button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { FavoriteCollection } from './favorites'

const props = defineProps<{
  show: boolean
  collections: FavoriteCollection[]
  initialIds: string[]
}>()

const emit = defineEmits<{
  close: []
  create: [name: string]
  confirm: [ids: string[]]
}>()

const { t } = useI18n()
const draft = ref('')
const checkedIds = ref<string[]>([])
const idsAtOpen = ref<string[]>([])

watch(() => props.show, (open) => {
  if (!open) return
  draft.value = ''
  checkedIds.value = [...props.initialIds]
  idsAtOpen.value = props.collections.map((collection) => collection.id)
})

watch(() => props.collections.map((collection) => collection.id).join(','), () => {
  if (!props.show) return
  for (const collection of props.collections) {
    if (!idsAtOpen.value.includes(collection.id) && !checkedIds.value.includes(collection.id)) {
      checkedIds.value.push(collection.id)
    }
  }
})

function create() {
  if (!draft.value.trim()) return
  emit('create', draft.value)
  draft.value = ''
}

function confirm() {
  emit('confirm', [...checkedIds.value])
}
</script>

<style scoped>
.input-hint {
  margin: 0 0 0.85rem;
}

.collection-list {
  display: grid;
  gap: 0.35rem;
  margin: 0 0 1rem;
  padding: 0;
  list-style: none;
}

.collection-check {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.5rem 0.6rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-subtle);
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
}

.collection-create {
  display: flex;
  gap: 0.5rem;
}
</style>
