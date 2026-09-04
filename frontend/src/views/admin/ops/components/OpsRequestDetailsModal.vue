<script setup lang="ts">
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { computed, ref, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsRequestDetailsParams, type OpsRequestDetail } from '@/api/admin/ops'
import { parseTimeRangeMinutes, formatDateTime } from '../utils/opsFormatters'

export interface OpsRequestDetailsPreset {
  title: string
  kind?: OpsRequestDetailsParams['kind']
  sort?: OpsRequestDetailsParams['sort']
  min_duration_ms?: number
  max_duration_ms?: number
}

interface Props {
  modelValue: boolean
  timeRange: string
  preset: OpsRequestDetailsPreset
  platform?: string
  groupId?: string | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'openErrorDetail', errorId: string): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

// 与 DataTable 一致：< 768px 切换为卡片视图，避免宽表在移动端被截断。
const isDesktopViewport = useMediaQuery('(min-width: 768px)')

const loading = ref(false)
const items = ref<OpsRequestDetail[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const close = () => emit('update:modelValue', false)

const rangeLabel = computed(() => {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  if (minutes >= 60) return t('admin.ops.requestDetails.rangeHours', { n: Math.round(minutes / 60) })
  return t('admin.ops.requestDetails.rangeMinutes', { n: minutes })
})

function buildTimeParams(): Pick<OpsRequestDetailsParams, 'start_time' | 'end_time'> {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - minutes * 60 * 1000)
  return {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString()
  }
}

const fetchData = async () => {
  if (!props.modelValue) return
  loading.value = true
  try {
    const params: OpsRequestDetailsParams = {
      ...buildTimeParams(),
      page: page.value,
      page_size: pageSize.value,
      kind: props.preset.kind ?? 'all',
      sort: props.preset.sort ?? 'created_at_desc'
    }

    const platform = (props.platform || '').trim()
    if (platform) params.platform = platform
    if (props.groupId) params.group_id = props.groupId

    if (typeof props.preset.min_duration_ms === 'number') params.min_duration_ms = props.preset.min_duration_ms
    if (typeof props.preset.max_duration_ms === 'number') params.max_duration_ms = props.preset.max_duration_ms

    const res = await opsAPI.listRequestDetails(params)
    items.value = res.items || []
    total.value = res.total || 0
  } catch (e: any) {
    console.error('[OpsRequestDetailsModal] Failed to fetch request details', e)
    appStore.showError(e?.message || t('admin.ops.requestDetails.failedToLoad'))
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      page.value = 1
      pageSize.value = 10
      fetchData()
    }
  }
)

watch(
  () => [
    props.timeRange,
    props.platform,
    props.groupId,
    props.preset.kind,
    props.preset.sort,
    props.preset.min_duration_ms,
    props.preset.max_duration_ms
  ],
  () => {
    if (!props.modelValue) return
    page.value = 1
    fetchData()
  }
)

function handlePageChange(next: number) {
  page.value = next
  fetchData()
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
  fetchData()
}

async function handleCopyRequestId(requestId: string) {
  const ok = await copyToClipboard(requestId, t('admin.ops.requestDetails.requestIdCopied'))
  if (ok) return
  // `useClipboard` already shows toast on failure; this keeps UX consistent with older ops modal.
  appStore.showWarning(t('admin.ops.requestDetails.copyFailed'))
}

function openErrorDetail(errorId: string | null | undefined) {
  if (!errorId) return
  close()
  emit('openErrorDetail', errorId)
}

const kindBadgeClass = (kind: string) => {
  if (kind === 'error') return 'views-admin-ops-components-ops-request-details-modal__state'
  return 'views-admin-ops-components-ops-request-details-modal__state-2'
}
</script>

