<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="components-admin-account-account-test-modal__panel">
      <!-- Account Info Card -->
      <div
        v-if="account"
        class="components-admin-account-account-test-modal__panel-2"
      >
        <div class="components-admin-account-account-test-modal__panel-3">
          <div
            class="components-admin-account-account-test-modal__panel-4"
          >
            <Icon name="play" size="md" class="components-admin-account-account-test-modal__icon" :stroke-width="2" />
          </div>
          <div>
            <div class="components-admin-account-account-test-modal__panel-5">{{ account.name }}</div>
            <div class="components-admin-account-account-test-modal__panel-6">
              <span
                class="components-admin-account-account-test-modal__text"
              >
                {{ account.type }}
              </span>
              <span>{{ t('admin.accounts.account') }}</span>
            </div>
          </div>
        </div>
        <span
          :class="[
            'components-admin-account-account-test-modal__text-5',
            account.status === 'active'
              ? 'components-admin-account-account-test-modal__text-7'
              : 'components-admin-account-account-test-modal__text-8'
          ]"
        >
          {{ account.status }}
        </span>
      </div>

      <!-- Grok: mode first, then optional model / mode params -->
      <div v-if="isGrokAccount" class="components-admin-account-account-test-modal__panel-7">
        <label class="components-admin-account-account-test-modal__label">
          {{ t('admin.accounts.grok.testMode') }}
        </label>
        <Select
          v-model="grokTestMode"
          :options="grokTestModeOptions"
          :disabled="status === 'connecting'"
        />
        <p class="components-admin-account-account-test-modal__description">
          {{ t('admin.accounts.grok.testModeHint') }}
        </p>
      </div>

      <div v-if="showModelSelect" class="components-admin-account-account-test-modal__panel-7">
        <label class="components-admin-account-account-test-modal__label">
          {{ t('admin.accounts.selectTestModel') }}
        </label>
        <Select
          v-model="selectedModelId"
          :options="modelOptionsForMode"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? t('common.loading') + '...' : t('admin.accounts.selectTestModel')"
        />
      </div>

      <div v-if="isOpenAIAccount" class="components-admin-account-account-test-modal__panel-7">
        <label class="components-admin-account-account-test-modal__label">
          {{ t('admin.accounts.openai.testMode') }}
        </label>
        <Select
          v-model="testMode"
          :options="openAITestModeOptions"
          :disabled="status === 'connecting'"
        />
      </div>

      <div v-if="supportsPromptInput" class="components-admin-account-account-test-modal__panel-7">
        <TextArea
          v-model="testPrompt"
          :label="promptInputLabel"
          :placeholder="promptInputPlaceholder"
          :hint="promptInputHint"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>
      <p
        v-else-if="isGrokAccount && promptInputHint"
        class="components-admin-account-account-test-modal__description"
      >
        {{ promptInputHint }}
      </p>

      <!-- Optional media uploads for real generation / transcription -->
      <div v-if="supportsImageUpload" class="components-admin-account-account-test-modal__panel-7">
        <label class="components-admin-account-account-test-modal__label">
          {{ imageUploadLabel }}
        </label>
        <div class="components-admin-account-account-test-modal__panel-3">
          <button
            type="button"
            class="components-admin-account-account-test-modal__action btn btn-secondary btn-sm"
            :disabled="status === 'connecting'"
            @click="imageFileInput?.click()"
          >
            {{ t('admin.accounts.grok.chooseImageFile') }}
          </button>
          <span class="components-admin-account-account-test-modal__text-2">
            {{
              uploadImageName
                ? t('common.selectedFile', { name: uploadImageName })
                : t('common.noFileSelected')
            }}
          </span>
          <input
            ref="imageFileInput"
            type="file"
            accept="image/png,image/jpeg,image/webp,image/gif"
            class="components-admin-account-account-test-modal__field"
            :disabled="status === 'connecting'"
            @change="onImageFileChange"
          />
        </div>
        <p class="components-admin-account-account-test-modal__description">{{ imageUploadHint }}</p>
        <div v-if="uploadImagePreview" class="components-admin-account-account-test-modal__panel-8">
          <img
            :src="uploadImagePreview"
            :alt="t('admin.accounts.grok.uploadPreviewAlt')"
            class="components-admin-account-account-test-modal__image"
          />
        </div>
      </div>

      <div v-if="supportsAudioUpload" class="components-admin-account-account-test-modal__panel-7">
        <label class="components-admin-account-account-test-modal__label">
          {{ t('admin.accounts.grok.audioUploadLabel') }}
        </label>
        <div class="components-admin-account-account-test-modal__panel-3">
          <button
            type="button"
            class="components-admin-account-account-test-modal__action btn btn-secondary btn-sm"
            :disabled="status === 'connecting'"
            @click="audioFileInput?.click()"
          >
            {{ t('admin.accounts.grok.chooseAudioFile') }}
          </button>
          <span class="components-admin-account-account-test-modal__text-2">
            {{
              uploadAudioName
                ? t('common.selectedFile', { name: uploadAudioName })
                : t('common.noFileSelected')
            }}
          </span>
          <input
            ref="audioFileInput"
            type="file"
            accept="audio/*,.wav,.mp3,.m4a,.ogg,.webm"
            class="components-admin-account-account-test-modal__field"
            :disabled="status === 'connecting'"
            @change="onAudioFileChange"
          />
        </div>
        <p class="components-admin-account-account-test-modal__description">{{ t('admin.accounts.grok.audioUploadHint') }}</p>
      </div>

      <!-- Terminal Output -->
      <div class="components-admin-account-account-test-modal__panel-9">
        <div
          ref="terminalRef"
          class="components-admin-account-account-test-modal__panel-10"
        >
          <!-- Status Line -->
          <div v-if="status === 'idle'" class="components-admin-account-account-test-modal__panel-11">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.readyToTest') }}</span>
          </div>
          <div v-else-if="status === 'connecting'" class="components-admin-account-account-test-modal__panel-12">
            <Icon name="refresh" size="sm" class="components-admin-account-account-test-modal__icon-2" :stroke-width="2" />
            <span>{{ t('admin.accounts.connectingToApi') }}</span>
          </div>

          <!-- Output Lines -->
          <div v-for="(line, index) in outputLines" :key="index" :class="line.class">
            {{ line.text }}
          </div>

          <!-- Streaming Content -->
          <div v-if="streamingContent" class="components-admin-account-account-test-modal__panel-13">
            {{ streamingContent }}<span class="components-admin-account-account-test-modal__text-3">_</span>
          </div>

          <!-- Result Status -->
          <div
            v-if="status === 'success'"
            class="components-admin-account-account-test-modal__panel-14"
          >
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.testCompleted') }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="components-admin-account-account-test-modal__panel-15"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <!-- Copy Button -->
        <button
          v-if="outputLines.length > 0"
          @click="copyOutput"
          class="components-admin-account-account-test-modal__action-2"
          :title="t('admin.accounts.copyOutput')"
        >
          <Icon name="link" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div v-if="generatedImages.length > 0" class="components-admin-account-account-test-modal__panel-16">
        <div class="components-admin-account-account-test-modal__panel-17">
          {{ t('admin.accounts.imagePreview') }}
        </div>
        <div class="components-admin-account-account-test-modal__panel-18">
          <div
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            class="components-admin-account-account-test-modal__panel-19"
            @click="previewImageUrl = image.url"
          >
            <img
              :src="image.url"
              :alt="t('admin.accounts.imagePreviewAlt', { index: index + 1 })"
              class="components-admin-account-account-test-modal__image-2"
            />
            <div class="components-admin-account-account-test-modal__panel-20">
              <Icon name="eye" size="lg" class="components-admin-account-account-test-modal__icon-3" :stroke-width="2" />
            </div>
            <div class="components-admin-account-account-test-modal__panel-21">
              {{ image.mimeType || 'image/*' }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="generatedAudios.length > 0" class="components-admin-account-account-test-modal__panel-16">
        <div class="components-admin-account-account-test-modal__panel-17">
          {{ t('admin.accounts.audioPreview') }}
        </div>
        <div
          v-for="(audio, index) in generatedAudios"
          :key="`audio-${index}`"
          class="components-admin-account-account-test-modal__panel-22"
        >
          <audio :src="audio.url" controls class="components-admin-account-account-test-modal__audio" :type="audio.mimeType" />
          <div class="components-admin-account-account-test-modal__panel-23">{{ audio.mimeType || 'audio/*' }}</div>
        </div>
      </div>

      <div v-if="generatedVideos.length > 0" class="components-admin-account-account-test-modal__panel-16">
        <div class="components-admin-account-account-test-modal__panel-17">
          {{ t('admin.accounts.videoPreview') }}
        </div>
        <div
          v-for="(video, index) in generatedVideos"
          :key="`video-${index}`"
          class="components-admin-account-account-test-modal__panel-24"
        >
          <video :src="video.url" controls class="components-admin-account-account-test-modal__video" :type="video.mimeType" />
          <div class="components-admin-account-account-test-modal__panel-25">
            {{ video.mimeType || 'video/*' }}
          </div>
        </div>
      </div>

      <!-- Image Lightbox -->
      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="previewImageUrl"
            class="components-admin-account-account-test-modal__panel-26"
            @click.self="previewImageUrl = ''"
          >
            <button
              class="components-admin-account-account-test-modal__action-3"
              @click="previewImageUrl = ''"
            >
              <Icon name="x" size="lg" :stroke-width="2" />
            </button>
            <img
              :src="previewImageUrl"
              :alt="t('admin.accounts.imageLightboxAlt')"
              class="components-admin-account-account-test-modal__image-3"
            />
          </div>
        </Transition>
      </Teleport>

      <!-- Test Info -->
      <div class="components-admin-account-account-test-modal__panel-27">
        <div class="components-admin-account-account-test-modal__panel-3">
          <span class="components-admin-account-account-test-modal__text-4">
            <Icon name="grid" size="sm" :stroke-width="2" />
            {{ t('admin.accounts.testModel') }}
          </span>
        </div>
        <span class="components-admin-account-account-test-modal__text-4">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{ testModeSummary }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="components-admin-account-account-test-modal__panel-28">
        <button
          @click="handleClose"
          class="components-admin-account-account-test-modal__action-4"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="startTest"
          :disabled="!canStartTest"
          :class="[
            'components-admin-account-account-test-modal__action-5',
            !canStartTest
              ? 'components-admin-account-account-test-modal__action-6'
              : status === 'success'
                ? 'components-admin-account-account-test-modal__action-7'
                : status === 'error'
                  ? 'components-admin-account-account-test-modal__action-8'
                  : 'components-admin-account-account-test-modal__action-9'
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="components-admin-account-account-test-modal__icon-2"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('admin.accounts.testing')
                : status === 'idle'
                  ? t('admin.accounts.startTest')
                  : t('admin.accounts.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { buildApiUrl } from '@/api/client'
import { ADMIN_UI_REQUEST_HEADER } from '@/api/adminUIRequest'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

interface OutputLine {
  text: string
  class: string
}

interface PreviewMedia {
  url: string
  mimeType?: string
}

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
let abortController: AbortController | null = null
const generatedImages = ref<PreviewMedia[]>([])
const generatedAudios = ref<PreviewMedia[]>([])
const generatedVideos = ref<PreviewMedia[]>([])
const previewImageUrl = ref('')
const testMode = ref<'default' | 'compact'>('default')
const grokTestMode = ref<'text' | 'image' | 'video' | 'search' | 'tts' | 'stt' | 'realtime'>('text')
const uploadImageDataURL = ref('')
const uploadImagePreview = ref('')
const uploadImageName = ref('')
const uploadAudioDataURL = ref('')
const uploadAudioName = ref('')
const imageFileInput = ref<HTMLInputElement | null>(null)
const audioFileInput = ref<HTMLInputElement | null>(null)
const isOpenAIAccount = computed(() => props.account?.platform === 'openai')
const isGrokAccount = computed(() => props.account?.platform === 'grok')
const openAITestModeOptions = computed(() => [
  { value: 'default', label: t('admin.accounts.openai.testModeDefault') },
  { value: 'compact', label: t('admin.accounts.openai.testModeCompact') }
])
const grokTestModeOptions = computed(() => [
  { value: 'text', label: t('admin.accounts.grok.testModeText') },
  { value: 'image', label: t('admin.accounts.grok.testModeImage') },
  { value: 'video', label: t('admin.accounts.grok.testModeVideo') },
  { value: 'search', label: t('admin.accounts.grok.testModeSearch') },
  { value: 'tts', label: t('admin.accounts.grok.testModeTTS') },
  { value: 'stt', label: t('admin.accounts.grok.testModeSTT') },
  { value: 'realtime', label: t('admin.accounts.grok.testModeRealtime') }
])
const supportsOpenAIImageTest = computed(() => {
  const modelID = selectedModelId.value.toLowerCase()
  if (!modelID.startsWith('gpt-image-')) return false
  return props.account?.platform === 'openai'
})

const isGrokImageModel = (id: string) => {
  const modelID = id.toLowerCase()
  return (
    modelID === 'grok-imagine' ||
    modelID === 'grok-imagine-edit' ||
    modelID.startsWith('grok-imagine-image')
  )
}
const isGrokVideoModel = (id: string) => {
  const modelID = id.toLowerCase()
  return modelID.startsWith('grok-imagine-video') || modelID.startsWith('grok-video')
}
const isGrokTextModel = (id: string) => !isGrokImageModel(id) && !isGrokVideoModel(id)

const supportsGrokImageTest = computed(
  () => isGrokAccount.value && grokTestMode.value === 'image'
)
const supportsGrokVideoTest = computed(
  () => isGrokAccount.value && grokTestMode.value === 'video'
)

const supportsImageTest = computed(
  () => supportsOpenAIImageTest.value || supportsGrokImageTest.value
)

// Model select only when the mode needs a model.
const showModelSelect = computed(() => {
  if (!isGrokAccount.value) return true
  return grokTestMode.value === 'text' || grokTestMode.value === 'image' || grokTestMode.value === 'video'
})

const modelOptionsForMode = computed(() => {
  if (!isGrokAccount.value) return availableModels.value
  if (grokTestMode.value === 'image') {
    return availableModels.value.filter((m) => isGrokImageModel(m.id))
  }
  if (grokTestMode.value === 'video') {
    return availableModels.value.filter((m) => isGrokVideoModel(m.id))
  }
  if (grokTestMode.value === 'text') {
    return availableModels.value.filter((m) => isGrokTextModel(m.id))
  }
  return []
})

const supportsPromptInput = computed(() => {
  if (!isGrokAccount.value) {
    return supportsImageTest.value
  }
  return (
    grokTestMode.value === 'image' ||
    grokTestMode.value === 'video' ||
    grokTestMode.value === 'search' ||
    grokTestMode.value === 'tts'
  )
})

const supportsImageUpload = computed(
  () => isGrokAccount.value && (grokTestMode.value === 'image' || grokTestMode.value === 'video')
)
const supportsAudioUpload = computed(() => isGrokAccount.value && grokTestMode.value === 'stt')
const imageUploadLabel = computed(() =>
  grokTestMode.value === 'video'
    ? t('admin.accounts.grok.videoFirstFrameLabel')
    : t('admin.accounts.grok.imageUploadLabel')
)

const imageUploadHint = computed(() =>
  grokTestMode.value === 'video'
    ? t('admin.accounts.grok.videoFirstFrameHint')
    : t('admin.accounts.grok.imageUploadHint')
)

const readFileAsDataURL = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error(t('admin.accounts.grok.fileReadFailed')))
    reader.readAsDataURL(file)
  })

const onImageFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    uploadImageDataURL.value = ''
    uploadImagePreview.value = ''
    uploadImageName.value = ''
    return
  }
  if (file.size > 6 * 1024 * 1024) {
    errorMessage.value = t('admin.accounts.grok.mediaTooLarge')
    status.value = 'error'
    input.value = ''
    return
  }
  try {
    const dataURL = await readFileAsDataURL(file)
    uploadImageDataURL.value = dataURL
    uploadImagePreview.value = dataURL
    uploadImageName.value = file.name
  } catch {
    uploadImageDataURL.value = ''
    uploadImagePreview.value = ''
    uploadImageName.value = ''
    errorMessage.value = t('admin.accounts.grok.fileReadFailed')
    status.value = 'error'
  }
}

const onAudioFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    uploadAudioDataURL.value = ''
    uploadAudioName.value = ''
    return
  }
  if (file.size > 6 * 1024 * 1024) {
    errorMessage.value = t('admin.accounts.grok.mediaTooLarge')
    status.value = 'error'
    input.value = ''
    return
  }
  try {
    uploadAudioDataURL.value = await readFileAsDataURL(file)
    uploadAudioName.value = file.name
  } catch {
    uploadAudioDataURL.value = ''
    uploadAudioName.value = ''
    errorMessage.value = t('admin.accounts.grok.fileReadFailed')
    status.value = 'error'
  }
}

