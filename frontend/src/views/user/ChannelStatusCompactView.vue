<template>
  <component :is="layoutComponent">
    <div class="channel-status-compact">
      <!-- 页头:标题 + 更新时间 + 手动刷新(对齐全站卡片头标准) -->
      <section class="channel-status-compact__head card">
        <div class="channel-status-compact__head-main">
          <span class="channel-status-compact__head-icon">
            <Icon name="chartBar" size="md" />
          </span>
          <div>
            <h1 class="channel-status-compact__title">{{ t('channelMonitorV2.compact.pageTitle') }}</h1>
            <p class="channel-status-compact__description">{{ t('channelMonitorV2.compact.pageDescription') }}</p>
          </div>
        </div>
        <div class="channel-status-compact__head-meta">
          <div class="channel-status-compact__updated">
            <span v-if="lastUpdated">{{ t('channelMonitorV2.updatedTo', { time: formatClock(lastUpdated) }) }}</span>
            <span v-else>{{ t('common.loading') }}</span>
            <small>{{ t('channelMonitorV2.compact.autoRefresh') }}</small>
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-icon"
            :title="t('common.refresh')"
            :disabled="loading"
            @click="load()"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'channel-status-compact__spin' : ''" />
          </button>
        </div>
      </section>

      <LoadingState v-if="loading && rows.length === 0" variant="section" class="channel-status-compact__loading" />

      <section v-else-if="rows.length" class="channel-status-compact__list" aria-live="polite">
        <!-- 共享时间刻度:与卡片时间轴同一水平区间 -->
        <div class="channel-status-compact__axis" aria-hidden="true">
          <span>{{ axisLabels.start }}</span>
          <span>{{ axisLabels.middle }}</span>
          <span>{{ t('channelMonitorV2.compact.timeAxisNow') }}</span>
        </div>

        <article v-for="row in rows" :key="rowKey(row)" class="channel-status-compact__card">
          <header class="channel-status-compact__header">
            <div class="channel-status-compact__identity">
              <Icon
                :name="statusIcon(row)"
                size="sm"
                class="channel-status-compact__status-icon"
                :class="statusClass(row)"
              />
              <strong>{{ row.group_name || t('channelMonitorV2.compact.unnamedGroup') }}</strong>
              <span aria-hidden="true">/</span>
              <span>{{ platformLabel(row.platform) }}</span>
              <span class="channel-status-compact__multiplier">×{{ formatMultiplier(row.rate_multiplier) }}</span>
            </div>
            <span class="channel-status-compact__availability">
              {{ availabilityLabel(row) }}
            </span>
          </header>

          <div
            class="channel-status-compact__pulse"
            role="img"
            :aria-label="t('channelMonitorV2.compact.timelineAria', { group: row.group_name || t('channelMonitorV2.compact.unnamedGroup') })"
          >
            <span
              v-for="slot in alignedSlots(row)"
              :key="slot.start"
              class="channel-status-compact__cell"
              :class="bucketClass(slot.bucket)"
              :title="bucketTitle(slot.start, slot.bucket)"
            />
          </div>
        </article>

        <!-- 图例 -->
        <div class="channel-status-compact__legend" aria-hidden="true">
          <span><i class="channel-status-compact__dot is-healthy" />{{ t('channelMonitorV2.compact.legendHealthy') }}</span>
          <span><i class="channel-status-compact__dot is-warning" />{{ t('channelMonitorV2.compact.legendWarning') }}</span>
          <span><i class="channel-status-compact__dot is-critical" />{{ t('channelMonitorV2.compact.legendCritical') }}</span>
          <span><i class="channel-status-compact__dot is-unknown" />{{ t('channelMonitorV2.compact.legendUnknown') }}</span>
        </div>
      </section>

      <div v-else class="channel-status-compact__empty card">
        <Icon name="chart" size="lg" />
        <strong>{{ t('channelMonitorV2.compact.emptyTitle') }}</strong>
        <p>{{ t('channelMonitorV2.compact.emptyDescription') }}</p>
      </div>
    </div>
  </component>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import PublicMonitorLayout from '@/components/layout/PublicMonitorLayout.vue'
import Icon from '@/components/icons/Icon.vue'

import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getPublicMatrix, type MonitorMatrixBucket, type MonitorMatrixRow } from '@/api/channelMonitorV2'

type StatusIcon = 'checkCircle' | 'exclamationCircle' | 'xCircle' | 'clock'
type TimelineSlot = { start: string; bucket?: MonitorMatrixBucket }

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const layoutComponent = computed(() => authStore.isAuthenticated ? AppLayout : PublicMonitorLayout)
const loading = ref(true)
const rows = ref<MonitorMatrixRow[]>([])
const bucketStarts = ref<string[]>([])
const lastUpdated = ref<Date | null>(null)
let controller: AbortController | null = null
let refreshTimer: number | null = null

