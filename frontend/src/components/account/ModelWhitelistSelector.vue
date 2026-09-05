<template>
  <div>
    <!-- Multi-select Dropdown -->
    <div ref="containerRef" class="components-account-model-whitelist-selector__panel" @keydown.esc.prevent.stop="closeDropdown">
      <div
        ref="triggerRef"
        @click="toggleDropdown"
        class="components-account-model-whitelist-selector__panel-2"
      >
        <div class="components-account-model-whitelist-selector__panel-3">
          <span
            v-for="model in modelValue"
            :key="model"
            class="components-account-model-whitelist-selector__text"
          >
            <span class="components-account-model-whitelist-selector__text-2">
              <ModelIcon :model="model" size="14px" />
              <span class="components-account-model-whitelist-selector__text-3">{{ model }}</span>
            </span>
            <button
              type="button"
              @click.stop="removeModel(model)"
              class="components-account-model-whitelist-selector__action"
            >
              <Icon name="x" size="xs" class="components-account-model-whitelist-selector__icon" :stroke-width="2" />
            </button>
          </span>
        </div>
        <div class="components-account-model-whitelist-selector__panel-4">
          <span class="components-account-model-whitelist-selector__text-4">{{ t('admin.accounts.modelCount', { count: modelValue.length }) }}</span>
          <svg class="components-account-model-whitelist-selector__icon-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
          class="components-account-model-whitelist-selector__panel-5"
          @click.stop
        >
        <div class="components-account-model-whitelist-selector__panel-6">
          <input
            v-model="searchQuery"
            type="text"
            class="components-account-model-whitelist-selector__field input"
            :placeholder="t('admin.accounts.searchModels')"
            @click.stop
          />
        </div>
        <div class="components-account-model-whitelist-selector__panel-7">
          <div
            v-for="model in filteredModels"
            :key="model.value"
            data-testid="model-option"
            class="components-account-model-whitelist-selector__panel-8"
          >
            <button
              type="button"
              data-testid="select-model"
              class="components-account-model-whitelist-selector__action-2"
              @click="toggleModel(model.value)"
            >
              <span
                :class="[
                  'components-account-model-whitelist-selector__text-6',
                  modelValue.includes(model.value)
                    ? 'components-account-model-whitelist-selector__text-7'
                    : 'components-account-model-whitelist-selector__text-8'
                ]"
              >
                <svg v-if="modelValue.includes(model.value)" class="components-account-model-whitelist-selector__icon-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                </svg>
              </span>
              <ModelIcon :model="model.value" size="18px" />
              <span class="components-account-model-whitelist-selector__text-5">{{ model.value }}</span>
            </button>
            <button
              type="button"
              data-testid="copy-model-id"
              class="components-account-model-whitelist-selector__action-3"
              :title="`${t('common.copy')} ${model.value}`"
              :aria-label="`${t('common.copy')} ${model.value}`"
              @click="copyModelId(model.value)"
            >
              <Icon name="copy" size="sm" />
            </button>
          </div>
          <div v-if="filteredModels.length === 0" class="components-account-model-whitelist-selector__panel-9">
            {{ t('admin.accounts.noMatchingModels') }}
          </div>
        </div>
        </div>
      </Teleport>
    </div>

    <!-- Quick Actions -->
    <div class="components-account-model-whitelist-selector__panel-10">
      <button
        type="button"
        @click="fillRelated"
        class="components-account-model-whitelist-selector__action-4"
      >
        {{ t('admin.accounts.fillRelatedModels') }}
      </button>
      <button
        v-if="canSyncUpstream"
        type="button"
        data-testid="sync-upstream-models"
        @click="syncUpstreamModels"
        :disabled="isSyncingUpstream"
        class="components-account-model-whitelist-selector__action-5"
      >
        {{ isSyncingUpstream ? t('admin.accounts.syncUpstreamModelsLoading') : t('admin.accounts.syncUpstreamModels') }}
      </button>
      <button
        type="button"
        @click="clearAll"
        class="components-account-model-whitelist-selector__action-6"
      >
        {{ t('admin.accounts.clearAllModels') }}
      </button>
    </div>

    <!-- Custom Model Input -->
    <div class="components-account-model-whitelist-selector__panel-11">
      <label class="components-account-model-whitelist-selector__label">{{ t('admin.accounts.customModelName') }}</label>
      <div class="components-account-model-whitelist-selector__panel-12">
        <input
          v-model="customModel"
          type="text"
          class="components-account-model-whitelist-selector__field-2 input"
          :placeholder="t('admin.accounts.enterCustomModelName')"
          @keydown.enter.prevent="handleEnter"
          @compositionstart="isComposing = true"
          @compositionend="isComposing = false"
        />
        <button
          type="button"
          @click="addCustom"
          class="components-account-model-whitelist-selector__action-7"
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

const props = defineProps<{
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
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
  'upstream-synced': []
}>()

const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const showDropdown = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, showDropdown, {
  maxWidth: 480,
  align: 'start',
  minComfortableHeight: 220
})
const searchQuery = ref('')
const customModel = ref('')
const isComposing = ref(false)
const isSyncingUpstream = ref(false)
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

const fillRelated = () => {
  const newModels = [...props.modelValue]
  for (const platform of normalizedPlatforms.value) {
    for (const model of getModelsByPlatform(platform)) {
      if (!newModels.includes(model)) {
        newModels.push(model)
      }
    }
  }
  emit('update:modelValue', newModels)
}

const syncUpstreamModels = async () => {
  if (isSyncingUpstream.value) return
  if (!props.accountId && !props.syncCredentials) return

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
    const message = extractApiErrorMessage(error, t('admin.accounts.syncUpstreamModelsFailed'))
    appStore.showError(t('admin.accounts.syncUpstreamModelsError', { message }))
  } finally {
    isSyncingUpstream.value = false
  }
}

watch(() => props.accountId, () => {
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
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})

</script>

<style scoped>
.components-account-model-whitelist-selector__panel-5 {
  background-color: var(--glass-layer-floating-bg);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  box-shadow:
    0 18px 42px rgba(15, 23, 42, 0.16),
    0 4px 12px rgba(15, 23, 42, 0.10),
    0 1px 0 var(--glass-highlight) inset;
  color: var(--color-text-primary);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
}

:global(.dark) .components-account-model-whitelist-selector__panel-5 {
  box-shadow:
    0 18px 42px rgba(0, 0, 0, 0.34),
    0 4px 12px rgba(0, 0, 0, 0.24),
    0 1px 0 var(--glass-highlight) inset;
}
</style>
