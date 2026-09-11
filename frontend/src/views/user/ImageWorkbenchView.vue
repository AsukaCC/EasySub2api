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
        <div v-else-if="filteredHistory.length === 0 && !generating" class="empty-state">
          <Icon name="photo" size="xl" />
          <strong>{{ emptyMessage }}</strong>
        </div>
        <div v-else class="gallery-grid">
          <article v-if="generating" class="history-card history-card--pending">
            <div class="history-card__thumb history-card__thumb--pending">
              <Icon name="sparkles" size="md" />
            </div>
            <div class="history-card__body">
              <h3>{{ prompt.trim() || t('imageWorkbench.generating') }}</h3>
              <p>{{ t('imageWorkbench.generating') }}</p>
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
              <span v-if="aspectLabel(item)" class="thumb-badge">{{ aspectLabel(item) }}</span>
              <span v-if="resolutionLabel(item)" class="thumb-badge thumb-badge--size">{{ resolutionLabel(item) }}</span>
            </button>
            <div class="history-card__body">
              <h3>{{ item.prompt }}</h3>
              <div class="history-card__tags">
                <span><Icon name="key" size="xs" />{{ item.keyName || t('imageWorkbench.apiKey') }}</span>
                <span><Icon name="sparkles" size="xs" />{{ item.model }}</span>
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
                <button type="button" :title="t('imageWorkbench.delete')" @click.stop="removeItem(item)">
                  <Icon name="trash" size="xs" />
                </button>
              </div>
            </div>
          </article>
        </div>
      </section>

      <form class="composer" @submit.prevent="submitGeneration">
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

        <div class="composer__params">
          <button class="icon-action composer__settings" type="button" :title="t('imageWorkbench.settings')" @click="openSettings">
            <Icon name="cog" size="sm" />
          </button>
          <button class="param-field param-field--button" type="button" @click="openSizePicker('params')">
            <span>{{ t('imageWorkbench.size') }}</span>
            <span class="param-field__value">{{ params.size || 'auto' }}</span>
          </button>
          <label v-if="adapter.capabilities.quality" class="param-field">
            <span>{{ t('imageWorkbench.quality') }}</span>
            <select v-model="params.quality">
              <option value="auto">auto</option>
              <option value="low">low</option>
              <option value="medium">medium</option>
              <option value="high">high</option>
            </select>
          </label>
          <label v-if="adapter.capabilities.outputFormat" class="param-field">
            <span>{{ t('imageWorkbench.format') }}</span>
            <select v-model="params.output_format">
              <option value="png">PNG</option>
              <option value="jpeg">JPEG</option>
              <option value="webp">WebP</option>
            </select>
          </label>
          <label v-if="adapter.capabilities.transparency" class="param-field">
            <span>{{ t('imageWorkbench.transparent') }}</span>
            <select v-model="params.background">
              <option value="auto">auto</option>
              <option value="transparent">true</option>
              <option value="opaque">false</option>
            </select>
          </label>
          <label v-if="adapter.capabilities.quality" class="param-field">
            <span>{{ t('imageWorkbench.moderation') }}</span>
            <select v-model="params.moderation">
              <option value="auto">auto</option>
              <option value="low">low</option>
            </select>
          </label>
          <label class="param-field param-field--quantity">
            <span>{{ t('imageWorkbench.quantity') }}</span>
            <input v-model.number="params.n" type="number" min="1" max="4" />
          </label>
          <label class="icon-action" :class="{ 'icon-action--filled': referenceFile }" :title="referenceFile?.name || t('imageWorkbench.uploadReference')">
            <Icon name="paperclip" size="sm" />
            <input type="file" accept="image/png,image/jpeg,image/webp" hidden @change="onReferenceChange" />
          </label>
          <button class="send-button" type="submit" :disabled="generating || !prompt.trim()" :title="generating ? t('imageWorkbench.generating') : t('imageWorkbench.generate')">
            <Icon name="arrowRight" size="sm" />
          </button>
        </div>

        <div v-if="referenceFile || errorMessage" class="composer__meta">
          <button v-if="referenceFile" class="reference-chip" type="button" @click="clearReference">
            <Icon name="photo" size="xs" />
            <span>{{ referenceFile.name }}</span>
            <Icon name="x" size="xs" />
          </button>
          <span v-if="errorMessage" class="error-message">{{ errorMessage }}</span>
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
  imagePlatformAdapters,
  listImageModels,
  loadWorkbenchCredentials,
  type ImageModel,
  type ImagePlatform,
  type ImagePlatformAdapter,
} from '@/api/imageWorkbench'
import { clearHistory, deleteHistory, exportHistory, importHistory, listHistory, putHistory, type ImageHistoryItem } from '@/features/image-workbench/storage'
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
const activePlatform = ref<ImagePlatform>('openai')
const credentials = ref<{ keys: import('@/types').ApiKey[]; groups: import('@/types').Group[] }>({ keys: [], groups: [] })
const loadingHistory = ref(true)
const generating = ref(false)
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

