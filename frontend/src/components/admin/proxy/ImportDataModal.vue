<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-proxy-data-form" class="components-admin-proxy-import-data-modal__form" @submit.prevent="handleImport">
      <div class="components-admin-proxy-import-data-modal__panel">
        {{ t('admin.proxies.dataImportHint') }}
      </div>
      <div
        class="components-admin-proxy-import-data-modal__panel-2"
      >
        {{ t('admin.proxies.dataImportWarning') }}
      </div>

      <div>
        <label class="input-label">{{ t('admin.proxies.dataImportFile') }}</label>
        <div
          class="components-admin-proxy-import-data-modal__panel-3"
        >
          <div class="components-admin-proxy-import-data-modal__panel-4">
            <div class="components-admin-proxy-import-data-modal__panel-5">
              {{ fileName || t('admin.proxies.dataImportSelectFile') }}
            </div>
            <div class="components-admin-proxy-import-data-modal__panel-6">JSON (.json)</div>
          </div>
          <button type="button" class="components-admin-proxy-import-data-modal__action btn btn-secondary" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="components-admin-proxy-import-data-modal__field"
          accept="application/json,.json"
          @change="handleFileChange"
        />
      </div>

      <div
        v-if="result"
        class="components-admin-proxy-import-data-modal__panel-7"
      >
        <div class="components-admin-proxy-import-data-modal__panel-8">
          {{ t('admin.proxies.dataImportResult') }}
        </div>
        <div class="components-admin-proxy-import-data-modal__panel-9">
          {{ t('admin.proxies.dataImportResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="components-admin-proxy-import-data-modal__panel-10">
          <div class="components-admin-proxy-import-data-modal__panel-11">
            {{ t('admin.proxies.dataImportErrors') }}
          </div>
          <div
            class="components-admin-proxy-import-data-modal__panel-12"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="components-admin-proxy-import-data-modal__panel-13">
              {{ item.kind }} {{ item.name || item.proxy_key || '-' }} — {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="components-admin-proxy-import-data-modal__panel-14">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-proxy-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.proxies.dataImporting') : t('admin.proxies.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult } from '@/types'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const file = ref<File | null>(null)
const result = ref<AdminDataImportResult | null>(null)

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = computed(() => file.value?.name || '')

const errorItems = computed(() => result.value?.errors || [])

watch(
  () => props.show,
  (open) => {
    if (open) {
      file.value = null
      result.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] || null
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error(t('common.fileReadFailed')))
    reader.readAsText(sourceFile)
  })
}

const handleImport = async () => {
  if (!file.value) {
    appStore.showError(t('admin.proxies.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const text = await readFileAsText(file.value)
    const dataPayload = JSON.parse(text)

    const res = await adminAPI.proxies.importData({ data: dataPayload })

    result.value = res

    const msgParams: Record<string, unknown> = {
      proxy_created: res.proxy_created,
      proxy_reused: res.proxy_reused,
      proxy_failed: res.proxy_failed
    }

    if (res.proxy_failed > 0) {
      appStore.showError(t('admin.proxies.dataImportCompletedWithErrors', msgParams))
    } else {
      appStore.showSuccess(t('admin.proxies.dataImportSuccess', msgParams))
      emit('imported')
    }
  } catch (error: any) {
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.proxies.dataImportParseFailed'))
    } else {
      appStore.showError(error?.message || t('admin.proxies.dataImportFailed'))
    }
  } finally {
    importing.value = false
  }
}
</script>