function formatClock(value: Date): string {
  return new Intl.DateTimeFormat(locale.value || undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(value)
}

function formatAxisTime(iso: string | undefined): string {
  if (!iso) return ''
  return new Intl.DateTimeFormat(locale.value || undefined, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}

// 时间轴左端(约 90 分钟前)与中点(约 45 分钟前)的刻度
const axisLabels = computed(() => ({
  start: formatAxisTime(bucketStarts.value[0]),
  middle: formatAxisTime(bucketStarts.value[Math.floor(bucketStarts.value.length / 2)]),
}))

function buildBucketStarts(start: string, end: string | undefined, bucketSeconds: number): string[] {
  const step = Math.max(60, bucketSeconds || 60) * 1000
  const startMs = new Date(start).getTime()
  const endMs = end ? new Date(end).getTime() : Date.now()
  if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || startMs >= endMs) return []
  const result: string[] = []
  for (let cursor = Math.floor(startMs / step) * step; cursor < endMs; cursor += step) {
    result.push(new Date(cursor).toISOString())
  }
  return result.slice(-90)
}

async function load() {
  controller?.abort()
  const request = new AbortController()
  controller = request
  loading.value = true
  try {
    const result = await getPublicMatrix(
      { range: '90m', platforms: [], groupIds: [], models: [] },
      'platform_group',
      request.signal,
    )
    if (controller !== request) return
    rows.value = (result.items || []).filter((row) => Boolean(row.group_id))
    bucketStarts.value = buildBucketStarts(
      result.coverage.requested_start,
      result.coverage.requested_end,
      result.coverage.bucket_seconds,
    )
    lastUpdated.value = new Date()
  } catch (error) {
    const reason = error as { name?: string; code?: string }
    if (reason.name !== 'AbortError' && reason.name !== 'CanceledError' && reason.code !== 'ERR_CANCELED') {
      appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.compact.loadFailed')))
    }
  } finally {
    if (controller === request) {
      controller = null
      loading.value = false
    }
  }
}

function rowKey(row: MonitorMatrixRow): string {
  return `${row.platform}:${row.group_id || row.group_name || ''}`
}

function alignedSlots(row: MonitorMatrixRow): TimelineSlot[] {
  const buckets = new Map(
    (row.buckets || []).map((bucket) => [new Date(bucket.bucket_start).toISOString(), bucket]),
  )
  return bucketStarts.value.map((start) => ({ start, bucket: buckets.get(start) }))
}

function hasData(row: MonitorMatrixRow): boolean {
  return (row.buckets || []).length > 0
}

function availability(row: MonitorMatrixRow): number | null {
  if (!hasData(row)) return null
  return Math.max(0, Math.min(1, 1 - (row.metrics.error_rate || 0)))
}

function availabilityLabel(row: MonitorMatrixRow): string {
  const value = availability(row)
  if (value == null) return t('channelMonitorV2.compact.noData')
  const formatted = new Intl.NumberFormat(locale.value || undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value * 100)
  return t('channelMonitorV2.compact.availability', { value: formatted })
}

function latestBucket(row: MonitorMatrixRow): MonitorMatrixBucket | undefined {
  return row.buckets?.[row.buckets.length - 1]
}

function statusIcon(row: MonitorMatrixRow): StatusIcon {
  const bucket = latestBucket(row)
  if (!bucket) return 'clock'
  const success = 1 - (bucket.metrics.error_rate || 0)
  if (success >= 0.99) return 'checkCircle'
  if (success >= 0.95) return 'exclamationCircle'
  return 'xCircle'
}

function statusClass(row: MonitorMatrixRow): string {
  const bucket = latestBucket(row)
  if (!bucket) return 'is-unknown'
  return bucketClass(bucket)
}

function bucketClass(bucket?: MonitorMatrixBucket): string {
  if (!bucket) return 'is-unknown'
  const success = 1 - (bucket.metrics.error_rate || 0)
  if (success >= 0.99) return 'is-healthy'
  if (success >= 0.95) return 'is-warning'
  return 'is-critical'
}

function bucketTitle(start: string, bucket?: MonitorMatrixBucket): string {
  const time = new Intl.DateTimeFormat(locale.value || undefined, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(start))
  if (!bucket) return t('channelMonitorV2.compact.noTrafficAt', { time })
  const success = new Intl.NumberFormat(locale.value || undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format((1 - (bucket.metrics.error_rate || 0)) * 100)
  return t('channelMonitorV2.compact.bucketAt', { time, value: success })
}

function platformLabel(platform: string): string {
  return ({
    anthropic: 'Claude',
    openai: 'OpenAI',
    grok: 'Grok',
    kimi: 'Kimi',
    zhipu: 'Zhipu',
    deepseek: 'DeepSeek',
  } as Record<string, string>)[platform] || platform
}

function formatMultiplier(value: number | undefined): string {
  const multiplier = Number(value)
  if (!Number.isFinite(multiplier)) return '1.00'
  return multiplier.toFixed(2)
}

onMounted(() => {
  void load()
  refreshTimer = window.setInterval(() => void load(), 60_000)
})

onBeforeUnmount(() => {
  controller?.abort()
  if (refreshTimer) window.clearInterval(refreshTimer)
})
</script>

<style scoped>
.channel-status-compact {
  display: grid;
  gap: 1rem;
  width: min(1120px, 100%);
  margin: 0 auto;
  padding: 1.25rem 1rem 2rem;
}

/* ---- 页头卡:对齐全站卡片头标准 ---- */
.channel-status-compact__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  padding: 1rem 1.5rem;
}

.channel-status-compact__head-main {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 0.75rem;
}

.channel-status-compact__head-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-md);
  color: var(--color-text-brand);
  background: var(--color-primary-subtle);
}