const clearMediaUploads = () => {
  uploadImageDataURL.value = ''
  uploadImagePreview.value = ''
  uploadImageName.value = ''
  uploadAudioDataURL.value = ''
  uploadAudioName.value = ''
  if (imageFileInput.value) imageFileInput.value.value = ''
  if (audioFileInput.value) audioFileInput.value.value = ''
}

const promptInputLabel = computed(() => {
  if (supportsGrokVideoTest.value || grokTestMode.value === 'video') {
    return t('admin.accounts.videoPromptLabel')
  }
  if (supportsImageTest.value || grokTestMode.value === 'image') {
    return t('admin.accounts.imagePromptLabel')
  }
  if (grokTestMode.value === 'search') {
    return t('admin.accounts.grok.searchQueryLabel')
  }
  if (grokTestMode.value === 'tts') {
    return t('admin.accounts.grok.ttsTextLabel')
  }
  return t('admin.accounts.imagePromptLabel')
})

const promptInputPlaceholder = computed(() => {
  if (grokTestMode.value === 'video') {
    return t('admin.accounts.videoPromptPlaceholder')
  }
  if (grokTestMode.value === 'image' || supportsImageTest.value) {
    return t('admin.accounts.imagePromptPlaceholder')
  }
  if (grokTestMode.value === 'search') {
    return t('admin.accounts.grok.searchQueryPlaceholder')
  }
  if (grokTestMode.value === 'tts') {
    return t('admin.accounts.grok.ttsTextPlaceholder')
  }
  return ''
})

