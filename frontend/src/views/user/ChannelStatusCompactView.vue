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
        <article v-for="entry in heatmapRows" :key="rowKey(entry.row)" class="channel-status-compact__card">
          <header class="channel-status-compact__header">
            <div class="channel-status-compact__identity">
              <Icon
                :name="statusIcon(entry.row)"
                size="sm"
                class="channel-status-compact__status-icon"
                :class="statusClass(entry.row)"
              />
              <strong>{{ entry.row.group_name || t('channelMonitorV2.compact.unnamedGroup') }}</strong>
              <span aria-hidden="true">/</span>
              <span>{{ platformLabel(entry.row.platform) }}</span>
              <span class="channel-status-compact__multiplier">×{{ formatMultiplier(entry.row.rate_multiplier) }}</span>
            </div>
            <span class="channel-status-compact__error-rate">
              {{ errorRateLabel(entry.row) }}
            </span>
          </header>

          <!-- 时间轴属于当前分组，避免多个分组共用一个悬空刻度。 -->
          <div class="channel-status-compact__axis" aria-hidden="true">
            <span class="channel-status-compact__axis-spacer" />
            <div class="channel-status-compact__day-labels">
              <span v-for="day in heatmapDays" :key="day.key" :title="day.title">{{ day.label }}</span>
            </div>
          </div>

          <div
            class="channel-status-compact__heatmap"
            role="img"
            :aria-label="t('channelMonitorV2.compact.timelineAria', { group: entry.row.group_name || t('channelMonitorV2.compact.unnamedGroup') })"
          >
            <div v-for="slot in heatmapSlots" :key="slot.index" class="channel-status-compact__heatmap-row">
              <span class="channel-status-compact__time-label">{{ slot.label }}</span>
              <div class="channel-status-compact__cells">
                <span
                  v-for="day in heatmapDays"
                  :key="`${day.key}:${slot.index}`"
                  class="channel-status-compact__cell"
                  :class="bucketClass(bucketFor(entry, day.key, slot.index))"
                  :title="cellTooltipLabel(entry, day.key, slot.index)"
                  :aria-label="cellTooltipLabel(entry, day.key, slot.index)"
                  tabindex="0"
                  role="img"
                  @mouseenter="showTooltip($event, entry, day.key, slot.index)"
                  @mousemove="moveTooltip"
                  @mouseleave="hideTooltip"
                  @focus="showTooltip($event, entry, day.key, slot.index)"
                  @blur="hideTooltip"
                  @keydown.esc="hideTooltip"
                />
              </div>
            </div>
          </div>
        </article>

        <!-- 图例 -->
        <div class="channel-status-compact__legend" :aria-label="t('channelMonitorV2.compact.legendAria')">
          <span class="channel-status-compact__legend-label">{{ t('channelMonitorV2.compact.legendLowError') }}</span>
          <span class="channel-status-compact__legend-scale" aria-hidden="true">
            <i class="channel-status-compact__dot error-band-0" />
            <i class="channel-status-compact__dot error-band-1" />
            <i class="channel-status-compact__dot error-band-2" />
            <i class="channel-status-compact__dot error-band-3" />
            <i class="channel-status-compact__dot error-band-4" />
            <i class="channel-status-compact__dot error-band-5" />
            <i class="channel-status-compact__dot error-band-6" />
            <i class="channel-status-compact__dot error-band-7" />
          </span>
          <span class="channel-status-compact__legend-label">{{ t('channelMonitorV2.compact.legendHighError') }}</span>
          <span class="channel-status-compact__legend-unknown">
            <i class="channel-status-compact__dot error-band-unknown" />
            {{ t('channelMonitorV2.compact.legendUnknown') }}
          </span>
        </div>
      </section>

      <div v-else class="channel-status-compact__empty card">
        <Icon name="chart" size="lg" />
        <strong>{{ t('channelMonitorV2.compact.emptyTitle') }}</strong>
        <p>{{ t('channelMonitorV2.compact.emptyDescription') }}</p>
      </div>

      <Teleport to="body">
        <div
          v-if="floatingTooltip.visible"
          class="channel-status-compact__floating-tooltip"
          :style="{ left: `${floatingTooltip.x}px`, top: `${floatingTooltip.y}px` }"
          role="tooltip"
        >
          <span
            v-for="(line, index) in floatingTooltip.lines"
            :key="`${index}:${line}`"
            class="channel-status-compact__tooltip-line"
            :class="index === 0 ? 'channel-status-compact__tooltip-title' : ''"
          >
            {{ line }}
          </span>
        </div>
      </Teleport>
    </div>
  </component>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import PublicMonitorLayout from '@/components/layout/PublicMonitorLayout.vue'