.channel-status-compact__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
  font-weight: 650;
}

.channel-status-compact__description {
  margin: 0.125rem 0 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
}

.channel-status-compact__head-meta {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  gap: 0.75rem;
}

.channel-status-compact__updated {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.125rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-variant-numeric: tabular-nums;
}

.channel-status-compact__updated small {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
}

.channel-status-compact__spin {
  animation: channel-status-spin 1s linear infinite;
}

@keyframes channel-status-spin {
  to {
    transform: rotate(360deg);
  }
}

.channel-status-compact__loading {
  display: grid;
  min-height: 18rem;
  place-items: center;
}

.channel-status-compact__list {
  display: grid;
  gap: .75rem;
}

/* ---- 共享时间刻度:与卡片时间轴同一水平区间(1rem 内边距 + 1px 边框补偿) ---- */
.channel-status-compact__axis {
  display: flex;
  justify-content: space-between;
  margin-bottom: -0.25rem;
  padding: 0 calc(1rem + 1px);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-variant-numeric: tabular-nums;
}

.channel-status-compact__card {
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-bg);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
  transition: background-color 160ms ease, border-color 160ms ease;
}

.channel-status-compact__card:hover {
  border-color: var(--color-border);
  background: var(--glass-bg-thick);
}

.channel-status-compact__header {
  display: flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: .6rem 1rem .45rem;
}

.channel-status-compact__identity {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: .5rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.channel-status-compact__identity strong {
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.channel-status-compact__status-icon {
  flex: 0 0 auto;
}

.channel-status-compact__multiplier {
  color: var(--color-text-secondary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.channel-status-compact__availability {
  flex: 0 0 auto;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-variant-numeric: tabular-nums;
}

.channel-status-compact__pulse {
  display: grid;
  height: 1rem;
  grid-template-columns: repeat(90, minmax(2px, 1fr));
  gap: 2px;
  padding: 0 1rem .75rem;
  box-sizing: content-box;
}

.channel-status-compact__cell {
  display: block;
  min-width: 0;
  border-radius: 1px;
  background: #9ca3af;
}

.is-healthy {
  color: #10b981;
}

.channel-status-compact__cell.is-healthy {
  background: #10b981;
}

.is-warning {
  color: #f59e0b;
}

.channel-status-compact__cell.is-warning {
  background: #f59e0b;
}

.is-critical {
  color: #ef4444;
}

.channel-status-compact__cell.is-critical {
  background: #ef4444;
}

.is-unknown {
  color: #9ca3af;
}

.channel-status-compact__cell.is-unknown {
  background: #9ca3af;
  opacity: .45;
}

/* ---- 图例 ---- */
.channel-status-compact__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem 1rem;
  padding: 0.25rem calc(1rem + 1px) 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
}

.channel-status-compact__legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
}

.channel-status-compact__dot {
  display: inline-block;
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 2px;
}

.channel-status-compact__dot.is-healthy { background: #10b981; }
.channel-status-compact__dot.is-warning { background: #f59e0b; }
.channel-status-compact__dot.is-critical { background: #ef4444; }
.channel-status-compact__dot.is-unknown { background: #9ca3af; opacity: 0.45; }

.channel-status-compact__empty {
  display: grid;
  min-height: 12rem;
  place-items: center;
  align-content: center;
  gap: .5rem;
  padding: 2rem;
  color: var(--color-text-secondary);
  text-align: center;
}

.channel-status-compact__empty strong {
  color: var(--color-text-primary);
}

.channel-status-compact__empty p {
  margin: 0;
  font-size: var(--font-size-sm);
}

@media (max-width: 700px) {
  .channel-status-compact {
    padding: 0.75rem 0.75rem 1.5rem;
  }

  .channel-status-compact__head {
    padding: 0.875rem 1rem;
  }

  .channel-status-compact__updated {
    align-items: flex-start;
  }

  .channel-status-compact__header {
    align-items: flex-start;
    flex-direction: column;
    gap: .2rem;
  }

  .channel-status-compact__availability {
    padding-left: 1.5rem;
  }

  .channel-status-compact__pulse {
    gap: 1px;
  }
}
</style>
