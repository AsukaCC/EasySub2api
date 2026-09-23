<template>
  <div>
    <!-- Tags display -->
    <div class="components-admin-channel-model-tag-input__panel">
      <span
        v-for="(model, idx) in models"
        :key="idx"
        class="components-admin-channel-model-tag-input__text"
        :class="getPlatformTagClass(props.platform || '')"
      >
        {{ model }}
        <button
          type="button"
          @click="removeModel(idx)"
          class="components-admin-channel-model-tag-input__action"
        >
          <Icon name="x" size="xs" />
        </button>
      </span>
      <input
        ref="inputRef"
        v-model="inputValue"
        type="text"
        class="components-admin-channel-model-tag-input__field"
        :placeholder="models.length === 0 ? placeholder : ''"
        @keydown.enter="handleEnter"
        @keydown.tab="handleTab"
        @keydown.backspace="handleBackspace"
        @paste="handlePaste"
        @blur="addModel"
      />
    </div>
    <p class="components-admin-channel-model-tag-input__description">
      {{ t('admin.channels.form.modelInputHint', 'Press Enter to add, supports paste for batch import.') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getPlatformTagClass } from './types'

const { t } = useI18n()

const props = defineProps<{
  models: string[]
  placeholder?: string
  platform?: string
}>()

const emit = defineEmits<{
  'update:models': [models: string[]]
}>()

const inputValue = ref('')
const inputRef = ref<HTMLInputElement>()

function addModel() {
  const val = inputValue.value.trim()
  if (!val) return
  if (!props.models.includes(val)) {
    emit('update:models', [...props.models, val])
  }
  inputValue.value = ''
}

function removeModel(idx: number) {
  const newModels = [...props.models]
  newModels.splice(idx, 1)
  emit('update:models', newModels)
}

function handleBackspace(event: KeyboardEvent) {
  if (event.isComposing) return
  if (inputValue.value === '' && props.models.length > 0) {
    removeModel(props.models.length - 1)
  }
}

function handleEnter(event: KeyboardEvent) { if (event.isComposing) return; event.preventDefault(); addModel() }
function handleTab(event: KeyboardEvent) { if (event.isComposing || !inputValue.value.trim()) return; event.preventDefault(); addModel() }

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text') || ''
  const items = text.split(/[,\n;]+/).map(s => s.trim()).filter(Boolean)
  if (items.length === 0) return
  const unique = [...new Set([...props.models, ...items])]
  emit('update:models', unique)
  inputValue.value = ''
}
</script>