import Icon from '@/components/icons/Icon.vue'

import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getMatrix, getPublicMatrix, type MonitorMatrixBucket, type MonitorMatrixRow } from '@/api/channelMonitorV2'

type StatusIcon = 'checkCircle' | 'exclamationCircle' | 'xCircle' | 'clock'
type HeatmapDay = { key: string; label: string; title: string }
type HeatmapSlot = { index: number; label: string }
type HeatmapEntry = { row: MonitorMatrixRow; bucketMap: Map<string, MonitorMatrixBucket> }

const HEATMAP_DAY_COUNT = 14
const HEATMAP_SLOT_COUNT = 12
const HEATMAP_BUCKET_HOURS = 2

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const layoutComponent = computed(() => authStore.isAuthenticated ? AppLayout : PublicMonitorLayout)
const loading = ref(true)
const rows = ref<MonitorMatrixRow[]>([])
const heatmapAnchor = ref(new Date())
const lastUpdated = ref<Date | null>(null)
const floatingTooltip = reactive({
  visible: false,
  x: 0,
  y: 0,
  lines: [] as string[],
})
let controller: AbortController | null = null
let refreshTimer: number | null = null

function formatClock(value: Date): string {
  return new Intl.DateTimeFormat(locale.value || undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(value)
}

function localDateKey(date: Date): string {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

function formatHeatmapDay(date: Date): string {
  return new Intl.DateTimeFormat(locale.value || undefined, {
    month: 'short',
    day: 'numeric',
  }).format(date)
}

const heatmapDays = computed<HeatmapDay[]>(() => {
  const anchor = heatmapAnchor.value
  const anchorDay = new Date(anchor.getFullYear(), anchor.getMonth(), anchor.getDate())
  return Array.from({ length: HEATMAP_DAY_COUNT }, (_, index) => {
    const date = new Date(anchorDay)
    date.setDate(anchorDay.getDate() - (HEATMAP_DAY_COUNT - 1 - index))
    return { key: localDateKey(date), label: String(date.getDate()), title: formatHeatmapDay(date) }
  })
})

const heatmapSlots = computed<HeatmapSlot[]>(() => Array.from({ length: HEATMAP_SLOT_COUNT }, (_, index) => {
  const hour = index * HEATMAP_BUCKET_HOURS
  return {
    index,
    label: index % 2 === 0 ? `${String(hour).padStart(2, '0')}:00` : '',
  }
}))

async function load() {
  controller?.abort()
  const request = new AbortController()
  controller = request
  loading.value = true
  try {
    const filter = { range: '14d' as const, platforms: [], groupIds: [], models: [] }
    const result = authStore.isAdmin
      ? await getMatrix(filter, 'platform_group', true, request.signal)
      : await getPublicMatrix(filter, 'platform_group', request.signal)
    if (controller !== request) return
    rows.value = (result.items || []).filter((row) => Boolean(row.group_id))
    heatmapAnchor.value = new Date()
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

function buildBucketMap(row: MonitorMatrixRow): Map<string, MonitorMatrixBucket> {
  const result = new Map<string, MonitorMatrixBucket>()
  for (const bucket of row.buckets || []) {
    const date = new Date(bucket.bucket_start)
    if (!Number.isFinite(date.getTime())) continue
    const slotIndex = Math.floor(date.getHours() / HEATMAP_BUCKET_HOURS)
    result.set(`${localDateKey(date)}:${slotIndex}`, bucket)
  }
  return result
}

const heatmapRows = computed<HeatmapEntry[]>(() => rows.value.map((row) => ({
  row,
  bucketMap: buildBucketMap(row),
})))

function bucketFor(entry: HeatmapEntry, dayKey: string, slotIndex: number): MonitorMatrixBucket | undefined {
  return entry.bucketMap.get(`${dayKey}:${slotIndex}`)
}

function heatmapSlotStart(dayKey: string, slotIndex: number): Date {
  const [year, month, day] = dayKey.split('-').map(Number)
  return new Date(year, month - 1, day, slotIndex * HEATMAP_BUCKET_HOURS)
}

function hasBucketTraffic(bucket: MonitorMatrixBucket): boolean {
  const metrics = bucket.metrics
  return Boolean(
    metrics.request_count > 0 ||
      metrics.error_requests > 0 ||
      metrics.rpm > 0 ||
      metrics.tpm > 0 ||
      metrics.error_rate > 0,
  )
}

function hasData(row: MonitorMatrixRow): boolean {
  return row.metrics.request_count > 0 || (row.buckets || []).some(hasBucketTraffic)
}

function normalizedErrorRate(value: number | null | undefined): number {
  const rate = Number(value)
  if (!Number.isFinite(rate)) return 0
  return Math.max(0, Math.min(1, rate))
}

function formatErrorRate(value: number): string {
  const percent = normalizedErrorRate(value) * 100
  const fractionDigits = percent > 0 && percent < 1 ? 2 : 1
  return new Intl.NumberFormat(locale.value || undefined, {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(percent)
}

function errorRateLabel(row: MonitorMatrixRow): string {
  if (!hasData(row)) return t('channelMonitorV2.compact.noData')
  return t('channelMonitorV2.compact.errorRate', { value: formatErrorRate(row.metrics.error_rate) })
}

function statusIcon(row: MonitorMatrixRow): StatusIcon {
  const rate = hasData(row) ? normalizedErrorRate(row.metrics.error_rate) : null
  if (rate == null) return 'clock'
  if (rate <= 0.01) return 'checkCircle'
  if (rate <= 0.05) return 'exclamationCircle'
  return 'xCircle'
}

function statusClass(row: MonitorMatrixRow): string {
  return errorRateClass(hasData(row) ? row.metrics.error_rate : null)
}

function bucketClass(bucket?: MonitorMatrixBucket): string {
  if (!bucket || !hasBucketTraffic(bucket)) return 'error-band-unknown'
  return errorRateClass(bucket.metrics.error_rate)
}

function errorRateClass(value: number | null | undefined): string {
  if (value == null || !Number.isFinite(Number(value))) return 'error-band-unknown'
  const rate = normalizedErrorRate(value)
  if (rate <= 0.005) return 'error-band-0'
  if (rate <= 0.01) return 'error-band-1'
  if (rate <= 0.03) return 'error-band-2'
  if (rate <= 0.05) return 'error-band-3'
  if (rate <= 0.1) return 'error-band-4'
  if (rate <= 0.2) return 'error-band-5'
  if (rate <= 0.35) return 'error-band-6'
  return 'error-band-7'
}

function format24Hour(date: Date): string {
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

function formatBucketRange(start: Date): string {
  const end = new Date(start.getTime() + HEATMAP_BUCKET_HOURS * 60 * 60 * 1000)
  return `${formatHeatmapDay(start)} ${format24Hour(start)}-${format24Hour(end)}`
}

function formatCount(value: number): string {
  return new Intl.NumberFormat(locale.value || undefined, {
    maximumFractionDigits: 0,
  }).format(value)
}

function bucketTooltipLines(start: Date, bucket?: MonitorMatrixBucket): string[] {
  const time = formatBucketRange(start)
  if (!bucket) return [time, t('channelMonitorV2.compact.noTraffic')]

  const metrics = bucket.metrics
  const lines = [
    time,
    t('channelMonitorV2.compact.tooltipErrorRate', { value: formatErrorRate(metrics.error_rate) }),
  ]
  if (metrics.request_count > 0) {
    lines.push(t('channelMonitorV2.compact.requestCount', { value: formatCount(metrics.request_count) }))
  }
  if (metrics.error_requests > 0) {
    lines.push(t('channelMonitorV2.compact.errorRequests', { value: formatCount(metrics.error_requests) }))
  }
  if (metrics.request_count > 0) {
    const successRate = metrics.success_requests / metrics.request_count
    lines.push(t('channelMonitorV2.metrics.successRateValue', { value: formatErrorRate(successRate) + '%' }))
  }
  return lines
}

function cellTooltipLabel(entry: HeatmapEntry, dayKey: string, slotIndex: number): string {
  return bucketTooltipLines(
    heatmapSlotStart(dayKey, slotIndex),
    bucketFor(entry, dayKey, slotIndex),
  ).join('\n')
}

function showTooltip(event: MouseEvent | FocusEvent, entry: HeatmapEntry, dayKey: string, slotIndex: number) {
  floatingTooltip.lines = bucketTooltipLines(
    heatmapSlotStart(dayKey, slotIndex),
    bucketFor(entry, dayKey, slotIndex),
  )
  floatingTooltip.visible = true
  positionTooltip(event)
}

function moveTooltip(event: MouseEvent) {
  if (floatingTooltip.visible) positionTooltip(event)
}

function hideTooltip() {
  floatingTooltip.visible = false
}

function positionTooltip(event: MouseEvent | FocusEvent) {
  if ('clientX' in event) {
    floatingTooltip.x = Math.min(window.innerWidth - 12, Math.max(12, event.clientX))
    floatingTooltip.y = Math.min(window.innerHeight - 12, Math.max(12, event.clientY)) - 12
    return
  }
  const target = event.target as HTMLElement | null
  const rect = target?.getBoundingClientRect()
  if (!rect) return
  floatingTooltip.x = rect.left + rect.width / 2
  floatingTooltip.y = rect.top - 10
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
  gap: .75rem;
  width: min(1280px, 100%);
  margin: 0 auto;
  padding: .9rem .75rem 1.5rem;
}

/* ---- 页头卡:对齐全站卡片头标准 ---- */
.channel-status-compact__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: .5rem .75rem;
  padding: .75rem 1rem;
}

.channel-status-compact__head-main {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: .6rem;
}

.channel-status-compact__head-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 2rem;
  height: 2rem;
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
  grid-template-columns: repeat(auto-fit, minmax(17rem, 21rem));
  justify-content: start;
  align-items: start;
  gap: .6rem;
}

/* ---- 每个分组内部的最近 14 天日期轴，与 14 列网格对齐 ---- */
.channel-status-compact__axis {
  display: grid;
  grid-template-columns: 1.9rem max-content;
  align-items: end;
  gap: .3rem;
  padding: 0 .65rem .25rem;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-variant-numeric: tabular-nums;
  overflow-x: auto;
  scrollbar-width: none;
}

.channel-status-compact__axis::-webkit-scrollbar,
.channel-status-compact__heatmap::-webkit-scrollbar {
  display: none;
}

.channel-status-compact__axis-spacer {
  display: block;
}

.channel-status-compact__day-labels,
.channel-status-compact__cells {
  display: grid;
  grid-template-columns: repeat(14, .62rem);
  gap: .16rem;
  width: max-content;
}

.channel-status-compact__day-labels span {
  width: .62rem;
  text-align: center;
  white-space: nowrap;
  font-size: .58rem;
}

.channel-status-compact__card {
  min-width: 0;
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
  min-height: 2.35rem;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  padding: .45rem .65rem .35rem;
}

.channel-status-compact__identity {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  align-items: center;
  flex-wrap: wrap;
  gap: .35rem;
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

.channel-status-compact__error-rate {
  flex: 0 0 auto;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-variant-numeric: tabular-nums;
}

.channel-status-compact__heatmap {
  display: grid;
  gap: .08rem;
  padding: 0 .65rem .55rem;
  overflow-x: auto;
  scrollbar-width: none;
}

.channel-status-compact__heatmap-row {
  display: grid;
  grid-template-columns: 1.9rem max-content;
  align-items: center;
  gap: .3rem;
  min-height: .58rem;
}

.channel-status-compact__time-label {
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-variant-numeric: tabular-nums;
  text-align: right;
  white-space: nowrap;
}

.channel-status-compact__cells {
  align-items: center;
}

.channel-status-compact__cell {
  display: block;
  width: .62rem;
  height: .62rem;
  border-radius: 1px;
  background: var(--monitor-error-color);
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 0.2);
  transition: transform 120ms ease, box-shadow 120ms ease;
}

.channel-status-compact__cell:hover {
  position: relative;
  z-index: 1;
  transform: scale(1.2);
  box-shadow: 0 0 0 2px var(--monitor-error-color), 0 2px 6px rgb(0 0 0 / 0.18);
}

.channel-status-compact__status-icon {
  color: var(--monitor-error-ink);
}

/* Error-rate bands: low error is green, high error moves through yellow/orange to red. */
.error-band-0 {
  --monitor-error-color: #dcfce7;
  --monitor-error-ink: #166534;
}
.error-band-1 {
  --monitor-error-color: #86efac;
  --monitor-error-ink: #166534;
}
.error-band-2 {
  --monitor-error-color: #4ade80;
  --monitor-error-ink: #15803d;
}
.error-band-3 {
  --monitor-error-color: #22c55e;
  --monitor-error-ink: #166534;
}
.error-band-4 {
  --monitor-error-color: #facc15;
  --monitor-error-ink: #854d0e;
}
.error-band-5 {
  --monitor-error-color: #fb923c;
  --monitor-error-ink: #9a3412;
}
.error-band-6 {
  --monitor-error-color: #f97316;
  --monitor-error-ink: #9a3412;
}
.error-band-7 {
  --monitor-error-color: #ef4444;
  --monitor-error-ink: #991b1b;
}
.error-band-unknown {
  --monitor-error-color: #d1d5db;
  --monitor-error-ink: #6b7280;
}

.channel-status-compact__cell.error-band-unknown {
  opacity: .45;
}

/* ---- 图例 ---- */
.channel-status-compact__legend {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.375rem .5rem;
  padding: 0.25rem calc(1rem + 1px) 0;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
}

.channel-status-compact__legend-label,
.channel-status-compact__legend-unknown,
.channel-status-compact__legend-scale {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
}

.channel-status-compact__legend-unknown {
  margin-left: auto;
}

.channel-status-compact__dot {
  display: inline-block;
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 2px;
  background: var(--monitor-error-color);
}

.channel-status-compact__legend-scale {
  gap: 2px;
}

.channel-status-compact__floating-tooltip {
  pointer-events: none;
  position: fixed;
  z-index: 100000;
  min-width: 12rem;
  max-width: min(20rem, calc(100vw - 1.5rem));
  transform: translate(-50%, calc(-100% - .5rem));
  padding: .5rem .625rem;
  overflow-wrap: anywhere;
  border: 1px solid var(--glass-border-hover);
  border-radius: .65rem;
  background: var(--glass-layer-floating-bg);
  color: var(--color-text-secondary);
  box-shadow: var(--glass-shadow-hover);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  font-size: var(--font-size-2xs);
  line-height: 1.45;
  white-space: normal;
}

.channel-status-compact__tooltip-line {
  display: block;
}

.channel-status-compact__tooltip-title {
  margin-bottom: .2rem;
  color: var(--color-text-primary);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

:global(.dark) .error-band-0 {
  --monitor-error-color: #166534;
  --monitor-error-ink: #bbf7d0;
}
:global(.dark) .error-band-1 {
  --monitor-error-color: #15803d;
  --monitor-error-ink: #bbf7d0;
}
:global(.dark) .error-band-2 {
  --monitor-error-color: #22c55e;
  --monitor-error-ink: #dcfce7;
}
:global(.dark) .error-band-3 {
  --monitor-error-color: #4ade80;
  --monitor-error-ink: #14532d;
}
:global(.dark) .error-band-4 {
  --monitor-error-color: #ca8a04;
  --monitor-error-ink: #fef08a;
}
:global(.dark) .error-band-5 {
  --monitor-error-color: #ea580c;
  --monitor-error-ink: #fed7aa;
}
:global(.dark) .error-band-6 {
  --monitor-error-color: #f97316;
  --monitor-error-ink: #ffedd5;
}
:global(.dark) .error-band-7 {
  --monitor-error-color: #dc2626;
  --monitor-error-ink: #fecaca;
}
:global(.dark) .error-band-unknown {
  --monitor-error-color: #4b5563;
  --monitor-error-ink: #d1d5db;
}

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
    padding: .75rem .5rem 1.25rem;
  }

  .channel-status-compact__list {
    grid-template-columns: minmax(0, 1fr);
    gap: .5rem;
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

  .channel-status-compact__error-rate {
    padding-left: 1.5rem;
  }

  .channel-status-compact__axis {
    grid-template-columns: 1.65rem max-content;
    gap: .22rem;
    padding-right: .5rem;
    padding-left: .5rem;
  }

  .channel-status-compact__heatmap {
    padding-right: .5rem;
    padding-left: .5rem;
  }

  .channel-status-compact__heatmap-row {
    grid-template-columns: 1.65rem max-content;
    gap: .22rem;
  }

  .channel-status-compact__day-labels,
  .channel-status-compact__cells {
    grid-template-columns: repeat(14, .5rem);
    gap: .12rem;
  }

  .channel-status-compact__day-labels span {
    width: .5rem;
    font-size: .52rem;
  }

  .channel-status-compact__cell {
    width: .5rem;
    height: .5rem;
  }

  .channel-status-compact__floating-tooltip {
    min-width: 10rem;
  }

  .channel-status-compact__legend-unknown {
    margin-left: 0;
  }
}
</style>
