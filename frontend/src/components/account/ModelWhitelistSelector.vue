<template>
  <div>
    <!-- Multi-select Dropdown -->
    <div ref="containerRef" class="model-whitelist__root" @keydown.esc.prevent.stop="closeDropdown">
      <div
        ref="triggerRef"
        @click="toggleDropdown"
        class="model-whitelist__trigger"
        role="button"
        tabindex="0"
        :aria-expanded="showDropdown"
        aria-haspopup="listbox"
        :aria-label="t('admin.accounts.selectAllowedModels')"
        @keydown.enter.self.prevent="toggleDropdown"
        @keydown.space.self.prevent="toggleDropdown"
      >
        <div class="model-whitelist__tags">
          <span v-if="modelValue.length === 0" class="model-whitelist__placeholder">{{ t('common.selectOption') }}</span>
          <span
            v-for="model in modelValue"
            :key="model"
            class="model-whitelist__tag"
          >
            <span class="model-whitelist__tag-content">
              <ModelIcon :model="model" size="14px" />
              <span class="model-whitelist__tag-label">{{ model }}</span>
            </span>
            <button
              type="button"
              @click.stop="removeModel(model)"
              :aria-label="`${t('common.delete')} ${model}`"
              class="model-whitelist__remove"
            >
              <Icon name="x" size="xs" class="model-whitelist__remove-icon" :stroke-width="2" />
            </button>
          </span>
        </div>
        <div class="model-whitelist__summary">
          <span class="model-whitelist__count">{{ t('admin.accounts.modelCount', { count: modelValue.length }) }}</span>
          <svg class="model-whitelist__chevron" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      <!-- Dropdown List -->
      <Teleport to="body">
        <div
          v-if="showDropdown"
          ref="panelRef"
          :style="panelStyle"
          class="model-whitelist__dropdown dropdown dropdown--portal"
          role="listbox"
          aria-multiselectable="true"
          :aria-label="t('admin.accounts.selectAllowedModels')"
          @click.stop
        >
        <div class="model-whitelist__search">
          <input
            v-model="searchQuery"
            type="text"
            class="model-whitelist__search-input input"
            :placeholder="t('admin.accounts.searchModels')"
            :aria-label="t('admin.accounts.searchModels')"
            @click.stop
          />
        </div>
        <div class="model-whitelist__options">
          <div
            v-for="model in filteredModels"
            :key="model.value"
            data-testid="model-option"
            class="model-whitelist__option"
          >
            <button
              type="button"
              data-testid="select-model"
              class="model-whitelist__select"
              role="option"
              :aria-selected="modelValue.includes(model.value)"
              @click="toggleModel(model.value)"
            >
              <span
                :class="[
                  'model-whitelist__checkbox',
                  modelValue.includes(model.value)
                    ? 'model-whitelist__checked'
                    : 'model-whitelist__unchecked'
                ]"
              >
                <svg v-if="modelValue.includes(model.value)" class="model-whitelist__check-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                </svg>
              </span>
              <ModelIcon :model="model.value" size="18px" />
              <span class="model-whitelist__model-name">{{ model.value }}</span>
            </button>
            <button
              type="button"
              data-testid="copy-model-id"
              class="model-whitelist__copy"
              :title="`${t('common.copy')} ${model.value}`"
              :aria-label="`${t('common.copy')} ${model.value}`"
              @click="copyModelId(model.value)"
            >
              <Icon name="copy" size="sm" />
            </button>
          </div>
          <div v-if="filteredModels.length === 0" class="model-whitelist__empty">
            {{ t('admin.accounts.noMatchingModels') }}
          </div>
        </div>
        </div>
      </Teleport>
    </div>

    <!-- Quick Actions -->
    <div class="model-whitelist__actions">
      <button
        v-if="canSyncUpstream && showSyncButton"
        type="button"
        data-testid="sync-upstream-models"
        @click="syncUpstreamModels"
        :disabled="isSyncingUpstream"
        class="model-whitelist__sync"
      >
        {{ isSyncingUpstream ? t('admin.accounts.syncUpstreamModelsLoading') : t('admin.accounts.syncUpstreamModels') }}
      </button>
      <button
        type="button"
        @click="clearAll"
        class="model-whitelist__clear"
      >
        {{ t('admin.accounts.clearAllModels') }}
      </button>
    </div>

    <!-- Custom Model Input -->
    <div class="model-whitelist__custom">
      <label class="model-whitelist__label">{{ t('admin.accounts.customModelName') }}</label>
      <div class="model-whitelist__custom-row">
        <input
          v-model="customModel"
          type="text"
          class="model-whitelist__custom-input input"
          :placeholder="t('admin.accounts.enterCustomModelName')"
          @keydown.enter.prevent="handleEnter"
          @compositionstart="isComposing = true"
          @compositionend="isComposing = false"
        />
        <button
          type="button"
          @click="addCustom"
          class="model-whitelist__add"
        >
          {{ t('admin.accounts.addModel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { accountsAPI } from '@/api/admin/accounts'
import type { SyncUpstreamPreviewParams } from '@/api/admin/accounts'
import { useClipboard } from '@/composables/useClipboard'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { allModels, getModelsByPlatform } from '@/composables/useModelWhitelist'
import { useFloatingPanel } from '@/composables/useFloatingPanel'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  showSyncButton?: boolean
  modelValue: string[]
  platform?: string
  platforms?: string[]
  accountId?: string
  accountType?: string
  syncPlatform?: string
  syncCredentials?: {
    platform: string
    type: string
    base_url?: string
    api_key: string
    model_mapping?: Record<string, string>
  }
}>(), { showSyncButton: true })

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
  'upstream-synced': []
  'upstream-sync-finished': []
}>()