const promptInputHint = computed(() => {
  if (grokTestMode.value === 'video') {
    return t('admin.accounts.videoTestHint')
  }
  if (grokTestMode.value === 'image' || supportsImageTest.value) {
    return t('admin.accounts.imageTestHint')
  }
  if (grokTestMode.value === 'search') {
    return t('admin.accounts.grok.searchTestHint')
  }
  if (grokTestMode.value === 'tts') {
    return t('admin.accounts.grok.ttsTestHint')
  }
  if (grokTestMode.value === 'stt') {
    return t('admin.accounts.grok.sttTestHint')
  }
  if (grokTestMode.value === 'realtime') {
    return t('admin.accounts.grok.realtimeTestHint')
  }
  return ''
})

const testModeSummary = computed(() => {
  if (isGrokAccount.value) {
    switch (grokTestMode.value) {
      case 'video':
        return t('admin.accounts.videoTestMode')
      case 'image':
        return t('admin.accounts.imageTestMode')
      case 'search':
        return t('admin.accounts.grok.searchTestMode')
      case 'tts':
        return t('admin.accounts.grok.ttsTestMode')
      case 'stt':
        return t('admin.accounts.grok.sttTestMode')
      case 'realtime':
        return t('admin.accounts.grok.realtimeTestMode')
      default:
        return t('admin.accounts.grok.textTestMode')
    }
  }
  if (supportsImageTest.value) return t('admin.accounts.imageTestMode')
  return t('admin.accounts.testPrompt')
})

