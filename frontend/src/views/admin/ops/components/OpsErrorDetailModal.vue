<template>
  <BaseDialog :show="show" :title="title" width="full" :close-on-click-outside="true" @close="close">
    <div v-if="loading" class="views-admin-ops-components-ops-error-detail-modal__panel">
      <div class="views-admin-ops-components-ops-error-detail-modal__panel-2">
        <div class="views-admin-ops-components-ops-error-detail-modal__panel-3"></div>
        <div class="views-admin-ops-components-ops-error-detail-modal__panel-4">{{ t('admin.ops.errorDetail.loading') }}</div>
      </div>
    </div>

    <div v-else-if="!detail" class="views-admin-ops-components-ops-error-detail-modal__panel-5">
      {{ emptyText }}
    </div>

    <div v-else class="views-admin-ops-components-ops-error-detail-modal__panel-6 card-body">
      <!-- Summary -->
      <div class="views-admin-ops-components-ops-error-detail-modal__panel-7">
        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.requestId') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-10">
            {{ requestId || '—' }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.time') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            {{ formatDateTime(detail.created_at) }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">
            {{ isUpstreamError(detail) ? t('admin.ops.errorDetail.account') : t('admin.ops.errorDetail.user') }}
          </div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            <template v-if="isUpstreamError(detail)">
              {{ detail.account_name || (detail.account_id != null ? String(detail.account_id) : '—') }}
            </template>
            <template v-else>
              {{ detail.user_email || (detail.user_id != null ? String(detail.user_id) : '—') }}
            </template>
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.platform') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            {{ detail.platform || '—' }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.group') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            {{ detail.group_name || (detail.group_id != null ? String(detail.group_id) : '—') }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.model') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            <template v-if="hasModelMapping(detail)">
              <span class="views-admin-ops-components-ops-error-detail-modal__text">{{ detail.requested_model }}</span>
              <span class="views-admin-ops-components-ops-error-detail-modal__text-2">→</span>
              <span class="views-admin-ops-components-ops-error-detail-modal__text-3">{{ detail.upstream_model }}</span>
            </template>
            <template v-else>
              {{ displayModel(detail) || '—' }}
            </template>
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.inboundEndpoint') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-10">
            {{ detail.inbound_endpoint || '—' }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.upstreamEndpoint') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-10">
            {{ detail.upstream_endpoint || '—' }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.status') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-12">
            <span :class="['views-admin-ops-components-ops-error-detail-modal__text-7', statusClass]">
              {{ detail.status_code }}
            </span>
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.requestType') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-11">
            {{ formatRequestTypeLabel(detail.request_type) }}
          </div>
        </div>

        <div class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.message') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-13" :title="detail.message">
            {{ detail.message || '—' }}
          </div>
        </div>

        <div v-if="detail.api_key_prefix" class="views-admin-ops-components-ops-error-detail-modal__panel-8">
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-9">{{ t('admin.ops.errorDetail.apiKeyPrefix') }}</div>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-14">
            {{ detail.api_key_prefix }}
          </div>
        </div>

      </div>

      <!-- Response content (client request -> error_body; upstream -> upstream_error_detail/message) -->
      <div class="views-admin-ops-components-ops-error-detail-modal__panel-15 card-body">
        <h3 class="views-admin-ops-components-ops-error-detail-modal__heading">{{ t('admin.ops.errorDetail.responseBody') }}</h3>
        <pre class="views-admin-ops-components-ops-error-detail-modal__pre"><code>{{ prettyJSON(primaryResponseBody || '') }}</code></pre>
      </div>

      <!-- Upstream errors list (only for request errors) -->
      <div v-if="showUpstreamList" class="views-admin-ops-components-ops-error-detail-modal__panel-15 card-body">
        <div class="views-admin-ops-components-ops-error-detail-modal__panel-16">
          <h3 class="views-admin-ops-components-ops-error-detail-modal__heading">{{ t('admin.ops.errorDetails.upstreamErrors') }}</h3>
          <div class="views-admin-ops-components-ops-error-detail-modal__panel-17" v-if="correlatedUpstreamLoading">{{ t('common.loading') }}</div>
        </div>

        <div v-if="!correlatedUpstreamLoading && !correlatedUpstreamErrors.length" class="views-admin-ops-components-ops-error-detail-modal__panel-18">
          {{ t('common.noData') }}
        </div>

        <div v-else class="views-admin-ops-components-ops-error-detail-modal__panel-19">
          <div
            v-for="(ev, idx) in correlatedUpstreamErrors"
            :key="ev.id"
            class="views-admin-ops-components-ops-error-detail-modal__panel-20"
          >
            <div class="views-admin-ops-components-ops-error-detail-modal__panel-16">
              <div class="views-admin-ops-components-ops-error-detail-modal__panel-21">
                #{{ idx + 1 }}
                <span v-if="ev.type" class="views-admin-ops-components-ops-error-detail-modal__text-4">{{ ev.type }}</span>
              </div>
              <div class="views-admin-ops-components-ops-error-detail-modal__panel-22">
                <div class="views-admin-ops-components-ops-error-detail-modal__panel-23">
                  {{ ev.status_code ?? '—' }}
                </div>
                <button
                  type="button"
                  class="views-admin-ops-components-ops-error-detail-modal__action"
                  :disabled="!getUpstreamResponsePreview(ev)"
                  :title="getUpstreamResponsePreview(ev) ? '' : t('common.noData')"
                  @click="toggleUpstreamDetail(ev.id)"
                >
                  <Icon
                    :name="expandedUpstreamDetailIds.has(ev.id) ? 'chevronDown' : 'chevronRight'"
                    size="xs"
                    :stroke-width="2"
                  />
                  <span>
                    {{
                      expandedUpstreamDetailIds.has(ev.id)
                        ? t('admin.ops.errorDetail.responsePreview.collapse')
                        : t('admin.ops.errorDetail.responsePreview.expand')
                    }}
                  </span>
                </button>
              </div>
            </div>

            <div class="views-admin-ops-components-ops-error-detail-modal__panel-24">
              <div>
                <span class="views-admin-ops-components-ops-error-detail-modal__text-5">{{ t('admin.ops.errorDetail.upstreamEvent.status') }}:</span>
                <span class="views-admin-ops-components-ops-error-detail-modal__text-6">{{ ev.status_code ?? '—' }}</span>
              </div>
              <div>
                <span class="views-admin-ops-components-ops-error-detail-modal__text-5">{{ t('admin.ops.errorDetail.upstreamEvent.requestId') }}:</span>
                <span class="views-admin-ops-components-ops-error-detail-modal__text-6">{{ ev.request_id || ev.client_request_id || '—' }}</span>
              </div>
            </div>

            <div v-if="ev.message" class="views-admin-ops-components-ops-error-detail-modal__panel-25">{{ ev.message }}</div>

            <pre
              v-if="expandedUpstreamDetailIds.has(ev.id)"
              class="views-admin-ops-components-ops-error-detail-modal__pre-2"
            ><code>{{ prettyJSON(getUpstreamResponsePreview(ev)) }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsErrorDetail } from '@/api/admin/ops'
import { formatDateTime } from '@/utils/format'
import { resolvePrimaryResponseBody, resolveUpstreamPayload } from '../utils/errorDetailResponse'

interface Props {
  show: boolean
  errorId: string | null
  errorType?: 'request' | 'upstream'
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const detail = ref<OpsErrorDetail | null>(null)

const showUpstreamList = computed(() => props.errorType === 'request')

const requestId = computed(() => detail.value?.request_id || detail.value?.client_request_id || '')

const primaryResponseBody = computed(() => {
  return resolvePrimaryResponseBody(detail.value, props.errorType)
})




const title = computed(() => {
  if (!props.errorId) return t('admin.ops.errorDetail.title')
  return t('admin.ops.errorDetail.titleWithId', { id: String(props.errorId) })
})

const emptyText = computed(() => t('admin.ops.errorDetail.noErrorSelected'))

function isUpstreamError(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const phase = String(d.phase || '').toLowerCase()
  const owner = String(d.error_owner || '').toLowerCase()
  return phase === 'upstream' && owner === 'provider'
}

function formatRequestTypeLabel(type: number | null | undefined): string {
  switch (type) {
    case 1: return t('admin.ops.errorDetail.requestTypeSync')
    case 2: return t('admin.ops.errorDetail.requestTypeStream')
    case 3: return t('admin.ops.errorDetail.requestTypeWs')
    default: return t('admin.ops.errorDetail.requestTypeUnknown')
  }
}

function hasModelMapping(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const requested = String(d.requested_model || '').trim()
  const upstream = String(d.upstream_model || '').trim()
  return !!requested && !!upstream && requested !== upstream
}

function displayModel(d: OpsErrorDetail | null): string {
  if (!d) return ''
  const upstream = String(d.upstream_model || '').trim()
  if (upstream) return upstream
  const requested = String(d.requested_model || '').trim()
  if (requested) return requested
  return String(d.model || '').trim()
}

const correlatedUpstream = ref<OpsErrorDetail[]>([])
const correlatedUpstreamLoading = ref(false)

const correlatedUpstreamErrors = computed<OpsErrorDetail[]>(() => correlatedUpstream.value)

const expandedUpstreamDetailIds = ref(new Set<string>())

function getUpstreamResponsePreview(ev: OpsErrorDetail): string {
  const upstreamPayload = resolveUpstreamPayload(ev)
  if (upstreamPayload) return upstreamPayload
  return String(ev.error_body || '').trim()
}

function toggleUpstreamDetail(id: string) {
  const next = new Set(expandedUpstreamDetailIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedUpstreamDetailIds.value = next
}

async function fetchCorrelatedUpstreamErrors(requestErrorId: string) {
  correlatedUpstreamLoading.value = true
  try {
    const res = await opsAPI.listRequestErrorUpstreamErrors(
      requestErrorId,
      { page: 1, page_size: 100, view: 'all' },
      { include_detail: true }
    )
    correlatedUpstream.value = res.items || []
  } catch (err) {
    console.error('[OpsErrorDetailModal] Failed to load correlated upstream errors', err)
    correlatedUpstream.value = []
  } finally {
    correlatedUpstreamLoading.value = false
  }
}

function close() {
  emit('update:show', false)
}

function prettyJSON(raw?: string): string {
  if (!raw) return 'N/A'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function fetchDetail(id: string) {
  loading.value = true
  try {
    const kind = props.errorType || (detail.value?.phase === 'upstream' ? 'upstream' : 'request')
    const d = kind === 'upstream' ? await opsAPI.getUpstreamErrorDetail(id) : await opsAPI.getRequestErrorDetail(id)
    detail.value = d
  } catch (err: any) {
    detail.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadErrorDetail'))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.errorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id) {
      expandedUpstreamDetailIds.value = new Set()
      fetchDetail(id)
      if (props.errorType === 'request') {
        fetchCorrelatedUpstreamErrors(id)
      } else {
        correlatedUpstream.value = []
      }
    }
  },
  { immediate: true }
)

const statusClass = computed(() => {
  const code = detail.value?.status_code ?? 0
  if (code >= 500) return 'views-admin-ops-components-ops-error-detail-modal__state'
  if (code === 429) return 'views-admin-ops-components-ops-error-detail-modal__state-2'
  if (code >= 400) return 'views-admin-ops-components-ops-error-detail-modal__state-3'
  return 'views-admin-ops-components-ops-error-detail-modal__state-4'
})

</script>