const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const showDropdown = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, showDropdown, {
  maxWidth: 480,
  align: 'start',
  minComfortableHeight: 220,
  zIndex: 100000020
})
const searchQuery = ref('')
const customModel = ref('')
const isComposing = ref(false)
const isSyncingUpstream = ref(false)
let upstreamSyncVersion = 0
const syncedModels = ref<string[]>([])
const normalizedPlatforms = computed(() => {
  const rawPlatforms =
    props.platforms && props.platforms.length > 0
      ? props.platforms
      : props.platform
        ? [props.platform]
        : []

  return Array.from(
    new Set(
      rawPlatforms
        .map(platform => platform?.trim())
        .filter((platform): platform is string => Boolean(platform))
    )
  )
})

const canSyncPlatformAccountType = (platform: string, accountType: string) => {
  const normalizedPlatform = platform.trim().toLowerCase()
  const normalizedType = accountType.trim().toLowerCase()

  switch (normalizedPlatform) {
    case 'anthropic':
      return ['apikey', 'oauth', 'setup-token'].includes(normalizedType)
    case 'openai':
      return ['apikey', 'oauth'].includes(normalizedType)
    case 'gemini':
      return ['apikey', 'oauth'].includes(normalizedType)
    case 'antigravity':
      return normalizedType === 'apikey'
    case 'grok':
      return ['apikey', 'oauth'].includes(normalizedType)
    case 'kimi':
    case 'zhipu':
    case 'deepseek':
    case 'minimax':
      return normalizedType === 'apikey'
    default:
      return false
  }
}
const canSyncUpstream = computed(() => {
  if (props.accountId) {
    const syncPlatform = props.syncPlatform?.trim() || normalizedPlatforms.value[0]
    if (!props.accountType || !syncPlatform) return true
    return canSyncPlatformAccountType(syncPlatform, props.accountType)
  }
  if (props.syncCredentials) {
    return canSyncPlatformAccountType(props.syncCredentials.platform, props.syncCredentials.type)
  }
  return false
})

const availableOptions = computed(() => {
  const allowedModels = new Set<string>()
  if (normalizedPlatforms.value.length === 0) {
    for (const model of allModels) allowedModels.add(model.value)
  } else {
    for (const platform of normalizedPlatforms.value) {
      for (const model of getModelsByPlatform(platform)) {
        allowedModels.add(model)
      }
    }
  }

  // Upstream model IDs and existing custom entries are valid whitelist values
  // even when they are not part of the static local catalog.
  for (const model of syncedModels.value) allowedModels.add(model)
  for (const model of props.modelValue) allowedModels.add(model)

  return Array.from(allowedModels, value => ({ value, label: value }))
})