const canStartTest = computed(() => {
  if (status.value === 'connecting') return false
  if (isGrokAccount.value) {
    if (
      grokTestMode.value === 'search' ||
      grokTestMode.value === 'tts' ||
      grokTestMode.value === 'stt' ||
      grokTestMode.value === 'realtime'
    ) {
      return true // standalone modes (prompt/model optional)
    }
    return Boolean(selectedModelId.value)
  }
  return Boolean(selectedModelId.value)
})

// Load available models when modal opens
const applyDefaultPromptForMode = () => {
  if (!supportsPromptInput.value) return
  if (testPrompt.value.trim()) return
  if (grokTestMode.value === 'video') {
    testPrompt.value = t('admin.accounts.videoPromptDefault')
  } else if (grokTestMode.value === 'image' || supportsImageTest.value) {
    testPrompt.value = t('admin.accounts.imagePromptDefault')
  } else if (grokTestMode.value === 'search') {
    testPrompt.value = t('admin.accounts.grok.searchQueryDefault')
  } else if (grokTestMode.value === 'tts') {
    testPrompt.value = t('admin.accounts.grok.ttsTextDefault')
  }
}

const pickDefaultModelForMode = () => {
  const opts = modelOptionsForMode.value
  if (!opts.length) {
    selectedModelId.value = ''
    return
  }
  if (opts.some((m) => m.id === selectedModelId.value)) return
  if (grokTestMode.value === 'text') {
    const preferred =
      opts.find((m) => m.id.includes('grok-4.5')) ||
      opts.find((m) => m.id === 'grok') ||
      opts[0]
    selectedModelId.value = preferred.id
    return
  }
  selectedModelId.value = opts[0].id
}

