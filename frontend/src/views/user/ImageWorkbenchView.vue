<template>
  <AppLayout>
    <div class="image-workbench">
      <div class="toolbar">
        <button
          class="pill-button"
          type="button"
          :class="{ 'pill-button--active': favoritesOnly }"
          :title="favoriteButtonTitle"
          :aria-pressed="favoritesOnly"
          @click="handleFavoriteClick"
        >
          <Icon :name="activeFavoriteCollectionId ? 'chevronLeft' : 'star'" size="sm" />
        </button>
        <button
          v-if="inCollectionOverview"
          class="pill-button"
          type="button"
          :title="t('imageWorkbench.manageCollections')"
          @click="manageCollectionsOpen = true"
        >
          <Icon name="folder" size="sm" />
        </button>
        <label v-if="!inCollectionOverview" class="pill-select">
          <span class="sr-only">{{ t('imageWorkbench.all') }}</span>
          <select v-model="modelFilter">
            <option value="">{{ t('imageWorkbench.all') }}</option>
            <option v-for="model in historyModels" :key="model" :value="model">{{ model }}</option>
          </select>
          <Icon name="chevronDown" size="xs" />
        </label>
        <label class="search-field">
          <Icon name="search" size="sm" />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="inCollectionOverview ? t('imageWorkbench.searchCollections') : t('imageWorkbench.searchPlaceholder')"
          />
        </label>
      </div>

      <section class="gallery" aria-live="polite">
        <div v-if="loadingHistory" class="empty-state">{{ t('imageWorkbench.loading') }}</div>
        <div v-else-if="inCollectionOverview && collectionCards.length === 0" class="empty-state">
          <Icon name="folder" size="xl" />
          <strong>{{ emptyMessage }}</strong>
        </div>
        <div v-else-if="inCollectionOverview" class="gallery-grid">
          <article
            v-for="card in collectionCards"
            :key="card.id"
            class="history-card collection-card"
            @click="openCollection(card.id)"
          >
            <div class="history-card__thumb">
              <img v-if="card.cover" :src="card.cover.src" :alt="card.name" />
              <div v-else class="history-card__thumb--pending">
                <Icon name="folder" size="md" />
              </div>
            </div>
            <div class="history-card__body">
              <h3>{{ card.name }}</h3>
              <p>{{ t('imageWorkbench.collectionCount', { count: card.items.length }) }}</p>
              <div v-if="card.collection" class="history-card__actions">
                <button type="button" :class="{ active: card.id === defaultFavoriteCollectionId }" :title="t('imageWorkbench.setDefaultCollection')" @click.stop="setDefaultCollection(card.id)">
                  <Icon name="star" size="xs" />
                </button>
                <button type="button" :title="t('common.delete')" :disabled="favoriteCollections.length <= 1" @click.stop="removeCollection(card.id)">
                  <Icon name="trash" size="xs" />
                </button>
              </div>
            </div>
          </article>
        </div>
        <div v-else-if="filteredHistory.length === 0 && taskItems.length === 0 && !generating" class="empty-state">
          <Icon name="photo" size="xl" />
          <strong>{{ emptyMessage }}</strong>
        </div>
        <div v-else class="gallery-grid">
          <article v-for="task in taskItems" :key="task.taskId" class="history-card history-card--pending">
            <div class="history-card__thumb history-card__thumb--pending">
              <Icon :name="task.status === 'failed' ? 'exclamationCircle' : task.status === 'canceled' ? 'x' : 'sparkles'" size="md" />
            </div>
            <div class="history-card__body">
              <h3>{{ task.prompt }}</h3>
              <p>{{ task.status === 'queued' ? t('imageWorkbench.queued') : task.status === 'processing' ? t('imageWorkbench.generating') : task.status === 'failed' ? (task.error?.message || t('imageWorkbench.generateFailed')) : t('imageWorkbench.canceled') }}</p>
              <div class="history-card__actions">
                <button v-if="task.status === 'queued' || task.status === 'processing'" type="button" :title="t('imageWorkbench.cancelTask')" @click.stop="cancelTask(task)"><Icon name="x" size="xs" /></button>
                <button v-if="task.status === 'failed' || task.status === 'canceled' || task.status === 'cancelled'" type="button" :title="t('imageWorkbench.retryTask')" @click.stop="retryTask(task)"><Icon name="refresh" size="xs" /></button>
                <button v-if="task.status === 'failed' || task.status === 'canceled' || task.status === 'cancelled'" class="action-btn-del" type="button" :title="t('imageWorkbench.deleteTask')" @click.stop="removeTask(task)"><Icon name="trash" size="xs" /></button>
              </div>
            </div>
          </article>
          <article
            v-for="item in filteredHistory"
            :key="item.id"
            class="history-card"
            @click="reuseItem(item)"
          >
            <button class="history-card__thumb" type="button" :title="t('imageWorkbench.preview')" @click.stop="previewItem = item">
              <img :src="item.src" :alt="item.prompt" loading="lazy" />
              <div v-if="aspectLabel(item) || resolutionLabel(item)" class="thumb-badge">
                <span v-if="aspectLabel(item)" class="thumb-badge__ratio">{{ aspectLabel(item) }}</span>
                <span v-if="aspectLabel(item) && resolutionLabel(item)" class="thumb-badge__divider">·</span>
                <span v-if="resolutionLabel(item)" class="thumb-badge__res">{{ resolutionLabel(item) }}</span>
              </div>
            </button>
            <div class="history-card__body">
              <h3>{{ item.prompt }}</h3>
              <div class="history-card__tags">
                <span :title="item.keyName"><Icon name="key" size="xs" />{{ item.keyName || t('imageWorkbench.apiKey') }}</span>
                <span :title="item.model"><Icon name="sparkles" size="xs" />{{ item.model }}</span>
                <span>{{ t('imageWorkbench.quality') }} {{ qualityLabel(item) }}</span>
              </div>
              <div class="history-card__actions">
                <button type="button" :class="{ active: isItemFavorited(item) }" :title="t('imageWorkbench.saveToCollection')" @click.stop="openFavoritePicker(item)">
                  <Icon name="star" size="xs" />
                </button>
                <button type="button" :title="t('imageWorkbench.download')" @click.stop="downloadItem(item)">
                  <Icon name="share" size="xs" />
                </button>
                <button type="button" :title="t('imageWorkbench.copyPrompt')" @click.stop="copyPrompt(item)">
                  <Icon name="copy" size="xs" />
                </button>
                <button class="action-btn-del" type="button" :title="t('imageWorkbench.delete')" @click.stop="removeItem(item)">
                  <Icon name="trash" size="xs" />
                </button>
              </div>
            </div>
          </article>
        </div>
      </section>

      <form class="composer" @submit.prevent="submitGeneration">
        <div v-if="referenceFile || errorMessage" class="composer__context">
          <button v-if="referenceFile" class="reference-chip" type="button" :title="t('imageWorkbench.clearReference')" @click="clearReference">
            <Icon name="photo" size="xs" />
            <span class="reference-chip__name">{{ referenceFile.name }}</span>
            <Icon name="x" size="xs" />
          </button>
          <span v-if="errorMessage" class="error-message">
            <Icon name="exclamationCircle" size="xs" />
            <span>{{ errorMessage }}</span>
          </span>
        </div>

        <div class="composer__prompt">
          <textarea
            ref="promptInput"
            v-model="prompt"
            rows="1"
            :placeholder="t('imageWorkbench.promptPlaceholder')"
            @input="autosizePrompt"
            @keydown="onPromptKeydown"
          />
          <button v-if="prompt" class="composer__clear" type="button" :title="t('imageWorkbench.clearPrompt')" @click="prompt = ''">
            <Icon name="x" size="xs" />
          </button>
        </div>

        <Transition name="advanced-slide">
          <div v-if="showAdvanced && hasAdvancedCapabilities" class="composer__advanced">
            <label v-if="adapter.capabilities.outputFormat" class="toolbar-pill toolbar-pill--select" :title="t('imageWorkbench.format')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.format') }}</span>
              <select v-model="params.output_format">
                <option value="png">PNG</option>
                <option value="jpeg">JPEG</option>
                <option value="webp">WebP</option>
              </select>
              <Icon name="chevronDown" size="xs" />
            </label>
            <label v-if="adapter.capabilities.outputFormat" class="toolbar-pill toolbar-pill--input" :title="t('imageWorkbench.compression')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.compression') }}</span>
              <input v-model.number="params.output_compression" type="number" min="0" max="100" step="1" />
              <span class="toolbar-pill__suffix">%</span>
            </label>
            <label v-if="adapter.capabilities.transparency" class="toolbar-pill toolbar-pill--select" :title="t('imageWorkbench.transparent')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.transparent') }}</span>
              <select v-model="params.background">
                <option value="auto">auto</option>
                <option value="transparent">transparent</option>
                <option value="opaque">opaque</option>
              </select>
              <Icon name="chevronDown" size="xs" />
            </label>
            <label v-if="adapter.capabilities.quality" class="toolbar-pill toolbar-pill--select" :title="t('imageWorkbench.moderation')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.moderation') }}</span>
              <select v-model="params.moderation">
                <option value="auto">auto</option>
                <option value="low">low</option>
              </select>
              <Icon name="chevronDown" size="xs" />
            </label>
          </div>
        </Transition>

        <div class="composer__toolbar">
          <div class="composer__toolbar-left">
            <button
              class="toolbar-pill toolbar-pill--model"
              type="button"
              :title="t('imageWorkbench.settings')"
              @click="openSettings"
            >
              <Icon name="cog" size="xs" />
              <span class="toolbar-pill__label">{{ activeModelBadge }}</span>
            </button>
            <button
              v-if="adapter.capabilities.size"
              class="toolbar-pill toolbar-pill--button"
              type="button"
              :title="t('imageWorkbench.sizePickerTitle')"
              @click="openSizePicker('params')"
            >
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.size') }}</span>
              <span class="toolbar-pill__val">{{ params.size || 'auto' }}</span>
              <Icon name="chevronDown" size="xs" />
            </button>
            <label v-if="adapter.capabilities.quality" class="toolbar-pill toolbar-pill--select" :title="t('imageWorkbench.quality')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.quality') }}</span>
              <select v-model="params.quality">
                <option value="auto">auto</option>
                <option value="low">low</option>
                <option value="medium">medium</option>
                <option value="high">high</option>
              </select>
              <Icon name="chevronDown" size="xs" />
            </label>
            <label class="toolbar-pill toolbar-pill--number" :title="t('imageWorkbench.quantity')">
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.quantity') }}</span>
              <input v-model.number="params.n" type="number" min="1" max="4" step="1" />
            </label>
            <button
              v-if="hasAdvancedCapabilities"
              class="toolbar-pill toolbar-pill--toggle"
              :class="{ 'toolbar-pill--active': showAdvanced, 'toolbar-pill--has-custom': hasCustomAdvanced }"
              type="button"
              :title="t('imageWorkbench.advancedParams')"
              @click="showAdvanced = !showAdvanced"
            >
              <span class="toolbar-pill__prefix">{{ t('imageWorkbench.advancedParams') }}</span>
              <span v-if="hasCustomAdvanced" class="toolbar-pill__dot"></span>
              <Icon :name="showAdvanced ? 'chevronUp' : 'chevronDown'" size="xs" />
            </button>
          </div>

          <div class="composer__toolbar-right">
            <label
              class="action-btn"
              :class="{ 'action-btn--filled': referenceFile }"
              :title="referenceFile?.name || t('imageWorkbench.uploadReference')"
            >
              <Icon name="paperclip" size="sm" />
              <input type="file" accept="image/png,image/jpeg,image/webp" hidden @change="onReferenceChange" />
            </label>
            <button
              class="action-btn action-btn--send"
              type="submit"
              :disabled="generating || !prompt.trim()"
              :title="generating ? t('imageWorkbench.generating') : `${t('imageWorkbench.generate')} (Ctrl+Enter)`"
            >
              <Icon v-if="!generating" name="arrowRight" size="sm" />
              <span v-else class="btn-spinner" aria-hidden="true"></span>
            </button>
          </div>
        </div>
      </form>
    </div>

    <div v-if="previewItem" class="lightbox" @click.self="previewItem = null">
      <button class="lightbox__close" type="button" @click="previewItem = null">
        <Icon name="x" size="sm" />
      </button>
      <img :src="previewItem.src" :alt="previewItem.prompt" />
      <p>{{ previewItem.prompt }}</p>
    </div>

    <div v-if="settingsOpen" class="settings-overlay" @click.self="settingsOpen = false">
      <section class="settings-panel" role="dialog" aria-modal="true">
        <div class="settings-panel__header">
          <h2>{{ t('imageWorkbench.settings') }}</h2>
          <button class="icon-button" type="button" @click="settingsOpen = false">
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="settings-tabs">
          <button type="button" :class="{ active: settingsTab === 'connection' }" @click="settingsTab = 'connection'">{{ t('imageWorkbench.connection') }}</button>
          <button type="button" :class="{ active: settingsTab === 'preferences' }" @click="settingsTab = 'preferences'">{{ t('imageWorkbench.preferences') }}</button>
          <button type="button" :class="{ active: settingsTab === 'data' }" @click="settingsTab = 'data'">{{ t('imageWorkbench.dataManagement') }}</button>
        </div>
        <div v-if="settingsTab === 'connection'" class="settings-form">
          <p class="settings-note">{{ t('imageWorkbench.platformHint') }}</p>
          <div class="settings-step">
            <span>{{ t('imageWorkbench.platform') }}</span>
            <div class="platform-picks">
              <button
                v-for="item in enabledAdapters"
                :key="item.id"
                type="button"
                :class="{ active: activePlatform === item.id }"
                @click="activePlatform = item.id"
              >
                {{ item.label }}
              </button>
            </div>
          </div>
          <label>
            <span>{{ t('imageWorkbench.apiKey') }}</span>
            <select v-model="selectedKeyId" :disabled="eligibleKeys.length === 0">
              <option value="">{{ eligibleKeys.length ? t('imageWorkbench.apiKey') : t('imageWorkbench.noKeys') }}</option>
              <option v-for="key in eligibleKeys" :key="key.id" :value="key.id">{{ key.name || key.id }}</option>
            </select>
          </label>
          <label>
            <span>{{ t('imageWorkbench.model') }}</span>
            <select v-model="selectedModel" :disabled="models.length === 0">
              <option value="">{{ models.length ? t('imageWorkbench.model') : t('imageWorkbench.noModels') }}</option>
              <option v-for="model in models" :key="model.id" :value="model.id">{{ model.name || model.id }}</option>
            </select>
          </label>
          <button class="primary-button" type="button" @click="saveConnection">{{ t('common.save') }}</button>
        </div>
        <div v-else-if="settingsTab === 'preferences'" class="settings-form">
          <div class="settings-step">
            <span>{{ t('imageWorkbench.size') }}</span>
            <button class="param-field param-field--button" type="button" @click="openSizePicker('preferences')">
              <span class="param-field__value">{{ preferences.size || 'auto' }}</span>
            </button>
          </div>
          <label>
            <span>{{ t('imageWorkbench.quality') }}</span>
            <select v-model="preferences.quality">
              <option value="auto">auto</option>
              <option value="low">low</option>
              <option value="medium">medium</option>
              <option value="high">high</option>
            </select>
          </label>
          <button class="primary-button" type="button" @click="savePreferences">{{ t('common.save') }}</button>
        </div>
        <div v-else class="settings-form">
          <p class="settings-note">{{ t('imageWorkbench.dataManagement') }}</p>
          <button class="secondary-button" type="button" @click="exportData">
            <Icon name="download" size="sm" />{{ t('imageWorkbench.exportData') }}
          </button>
          <label class="secondary-button">
            <Icon name="upload" size="sm" />{{ t('imageWorkbench.importData') }}
            <input type="file" accept="application/json" hidden @change="importData" />
          </label>
          <button class="danger-button" type="button" @click="clearAllHistory">{{ t('imageWorkbench.clearHistory') }}</button>
        </div>
      </section>
    </div>

    <SizePickerModal
      :show="sizePickerOpen"
      :current-size="sizePickerTarget === 'preferences' ? preferences.size : params.size"
      @select="onSizePicked"
      @close="sizePickerOpen = false"
    />
    <ManageCollectionsModal
      :show="manageCollectionsOpen"
      :collections="favoriteCollections"
      :default-favorite-collection-id="defaultFavoriteCollectionId"
      @close="manageCollectionsOpen = false"
      @create="createCollection"
      @rename="renameCollection"
      @remove="removeCollection"
      @set-default="setDefaultCollection"
    />
    <FavoritePickerModal
      :show="Boolean(favoritePickerItem)"
      :collections="favoriteCollections"
      :initial-ids="favoritePickerInitialIds"
      @close="favoritePickerItem = null"
      @create="createCollection"
      @confirm="saveFavoritePicker"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import {
  eligibleImageKeys,
  generateImage,
  submitImageTask,
  getImageTask,
  cancelImageTask,
  retryImageTask,
  deleteImageTask,
  isAsyncImageTaskUnavailable,
  imagePlatformAdapters,
  listImageModels,
  loadWorkbenchCredentials,
  type ImageGenerationParams,
  type ImageResult,
  type ImageTask,
  type ImageModel,
  type ImagePlatform,
  type ImagePlatformAdapter,
} from '@/api/imageWorkbench'
import { clearHistory, deleteHistory, exportHistory, importHistory, listHistory, putHistory, listTasks, putTask, deleteTask, type ImageHistoryItem, type ImageTaskItem } from '@/features/image-workbench/storage'
import {
  ALL_FAVORITES_COLLECTION_ID,
  createFavoriteCollection,
  ensureDefaultFavoriteCollection,
  isItemFavorite,
  itemCollectionIds,
  loadFavoriteState,
  mergeImportedFavoriteState,
  saveFavoriteState,
  type FavoriteCollection,
} from '@/features/image-workbench/favorites'
import SizePickerModal from '@/features/image-workbench/SizePickerModal.vue'
import ManageCollectionsModal from '@/features/image-workbench/ManageCollectionsModal.vue'
import FavoritePickerModal from '@/features/image-workbench/FavoritePickerModal.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const enabledAdapters = computed(() => imagePlatformAdapters.filter((adapter) => adapter.enabled))
const defaultPlatform = imagePlatformAdapters.find((item) => item.enabled)?.id || 'openai'
const activePlatform = ref<ImagePlatform>(defaultPlatform)
const credentials = ref<{ keys: import('@/types').ApiKey[]; groups: import('@/types').Group[] }>({ keys: [], groups: [] })
const loadingHistory = ref(true)
const taskItems = ref<ImageTaskItem[]>([])
const submitting = ref(false)
const generating = computed(() => submitting.value || taskItems.value.some((task) => task.status === 'queued' || task.status === 'processing'))
const pollTimers = new Map<string, number>()
const settingsOpen = ref(false)
const settingsTab = ref<'connection' | 'preferences' | 'data'>('connection')
const sizePickerOpen = ref(false)
const sizePickerTarget = ref<'params' | 'preferences'>('params')
const favoritesOnly = ref(false)
const activeFavoriteCollectionId = ref<string | null>(null)
const favoriteCollections = ref<FavoriteCollection[]>([])
const defaultFavoriteCollectionId = ref<string | null>(null)
const manageCollectionsOpen = ref(false)
const favoritePickerItem = ref<ImageHistoryItem | null>(null)
const modelFilter = ref('')
const searchQuery = ref('')
const prompt = ref('')
const promptInput = ref<HTMLTextAreaElement | null>(null)
const referenceFile = ref<File | null>(null)
const models = ref<ImageModel[]>([])
const history = ref<ImageHistoryItem[]>([])
const errorMessage = ref('')
const previewItem = ref<ImageHistoryItem | null>(null)
const selectedByPlatform = reactive<Record<ImagePlatform, { keyId: string; model: string }>>({
  openai: { keyId: '', model: '' },
  grok: { keyId: '', model: '' },
})
const preferences = reactive({ size: '1024x1024', quality: 'auto' })
const params = reactive({
  size: preferences.size,
  quality: preferences.quality,
  output_format: 'png',
  output_compression: 100,
  background: 'opaque',
  moderation: 'auto',
  n: 1,
})