const filteredModels = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  if (!query) return availableOptions.value
  return availableOptions.value.filter(
    m => m.value.toLowerCase().includes(query) || m.label.toLowerCase().includes(query)
  )
})

const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value
  if (!showDropdown.value) searchQuery.value = ''
}

const closeDropdown = () => {
  showDropdown.value = false
  searchQuery.value = ''
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node
  if (containerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  closeDropdown()
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && showDropdown.value) closeDropdown()
}

const removeModel = (model: string) => {
  emit('update:modelValue', props.modelValue.filter(m => m !== model))
}

const toggleModel = (model: string) => {
  if (props.modelValue.includes(model)) {
    removeModel(model)
  } else {
    emit('update:modelValue', [...props.modelValue, model])
  }
}

const copyModelId = async (model: string) => {
  await copyToClipboard(model)
}

const addCustom = () => {
  const model = customModel.value.trim()
  if (!model) return
  if (props.modelValue.includes(model)) {
    appStore.showInfo(t('admin.accounts.modelExists'))
    return
  }
  emit('update:modelValue', [...props.modelValue, model])
  customModel.value = ''
}

const handleEnter = () => {
  if (!isComposing.value) addCustom()
}

const syncUpstreamModels = async () => {
  if (isSyncingUpstream.value) return
  if (!props.accountId && !props.syncCredentials) return

  const syncVersion = upstreamSyncVersion
  const accountId = props.accountId
  const isCurrentSync = () => syncVersion === upstreamSyncVersion
  isSyncingUpstream.value = true
  try {
    let result
    if (props.accountId) {
      result = await accountsAPI.syncUpstreamModels(props.accountId)
    } else if (props.syncCredentials) {
      result = await accountsAPI.syncUpstreamModelsPreview(props.syncCredentials as SyncUpstreamPreviewParams)
    } else {
      return
    }

    if (!isCurrentSync()) return
    if (result.discovery_source === "configured_fallback") { appStore.showWarning(t("admin.accounts.grokModelDiagnostics.configuredFallback")); return }
    const upstreamModels = result.models.map(model => model.trim()).filter(Boolean)
    if (upstreamModels.length === 0) {
      appStore.showInfo(t('admin.accounts.syncUpstreamModelsEmpty'))
      return
    }

    syncedModels.value = Array.from(new Set([...syncedModels.value, ...upstreamModels]))
    const newModels = [...props.modelValue]
    let addedCount = 0
    for (const model of upstreamModels) {
      if (!newModels.includes(model)) {
        newModels.push(model)
        addedCount += 1
      }
    }

    emit('update:modelValue', newModels)
    emit('upstream-synced')
    if (result.warnings?.some(warning => warning.code === 'upstream_model_metadata_incomplete')) {
      appStore.showWarning(t('admin.accounts.syncUpstreamModelsMetadataIncomplete'))
      return
    }
    if (addedCount > 0) {
      appStore.showSuccess(t('admin.accounts.syncUpstreamModelsSuccess', { count: addedCount, total: upstreamModels.length }))
    } else {
      appStore.showInfo(t('admin.accounts.syncUpstreamModelsNoChanges', { count: upstreamModels.length }))
    }
  } catch (error) {
    if (!isCurrentSync()) return
    const message = extractApiErrorMessage(error, t('admin.accounts.syncUpstreamModelsFailed'))
    appStore.showError(t('admin.accounts.syncUpstreamModelsError', { message }))
  } finally {
    if (isCurrentSync()) { isSyncingUpstream.value = false; if (accountId) emit("upstream-sync-finished") }
  }
}

defineExpose({ syncUpstreamModels, isSyncingUpstream })

watch(() => props.accountId, () => {
  upstreamSyncVersion += 1
  isSyncingUpstream.value = false
  syncedModels.value = []
})

const clearAll = () => {
  emit('update:modelValue', [])
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  upstreamSyncVersion += 1
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})

</script>