function aspectLabel(item: ImageHistoryItem): string {
  const size = itemSize(item)
  if (!size) return ''
  const divisor = gcd(size.width, size.height)
  return `${size.width / divisor}:${size.height / divisor}`
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

function reuseItem(item: ImageHistoryItem) {
  prompt.value = item.prompt
  if (typeof item.params?.size === 'string') params.size = item.params.size
  if (typeof item.params?.quality === 'string') params.quality = item.params.quality
  if (typeof item.params?.output_format === 'string') params.output_format = item.params.output_format
  if (typeof item.params?.background === 'string') params.background = item.params.background
  if (typeof item.params?.moderation === 'string') params.moderation = item.params.moderation
  if (typeof item.params?.n === 'number') params.n = item.params.n
  if (item.model) selectedByPlatform[activePlatform.value].model = item.model
  void nextTick(() => autosizePrompt())
}

async function submitGeneration() {
  if (!selectedKey.value?.key || !selectedModel.value) {
    errorMessage.value = t('imageWorkbench.configureApi')
    openSettings()
    return
  }
  if (!prompt.value.trim()) {
    errorMessage.value = t('imageWorkbench.promptRequired')
    return
  }
  generating.value = true
  errorMessage.value = ''
  try {
    const results = await generateImage(
      selectedKey.value.key,
      {
        prompt: prompt.value.trim(),
        model: selectedModel.value,
        size: params.size,
        quality: params.quality,
        output_format: params.output_format,
        output_compression: params.output_compression,
        background: params.background,
        moderation: params.moderation,
        n: Math.min(4, Math.max(1, Number(params.n) || 1)),
      },
      referenceFile.value || undefined,
    )
    for (const result of results) {
      const src = imageSource(result)
      if (!src) continue
      const measured = await measureImage(src)
      await putHistory({
        id: crypto.randomUUID(),
        createdAt: Date.now(),
        prompt: prompt.value.trim(),
        model: selectedModel.value,
        platform: activePlatform.value,
        src,
        revisedPrompt: result.revised_prompt,
        favorite: false,
        keyName: selectedKey.value.name,
        width: measured?.width,
        height: measured?.height,
        params: { ...params },
      })
    }
    history.value = await listHistory()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('imageWorkbench.generateFailed')
  } finally {
    generating.value = false
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
    if (saved.platform === 'openai' || saved.platform === 'grok') activePlatform.value = saved.platform
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
  } finally {
    loadingHistory.value = false
  }
  persistFavorites()
  if (!appStore.publicSettingsLoaded) await appStore.fetchPublicSettings().catch(() => undefined)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
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
  padding: 0 1.65rem 0 .9rem;
}

.pill-select select {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  appearance: none;
}

.pill-select :deep(.app-icon) {
  position: absolute;
  right: .65rem;
  pointer-events: none;
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
  display: flex;
  flex-wrap: wrap;
  gap: .75rem;
  align-content: start;
}

.history-card {
  display: flex;
  overflow: hidden;
  width: min(100%, 24.5rem);
  min-height: 6.6rem;
  border: 1px solid color-mix(in srgb, var(--color-border) 80%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  cursor: pointer;
}

.history-card--pending {
  opacity: .72;
  pointer-events: none;
}

.history-card__thumb {
  position: relative;
  flex: 0 0 9.4rem;
  width: 9.4rem;
  padding: 0;
  border: 0;
  background: var(--color-surface-muted);
  cursor: zoom-in;
}

.history-card__thumb img,
.history-card__thumb--pending {
  display: grid;
  width: 100%;
  height: 100%;
  min-height: 6.6rem;
  object-fit: cover;
  place-items: center;
  color: var(--color-text-secondary);
}

.thumb-badge {
  position: absolute;
  top: .4rem;
  left: .4rem;
  border-radius: .4rem;
  background: rgba(0, 0, 0, .55);
  color: #fff;
  padding: .12rem .38rem;
  font-size: .68rem;
  line-height: 1.2;
}

.thumb-badge--size {
  left: auto;
  right: .4rem;
}

.history-card__body {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  flex-direction: column;
  gap: .45rem;
  padding: .7rem .75rem .55rem;
}

.history-card__body h3 {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: .92rem;
  font-weight: 600;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
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
  gap: .4rem;
}

.history-card__tags {
  flex-wrap: wrap;
  color: var(--color-text-secondary);
  font-size: .72rem;
}

.history-card__tags span {
  display: inline-flex;
  align-items: center;
  gap: .22rem;
  max-width: 100%;
  overflow: hidden;
  border-radius: .4rem;
  background: color-mix(in srgb, var(--color-surface-muted) 80%, transparent);
  padding: .12rem .4rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-card__actions {
  justify-content: flex-end;
  margin-top: auto;
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
  width: 1.7rem;
  height: 1.7rem;
  border-radius: var(--radius-full);
}

.history-card__actions button:hover,
.history-card__actions button.active {
  color: var(--theme-accent);
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
  display: grid;
  width: min(46rem, calc(100% - 1.5rem));
  box-sizing: border-box;
  transform: translateX(-50%);
  border: 1px solid color-mix(in srgb, var(--color-border) 88%, transparent);
  border-radius: 1.35rem;
  background: color-mix(in srgb, var(--color-surface) 94%, transparent);
  padding: .7rem 1rem .75rem 3.15rem;
  box-shadow: 0 18px 48px rgba(0, 0, 0, .28);
  backdrop-filter: blur(18px);
}

.composer__params {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: .45rem;
}

.composer__settings {
  position: absolute;
  left: .65rem;
  bottom: .65rem;
  z-index: 1;
}

.composer__creds label,
.param-field,
.settings-form label,
.settings-step {
  display: grid;
  gap: .28rem;
  color: var(--color-text-secondary);
  font-size: .72rem;
}

.composer__creds select,
.param-field select,
.param-field input,
.param-field__value,
.settings-form select,
.settings-form input,
.composer__prompt textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-surface-muted) 88%, transparent);
  color: var(--color-text-primary);
  padding: .42rem .7rem;
  font: inherit;
  font-size: .82rem;
}

