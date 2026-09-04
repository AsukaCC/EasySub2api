<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.tempUnschedulable.statusTitle')"
    width="normal"
    @close="handleClose"
  >
    <div class="components-account-temp-unsched-status-modal__panel">
      <LoadingState v-if="loading" variant="section" size="sm" class="components-account-temp-unsched-status-modal__panel-2" />

      <div v-else-if="!isActive" class="components-account-temp-unsched-status-modal__panel-3">
        {{ t('admin.accounts.tempUnschedulable.notActive') }}
      </div>

      <div v-else class="components-account-temp-unsched-status-modal__panel">
        <div class="components-account-temp-unsched-status-modal__panel-4">
          {{ t('admin.accounts.recoverStateHint') }}
        </div>

        <div class="components-account-temp-unsched-status-modal__panel-5">
          <p class="components-account-temp-unsched-status-modal__description">
            {{ t('admin.accounts.tempUnschedulable.accountName') }}
          </p>
          <p class="components-account-temp-unsched-status-modal__description-2">
            {{ account?.name || '-' }}
          </p>
        </div>

        <div class="components-account-temp-unsched-status-modal__panel-6">
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.triggeredAt') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ triggeredAtText }}
            </p>
          </div>
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.until') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ untilText }}
            </p>
          </div>
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.remaining') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ remainingText }}
            </p>
          </div>
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.errorCode') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ state?.status_code || '-' }}
            </p>
          </div>
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.matchedKeyword') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ state?.matched_keyword || '-' }}
            </p>
          </div>
          <div class="components-account-temp-unsched-status-modal__panel-7">
            <p class="components-account-temp-unsched-status-modal__description">
              {{ t('admin.accounts.tempUnschedulable.ruleOrder') }}
            </p>
            <p class="components-account-temp-unsched-status-modal__description-2">
              {{ ruleIndexDisplay }}
            </p>
          </div>
        </div>

        <div class="components-account-temp-unsched-status-modal__panel-7">
          <p class="components-account-temp-unsched-status-modal__description">
            {{ t('admin.accounts.tempUnschedulable.errorMessage') }}
          </p>
          <div class="components-account-temp-unsched-status-modal__panel-8">
            {{ state?.error_message || '-' }}
          </div>
        </div>

        <div
          v-if="hasThresholdEvidence"
          class="components-account-temp-unsched-status-modal__panel-9"
          data-testid="temp-unsched-trigger-evidence"
        >
          {{ triggerEvidenceText }}
        </div>
      </div>
    </div>

    <template #footer>
      <div class="components-account-temp-unsched-status-modal__panel-10">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.close') }}
        </button>
        <button :aria-busy="resetting"
          type="button"
          class="btn btn-primary"
          :disabled="!isActive || resetting"
          @click="handleReset"
        >
<LoadingSpinner v-if="resetting" size="sm" color="inherit" decorative />
          {{ t('admin.accounts.recoverState') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Account, TempUnschedulableStatus } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  close: []
  reset: [account: Account]
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const resetting = ref(false)
const status = ref<TempUnschedulableStatus | null>(null)

const state = computed(() => status.value?.state || null)

const isActive = computed(() => {
  if (!status.value?.active || !state.value) return false
  return state.value.until_unix * 1000 > Date.now()
})

const ruleIndexDisplay = computed(() => {
  if (!state.value || !state.value.matched_keyword || state.value.rule_index < 0) return '-'
  return state.value.rule_index + 1
})

const hasThresholdEvidence = computed(() => (state.value?.trigger_count || 0) > 1)

const triggerEvidenceText = computed(() => {
  const count = state.value?.trigger_count || 0
  const threshold = state.value?.trigger_threshold || 0
  const minutes = state.value?.trigger_window_minutes || 0
  if (threshold > 0 && minutes > 0) {
    return t('admin.accounts.tempUnschedulable.multipleErrorTrigger', { count, threshold, minutes })
  }
  if (threshold > 0) {
    return t('admin.accounts.tempUnschedulable.multipleErrorTriggerNoWindow', { count, threshold })
  }
  if (minutes > 0) {
    return t('admin.accounts.tempUnschedulable.multipleErrorCountInWindow', { count, minutes })
  }
  return t('admin.accounts.tempUnschedulable.multipleErrorCount', { count })
})

const triggeredAtText = computed(() => {
  if (!state.value?.triggered_at_unix) return '-'
  return formatDateTime(new Date(state.value.triggered_at_unix * 1000))
})

const untilText = computed(() => {
  if (!state.value?.until_unix) return '-'
  return formatDateTime(new Date(state.value.until_unix * 1000))
})

const remainingText = computed(() => {
  if (!state.value) return '-'
  const remainingMs = state.value.until_unix * 1000 - Date.now()
  if (remainingMs <= 0) {
    return t('admin.accounts.tempUnschedulable.expired')
  }
  const minutes = Math.ceil(remainingMs / 60000)
  if (minutes < 60) {
    return t('admin.accounts.tempUnschedulable.remainingMinutes', { minutes })
  }
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  if (rest === 0) {
    return t('admin.accounts.tempUnschedulable.remainingHours', { hours })
  }
  return t('admin.accounts.tempUnschedulable.remainingHoursMinutes', { hours, minutes: rest })
})

const loadStatus = async () => {
  if (!props.account) return
  loading.value = true
  try {
    status.value = await adminAPI.accounts.getTempUnschedulableStatus(props.account.id)
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.tempUnschedulable.failedToLoad'))
    status.value = null
  } finally {
    loading.value = false
  }
}

const handleClose = () => {
  emit('close')
}

const handleReset = async () => {
  if (!props.account) return
  resetting.value = true
  try {
    const updated = await adminAPI.accounts.recoverState(props.account.id)
    appStore.showSuccess(t('admin.accounts.recoverStateSuccess'))
    emit('reset', updated)
    handleClose()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.recoverStateFailed'))
  } finally {
    resetting.value = false
  }
}

watch(
  () => [props.show, props.account?.id],
  ([visible]) => {
    if (visible && props.account) {
      loadStatus()
      return
    }
    status.value = null
  }
)
</script>
