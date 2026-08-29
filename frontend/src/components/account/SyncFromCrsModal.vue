<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.syncFromCrsTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <!-- Step 1: Input credentials -->
    <form
      v-if="currentStep === 'input'"
      id="sync-from-crs-form"
      class="components-account-sync-from-crs-modal__form"
      @submit.prevent="handlePreview"
    >
      <div class="components-account-sync-from-crs-modal__panel">
        {{ t('admin.accounts.syncFromCrsDesc') }}
      </div>
      <div
        class="components-account-sync-from-crs-modal__panel-2"
      >
        {{ t('admin.accounts.crsUpdateBehaviorNote') }}
      </div>
      <div
        class="components-account-sync-from-crs-modal__panel-3"
      >
        {{ t('admin.accounts.crsVersionRequirement') }}
      </div>

      <div class="components-account-sync-from-crs-modal__panel-4">
        <div>
          <label for="crs-base-url" class="input-label">{{ t('admin.accounts.crsBaseUrl') }}</label>
          <input
            id="crs-base-url"
            v-model="form.base_url"
            type="text"
            class="input"
            required
            :placeholder="t('admin.accounts.crsBaseUrlPlaceholder')"
          />
        </div>

        <div class="components-account-sync-from-crs-modal__panel-5">
          <div>
            <label for="crs-username" class="input-label">{{ t('admin.accounts.crsUsername') }}</label>
            <input id="crs-username" v-model="form.username" type="text" class="input" required autocomplete="username" />
          </div>
          <div>
            <label for="crs-password" class="input-label">{{ t('admin.accounts.crsPassword') }}</label>
            <input
              id="crs-password"
              v-model="form.password"
              type="password"
              class="input"
              required
              autocomplete="current-password"
            />
          </div>
        </div>

        <label class="components-account-sync-from-crs-modal__label">
          <input
            v-model="form.sync_proxies"
            type="checkbox"
            class="components-account-sync-from-crs-modal__field"
          />
          {{ t('admin.accounts.syncProxies') }}
        </label>
      </div>
    </form>

    <!-- Step 2: Preview & select -->
    <div v-else-if="currentStep === 'preview' && previewResult" class="components-account-sync-from-crs-modal__form">
      <!-- Existing accounts (read-only info) -->
      <div
        v-if="previewResult.existing_accounts.length"
        class="components-account-sync-from-crs-modal__panel-6"
      >
        <div class="components-account-sync-from-crs-modal__panel-7">
          {{ t('admin.accounts.crsExistingAccounts') }}
          <span class="components-account-sync-from-crs-modal__text">({{ previewResult.existing_accounts.length }})</span>
        </div>
        <div class="components-account-sync-from-crs-modal__panel-8">
          <div
            v-for="acc in previewResult.existing_accounts"
            :key="acc.crs_account_id"
            class="components-account-sync-from-crs-modal__panel-9"
          >
            <span
              class="components-account-sync-from-crs-modal__text-2"
            >{{ acc.platform }} / {{ acc.type }}</span>
            <span class="components-account-sync-from-crs-modal__text-3">{{ acc.name }}</span>
          </div>
        </div>
      </div>

      <!-- New accounts (selectable) -->
      <div v-if="previewResult.new_accounts.length">
        <div class="components-account-sync-from-crs-modal__panel-10">
          <div class="components-account-sync-from-crs-modal__panel-11">
            {{ t('admin.accounts.crsNewAccounts') }}
            <span class="components-account-sync-from-crs-modal__text">({{ previewResult.new_accounts.length }})</span>
          </div>
          <div class="components-account-sync-from-crs-modal__panel-12">
            <button
              type="button"
              class="components-account-sync-from-crs-modal__action"
              @click="selectAll"
            >{{ t('admin.accounts.crsSelectAll') }}</button>
            <button
              type="button"
              class="components-account-sync-from-crs-modal__action-2"
              @click="selectNone"
            >{{ t('admin.accounts.crsSelectNone') }}</button>
          </div>
        </div>
        <div
          class="components-account-sync-from-crs-modal__panel-13"
        >
          <label
            v-for="acc in previewResult.new_accounts"
            :key="acc.crs_account_id"
            class="components-account-sync-from-crs-modal__label-2"
          >
            <input
              type="checkbox"
              :checked="selectedIds.has(acc.crs_account_id)"
              class="components-account-sync-from-crs-modal__field"
              @change="toggleSelect(acc.crs_account_id)"
            />
            <span
              class="components-account-sync-from-crs-modal__text-4"
            >{{ acc.platform }} / {{ acc.type }}</span>
            <span class="components-account-sync-from-crs-modal__text-5">{{ acc.name }}</span>
          </label>
        </div>
        <div class="components-account-sync-from-crs-modal__panel-14">
          {{ t('admin.accounts.crsSelectedCount', { count: selectedIds.size }) }}
        </div>
      </div>

      <!-- Sync options summary -->
      <div class="components-account-sync-from-crs-modal__panel-15">
        <span>{{ t('admin.accounts.syncProxies') }}:</span>
        <span :class="form.sync_proxies ? 'components-account-sync-from-crs-modal__text-6' : 'components-account-sync-from-crs-modal__text-7'">
          {{ form.sync_proxies ? t('common.yes') : t('common.no') }}
        </span>
      </div>

      <!-- No new accounts -->
      <div
        v-if="!previewResult.new_accounts.length"
        class="components-account-sync-from-crs-modal__panel-16"
      >
        {{ t('admin.accounts.crsNoNewAccounts') }}
        <span v-if="previewResult.existing_accounts.length">
          {{ t('admin.accounts.crsWillUpdate', { count: previewResult.existing_accounts.length }) }}
        </span>
      </div>
    </div>

    <!-- Step 3: Result -->
    <div v-else-if="currentStep === 'result' && result" class="components-account-sync-from-crs-modal__form">
      <div
        class="components-account-sync-from-crs-modal__panel-17"
      >
        <div class="components-account-sync-from-crs-modal__panel-11">
          {{ t('admin.accounts.syncResult') }}
        </div>
        <div class="components-account-sync-from-crs-modal__panel-18">
          {{ t('admin.accounts.syncResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="components-account-sync-from-crs-modal__panel-19">
          <div class="components-account-sync-from-crs-modal__panel-20">
            {{ t('admin.accounts.syncErrors') }}
          </div>
          <div
            class="components-account-sync-from-crs-modal__panel-21"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="components-account-sync-from-crs-modal__panel-22">
              {{ item.kind }} {{ item.crs_account_id }} — {{ item.action
              }}{{ item.error ? `: ${item.error}` : '' }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="components-account-sync-from-crs-modal__panel-23">
        <!-- Step 1: Input -->
        <template v-if="currentStep === 'input'">
          <button
            class="btn btn-secondary"
            type="button"
            :disabled="previewing"
            @click="handleClose"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="btn btn-primary"
            type="submit"
            form="sync-from-crs-form"
            :disabled="previewing"
          >
            {{ previewing ? t('admin.accounts.crsPreviewing') : t('admin.accounts.crsPreview') }}
          </button>
        </template>

        <!-- Step 2: Preview -->
        <template v-else-if="currentStep === 'preview'">
          <button
            class="btn btn-secondary"
            type="button"
            :disabled="syncing"
            @click="handleBack"
          >
            {{ t('admin.accounts.crsBack') }}
          </button>
          <button
            class="btn btn-primary"
            type="button"
            :disabled="syncing || hasNewButNoneSelected"
            @click="handleSync"
          >
            {{ syncing ? t('admin.accounts.syncing') : t('admin.accounts.syncNow') }}
          </button>
        </template>

        <!-- Step 3: Result -->
        <template v-else-if="currentStep === 'result'">
          <button class="btn btn-secondary" type="button" @click="handleClose">
            {{ t('common.close') }}
          </button>
        </template>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { PreviewFromCRSResult } from '@/api/admin/accounts'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'synced'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

type Step = 'input' | 'preview' | 'result'
const currentStep = ref<Step>('input')
const previewing = ref(false)
const syncing = ref(false)
const previewResult = ref<PreviewFromCRSResult | null>(null)
const selectedIds = ref(new Set<string>())
const result = ref<Awaited<ReturnType<typeof adminAPI.accounts.syncFromCrs>> | null>(null)

const form = reactive({
  base_url: '',
  username: '',
  password: '',
  sync_proxies: true
})

const hasNewButNoneSelected = computed(() => {
  if (!previewResult.value) return false
  return previewResult.value.new_accounts.length > 0 && selectedIds.value.size === 0
})

const errorItems = computed(() => {
  if (!result.value?.items) return []
  return result.value.items.filter(
    (i) => i.action === 'failed' || (i.action === 'skipped' && i.error !== 'not selected')
  )
})

watch(
  () => props.show,
  (open) => {
    if (open) {
      currentStep.value = 'input'
      previewResult.value = null
      selectedIds.value = new Set()
      result.value = null
      form.base_url = ''
      form.username = ''
      form.password = ''
      form.sync_proxies = true
    }
  }
)

const handleClose = () => {
  if (syncing.value || previewing.value) {
    return
  }
  emit('close')
}

const handleBack = () => {
  currentStep.value = 'input'
  previewResult.value = null
  selectedIds.value = new Set()
}

const selectAll = () => {
  if (!previewResult.value) return
  selectedIds.value = new Set(previewResult.value.new_accounts.map((a) => a.crs_account_id))
}

const selectNone = () => {
  selectedIds.value = new Set()
}

const toggleSelect = (id: string) => {
  const s = new Set(selectedIds.value)
  if (s.has(id)) {
    s.delete(id)
  } else {
    s.add(id)
  }
  selectedIds.value = s
}

const handlePreview = async () => {
  if (!form.base_url.trim() || !form.username.trim() || !form.password.trim()) {
    appStore.showError(t('admin.accounts.syncMissingFields'))
    return
  }

  previewing.value = true
  try {
    const res = await adminAPI.accounts.previewFromCrs({
      base_url: form.base_url.trim(),
      username: form.username.trim(),
      password: form.password
    })
    previewResult.value = res
    // Auto-select all new accounts
    selectedIds.value = new Set(res.new_accounts.map((a) => a.crs_account_id))
    currentStep.value = 'preview'
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.crsPreviewFailed'))
  } finally {
    previewing.value = false
  }
}

const handleSync = async () => {
  if (!form.base_url.trim() || !form.username.trim() || !form.password.trim()) {
    appStore.showError(t('admin.accounts.syncMissingFields'))
    return
  }

  syncing.value = true
  try {
    const res = await adminAPI.accounts.syncFromCrs({
      base_url: form.base_url.trim(),
      username: form.username.trim(),
      password: form.password,
      sync_proxies: form.sync_proxies,
      selected_account_ids: [...selectedIds.value]
    })
    result.value = res
    currentStep.value = 'result'

    if (res.failed > 0) {
      appStore.showError(t('admin.accounts.syncCompletedWithErrors', res))
    } else {
      appStore.showSuccess(t('admin.accounts.syncCompleted', res))
    }
    emit('synced')
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.syncFailed'))
  } finally {
    syncing.value = false
  }
}
</script>