const adapter = computed<ImagePlatformAdapter>(() => imagePlatformAdapters.find((item) => item.id === activePlatform.value) || imagePlatformAdapters[0])
const eligibleKeys = computed(() => eligibleImageKeys(credentials.value.keys, adapter.value))
const selectedKeyId = computed({
  get: () => selectedByPlatform[activePlatform.value].keyId,
  set: (value: string) => {
    selectedByPlatform[activePlatform.value].keyId = value
    void refreshModels()
  },
})
const selectedModel = computed({
  get: () => selectedByPlatform[activePlatform.value].model,
  set: (value: string) => {
    selectedByPlatform[activePlatform.value].model = value
  },
})
const selectedKey = computed(() => eligibleKeys.value.find((key) => key.id === selectedKeyId.value))
const showAdvanced = ref(false)
const hasAdvancedCapabilities = computed(() => {
  return Boolean(
    adapter.value.capabilities.outputFormat ||
    adapter.value.capabilities.transparency ||
    adapter.value.capabilities.quality
  )
})
const hasCustomAdvanced = computed(() => {
  return (
    params.output_format !== 'png' ||
    params.output_compression !== 100 ||
    params.background !== 'opaque' ||
    params.moderation !== 'auto'
  )
})
const activeModelBadge = computed(() => {
  if (selectedModel.value) return selectedModel.value
  return adapter.value.label || activePlatform.value
})
const inCollectionOverview = computed(() => favoritesOnly.value && !activeFavoriteCollectionId.value)
const favoriteButtonTitle = computed(() => {
  if (activeFavoriteCollectionId.value) return t('imageWorkbench.backToCollections')
  if (favoritesOnly.value) return t('imageWorkbench.exitFavorites')
  return t('imageWorkbench.favorites')
})
const historyModels = computed(() => [...new Set(history.value.map((item) => item.model).filter(Boolean))])
const favoriteItems = computed(() => history.value.filter((item) => isItemFavorite(item, defaultFavoriteCollectionId.value)))
const collectionCards = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const cards = [
    {
      id: ALL_FAVORITES_COLLECTION_ID,
      name: t('imageWorkbench.allFavorites'),
      collection: null as FavoriteCollection | null,
      items: favoriteItems.value,
      cover: favoriteItems.value[0] || null,
    },
    ...favoriteCollections.value.map((collection) => {
      const items = favoriteItems.value.filter((item) => itemCollectionIds(item, defaultFavoriteCollectionId.value).includes(collection.id))
      return { id: collection.id, name: collection.name, collection, items, cover: items[0] || null }
    }),
  ]
  if (!query) return cards
  return cards.filter((card) => card.name.toLowerCase().includes(query))
})
const favoritePickerInitialIds = computed(() => {
  if (!favoritePickerItem.value) return defaultFavoriteCollectionId.value ? [defaultFavoriteCollectionId.value] : []
  const ids = itemCollectionIds(favoritePickerItem.value, defaultFavoriteCollectionId.value)
  return ids.length ? ids : (defaultFavoriteCollectionId.value ? [defaultFavoriteCollectionId.value] : [])
})
const filteredHistory = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return history.value.filter((item) => {
    if (favoritesOnly.value) {
      if (!isItemFavorite(item, defaultFavoriteCollectionId.value)) return false
      if (
        activeFavoriteCollectionId.value
        && activeFavoriteCollectionId.value !== ALL_FAVORITES_COLLECTION_ID
        && !itemCollectionIds(item, defaultFavoriteCollectionId.value).includes(activeFavoriteCollectionId.value)
      ) return false
    }
    if (modelFilter.value && item.model !== modelFilter.value) return false
    if (!query) return true
    const haystack = [
      item.prompt,
      item.model,
      item.keyName,
      item.platform,
      JSON.stringify(item.params || {}),
      resolutionLabel(item),
    ].join(' ').toLowerCase()
    return haystack.includes(query)
  })
})
const emptyMessage = computed(() => {
  if (inCollectionOverview.value) {
    if (favoriteCollections.value.length === 0) return t('imageWorkbench.emptyFavorites')
    return t('imageWorkbench.searchEmpty')
  }
  if (history.value.length === 0) return t('imageWorkbench.emptyHistory')
  if (favoritesOnly.value) return t('imageWorkbench.emptyFavorites')
  return t('imageWorkbench.searchEmpty')
})