watch(
  () => props.show,
  async (newVal) => {
    if (newVal && props.account) {
      testPrompt.value = ''
      testMode.value = 'default'
      grokTestMode.value = 'text'
      resetState()
      await loadAvailableModels()
      if (isGrokAccount.value) {
        pickDefaultModelForMode()
        applyDefaultPromptForMode()
      }
    } else {
      abortStream()
    }
  }
)

watch(grokTestMode, () => {
  if (!isGrokAccount.value) return
  testPrompt.value = ''
  clearMediaUploads()
  pickDefaultModelForMode()
  applyDefaultPromptForMode()
})

const loadAvailableModels = async () => {
  if (!props.account) return

  loadingModels.value = true
  selectedModelId.value = '' // Reset selection before loading
  try {
    const models = await adminAPI.accounts.getAvailableModels(props.account.id)
    availableModels.value = models
    // Default selection by platform
    if (availableModels.value.length > 0) {
      // Try to select Sonnet as default, otherwise use first model.
      const sonnetModel = availableModels.value.find((m) => m.id.includes('sonnet'))
      selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
    }
  } catch (error) {
    console.error('Failed to load available models:', error)
    // Fallback to empty list
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  generatedAudios.value = []
  generatedVideos.value = []
  previewImageUrl.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

const addLine = (text: string, className: string = 'status-text--neutral') => {
  outputLines.value.push({ text, class: className })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

const startTest = async () => {
  if (!props.account || !canStartTest.value) return

  resetState()
  status.value = 'connecting'
  addLine(t('admin.accounts.startingTestForAccount', { name: props.account.name }), 'status-text--info')
  addLine(t('admin.accounts.testAccountTypeLabel', { type: props.account.type }), 'status-text--neutral')
  if (isGrokAccount.value) {
    const modeLabel =
      grokTestModeOptions.value.find((o) => o.value === grokTestMode.value)?.label || grokTestMode.value
    addLine(t('admin.accounts.grok.selectedTestMode', { mode: modeLabel }), 'status-text--neutral')
  }
  addLine('', 'status-text--neutral')

  abortStream()

  abortController = new AbortController()

  try {
    const requestBody: {
      model_id: string
      prompt: string
      mode?: string
      image_data_url?: string
      audio_data_url?: string
    } = {
      model_id: showModelSelect.value ? selectedModelId.value : '',
      prompt: supportsPromptInput.value ? testPrompt.value.trim() : ''
    }
    if (isOpenAIAccount.value) {
      requestBody.mode = testMode.value
    }
    if (isGrokAccount.value) {
      // Always send explicit Grok mode. search/tts/stt/realtime are standalone
      // endpoints (no free-form model select). text/image/video use optional model.
      requestBody.mode = grokTestMode.value
      if (
        grokTestMode.value === 'search' ||
        grokTestMode.value === 'tts' ||
        grokTestMode.value === 'stt' ||
        grokTestMode.value === 'realtime'
      ) {
        requestBody.model_id = ''
      }
      if (uploadImageDataURL.value && (grokTestMode.value === 'image' || grokTestMode.value === 'video')) {
        requestBody.image_data_url = uploadImageDataURL.value
      }
      if (uploadAudioDataURL.value && grokTestMode.value === 'stt') {
        requestBody.audio_data_url = uploadAudioDataURL.value
      }
    }

    // Use the configured API base; EventSource does not support POST.
    const url = buildApiUrl(`/admin/accounts/${props.account.id}/test`)

    // Use fetch with streaming for SSE since EventSource doesn't support POST
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json',
        [ADMIN_UI_REQUEST_HEADER]: '1'
      },
      body: JSON.stringify(requestBody),
      signal: abortController.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error(t('admin.accounts.grok.noResponseBody'))
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const jsonStr = line.slice(6).trim()
          if (jsonStr) {
            try {
              const event = JSON.parse(jsonStr)
              handleEvent(event)
            } catch (e) {
              console.error('Failed to parse SSE event:', e)
            }
          }
        }
      }
    }
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      status.value = 'idle'
      return
    }
    status.value = 'error'
    const msg = error instanceof Error ? error.message : t('common.unknownError')
    errorMessage.value = msg
    addLine(t('admin.accounts.errorPrefix', { message: msg }), 'status-text--danger')
  }
}