<template>
  <BaseDialog :show="modelValue" :title="props.preset.title || t('admin.ops.requestDetails.title')" width="full" @close="close">
    <template #default>
      <div class="views-admin-ops-components-ops-request-details-modal__panel">
        <div class="views-admin-ops-components-ops-request-details-modal__panel-2">
          <div class="views-admin-ops-components-ops-request-details-modal__panel-3">
            {{ t('admin.ops.requestDetails.rangeLabel', { range: rangeLabel }) }}
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            @click="fetchData"
          >
            {{ t('common.refresh') }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="views-admin-ops-components-ops-request-details-modal__panel-4">
          <div class="views-admin-ops-components-ops-request-details-modal__panel-5">
            <LoadingSpinner size="sm" color="inherit" decorative />
            <span class="views-admin-ops-components-ops-request-details-modal__text">{{ t('common.loading') }}</span>
          </div>
        </div>

        <!-- Table -->
        <div v-else class="views-admin-ops-components-ops-request-details-modal__panel-6">
          <div v-if="items.length === 0" class="views-admin-ops-components-ops-request-details-modal__panel-7">
            <div class="views-admin-ops-components-ops-request-details-modal__panel-8">{{ t('admin.ops.requestDetails.empty') }}</div>
            <div class="views-admin-ops-components-ops-request-details-modal__panel-9">{{ t('admin.ops.requestDetails.emptyHint') }}</div>
          </div>

          <div v-else class="views-admin-ops-components-ops-request-details-modal__panel-10">
            <div class="views-admin-ops-components-ops-request-details-modal__panel-11">
              <div v-if="!isDesktopViewport" class="views-admin-ops-components-ops-request-details-modal__panel-12">
                <div v-for="(row, idx) in items" :key="idx" class="views-admin-ops-components-ops-request-details-modal__panel-13">
                  <div class="views-admin-ops-components-ops-request-details-modal__panel-14">
                    <span class="views-admin-ops-components-ops-request-details-modal__text-2" :class="kindBadgeClass(row.kind)">
                      {{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}
                    </span>
                    <span class="views-admin-ops-components-ops-request-details-modal__text-3">{{ (row.platform || 'unknown').toUpperCase() }}</span>
                    <span class="views-admin-ops-components-ops-request-details-modal__text-4">{{ formatDateTime(row.created_at) }}</span>
                  </div>
                  <div class="views-admin-ops-components-ops-request-details-modal__panel-15">{{ row.model || '-' }}</div>
                  <div class="views-admin-ops-components-ops-request-details-modal__panel-16">
                    <span>{{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '-' }}</span>
                    <span>{{ row.status_code ?? '-' }}</span>
                  </div>
                  <div v-if="row.request_id" class="views-admin-ops-components-ops-request-details-modal__panel-17">
                    <span class="views-admin-ops-components-ops-request-details-modal__text-5" :title="row.request_id">
                      {{ row.request_id }}
                    </span>
                    <button
                      class="views-admin-ops-components-ops-request-details-modal__action"
                      @click="handleCopyRequestId(row.request_id)"
                    >
                      {{ t('admin.ops.requestDetails.copy') }}
                    </button>
                  </div>
                  <button
                    v-if="row.kind === 'error' && row.error_id"
                    class="views-admin-ops-components-ops-request-details-modal__action-2"
                    @click="openErrorDetail(row.error_id)"
                  >
                    {{ t('admin.ops.requestDetails.viewError') }}
                  </button>
                </div>
              </div>
              <table v-else class="views-admin-ops-components-ops-request-details-modal__table">
                <thead class="views-admin-ops-components-ops-request-details-modal__header">
                <tr>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.time') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.kind') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.platform') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.model') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.duration') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.status') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading">
                    {{ t('admin.ops.requestDetails.table.requestId') }}
                  </th>
                  <th class="views-admin-ops-components-ops-request-details-modal__heading-2">
                    {{ t('admin.ops.requestDetails.table.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody class="views-admin-ops-components-ops-request-details-modal__body">
                <tr v-for="(row, idx) in items" :key="idx" class="views-admin-ops-components-ops-request-details-modal__row">
                  <td class="views-admin-ops-components-ops-request-details-modal__cell">
                    {{ formatDateTime(row.created_at) }}
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell-2">
                    <span class="views-admin-ops-components-ops-request-details-modal__text-2" :class="kindBadgeClass(row.kind)">
                      {{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}
                    </span>
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell-3">
                    {{ (row.platform || 'unknown').toUpperCase() }}
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell-4" :title="row.model || ''">
                    {{ row.model || '-' }}
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell">
                    {{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '-' }}
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell">
                    {{ row.status_code ?? '-' }}
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell-5">
                    <div v-if="row.request_id" class="views-admin-ops-components-ops-request-details-modal__panel-17">
                      <span class="views-admin-ops-components-ops-request-details-modal__text-6" :title="row.request_id">
                        {{ row.request_id }}
                      </span>
                      <button
                        class="views-admin-ops-components-ops-request-details-modal__action-3"
                        @click="handleCopyRequestId(row.request_id)"
                      >
                        {{ t('admin.ops.requestDetails.copy') }}
                      </button>
                    </div>
                    <span v-else class="views-admin-ops-components-ops-request-details-modal__text-7">-</span>
                  </td>
                  <td class="views-admin-ops-components-ops-request-details-modal__cell-6">
                    <button
                      v-if="row.kind === 'error' && row.error_id"
                      class="views-admin-ops-components-ops-request-details-modal__action-4"
                      @click="openErrorDetail(row.error_id)"
                    >
                      {{ t('admin.ops.requestDetails.viewError') }}
                    </button>
                    <span v-else class="views-admin-ops-components-ops-request-details-modal__text-7">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
            </div>

            <Pagination
              :total="total"
              :page="page"
              :page-size="pageSize"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
          </div>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>
