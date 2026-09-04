<script setup lang="ts">
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { opsAPI, type AlertEventsQuery } from '@/api/admin/ops'
import type { AlertEvent } from '../types'
import { formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()
const appStore = useAppStore()

// 与 DataTable 一致：< 768px 切换为卡片视图，避免宽表在移动端被截断。
const isDesktopViewport = useMediaQuery('(min-width: 768px)')

const PAGE_SIZE = 10

const loading = ref(false)
const loadingMore = ref(false)
const events = ref<AlertEvent[]>([])
const hasMore = ref(true)

// Detail modal
const showDetail = ref(false)
const selected = ref<AlertEvent | null>(null)
const detailLoading = ref(false)
const detailActionLoading = ref(false)
const historyLoading = ref(false)
const history = ref<AlertEvent[]>([])
const historyRange = ref('7d')
const historyRangeOptions = computed(() => [
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const silenceDuration = ref('1h')
const silenceDurationOptions = computed(() => [
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') }
])

// Filters
const timeRange = ref('24h')
const timeRangeOptions = computed(() => [
  { value: '5m', label: t('admin.ops.timeRange.5m') },
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '6h', label: t('admin.ops.timeRange.6h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const severity = ref<string>('')
const severityOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'P0', label: 'P0' },
  { value: 'P1', label: 'P1' },
  { value: 'P2', label: 'P2' },
  { value: 'P3', label: 'P3' }
])

const status = ref<string>('')
const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'firing', label: t('admin.ops.alertEvents.status.firing') },
  { value: 'resolved', label: t('admin.ops.alertEvents.status.resolved') },
  { value: 'manual_resolved', label: t('admin.ops.alertEvents.status.manualResolved') }
])

const emailSent = ref<string>('')
const emailSentOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'true', label: t('admin.ops.alertEvents.table.emailSent') },
  { value: 'false', label: t('admin.ops.alertEvents.table.emailIgnored') }
])

function buildQuery(overrides: Partial<AlertEventsQuery> = {}): AlertEventsQuery {
  const q: AlertEventsQuery = {
    limit: PAGE_SIZE,
    time_range: timeRange.value
  }
  if (severity.value) q.severity = severity.value
  if (status.value) q.status = status.value
  if (emailSent.value === 'true') q.email_sent = true
  if (emailSent.value === 'false') q.email_sent = false
  return { ...q, ...overrides }
}

async function loadFirstPage() {
  loading.value = true
  try {
    const data = await opsAPI.listAlertEvents(buildQuery())
    events.value = data
    hasMore.value = data.length === PAGE_SIZE
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to load alert events', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.loadFailed'))
    events.value = []
    hasMore.value = false
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || loading.value) return
  if (!hasMore.value) return
  const last = events.value[events.value.length - 1]
  if (!last) return

  loadingMore.value = true
  try {
    const data = await opsAPI.listAlertEvents(
      buildQuery({ before_fired_at: last.fired_at || last.created_at, before_id: last.id })
    )
    if (!data.length) {
      hasMore.value = false
      return
    }
    events.value = [...events.value, ...data]
    if (data.length < PAGE_SIZE) hasMore.value = false
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to load more alert events', err)
    hasMore.value = false
  } finally {
    loadingMore.value = false
  }
}

function onScroll(e: Event) {
  const el = e.target as HTMLElement | null
  if (!el) return
  const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 120
  if (nearBottom) loadMore()
}

function getDimensionString(event: AlertEvent | null | undefined, key: string): string {
  const v = event?.dimensions?.[key]
  if (v == null) return ''
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  return ''
}