const handleEvent = (event: {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  audio_url?: string
  video_url?: string
  mime_type?: string
}) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'status-text--success')
      if (event.model) {
        addLine(t('admin.accounts.usingModel', { model: event.model }), 'status-text--info')
      }
      addLine(
        isGrokAccount.value
          ? grokTestMode.value === 'video'
            ? t('admin.accounts.sendingVideoRequest')
            : grokTestMode.value === 'image'
              ? t('admin.accounts.sendingImageRequest')
              : grokTestMode.value === 'search'
                ? t('admin.accounts.grok.sendingSearchRequest')
                : grokTestMode.value === 'tts'
                  ? t('admin.accounts.grok.sendingTTSRequest')
                  : grokTestMode.value === 'stt'
                    ? t('admin.accounts.grok.sendingSTTRequest')
                    : grokTestMode.value === 'realtime'
                      ? t('admin.accounts.grok.sendingRealtimeRequest')
                      : t('admin.accounts.sendingTestMessage')
          : supportsImageTest.value
            ? t('admin.accounts.sendingImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'status-text--neutral'
      )
      addLine('', 'status-text--neutral')
      addLine(t('admin.accounts.response'), 'status-text--warning')
      break

    case 'content':
      if (event.text) {
        streamingContent.value += event.text
        scrollToBottom()
      }
      break

    case 'image':
      if (event.image_url) {
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.imageReceived', { count: generatedImages.value.length }), 'status-text--accent')
      }
      break

    case 'audio':
      if (event.audio_url) {
        generatedAudios.value.push({
          url: event.audio_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.audioReceived', { count: generatedAudios.value.length }), 'status-text--accent')
      }
      break

    case 'video':
      if (event.video_url) {
        generatedVideos.value.push({
          url: event.video_url,
          mimeType: event.mime_type
        })
        addLine(t('admin.accounts.videoReceived', { count: generatedVideos.value.length }), 'status-text--accent')
      }
      break

    case 'status':
      if (event.text) {
        addLine(event.text, 'status-text--info')
      }
      break

    case 'test_complete':
      // Move streaming content to output lines
      if (streamingContent.value) {
        addLine(streamingContent.value, 'status-text--success')
        streamingContent.value = ''
      }
      if (event.success) {
        status.value = 'success'
      } else {
        status.value = 'error'
        errorMessage.value = event.error || t('admin.accounts.testFailed')
      }
      break

    case 'error':
      status.value = 'error'
      errorMessage.value = event.error || t('common.unknownError')
      if (streamingContent.value) {
        addLine(streamingContent.value, 'status-text--success')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const text = outputLines.value.map((l) => l.text).join('\n')
  copyToClipboard(text, t('admin.accounts.outputCopied'))
}
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
