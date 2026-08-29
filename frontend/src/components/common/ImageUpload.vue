<template>
  <div class="components-common-image-upload__panel">
    <!-- Preview Box -->
    <div class="components-common-image-upload__panel-2">
      <div
        class="components-common-image-upload__panel-3"
        :class="[previewSizeClass, { 'components-common-image-upload__panel-6': !!modelValue }]"
      >
        <!-- SVG mode: render inline -->
        <span
          v-if="mode === 'svg' && modelValue"
          class="components-common-image-upload__text"
          :class="innerSizeClass"
          v-html="sanitizedValue"
        ></span>
        <!-- Image mode: show as img -->
        <img
          v-else-if="mode === 'image' && modelValue"
          :src="modelValue"
          alt=""
          class="components-common-image-upload__image"
        />
        <!-- Empty placeholder -->
        <svg
          v-else
          class="components-common-image-upload__icon"
          :class="placeholderSizeClass"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      </div>
    </div>

    <!-- Controls -->
    <div class="components-common-image-upload__panel-4">
      <div class="components-common-image-upload__panel-5">
        <label class="components-common-image-upload__label btn btn-secondary btn-sm">
          <input
            type="file"
            :accept="acceptTypes"
            class="components-common-image-upload__field"
            @change="handleUpload"
          />
          <Icon name="upload" size="sm" class="components-common-image-upload__icon-2" :stroke-width="2" />
          {{ resolvedUploadLabel }}
        </label>
        <button
          v-if="modelValue"
          type="button"
          class="components-common-image-upload__action btn btn-secondary btn-sm"
          @click="$emit('update:modelValue', '')"
        >
          <Icon name="trash" size="sm" class="components-common-image-upload__icon-2" :stroke-width="2" />
          {{ resolvedRemoveLabel }}
        </button>
      </div>
      <p v-if="hint" class="components-common-image-upload__description">{{ hint }}</p>
      <p v-if="error" class="components-common-image-upload__description-2">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeSvg } from '@/utils/sanitize'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  modelValue: string
  mode?: 'image' | 'svg'
  size?: 'sm' | 'md'
  uploadLabel?: string
  removeLabel?: string
  hint?: string
  maxSize?: number // bytes
}>(), {
  mode: 'image',
  size: 'md',
  uploadLabel: '',
  removeLabel: '',
  hint: '',
  maxSize: 300 * 1024,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const error = ref('')

const resolvedUploadLabel = computed(() => props.uploadLabel || t('common.upload'))
const resolvedRemoveLabel = computed(() => props.removeLabel || t('common.remove'))

const acceptTypes = computed(() => props.mode === 'svg' ? '.svg' : 'image/*')

const sanitizedValue = computed(() =>
  props.mode === 'svg' ? sanitizeSvg(props.modelValue ?? '') : ''
)

const previewSizeClass = computed(() => props.size === 'sm' ? 'components-common-image-upload__state' : 'components-common-image-upload__state-2')
const innerSizeClass = computed(() => props.size === 'sm' ? 'components-common-image-upload__state-3' : 'components-common-image-upload__state-4')
const placeholderSizeClass = computed(() => props.size === 'sm' ? 'components-common-image-upload__state-5' : 'components-common-image-upload__state-6')

function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  error.value = ''

  if (!file) return

  if (props.maxSize && file.size > props.maxSize) {
    error.value = t('common.fileTooLargeKb', {
      size: (file.size / 1024).toFixed(1),
      max: (props.maxSize / 1024).toFixed(0)
    })
    input.value = ''
    return
  }

  const reader = new FileReader()
  if (props.mode === 'svg') {
    reader.onload = (e) => {
      const text = e.target?.result as string
      if (text) emit('update:modelValue', text.trim())
    }
    reader.readAsText(file)
  } else {
    if (!file.type.startsWith('image/')) {
      error.value = t('common.selectImageFile')
      input.value = ''
      return
    }
    reader.onload = (e) => {
      emit('update:modelValue', e.target?.result as string)
    }
    reader.readAsDataURL(file)
  }

  reader.onerror = () => {
    error.value = t('common.fileReadFailed')
  }
  input.value = ''
}
</script>