function formatDurationMs(ms: number): string {
  const safe = Math.max(0, Math.floor(ms))
  const sec = Math.floor(safe / 1000)
  if (sec < 60) return `${sec}s`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h`
  const day = Math.floor(hr / 24)
  return `${day}d`
}

function formatDurationLabel(event: AlertEvent): string {
  const firedAt = new Date(event.fired_at || event.created_at)
  if (Number.isNaN(firedAt.getTime())) return '-'
  const resolvedAtStr = event.resolved_at || null
  const status = String(event.status || '').trim().toLowerCase()

  if (resolvedAtStr) {
    const resolvedAt = new Date(resolvedAtStr)
    if (!Number.isNaN(resolvedAt.getTime())) {
      const ms = resolvedAt.getTime() - firedAt.getTime()
      const prefix = status === 'manual_resolved'
        ? t('admin.ops.alertEvents.status.manualResolved')
        : t('admin.ops.alertEvents.status.resolved')
      return `${prefix} ${formatDurationMs(ms)}`
    }
  }

  const now = Date.now()
  const ms = now - firedAt.getTime()
  return `${t('admin.ops.alertEvents.status.firing')} ${formatDurationMs(ms)}`
}

function formatDimensionsSummary(event: AlertEvent): string {
  const parts: string[] = []
  const platform = getDimensionString(event, 'platform')
  if (platform) parts.push(`platform=${platform}`)
  const groupId = event.dimensions?.group_id
  if (groupId != null && groupId !== '') parts.push(`group_id=${String(groupId)}`)
  const region = getDimensionString(event, 'region')
  if (region) parts.push(`region=${region}`)
  return parts.length ? parts.join(' ') : '-'
}

function closeDetail() {
  showDetail.value = false
  selected.value = null
  history.value = []
}

async function openDetail(row: AlertEvent) {
  showDetail.value = true
  selected.value = row
  detailLoading.value = true
  historyLoading.value = true

  try {
    const detail = await opsAPI.getAlertEvent(row.id)
    selected.value = detail
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to load alert detail', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.loadFailed'))
  } finally {
    detailLoading.value = false
  }

  await loadHistory()
}

async function loadHistory() {
  const ev = selected.value
  if (!ev) {
    history.value = []
    historyLoading.value = false
    return
  }

  historyLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'string' && groupIdRaw ? groupIdRaw : undefined

    const items = await opsAPI.listAlertEvents({
      limit: 20,
      time_range: historyRange.value,
      platform: platform || undefined,
      group_id: groupId,
      status: ''
    })

    // Best-effort: narrow to same rule_id + dimensions
    history.value = items.filter((it) => {
      if (it.rule_id !== ev.rule_id) return false
      const p1 = getDimensionString(it, 'platform')
      const p2 = getDimensionString(ev, 'platform')
      if ((p1 || '') !== (p2 || '')) return false
      const g1 = it.dimensions?.group_id
      const g2 = ev.dimensions?.group_id
      return (g1 ?? null) === (g2 ?? null)
    })
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to load alert history', err)
    history.value = []
  } finally {
    historyLoading.value = false
  }
}

function durationToUntilRFC3339(duration: string): string {
  const now = Date.now()
  if (duration === '1h') return new Date(now + 60 * 60 * 1000).toISOString()
  if (duration === '24h') return new Date(now + 24 * 60 * 60 * 1000).toISOString()
  if (duration === '7d') return new Date(now + 7 * 24 * 60 * 60 * 1000).toISOString()
  return new Date(now + 60 * 60 * 1000).toISOString()
}

async function silenceAlert() {
  const ev = selected.value
  if (!ev) return
  if (detailActionLoading.value) return
  detailActionLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'string' && groupIdRaw ? groupIdRaw : null
    const region = getDimensionString(ev, 'region') || null

    await opsAPI.createAlertSilence({
      rule_id: ev.rule_id,
      platform: platform || '',
      group_id: groupId ?? undefined,
      region: region ?? undefined,
      until: durationToUntilRFC3339(silenceDuration.value),
      reason: `silence from UI (${silenceDuration.value})`
    })

    appStore.showSuccess(t('admin.ops.alertEvents.detail.silenceSuccess'))
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to silence alert', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.silenceFailed'))
  } finally {
    detailActionLoading.value = false
  }
}

async function manualResolve() {
  if (!selected.value) return
  if (detailActionLoading.value) return
  detailActionLoading.value = true
  try {
    await opsAPI.updateAlertEventStatus(selected.value.id, 'manual_resolved')
    appStore.showSuccess(t('admin.ops.alertEvents.detail.manualResolvedSuccess'))

    // Refresh detail + first page to reflect new status
    const detail = await opsAPI.getAlertEvent(selected.value.id)
    selected.value = detail
    await loadFirstPage()
    await loadHistory()
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to resolve alert', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.manualResolvedFailed'))
  } finally {
    detailActionLoading.value = false
  }
}

onMounted(() => {
  loadFirstPage()
})

watch([timeRange, severity, status, emailSent], () => {
  events.value = []
  hasMore.value = true
  loadFirstPage()
})

watch(historyRange, () => {
  if (showDetail.value) loadHistory()
})

function severityBadgeClass(severity: string | undefined): string {
  const s = String(severity || '').trim().toLowerCase()
  if (s === 'p0' || s === 'critical') return 'views-admin-ops-components-ops-alert-events-card__state'
  if (s === 'p1' || s === 'warning') return 'views-admin-ops-components-ops-alert-events-card__state-2'
  if (s === 'p2' || s === 'info') return 'views-admin-ops-components-ops-alert-events-card__state-3'
  if (s === 'p3') return 'views-admin-ops-components-ops-alert-events-card__state-4'
  return 'views-admin-ops-components-ops-alert-events-card__state-4'
}

function statusBadgeClass(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (s === 'firing') return 'views-admin-ops-components-ops-alert-events-card__state-5'
  if (s === 'resolved') return 'views-admin-ops-components-ops-alert-events-card__state-6'
  if (s === 'manual_resolved') return 'views-admin-ops-components-ops-alert-events-card__state-7'
  return 'views-admin-ops-components-ops-alert-events-card__state-8'
}

function formatStatusLabel(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (!s) return '-'
  if (s === 'firing') return t('admin.ops.alertEvents.status.firing')
  if (s === 'resolved') return t('admin.ops.alertEvents.status.resolved')
  if (s === 'manual_resolved') return t('admin.ops.alertEvents.status.manualResolved')
  return s.toUpperCase()
}

const empty = computed(() => events.value.length === 0 && !loading.value)
</script>

<template>
  <div class="views-admin-ops-components-ops-alert-events-card__panel card-body">
    <div class="views-admin-ops-components-ops-alert-events-card__panel-2">
      <div>
        <h3 class="views-admin-ops-components-ops-alert-events-card__heading">{{ t('admin.ops.alertEvents.title') }}</h3>
        <p class="views-admin-ops-components-ops-alert-events-card__description">{{ t('admin.ops.alertEvents.description') }}</p>
      </div>

      <div class="views-admin-ops-components-ops-alert-events-card__panel-3">
        <Select :model-value="timeRange" :options="timeRangeOptions" class="views-admin-ops-components-ops-alert-events-card__field" @change="timeRange = String($event || '24h')" />
        <Select :model-value="severity" :options="severityOptions" class="views-admin-ops-components-ops-alert-events-card__field-2" @change="severity = String($event || '')" />
        <Select :model-value="status" :options="statusOptions" class="views-admin-ops-components-ops-alert-events-card__field-3" @change="status = String($event || '')" />
        <Select :model-value="emailSent" :options="emailSentOptions" class="views-admin-ops-components-ops-alert-events-card__field-3" @change="emailSent = String($event || '')" />
        <button
          class="views-admin-ops-components-ops-alert-events-card__action"
          :disabled="loading"
          @click="loadFirstPage"
        >
          <svg class="views-admin-ops-components-ops-alert-events-card__icon" :class="{ 'views-admin-ops-components-ops-alert-events-card__icon-5': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="views-admin-ops-components-ops-alert-events-card__panel-4">
      <LoadingSpinner size="sm" color="inherit" decorative />
      {{ t('admin.ops.alertEvents.loading') }}
    </div>

    <div v-else-if="empty" class="views-admin-ops-components-ops-alert-events-card__panel-5">
      {{ t('admin.ops.alertEvents.empty') }}
    </div>

    <div v-else class="views-admin-ops-components-ops-alert-events-card__panel-6">
      <div class="views-admin-ops-components-ops-alert-events-card__panel-7" @scroll="onScroll">
        <div v-if="!isDesktopViewport" class="views-admin-ops-components-ops-alert-events-card__panel-8">
          <div
            v-for="row in events"
            :key="row.id"
            class="views-admin-ops-components-ops-alert-events-card__panel-9"
            @click="openDetail(row)"
          >
            <div class="views-admin-ops-components-ops-alert-events-card__panel-3">
              <span class="views-admin-ops-components-ops-alert-events-card__text" :class="severityBadgeClass(String(row.severity || ''))">
                {{ row.severity || '-' }}
              </span>
              <span class="views-admin-ops-components-ops-alert-events-card__text-2" :class="statusBadgeClass(row.status)">
                {{ formatStatusLabel(row.status) }}
              </span>
              <span class="views-admin-ops-components-ops-alert-events-card__text-3">
                {{ formatDateTime(row.fired_at || row.created_at) }}
              </span>
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-10">{{ row.title || '-' }}</div>
            <div v-if="row.description" class="views-admin-ops-components-ops-alert-events-card__panel-11">
              {{ row.description }}
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-12">
              <span><span class="views-admin-ops-components-ops-alert-events-card__text-4">#{{ row.rule_id }}</span> · {{ formatDurationLabel(row) }}</span>
              <span class="views-admin-ops-components-ops-alert-events-card__text-5">
                <Icon
                  v-if="row.email_sent"
                  name="checkCircle"
                  size="xs"
                  class="views-admin-ops-components-ops-alert-events-card__icon-3"
                />
                <Icon
                  v-else
                  name="ban"
                  size="xs"
                  class="views-admin-ops-components-ops-alert-events-card__icon-4"
                />
                {{ row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored') }}
              </span>
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-13">{{ formatDimensionsSummary(row) }}</div>
          </div>
        </div>
        <table v-else class="views-admin-ops-components-ops-alert-events-card__table">
          <thead class="views-admin-ops-components-ops-alert-events-card__header">
            <tr>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.time') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.severity') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.platform') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.ruleId') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.title') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.duration') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-2">
                {{ t('admin.ops.alertEvents.table.dimensions') }}
              </th>
              <th class="views-admin-ops-components-ops-alert-events-card__heading-3">
                {{ t('admin.ops.alertEvents.table.email') }}
              </th>
            </tr>
          </thead>
          <tbody class="views-admin-ops-components-ops-alert-events-card__body">
            <tr
              v-for="row in events"
              :key="row.id"
              class="views-admin-ops-components-ops-alert-events-card__row"
              @click="openDetail(row)"
              :title="row.title || ''"
            >
              <td class="views-admin-ops-components-ops-alert-events-card__cell">
                {{ formatDateTime(row.fired_at || row.created_at) }}
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell-2">
                <div class="views-admin-ops-components-ops-alert-events-card__panel-14">
                  <span class="views-admin-ops-components-ops-alert-events-card__text" :class="severityBadgeClass(String(row.severity || ''))">
                    {{ row.severity || '-' }}
                  </span>
                  <span class="views-admin-ops-components-ops-alert-events-card__text-2" :class="statusBadgeClass(row.status)">
                    {{ formatStatusLabel(row.status) }}
                  </span>
                </div>
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell">
                {{ getDimensionString(row, 'platform') || '-' }}
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell">
                <span class="views-admin-ops-components-ops-alert-events-card__text-4">#{{ row.rule_id }}</span>
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell-3">
                <div class="views-admin-ops-components-ops-alert-events-card__panel-15">{{ row.title || '-' }}</div>
                <div v-if="row.description" class="views-admin-ops-components-ops-alert-events-card__panel-16">
                  {{ row.description }}
                </div>
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell">
                {{ formatDurationLabel(row) }}
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell-4">
                {{ formatDimensionsSummary(row) }}
              </td>
              <td class="views-admin-ops-components-ops-alert-events-card__cell-5">
                <span
                  class="views-admin-ops-components-ops-alert-events-card__text-6"
                  :title="row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored')"
                >
                  <Icon
                    v-if="row.email_sent"
                    name="checkCircle"
                    size="sm"
                    class="views-admin-ops-components-ops-alert-events-card__icon-3"
                  />
                  <Icon
                    v-else
                    name="ban"
                    size="sm"
                    class="views-admin-ops-components-ops-alert-events-card__icon-4"
                  />
                  <span class="views-admin-ops-components-ops-alert-events-card__text-7">
                    {{ row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored') }}
                  </span>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="loadingMore" class="views-admin-ops-components-ops-alert-events-card__panel-17">
          <LoadingSpinner size="sm" color="inherit" decorative />
          {{ t('admin.ops.alertEvents.loading') }}
        </div>
        <div v-else-if="!hasMore && events.length > 0" class="views-admin-ops-components-ops-alert-events-card__panel-18">
          -
        </div>
      </div>
    </div>

    <BaseDialog
      :show="showDetail"
      :title="t('admin.ops.alertEvents.detail.title')"
      width="wide"
      :close-on-click-outside="true"
      @close="closeDetail"
    >
      <div v-if="detailLoading" class="views-admin-ops-components-ops-alert-events-card__panel-19">
        {{ t('admin.ops.alertEvents.detail.loading') }}
      </div>

      <div v-else-if="!selected" class="views-admin-ops-components-ops-alert-events-card__panel-20">
        {{ t('admin.ops.alertEvents.detail.empty') }}
      </div>

      <div v-else class="views-admin-ops-components-ops-alert-events-card__panel-21">
        <div class="views-admin-ops-components-ops-alert-events-card__panel-22">
          <div class="views-admin-ops-components-ops-alert-events-card__panel-23">
            <div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-3">
                <span class="views-admin-ops-components-ops-alert-events-card__text-8" :class="severityBadgeClass(String(selected.severity || ''))">
                  {{ selected.severity || '-' }}
                </span>
                <span class="views-admin-ops-components-ops-alert-events-card__text-2" :class="statusBadgeClass(selected.status)">
                  {{ formatStatusLabel(selected.status) }}
                </span>
              </div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-24">
                {{ selected.title || '-' }}
              </div>
              <div v-if="selected.description" class="views-admin-ops-components-ops-alert-events-card__panel-25">
                {{ selected.description }}
              </div>
            </div>

            <div class="views-admin-ops-components-ops-alert-events-card__panel-26">
              <div class="views-admin-ops-components-ops-alert-events-card__panel-27">
                <span class="views-admin-ops-components-ops-alert-events-card__text-7">{{ t('admin.ops.alertEvents.detail.silence') }}</span>
                <Select
                  :model-value="silenceDuration"
                  :options="silenceDurationOptions"
                  class="views-admin-ops-components-ops-alert-events-card__field-3"
                  @change="silenceDuration = String($event || '1h')"
                />
                <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="silenceAlert">
                  <Icon name="ban" size="sm" />
                  {{ t('common.apply') }}
                </button>
              </div>

              <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="manualResolve">
                <Icon name="checkCircle" size="sm" />
                {{ t('admin.ops.alertEvents.detail.manualResolve') }}
              </button>
            </div>
          </div>
        </div>

          <div class="views-admin-ops-components-ops-alert-events-card__panel-28">
            <div class="views-admin-ops-components-ops-alert-events-card__panel-22">
              <div class="views-admin-ops-components-ops-alert-events-card__panel-29">{{ t('admin.ops.alertEvents.detail.firedAt') }}</div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-30">{{ formatDateTime(selected.fired_at || selected.created_at) }}</div>
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-22">
              <div class="views-admin-ops-components-ops-alert-events-card__panel-29">{{ t('admin.ops.alertEvents.detail.resolvedAt') }}</div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-30">{{ selected.resolved_at ? formatDateTime(selected.resolved_at) : '-' }}</div>
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-22">
              <div class="views-admin-ops-components-ops-alert-events-card__panel-29">{{ t('admin.ops.alertEvents.detail.ruleId') }}</div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-31">
                <div class="views-admin-ops-components-ops-alert-events-card__panel-32">#{{ selected.rule_id }}</div>
                <a
                  class="views-admin-ops-components-ops-alert-events-card__link"
                  :href="`/admin/ops?open_alert_rules=1&alert_rule_id=${selected.rule_id}`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewRule') }}
                </a>
                <a
                  class="views-admin-ops-components-ops-alert-events-card__link"
                  :href="`/admin/ops?platform=${encodeURIComponent(getDimensionString(selected,'platform')||'')}&group_id=${selected.dimensions?.group_id || ''}&error_type=request&open_error_details=1`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewLogs') }}
                </a>
              </div>
            </div>
            <div class="views-admin-ops-components-ops-alert-events-card__panel-22">
              <div class="views-admin-ops-components-ops-alert-events-card__panel-29">{{ t('admin.ops.alertEvents.detail.dimensions') }}</div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-33">
                <div v-if="getDimensionString(selected, 'platform')">platform={{ getDimensionString(selected, 'platform') }}</div>
                <div v-if="selected.dimensions?.group_id">group_id={{ selected.dimensions.group_id }}</div>
                <div v-if="getDimensionString(selected, 'region')">region={{ getDimensionString(selected, 'region') }}</div>
              </div>
            </div>
          </div>


        <div class="views-admin-ops-components-ops-alert-events-card__panel-34">
          <div class="views-admin-ops-components-ops-alert-events-card__panel-35">
            <div>
              <div class="views-admin-ops-components-ops-alert-events-card__heading">{{ t('admin.ops.alertEvents.detail.historyTitle') }}</div>
              <div class="views-admin-ops-components-ops-alert-events-card__panel-36">{{ t('admin.ops.alertEvents.detail.historyHint') }}</div>
            </div>
            <Select :model-value="historyRange" :options="historyRangeOptions" class="views-admin-ops-components-ops-alert-events-card__field-4" @change="historyRange = String($event || '7d')" />
          </div>

          <div v-if="historyLoading" class="views-admin-ops-components-ops-alert-events-card__panel-37">
            {{ t('admin.ops.alertEvents.detail.historyLoading') }}
          </div>
          <div v-else-if="history.length === 0" class="views-admin-ops-components-ops-alert-events-card__panel-37">
            {{ t('admin.ops.alertEvents.detail.historyEmpty') }}
          </div>
          <div v-else class="views-admin-ops-components-ops-alert-events-card__panel-38">
            <table class="views-admin-ops-components-ops-alert-events-card__table-2">
              <thead class="views-admin-ops-components-ops-alert-events-card__header-2">
                <tr>
                  <th class="views-admin-ops-components-ops-alert-events-card__heading-4">{{ t('admin.ops.alertEvents.table.time') }}</th>
                  <th class="views-admin-ops-components-ops-alert-events-card__heading-4">{{ t('admin.ops.alertEvents.table.status') }}</th>
                  <th class="views-admin-ops-components-ops-alert-events-card__heading-4">{{ t('admin.ops.alertEvents.table.metric') }}</th>
                </tr>
              </thead>
              <tbody class="views-admin-ops-components-ops-alert-events-card__body-2">
                <tr v-for="it in history" :key="it.id" class="views-admin-ops-components-ops-alert-events-card__row-2">
                  <td class="views-admin-ops-components-ops-alert-events-card__cell-6">{{ formatDateTime(it.fired_at || it.created_at) }}</td>
                  <td class="views-admin-ops-components-ops-alert-events-card__cell-7">
                    <span class="views-admin-ops-components-ops-alert-events-card__text-2" :class="statusBadgeClass(it.status)">
                      {{ formatStatusLabel(it.status) }}
                    </span>
                  </td>
                  <td class="views-admin-ops-components-ops-alert-events-card__cell-6">
                    <span v-if="typeof it.metric_value === 'number' && typeof it.threshold_value === 'number'">
                      {{ it.metric_value.toFixed(2) }} / {{ it.threshold_value.toFixed(2) }}
                    </span>
                    <span v-else>-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </BaseDialog>
  </div>
</template>