.param-field select,
.composer__creds select,
.settings-form select,
.param-field__value {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2394a3b8'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='1.8' d='M19.5 8.25l-7.5 7.5-7.5-7.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right .55rem center;
  background-size: .8rem;
  padding-right: 1.55rem;
}

.param-field input[type='number'] {
  appearance: textfield;
}

.param-field input[type='number']::-webkit-outer-spin-button,
.param-field input[type='number']::-webkit-inner-spin-button {
  appearance: none;
}

.composer__prompt {
  position: relative;
  margin-bottom: .55rem;
}

.composer__prompt textarea {
  min-height: 2.4rem;
  max-height: 7rem;
  border-radius: .9rem;
  padding: .65rem 2.1rem .65rem .9rem;
  resize: none;
  line-height: 1.45;
}

.composer__clear {
  position: absolute;
  top: .55rem;
  right: .45rem;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: var(--radius-full);
}

.param-field {
  flex: 1 1 0;
  min-width: 5.1rem;
}

.param-field--quantity {
  flex: 0 0 4.1rem;
  min-width: 3.8rem;
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

.icon-action,
.send-button {
  flex: 0 0 auto;
  width: 2.35rem;
  height: 2.35rem;
  margin-bottom: .05rem;
}

.icon-action {
  border-radius: var(--radius-full);
}

.send-button {
  border-radius: .8rem;
}

.icon-action {
  border: 1px solid var(--color-border);
  background: color-mix(in srgb, var(--color-surface-muted) 88%, transparent);
}

.icon-action--filled,
.icon-action:hover {
  color: var(--theme-accent);
  border-color: color-mix(in srgb, var(--theme-accent) 45%, var(--color-border));
}

.send-button {
  background: #3b82f6;
  color: #fff;
}

.send-button:hover {
  background: #2563eb;
  color: #fff;
}

.send-button:disabled {
  opacity: .45;
  cursor: not-allowed;
}

.composer__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: .5rem;
  margin-top: .55rem;
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
    padding-bottom: 11rem;
  }

  .history-card {
    width: 100%;
  }

  .composer {
    width: calc(100% - .75rem);
    bottom: .5rem;
    padding: .6rem .7rem .7rem 3rem;
  }

  .param-field {
    flex: 1 1 calc(50% - .45rem);
  }
}
</style>
