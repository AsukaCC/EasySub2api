<template>
  <BaseDialog :show="show" :title="t('usage.errors.detail.title')" width="wide" @close="emit('update:show', false)">
    <!-- Loading -->
    <LoadingState v-if="loading" variant="section" size="sm" class="components-user-user-error-detail-modal__panel" />

    <!-- Error state -->
    <div v-else-if="loadError" class="components-user-user-error-detail-modal__panel-2">
      {{ t('usage.errors.detail.loadFailed') }}
    </div>

    <!-- Detail content -->
    <div v-else-if="detail" class="components-user-user-error-detail-modal__panel-3">
      <div class="components-user-user-error-detail-modal__panel-4">
        <!-- Time -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.time') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ formatDateTime(detail.created_at) }}</p>
        </div>
        <!-- Model -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.model') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ detail.model || '-' }}</p>
        </div>
        <!-- Endpoint -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.endpoint') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ detail.inbound_endpoint || '-' }}</p>
        </div>
        <!-- Status Code -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.status') }}</span>
          <p class="components-user-user-error-detail-modal__description-2">
            <span class="badge" :class="statusClass(detail.status_code)">{{ detail.status_code || '-' }}</span>
          </p>
        </div>
        <!-- Category -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.category') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ t('usage.errors.categories.' + detail.category) }}</p>
        </div>
        <!-- Platform -->
        <div>
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.platform') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ detail.platform || '-' }}</p>
        </div>
        <!-- Upstream status code -->
        <div v-if="detail.upstream_status_code != null">
          <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.detail.upstreamStatus') }}</span>
          <p class="components-user-user-error-detail-modal__description">{{ detail.upstream_status_code }}</p>
        </div>
      </div>

      <!-- Message -->
      <div v-if="detail.message">
        <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.message') }}</span>
        <p class="components-user-user-error-detail-modal__description-3">{{ detail.message }}</p>
      </div>

      <!-- Error Body -->
      <div v-if="detail.error_body">
        <span class="components-user-user-error-detail-modal__text">{{ t('usage.errors.detail.responseBody') }}</span>
        <pre class="components-user-user-error-detail-modal__pre">{{ detail.error_body }}</pre>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'

import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { getMyErrorDetail } from '@/api/usage'
import { formatDateTime } from '@/utils/format'
import type { UserErrorRequestDetail } from '@/types'

const props = defineProps<{
  show: boolean
  errorId: string | null
}>()

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
}>()

const { t } = useI18n()

const loading = ref(false)
const loadError = ref(false)
const detail = ref<UserErrorRequestDetail | null>(null)

watch(
  () => [props.show, props.errorId] as const,
  ([show, id]) => {
    if (show && id != null) {
      fetchDetail(id)
    } else if (!show) {
      detail.value = null
      loadError.value = false
    }
  }
)

async function fetchDetail(id: string) {
  loading.value = true
  loadError.value = false
  detail.value = null
  try {
    detail.value = await getMyErrorDetail(id)
  } catch (e) {
    console.error('[UserErrorDetailModal] Failed to load error detail:', e)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function statusClass(code: number) {
  if (code >= 500) return 'badge-danger'
  if (code === 429) return 'badge-warning'
  return 'badge-gray'
}
</script>