function gcd(a: number, b: number): number {
  return b === 0 ? a : gcd(b, a % b)
}

function parseSize(value: unknown): { width: number; height: number } | null {
  const match = String(value || '').match(/^(\d+)\s*[x×]\s*(\d+)$/i)
  if (!match) return null
  const width = Number(match[1])
  const height = Number(match[2])
  if (!width || !height) return null
  return { width, height }
}

function itemSize(item: ImageHistoryItem): { width: number; height: number } | null {
  if (item.width && item.height) return { width: item.width, height: item.height }
  return parseSize(item.params?.size)
}

const COMMON_ASPECT_RATIOS = [
  { ratio: 1, label: '1:1' },
  { ratio: 16 / 9, label: '16:9' },
  { ratio: 9 / 16, label: '9:16' },
  { ratio: 4 / 3, label: '4:3' },
  { ratio: 3 / 4, label: '3:4' },
  { ratio: 3 / 2, label: '3:2' },
  { ratio: 2 / 3, label: '2:3' },
  { ratio: 21 / 9, label: '21:9' },
  { ratio: 5 / 4, label: '5:4' },
  { ratio: 4 / 5, label: '4:5' },
  { ratio: 6 / 5, label: '6:5' },
  { ratio: 5 / 6, label: '5:6' },
]

