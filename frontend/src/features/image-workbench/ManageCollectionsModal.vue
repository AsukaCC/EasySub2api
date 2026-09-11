<template>
  <BaseDialog
    :show="show"
    :title="t('imageWorkbench.manageCollections')"
    width="normal"
    close-on-click-outside
    :z-index="60"
    @close="emit('close')"
  >
    <p class="input-hint">{{ t('imageWorkbench.manageCollectionsHint') }}</p>
    <ul class="collection-list">
      <li v-for="collection in collections" :key="collection.id" class="collection-list__item">
        <input
          v-if="editingId === collection.id"
          v-model="editingName"
          class="input"
          maxlength="60"
          @keydown.enter.prevent="confirmRename"
          @keydown.esc.prevent="cancelRename"
          @blur="confirmRename"
        />
        <span v-else>{{ collection.name }}</span>
        <div class="collection-list__actions">
          <button type="button" :class="{ active: collection.id === defaultFavoriteCollectionId }" :title="t('imageWorkbench.setDefaultCollection')" @click="emit('setDefault', collection.id)">
            <Icon name="star" size="xs" />
          </button>
          <button type="button" :title="t('common.edit')" @click="startRename(collection)">
            <Icon name="edit" size="xs" />
          </button>
          <button type="button" :disabled="collections.length <= 1" :title="t('common.delete')" @click="emit('remove', collection.id)">
            <Icon name="trash" size="xs" />
          </button>
        </div>
      </li>
    </ul>
    <div class="collection-create">
      <input v-model="draft" class="input" :placeholder="t('imageWorkbench.newCollectionPlaceholder')" @keydown.enter.prevent="create" />
      <button type="button" class="btn btn-secondary" :disabled="!draft.trim()" @click="create">{{ t('common.create') }}</button>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { FavoriteCollection } from './favorites'

const props = defineProps<{
  show: boolean
  collections: FavoriteCollection[]
  defaultFavoriteCollectionId: string | null
}>()

const emit = defineEmits<{
  close: []
  create: [name: string]
  rename: [id: string, name: string]
  remove: [id: string]
  setDefault: [id: string]
}>()

const { t } = useI18n()
const draft = ref('')
const editingId = ref<string | null>(null)
const editingName = ref('')

watch(() => props.show, (open) => {
  if (!open) return
  draft.value = ''
  editingId.value = null
  editingName.value = ''
})

function startRename(collection: FavoriteCollection) {
  editingId.value = collection.id
  editingName.value = collection.name
}

function confirmRename() {
  if (editingId.value && editingName.value.trim()) emit('rename', editingId.value, editingName.value.trim())
  cancelRename()
}

function cancelRename() {
  editingId.value = null
  editingName.value = ''
}

function create() {
  if (!draft.value.trim()) return
  emit('create', draft.value)
  draft.value = ''
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

.collection-list__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  padding: 0.45rem 0.55rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg-subtle);
}

.collection-list__item span {
  min-width: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-list__actions {
  display: flex;
  gap: 0.15rem;
}

.collection-list__actions button {
  display: inline-grid;
  place-items: center;
  width: 1.8rem;
  height: 1.8rem;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.collection-list__actions button.active,
.collection-list__actions button:hover:not(:disabled) {
  color: var(--theme-accent);
}

.collection-list__actions button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.collection-create {
  display: flex;
  gap: 0.5rem;
}
</style>