<style scoped>
.model-whitelist__root { position: relative; min-width: 0; margin-bottom: 0.75rem; }
.model-whitelist__trigger {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.75rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-md);
  background: var(--color-surface);
  cursor: pointer;
}
.model-whitelist__trigger:focus-visible { outline: 2px solid var(--color-primary-border); outline-offset: 2px; }
.model-whitelist__tags { display: flex; flex: 1; flex-wrap: wrap; gap: 0.375rem; min-width: 0; max-height: 9rem; overflow: auto; }
.model-whitelist__placeholder { color: var(--color-text-placeholder); font-size: var(--font-size-sm); }
.model-whitelist__tag { display: inline-flex; align-items: center; gap: 0.375rem; max-width: 100%; padding: 0.25rem 0.5rem; border-radius: var(--radius-sm); background: var(--color-primary-subtle); color: var(--color-text-primary); font-size: var(--font-size-xs); }
.model-whitelist__tag-content { display: flex; align-items: center; gap: 0.375rem; min-width: 0; }
.model-whitelist__tag-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.model-whitelist__remove { display: inline-flex; flex-shrink: 0; padding: 0.125rem; border-radius: var(--radius-full); }
.model-whitelist__remove:hover, .model-whitelist__copy:hover { background: var(--color-surface-hover); }
.model-whitelist__summary { display: flex; flex-shrink: 0; align-items: center; gap: 0.5rem; color: var(--color-text-tertiary); }
.model-whitelist__count { font-size: var(--font-size-xs); white-space: nowrap; }
.model-whitelist__chevron { width: 1.25rem; height: 1.25rem; }
.model-whitelist__dropdown {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--glass-layer-floating-bg);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--glass-shadow-hover);
  color: var(--color-text-primary);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
}
.model-whitelist__search { flex-shrink: 0; padding: 0.5rem; border-bottom: 1px solid var(--color-border); }
.model-whitelist__search-input { width: 100%; }
.model-whitelist__options { min-height: 0; max-height: 16rem; overflow-y: auto; overscroll-behavior: contain; padding: 0.25rem; }
.model-whitelist__option { display: flex; align-items: center; border-radius: var(--radius-sm); }
.model-whitelist__option:hover, .model-whitelist__option:focus-within { background: var(--color-surface-hover); }
.model-whitelist__select { display: flex; flex: 1; align-items: center; gap: 0.5rem; min-width: 0; padding: 0.625rem 0.5rem; text-align: left; font-size: var(--font-size-sm); color: var(--color-text-primary); }
.model-whitelist__select[aria-selected="true"] { background: var(--color-primary-subtle); border-radius: var(--radius-sm); }
.model-whitelist__checkbox { display: flex; flex-shrink: 0; align-items: center; justify-content: center; width: 1rem; height: 1rem; border: 1px solid var(--color-border-strong); border-radius: var(--radius-xs); }
.model-whitelist__checked { background: var(--color-primary); border-color: var(--color-primary); color: var(--color-text-inverse); }
.model-whitelist__check-icon { width: 0.75rem; height: 0.75rem; }
.model-whitelist__model-name { min-width: 0; overflow-wrap: anywhere; }
.model-whitelist__copy { display: inline-flex; flex-shrink: 0; padding: 0.5rem; border-radius: var(--radius-sm); color: var(--color-text-tertiary); }
.model-whitelist__empty { padding: 1rem; text-align: center; color: var(--color-text-tertiary); font-size: var(--font-size-sm); }
.model-whitelist__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 1rem; }
.model-whitelist__sync, .model-whitelist__clear, .model-whitelist__add { padding: 0.375rem 0.75rem; border: 1px solid var(--color-border); border-radius: var(--radius-md); color: var(--color-text-primary); font-size: var(--font-size-sm); }
.model-whitelist__sync:hover, .model-whitelist__clear:hover, .model-whitelist__add:hover { background: var(--color-surface-hover); }
.model-whitelist__sync:disabled { cursor: not-allowed; opacity: 0.6; }
.model-whitelist__label { display: block; margin-bottom: 0.375rem; color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.model-whitelist__custom-row { display: flex; gap: 0.5rem; }
.model-whitelist__custom-input { flex: 1; min-width: 0; }
.model-whitelist__add { flex-shrink: 0; }
</style>