function aspectLabel(item: ImageHistoryItem): string {
  const size = itemSize(item)
  if (!size || !size.width || !size.height) return ''
  const r = size.width / size.height
  for (const entry of COMMON_ASPECT_RATIOS) {
    if (Math.abs(r - entry.ratio) / entry.ratio < 0.035) {
      return entry.label
    }
  }
  const divisor = gcd(size.width, size.height)
  const sw = size.width / divisor
  const sh = size.height / divisor
  if (sw <= 20 && sh <= 20) {
    return `${sw}:${sh}`
  }
  return `${r.toFixed(2)}:1`
}

function resolutionLabel(item: ImageHistoryItem): string {
  const size = itemSize(item)
  if (size) return `${size.width}x${size.height}`
  const raw = String(item.params?.size || '')
  return raw && raw !== 'auto' ? raw : ''
}

function qualityLabel(item: ImageHistoryItem): string {
  return String(item.params?.quality || 'auto')
}

function imageSource(result: { url?: string; b64_json?: string }): string {
  return result.url || (result.b64_json ? `data:image/png;base64,${result.b64_json}` : '')
}

function measureImage(src: string): Promise<{ width: number; height: number } | null> {
  return new Promise((resolve) => {
    const image = new Image()
    image.onload = () => resolve({ width: image.naturalWidth, height: image.naturalHeight })
    image.onerror = () => resolve(null)
    image.src = src
  })
}

async function saveImageResults(
  results: ImageResult[],
  metadata: {
    prompt: string
    model: string
    platform: string
    keyName?: string
    params: Record<string, unknown>
  },
) {
  for (const result of results) {
    const src = imageSource(result)
    if (!src) continue
    const measured = await measureImage(src)
    await putHistory({
      id: crypto.randomUUID(),
      createdAt: Date.now(),
      prompt: metadata.prompt,
      model: metadata.model,
      platform: metadata.platform,
      src,
      revisedPrompt: result.revised_prompt,
      favorite: false,
      keyName: metadata.keyName,
      width: measured?.width,
      height: measured?.height,
      params: { ...metadata.params },
    })
  }
  history.value = await listHistory()
}

async function refreshModels() {
  const key = selectedKey.value?.key
  if (!key) {
    models.value = []
    selectedByPlatform[activePlatform.value].model = ''
    return
  }
  try {
    models.value = await listImageModels(key, adapter.value)
    const current = selectedByPlatform[activePlatform.value].model
    if (!models.value.some((model) => model.id === current)) selectedByPlatform[activePlatform.value].model = models.value[0]?.id || ''
  } catch {
    models.value = []
    errorMessage.value = t('imageWorkbench.noModels')
  }
}

function onReferenceChange(event: Event) {
  referenceFile.value = (event.target as HTMLInputElement).files?.[0] || null
}

function autosizePrompt(event?: Event) {
  const el = (event?.target as HTMLTextAreaElement | undefined) || promptInput.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${Math.min(el.scrollHeight, 112)}px`
}

function onPromptKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.isComposing || event.keyCode === 229) return
  if (!event.ctrlKey && !event.metaKey) return
  event.preventDefault()
  void submitGeneration()
}

function openSizePicker(target: 'params' | 'preferences') {
  sizePickerTarget.value = target
  sizePickerOpen.value = true
}

function onSizePicked(size: string) {
  if (sizePickerTarget.value === 'preferences') preferences.size = size
  else params.size = size
}

function clearReference() {
  referenceFile.value = null
}

function resetComposerAfterSubmit() {
  prompt.value = ''
  clearReference()
  void nextTick(() => autosizePrompt())
}

function reuseItem(item: ImageHistoryItem) {
  prompt.value = item.prompt
  if (typeof item.params?.size === 'string') params.size = item.params.size
  if (typeof item.params?.quality === 'string') params.quality = item.params.quality
  if (typeof item.params?.output_format === 'string') params.output_format = item.params.output_format
  if (typeof item.params?.output_compression === 'number') params.output_compression = item.params.output_compression
  if (typeof item.params?.background === 'string') params.background = item.params.background
  if (typeof item.params?.moderation === 'string') params.moderation = item.params.moderation
  if (typeof item.params?.n === 'number') params.n = item.params.n
  if (item.model) selectedByPlatform[activePlatform.value].model = item.model
  void nextTick(() => autosizePrompt())
}

async function submitGeneration() {
  if (submitting.value) return
  const key = selectedKey.value
  const model = selectedModel.value
  const promptText = prompt.value.trim()
  if (!key?.key || !model) {
    errorMessage.value = t('imageWorkbench.configureApi')
    openSettings()
    return
  }
  if (!promptText) {
    errorMessage.value = t('imageWorkbench.promptRequired')
    return
  }

  const compression = Number(params.output_compression)
  const generationParams: ImageGenerationParams = {
    prompt: promptText,
    model,
    size: params.size,
    quality: params.quality,
    output_format: params.output_format,
    output_compression: Number.isFinite(compression) ? Math.min(100, Math.max(0, compression)) : 100,
    background: params.background,
    moderation: params.moderation,
    n: Math.min(4, Math.max(1, Number(params.n) || 1)),
  }
  const paramsSnapshot: Record<string, unknown> = {
    ...params,
    output_compression: generationParams.output_compression,
    n: generationParams.n,
  }
  const reference = referenceFile.value || undefined
  const platform = activePlatform.value
  errorMessage.value = ''
  submitting.value = true
  try {
    let accepted: ImageTask
    try {
      accepted = await submitImageTask(key.key, generationParams, reference)
    } catch (error) {
      if (!isAsyncImageTaskUnavailable(error)) throw error

      // Async tasks require Redis-backed task state and object storage. A
      // deployment that does not provide either can still use the original
      // synchronous image endpoint, which returns the result directly.
      const results = await generateImage(key.key, generationParams, reference)
      await saveImageResults(results, {
        prompt: promptText,
        model,
        platform,
        keyName: key.name,
        params: paramsSnapshot,
      })
      resetComposerAfterSubmit()
      return
    }

    resetComposerAfterSubmit()

    if (accepted.status === 'completed' && accepted.result?.data?.length) {
      await saveImageResults(accepted.result.data, {
        prompt: promptText,
        model,
        platform,
        keyName: key.name,
        params: paramsSnapshot,
      })
      return
    }

    const taskId = accepted.task_id || accepted.id
    if (!taskId) throw new Error(t('imageWorkbench.generateFailed'))
    const task: ImageTaskItem = {
      taskId,
      keyId: key.id,
      keyName: key.name,
      platform,
      prompt: promptText,
      model,
      params: paramsSnapshot,
      status: accepted.status || 'processing',
      createdAt: Date.now(),
    }
    taskItems.value = [task, ...taskItems.value.filter((item) => item.taskId !== task.taskId)]
    await putTask(task)
    startTaskPolling(task)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('imageWorkbench.generateFailed')
  } finally {
    submitting.value = false
  }
}

function startTaskPolling(task: ImageTaskItem) {
  if (pollTimers.has(task.taskId)) return
  const poll = async () => {
    pollTimers.delete(task.taskId)
    const key = credentials.value.keys.find((item) => item.id === task.keyId)?.key
    if (!key) { task.status = 'failed'; task.error = { message: t('imageWorkbench.keyUnavailable') }; await putTask(task); return }
    try {
      const remote = await getImageTask(key, task.taskId)
      task.status = remote.status; task.completedAt = remote.completed_at
      if (remote.error) task.error = remote.error
      if (remote.status === 'completed') {
        await saveImageResults(remote.result?.data || [], {
          prompt: task.prompt,
          model: task.model,
          platform: task.platform,
          keyName: task.keyName,
          params: task.params,
        })
        await deleteTask(task.taskId)
        taskItems.value = taskItems.value.filter((item) => item.taskId !== task.taskId)
        pollTimers.delete(task.taskId)
        return
      }
      await putTask(task); taskItems.value = [...taskItems.value]
      if (['failed', 'canceled', 'cancelled'].includes(task.status)) return
    } catch (error) { task.error = { message: error instanceof Error ? error.message : t('imageWorkbench.generateFailed') }; await putTask(task) }
    if (['failed', 'canceled', 'cancelled'].includes(task.status)) return
    const timer = window.setTimeout(poll, 3000); pollTimers.set(task.taskId, timer)
  }
  void poll()
}

async function cancelTask(task: ImageTaskItem) {
  const key = credentials.value.keys.find((item) => item.id === task.keyId)?.key
  if (!key) return
  stopTaskPolling(task.taskId)
  try {
    const remote = await cancelImageTask(key, task.taskId)
    if (!remote) {
      await deleteTask(task.taskId)
      taskItems.value = taskItems.value.filter((item) => item.taskId !== task.taskId)
      return
    }
    task.status = remote.status
    task.error = remote.error
    await putTask(task)
    taskItems.value = [...taskItems.value]
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('imageWorkbench.generateFailed')
    if (task.status === 'queued' || task.status === 'processing') startTaskPolling(task)
  }
}

function stopTaskPolling(taskId: string) {
  const timer = pollTimers.get(taskId)
  if (timer !== undefined) window.clearTimeout(timer)
  pollTimers.delete(taskId)
}

async function retryTask(task: ImageTaskItem) {
  const key = credentials.value.keys.find((item) => item.id === task.keyId)?.key
  if (!key || submitting.value) return
  stopTaskPolling(task.taskId)
  submitting.value = true
  try {
    // The backend reuses this task's retained request artifact/body. Sending
    // the current composer values here could turn an edit retry into a new
    // request when the original multipart image is no longer selected.
    const remote = await retryImageTask(key, task.taskId)
    task.status = remote.status
    task.error = remote.error
    task.completedAt = remote.completed_at
    await putTask(task)
    taskItems.value = [...taskItems.value]
    startTaskPolling(task)
  } catch (error) {
    task.error = { message: error instanceof Error ? error.message : t('imageWorkbench.generateFailed') }
    task.status = 'failed'
    await putTask(task)
    taskItems.value = [...taskItems.value]
  } finally {
    submitting.value = false
  }
}

async function removeTask(task: ImageTaskItem) {
  const key = credentials.value.keys.find((item) => item.id === task.keyId)?.key
  stopTaskPolling(task.taskId)
  if (!key) {
    await deleteTask(task.taskId)
    taskItems.value = taskItems.value.filter((item) => item.taskId !== task.taskId)
    return
  }
  try {
    await deleteImageTask(key, task.taskId)
    await deleteTask(task.taskId)
    taskItems.value = taskItems.value.filter((item) => item.taskId !== task.taskId)
  } catch (error) {
    if (error && typeof error === 'object' && 'status' in error && error.status === 404) {
      await deleteTask(task.taskId)
      taskItems.value = taskItems.value.filter((item) => item.taskId !== task.taskId)
      return
    }
    errorMessage.value = error instanceof Error ? error.message : t('imageWorkbench.generateFailed')
    taskItems.value = [...taskItems.value]
  }
}

function persistFavorites() {
  saveFavoriteState({
    collections: favoriteCollections.value,
    defaultFavoriteCollectionId: defaultFavoriteCollectionId.value,
  })
}

function handleFavoriteClick() {
  if (activeFavoriteCollectionId.value) {
    activeFavoriteCollectionId.value = null
    return
  }
  favoritesOnly.value = !favoritesOnly.value
}

function openCollection(id: string) {
  activeFavoriteCollectionId.value = id
}

function isItemFavorited(item: ImageHistoryItem) {
  return isItemFavorite(item, defaultFavoriteCollectionId.value)
}

function createCollection(name: string) {
  const created = createFavoriteCollection(name, favoriteCollections.value)
  if (!created) {
    appStore.showError(t('imageWorkbench.collectionExists'))
    return
  }
  favoriteCollections.value = [...favoriteCollections.value, created]
  persistFavorites()
}

function renameCollection(id: string, name: string) {
  favoriteCollections.value = favoriteCollections.value.map((collection) => (
    collection.id === id ? { ...collection, name, updatedAt: Date.now() } : collection
  ))
  persistFavorites()
}

function setDefaultCollection(id: string) {
  defaultFavoriteCollectionId.value = defaultFavoriteCollectionId.value === id ? favoriteCollections.value[0]?.id || id : id
  persistFavorites()
}

async function removeCollection(id: string) {
  if (favoriteCollections.value.length <= 1) return
  const collection = favoriteCollections.value.find((item) => item.id === id)
  if (!collection) return
  if (!window.confirm(t('imageWorkbench.confirmDeleteCollection', { name: collection.name }))) return
  favoriteCollections.value = favoriteCollections.value.filter((item) => item.id !== id)
  if (defaultFavoriteCollectionId.value === id) defaultFavoriteCollectionId.value = favoriteCollections.value[0]?.id || null
  if (activeFavoriteCollectionId.value === id) activeFavoriteCollectionId.value = null
  const remainingIds = new Set(favoriteCollections.value.map((item) => item.id))
  for (const item of history.value) {
    const ids = itemCollectionIds(item, defaultFavoriteCollectionId.value).filter((collectionId) => remainingIds.has(collectionId))
    if (ids.join() === (item.favoriteCollectionIds || []).join() && item.favorite === ids.length > 0) continue
    await putHistory({ ...item, favorite: ids.length > 0, favoriteCollectionIds: ids })
  }
  persistFavorites()
  history.value = await listHistory()
}

function openFavoritePicker(item: ImageHistoryItem) {
  favoritePickerItem.value = item
}

async function saveFavoritePicker(ids: string[]) {
  const item = favoritePickerItem.value
  if (!item) return
  await putHistory({ ...item, favorite: ids.length > 0, favoriteCollectionIds: ids })
  favoritePickerItem.value = null
  history.value = await listHistory()
}

async function removeItem(item: ImageHistoryItem) {
  if (!window.confirm(t('imageWorkbench.confirmDelete'))) return
  await deleteHistory(item.id)
  if (previewItem.value?.id === item.id) previewItem.value = null
  history.value = await listHistory()
}

async function copyPrompt(item: ImageHistoryItem) {
  prompt.value = item.prompt
  await copyToClipboard(item.prompt)
}

async function downloadItem(item: ImageHistoryItem) {
  const extension = String(item.params?.output_format || 'png')
  const filename = `image-${item.id}.${extension}`
  try {
    if (item.src.startsWith('data:')) {
      const link = document.createElement('a')
      link.href = item.src
      link.download = filename
      link.click()
      return
    }
    const response = await fetch(item.src)
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  } catch {
    appStore.showError(t('imageWorkbench.generateFailed'))
  }
}

async function exportData() {
  const blob = new Blob([await exportHistory({
    collections: favoriteCollections.value,
    defaultFavoriteCollectionId: defaultFavoriteCollectionId.value,
  })], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `image-workbench-history-${new Date().toISOString().slice(0, 10)}.json`
  link.click()
  URL.revokeObjectURL(url)
}

async function importData(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  try {
    const imported = await importHistory(await file.text())
    if (imported.collections) {
      const merged = mergeImportedFavoriteState(
        { collections: favoriteCollections.value, defaultFavoriteCollectionId: defaultFavoriteCollectionId.value },
        imported.collections,
        imported.defaultFavoriteCollectionId,
      )
      favoriteCollections.value = merged.collections
      defaultFavoriteCollectionId.value = merged.defaultFavoriteCollectionId
      persistFavorites()
    }
    history.value = await listHistory()
    appStore.showSuccess(t('imageWorkbench.dataImported'))
  } catch {
    errorMessage.value = t('imageWorkbench.invalidImport')
  }
}

async function clearAllHistory() {
  if (!window.confirm(t('imageWorkbench.confirmClear'))) return
  await clearHistory()
  history.value = []
}

function persistConnection() {
  localStorage.setItem('image-workbench-preferences', JSON.stringify({
    size: preferences.size,
    quality: preferences.quality,
    platform: activePlatform.value,
    selectedByPlatform,
  }))
}

function openSettings() {
  settingsTab.value = 'connection'
  settingsOpen.value = true
}

function saveConnection() {
  persistConnection()
  settingsOpen.value = false
  appStore.showSuccess(t('imageWorkbench.settingsSaved'))
}

function savePreferences() {
  params.size = preferences.size
  params.quality = preferences.quality
  persistConnection()
  settingsOpen.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'Escape') return
  if (sizePickerOpen.value) {
    sizePickerOpen.value = false
    return
  }
  if (favoritePickerItem.value) {
    favoritePickerItem.value = null
    return
  }
  if (manageCollectionsOpen.value) {
    manageCollectionsOpen.value = false
    return
  }
  previewItem.value = null
  settingsOpen.value = false
}

watch(eligibleKeys, (keys) => {
  if (!keys.some((key) => key.id === selectedKeyId.value)) selectedByPlatform[activePlatform.value].keyId = keys[0]?.id || ''
  void refreshModels()
}, { immediate: true })
watch(activePlatform, () => {
  errorMessage.value = ''
  persistConnection()
  void refreshModels()
})
watch(settingsOpen, (open) => {
  if (!open) persistConnection()
})
watch(() => params.background, (value) => {
  if (value === 'transparent' && params.output_format === 'jpeg') params.output_format = 'png'
})

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  try {
    const saved = JSON.parse(localStorage.getItem('image-workbench-preferences') || '{}') as {
      size?: string
      quality?: string
      platform?: ImagePlatform
      selectedByPlatform?: Record<ImagePlatform, { keyId: string; model: string }>
    }
    if (typeof saved.size === 'string') preferences.size = saved.size
    if (typeof saved.quality === 'string') preferences.quality = saved.quality
    if ((saved.platform === 'openai' || saved.platform === 'grok') && imagePlatformAdapters.some((item) => item.id === saved.platform && item.enabled)) {
      activePlatform.value = saved.platform
    }
    if (saved.selectedByPlatform) {
      if (saved.selectedByPlatform.openai) Object.assign(selectedByPlatform.openai, saved.selectedByPlatform.openai)
      if (saved.selectedByPlatform.grok) Object.assign(selectedByPlatform.grok, saved.selectedByPlatform.grok)
    }
    params.size = preferences.size
    params.quality = preferences.quality
  } catch {
    /* ignore malformed local preferences */
  }
  const favoriteState = loadFavoriteState()
  favoriteCollections.value = ensureDefaultFavoriteCollection(favoriteState.collections)
  defaultFavoriteCollectionId.value = favoriteState.defaultFavoriteCollectionId
  try {
    credentials.value = await loadWorkbenchCredentials()
  } catch {
    errorMessage.value = t('imageWorkbench.loadFailed')
  }
  try {
    history.value = await listHistory()
    let migrated = false
    for (const item of history.value) {
      const ids = itemCollectionIds(item, defaultFavoriteCollectionId.value)
      const favorite = ids.length > 0
      if (item.favorite === favorite && (item.favoriteCollectionIds || []).join() === ids.join()) continue
      await putHistory({ ...item, favorite, favoriteCollectionIds: ids })
      migrated = true
    }
    if (migrated) history.value = await listHistory()
    taskItems.value = await listTasks()
    for (const task of taskItems.value) {
      if (task.status === 'queued' || task.status === 'processing') startTaskPolling(task)
    }
  } finally {
    loadingHistory.value = false
  }
  persistFavorites()
  if (!appStore.publicSettingsLoaded) await appStore.fetchPublicSettings().catch(() => undefined)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  for (const timer of pollTimers.values()) window.clearTimeout(timer)
  pollTimers.clear()
})
</script>

<style scoped>
.image-workbench {
  position: relative;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  height: calc(100% + var(--app-page-padding-top, 1.25rem) + var(--app-page-padding-bottom, 2.5rem));
  min-height: 0;
  margin: calc(var(--app-page-padding-top, 1.25rem) * -1) 0 calc(var(--app-page-padding-bottom, 2.5rem) * -1);
  padding: .75rem 0 .35rem;
  color: var(--color-text-primary);
}

.toolbar {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: .55rem;
  margin-bottom: .85rem;
}

.pill-button,
.pill-select,
.search-field {
  display: inline-flex;
  align-items: center;
  min-height: 2.35rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 86%, transparent);
  color: var(--color-text-secondary);
}

.pill-button {
  justify-content: center;
  width: 2.35rem;
  padding: 0;
  cursor: pointer;
}

.pill-button:hover,
.pill-select:hover,
.search-field:focus-within,
.pill-button--active {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
}

.pill-button--active {
  color: var(--theme-accent);
  background: color-mix(in srgb, var(--theme-accent) 12%, var(--color-surface));
}

.pill-select {
  position: relative;
  min-width: 5.4rem;
  padding: 0 1.85rem 0 .95rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  color: var(--color-text-secondary);
  transition: border-color .15s ease, background-color .15s ease, box-shadow .15s ease;
}

.pill-select:hover,
.pill-select:focus-within {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--theme-accent) 55%, var(--color-border));
  background: color-mix(in srgb, var(--color-surface) 98%, transparent);
}

.pill-select:focus-within {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 25%, transparent);
}

.pill-select select,
.pill-select select:hover,
.pill-select select:focus,
.pill-select select:active {
  width: 100%;
  border: 0 !important;
  background: transparent !important;
  background-color: transparent !important;
  box-shadow: none !important;
  outline: none !important;
  -webkit-backdrop-filter: none !important;
  backdrop-filter: none !important;
  -webkit-appearance: none !important;
  -moz-appearance: none !important;
  appearance: none !important;
  color: var(--color-text-primary) !important;
  font: inherit;
  cursor: pointer;
}

.pill-select :deep(.app-icon) {
  position: absolute;
  right: .75rem;
  pointer-events: none;
  color: var(--color-text-tertiary);
  transition: color .15s ease;
}

.pill-select:hover :deep(.app-icon),
.pill-select:focus-within :deep(.app-icon) {
  color: var(--color-text-primary);
}

.search-field {
  flex: 1 1 auto;
  gap: .55rem;
  min-width: 0;
  padding: 0 .95rem;
}

.search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  background: transparent;
  color: var(--color-text-primary);
  font: inherit;
  outline: none;
}

.search-field input::-webkit-search-decoration,
.search-field input::-webkit-search-cancel-button {
  appearance: none;
}

.gallery {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  padding: .15rem .1rem 8.5rem;
  scrollbar-width: thin;
}

.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 22rem), 1fr));
  gap: .85rem;
  align-content: start;
  width: 100%;
}

.history-card {
  display: flex;
  overflow: hidden;
  width: 100%;
  min-height: 7.4rem;
  border: 1px solid var(--color-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  cursor: pointer;
  transition: transform .18s cubic-bezier(0.16, 1, 0.3, 1),
              border-color .18s ease,
              box-shadow .18s ease;
  box-sizing: border-box;
}

.history-card:hover {
  transform: translateY(-2px);
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  box-shadow: 0 8px 24px -4px rgba(0, 0, 0, .14), 0 2px 6px rgba(0, 0, 0, .04);
}

.history-card--pending {
  border-style: dashed;
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  animation: card-pulse 1.8s ease-in-out infinite alternate;
  cursor: default;
}

@keyframes card-pulse {
  from { opacity: .65; }
  to { opacity: .95; }
}

.history-card__thumb {
  position: relative;
  flex: 0 0 9.8rem;
  width: 9.8rem;
  padding: 0;
  border: 0;
  background: var(--color-surface-muted);
  cursor: zoom-in;
  overflow: hidden;
}

.history-card__thumb img,
.history-card__thumb--pending {
  display: grid;
  width: 100%;
  height: 100%;
  min-height: 7.4rem;
  object-fit: cover;
  place-items: center;
  color: var(--color-text-secondary);
  transition: transform .3s ease;
}

.history-card:hover .history-card__thumb img {
  transform: scale(1.03);
}

.thumb-badge {
  position: absolute;
  top: .45rem;
  left: .45rem;
  display: inline-flex;
  align-items: center;
  gap: .25rem;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, .68);
  border: 1px solid rgba(255, 255, 255, .16);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  color: #fff;
  padding: .15rem .45rem;
  font-size: .68rem;
  line-height: 1.2;
  letter-spacing: .02em;
  box-shadow: 0 2px 6px rgba(0, 0, 0, .25);
  pointer-events: none;
  max-width: calc(100% - .9rem);
}

.thumb-badge__ratio {
  font-weight: 600;
  color: #fff;
}

.thumb-badge__divider {
  opacity: .5;
}

.thumb-badge__res {
  opacity: .88;
  font-variant-numeric: tabular-nums;
}

.history-card__body {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  flex-direction: column;
  gap: .45rem;
  padding: .75rem .85rem .6rem;
}

.history-card__body h3 {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: .88rem;
  font-weight: 600;
  line-height: 1.36;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  transition: color .15s ease;
}

.history-card:hover .history-card__body h3 {
  color: var(--theme-accent);
}

.collection-card .history-card__body p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: .78rem;
}

.collection-card .history-card__thumb {
  cursor: inherit;
}

.history-card__tags,
.history-card__actions {
  display: flex;
  align-items: center;
  gap: .35rem;
}

.history-card__tags {
  flex-wrap: wrap;
  color: var(--color-text-secondary);
  font-size: .72rem;
}

.history-card__tags span {
  display: inline-flex;
  align-items: center;
  gap: .25rem;
  max-width: 100%;
  overflow: hidden;
  border-radius: var(--radius-sm, .375rem);
  background: color-mix(in srgb, var(--color-surface-muted) 80%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
  padding: .12rem .42rem;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: .7rem;
  line-height: 1.25;
}

.history-card__tags span :deep(.app-icon) {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.history-card__actions {
  justify-content: flex-end;
  margin-top: auto;
  gap: .25rem;
  padding-top: .2rem;
}

.history-card__actions button,
.icon-button,
.icon-action,
.send-button,
.composer__clear,
.lightbox__close {
  display: inline-grid;
  place-items: center;
  border: 0;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.history-card__actions button {
  width: 1.85rem;
  height: 1.85rem;
  border-radius: var(--radius-full);
  color: var(--color-text-tertiary);
  transition: all .15s ease;
}

.history-card__actions button:hover {
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-surface-muted) 90%, transparent);
}

.history-card__actions button.active {
  color: #f59e0b;
}

.history-card__actions button.active:hover {
  color: #d97706;
}

.history-card__actions button.action-btn-del:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, .12);
}

.empty-state {
  display: grid;
  min-height: 100%;
  place-content: center;
  justify-items: center;
  gap: .65rem;
  color: var(--color-text-secondary);
  text-align: center;
}

.composer {
  position: absolute;
  z-index: 5;
  left: 50%;
  bottom: .85rem;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  width: min(48rem, calc(100% - 1.5rem));
  box-sizing: border-box;
  transform: translateX(-50%);
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: 1.25rem;
  background: color-mix(in srgb, var(--color-surface) 90%, transparent);
  padding: .75rem .85rem .65rem;
  box-shadow: 0 16px 44px -8px rgba(0, 0, 0, .28), 0 2px 10px rgba(0, 0, 0, .06);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  transition: border-color .2s ease, box-shadow .2s ease;
}

.composer:focus-within {
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  box-shadow: 0 20px 50px -6px rgba(0, 0, 0, .32), 0 0 0 1px color-mix(in srgb, var(--theme-accent) 25%, transparent);
}

.composer__context {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: .45rem;
  min-height: 0;
}

.reference-chip {
  display: inline-flex;
  align-items: center;
  gap: .35rem;
  max-width: 100%;
  height: 1.75rem;
  padding: 0 .6rem;
  border: 1px solid color-mix(in srgb, var(--theme-accent) 35%, var(--color-border));
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--theme-accent) 12%, var(--color-surface));
  color: var(--color-text-primary);
  font-size: .75rem;
  cursor: pointer;
  transition: all .15s ease;
}

.reference-chip:hover {
  background: color-mix(in srgb, var(--theme-accent) 18%, var(--color-surface));
  border-color: var(--theme-accent);
}

.reference-chip__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 14rem;
}

.error-message {
  display: inline-flex;
  align-items: center;
  gap: .3rem;
  color: #ef4444;
  font-size: .75rem;
}

.composer__prompt {
  position: relative;
  width: 100%;
}

.composer__prompt textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 2.35rem;
  max-height: 7.5rem;
  border: 0;
  border-radius: .75rem;
  background: transparent;
  color: var(--color-text-primary);
  padding: .3rem 2rem .3rem .15rem;
  resize: none;
  font: inherit;
  font-size: .88rem;
  line-height: 1.5;
  outline: none;
}

.composer__prompt textarea::placeholder {
  color: color-mix(in srgb, var(--color-text-secondary) 85%, transparent);
}

.composer__clear {
  position: absolute;
  top: .25rem;
  right: .2rem;
  width: 1.5rem;
  height: 1.5rem;
  border: 0;
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface-muted) 80%, transparent);
  color: var(--color-text-secondary);
  cursor: pointer;
  display: inline-grid;
  place-items: center;
  transition: all .15s ease;
}

.composer__clear:hover {
  color: var(--color-text-primary);
  background: var(--color-surface-muted);
}

.composer__advanced {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: .45rem;
  padding: .45rem .6rem;
  border-radius: .75rem;
  background: color-mix(in srgb, var(--color-surface-muted) 55%, transparent);
  border: 1px dashed color-mix(in srgb, var(--color-border) 70%, transparent);
}

.advanced-slide-enter-active,
.advanced-slide-leave-active {
  transition: all .2s ease;
}

.advanced-slide-enter-from,
.advanced-slide-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.composer__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .65rem;
  width: 100%;
}

.composer__toolbar-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: .45rem;
  min-width: 0;
}

.composer__toolbar-right {
  display: flex;
  align-items: center;
  gap: .45rem;
  flex-shrink: 0;
}

.toolbar-pill {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: .35rem;
  height: 2rem;
  padding: 0 .65rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  color: var(--color-text-secondary);
  font: inherit;
  font-size: .78rem;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
  transition: border-color .15s ease, background-color .15s ease, box-shadow .15s ease, color .15s ease;
  box-sizing: border-box;
}

.toolbar-pill:hover,
.toolbar-pill:focus-within {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--theme-accent) 55%, var(--color-border));
  background: color-mix(in srgb, var(--color-surface) 98%, transparent);
}

.toolbar-pill:focus-within {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 25%, transparent);
}

.toolbar-pill__prefix {
  color: var(--color-text-secondary);
  font-size: .74rem;
  opacity: .9;
}

.toolbar-pill__val {
  color: var(--color-text-primary);
  font-weight: 500;
  max-width: 7.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toolbar-pill--model {
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  font-weight: 500;
  max-width: 11.5rem;
}

.toolbar-pill--model .toolbar-pill__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toolbar-pill--model:hover {
  color: var(--theme-accent);
}

.toolbar-pill--button {
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
}

.toolbar-pill--select select,
.toolbar-pill--select select:hover,
.toolbar-pill--select select:focus,
.toolbar-pill--select select:active {
  border: 0 !important;
  background: transparent !important;
  background-color: transparent !important;
  box-shadow: none !important;
  outline: none !important;
  -webkit-backdrop-filter: none !important;
  backdrop-filter: none !important;
  -webkit-appearance: none !important;
  -moz-appearance: none !important;
  appearance: none !important;
  color: var(--color-text-primary) !important;
  font: inherit;
  font-size: inherit;
  font-weight: 500;
  cursor: pointer;
  padding: 0 .15rem 0 0;
  margin: 0;
}

.toolbar-pill--select :deep(.app-icon) {
  color: var(--color-text-tertiary);
  pointer-events: none;
  transition: transform .15s ease, color .15s ease;
}

.toolbar-pill--select:hover :deep(.app-icon),
.toolbar-pill--select:focus-within :deep(.app-icon) {
  color: var(--color-text-primary);
}

.toolbar-pill--number input,
.toolbar-pill--input input {
  width: 2.2rem;
  border: 0;
  background: transparent;
  color: var(--color-text-primary);
  font: inherit;
  font-size: inherit;
  font-weight: 500;
  text-align: center;
  outline: none;
  appearance: textfield;
  -moz-appearance: textfield;
}

.toolbar-pill--number input::-webkit-outer-spin-button,
.toolbar-pill--number input::-webkit-inner-spin-button,
.toolbar-pill--input input::-webkit-outer-spin-button,
.toolbar-pill--input input::-webkit-inner-spin-button {
  appearance: none;
  margin: 0;
}

.toolbar-pill__suffix {
  color: var(--color-text-secondary);
  font-size: .74rem;
  margin-left: -.2rem;
}

.toolbar-pill--toggle {
  background: transparent;
}

.toolbar-pill--active {
  color: var(--theme-accent);
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  background: color-mix(in srgb, var(--theme-accent) 10%, var(--color-surface));
}

.toolbar-pill__dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--theme-accent);
  display: inline-block;
}

.action-btn {
  display: inline-grid;
  place-items: center;
  width: 2.2rem;
  height: 2.2rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface-muted) 80%, transparent);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all .15s ease;
  box-sizing: border-box;
}

.action-btn:hover {
  color: var(--theme-accent);
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  background: var(--color-surface);
}

.action-btn--filled {
  color: var(--theme-accent);
  border-color: var(--theme-accent);
  background: color-mix(in srgb, var(--theme-accent) 12%, var(--color-surface));
}

.action-btn--send {
  border: 0;
  border-radius: var(--radius-md, .75rem);
  background: var(--theme-accent);
  color: #fff;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--theme-accent) 35%, transparent);
}

.action-btn--send:hover:not(:disabled) {
  opacity: .92;
  transform: translateY(-1px);
  color: #fff;
  box-shadow: 0 6px 16px color-mix(in srgb, var(--theme-accent) 45%, transparent);
}

.action-btn--send:active:not(:disabled) {
  transform: translateY(0);
}

.action-btn--send:disabled {
  opacity: .4;
  cursor: not-allowed;
  box-shadow: none;
}

.btn-spinner {
  width: 1rem;
  height: 1rem;
  border: 2px solid rgba(255, 255, 255, .35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: btn-spin .7s linear infinite;
}

@keyframes btn-spin {
  to { transform: rotate(360deg); }
}

select {
  color-scheme: light dark;
}

select option {
  background-color: #ffffff;
  color: #09090b;
  padding: .4rem .6rem;
}

:global(html.dark) select option,
:global(.dark) select option {
  background-color: #1e1e28;
  color: #f4f4f6;
}

.param-field,
.settings-form label,
.settings-step {
  display: grid;
  gap: .28rem;
  color: var(--color-text-secondary);
  font-size: .72rem;
}

.param-field select,
.param-field input,
.param-field__value,
.settings-form select,
.settings-form input {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background-color: color-mix(in srgb, var(--color-surface) 92%, transparent);
  color: var(--color-text-primary);
  padding: .45rem .8rem;
  font: inherit;
  font-size: .82rem;
  transition: border-color .15s ease, background-color .15s ease;
}

.param-field select,
.settings-form select,
.param-field__value {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2394a3b8'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='1.8' d='M19.5 8.25l-7.5 7.5-7.5-7.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right .65rem center;
  background-size: .8rem;
  padding-right: 1.65rem;
}

.settings-form select:hover,
.param-field select:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
  background-color: color-mix(in srgb, var(--color-surface) 98%, transparent);
}

.settings-form select:focus,
.param-field select:focus {
  border-color: var(--theme-accent);
  background-color: var(--color-surface);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 25%, transparent);
  outline: none;
}

.settings-form select:disabled,
.param-field select:disabled {
  opacity: .5;
  cursor: not-allowed;
}

.param-field--button {
  min-width: 6.2rem;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.param-field__value {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reference-chip,
.primary-button,
.secondary-button,
.danger-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: .4rem;
  min-height: 2.1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  padding: .3rem .7rem;
  cursor: pointer;
  font: inherit;
  font-size: .8rem;
}

.reference-chip {
  background: color-mix(in srgb, var(--color-surface-muted) 88%, transparent);
  color: var(--color-text-secondary);
}

.error-message {
  color: #dc2626;
  font-size: .75rem;
}

.lightbox {
  position: fixed;
  inset: 0;
  z-index: 30;
  display: grid;
  place-items: center;
  padding: 2rem 1rem 3rem;
  background: rgba(0, 0, 0, .72);
}

.lightbox img {
  max-width: min(92vw, 72rem);
  max-height: calc(100vh - 7rem);
  border-radius: 1rem;
  object-fit: contain;
}

.lightbox p {
  max-width: min(42rem, 90vw);
  margin: .85rem 0 0;
  color: #fff;
  text-align: center;
  white-space: pre-wrap;
}

.lightbox__close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 2.4rem;
  height: 2.4rem;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, .12);
  color: #fff;
}

.settings-overlay {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: grid;
  place-items: center;
  padding: 1rem;
  background: rgba(0, 0, 0, .38);
}

.settings-panel {
  width: min(100%, 430px);
  border-radius: 12px;
  background: var(--color-surface);
  padding: 1rem;
  box-shadow: 0 20px 55px rgba(0, 0, 0, .22);
}

.settings-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.settings-panel h2 {
  margin: 0;
  font-size: 1.1rem;
}

.icon-button {
  width: 2.1rem;
  height: 2.1rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
}

.settings-tabs {
  display: flex;
  gap: .25rem;
  margin: 1rem 0;
  border-bottom: 1px solid var(--color-border);
}

.settings-tabs button {
  flex: 1;
  border: 0;
  border-bottom: 2px solid transparent;
  padding: .6rem;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.settings-tabs button.active {
  border-color: var(--theme-accent);
  color: var(--theme-accent);
}

.settings-form {
  display: grid;
  gap: .8rem;
}

.settings-note {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: .85rem;
}

.platform-picks {
  display: flex;
  flex-wrap: wrap;
  gap: .45rem;
}

.platform-picks button {
  min-height: 2.2rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface-muted) 88%, transparent);
  color: var(--color-text-secondary);
  padding: .35rem .95rem;
  cursor: pointer;
  font: inherit;
}

.platform-picks button.active,
.platform-picks button:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--theme-accent) 50%, var(--color-border));
}

.platform-picks button.active {
  background: color-mix(in srgb, var(--theme-accent) 14%, var(--color-surface));
  color: var(--theme-accent);
}

.primary-button {
  border-color: transparent;
  background: var(--theme-accent);
  color: #fff;
}

.secondary-button {
  background: var(--color-surface);
  color: var(--color-text-primary);
}

.danger-button {
  border-color: color-mix(in srgb, #ef4444 45%, var(--color-border));
  background: transparent;
  color: #dc2626;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
}

@media (max-width: 720px) {
  .gallery {
    padding-bottom: 9.5rem;
  }

  .history-card {
    width: 100%;
  }

  .composer {
    width: calc(100% - .75rem);
    bottom: .45rem;
    padding: .65rem .75rem .6rem;
    border-radius: 1rem;
  }

  .composer__toolbar {
    flex-wrap: wrap;
    gap: .5rem;
  }

  .composer__toolbar-left {
    flex: 1 1 100%;
    overflow-x: auto;
    flex-wrap: nowrap;
    padding-bottom: .25rem;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
  }

  .composer__toolbar-left::-webkit-scrollbar {
    display: none;
  }

  .composer__toolbar-right {
    margin-left: auto;
  }
}
</style>
