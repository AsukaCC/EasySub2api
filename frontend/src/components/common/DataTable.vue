<template>
  <div
    class="data-table"
    :class="{
      'data-table--mobile': !isDesktopViewport,
      'data-table--refreshing': isRefreshing,
      'data-table--resizing': Boolean(resizingColumnKey),
    }"
    :aria-busy="loading ? 'true' : undefined"
  >
    <div v-if="!isDesktopViewport" class="data-table__mobile-list">
      <template v-if="isInitialLoading">
        <div v-for="i in 5" :key="i" class="data-table__mobile-card data-table__mobile-card--skeleton">
          <div v-for="column in dataColumns" :key="column.key" class="data-table__mobile-field">
            <span class="data-table__skeleton data-table__skeleton--label"></span>
            <span class="data-table__skeleton data-table__skeleton--value"></span>
          </div>
          <div v-if="hasActionsColumn" class="data-table__mobile-actions">
            <span class="data-table__skeleton data-table__skeleton--action"></span>
          </div>
        </div>
      </template>

      <template v-else-if="!data || data.length === 0">
        <div class="data-table__mobile-empty">
          <slot name="empty">
            <div class="data-table__empty-content">
              <Icon name="inbox" size="xl" class="data-table__empty-icon" />
              <p class="data-table__empty-title">{{ t('empty.noData') }}</p>
            </div>
          </slot>
        </div>
      </template>

      <template v-else>
        <div v-if="selectable" class="data-table__mobile-selection">
          <label class="data-table__selection-label">
            <input
              type="checkbox"
              class="data-table__checkbox"
              :checked="allVisibleSelected"
              :indeterminate="someVisibleSelected"
              data-test="select-all-mobile"
              @change="toggleAllVisible(($event.target as HTMLInputElement).checked)"
            />
            <span>{{ t('common.selectAll') }}</span>
          </label>
        </div>
        <div
          v-for="(row, index) in sortedData"
          :key="resolveRowKey(row, index)"
          class="data-table__mobile-card"
          :class="{
            'is-clickable': clickableRows,
            'is-selected': selectable && isRowSelected(row, index),
          }"
          @click="clickableRows && emit('rowClick', row)"
        >
          <div v-if="selectable" class="data-table__mobile-row-selection">
            <input
              type="checkbox"
              class="data-table__checkbox"
              :checked="isRowSelected(row, index)"
              :aria-label="getRowSelectionLabel(row, index)"
              data-test="select-row"
              @click.stop
              @change="toggleRowSelection(row, index, ($event.target as HTMLInputElement).checked)"
            />
          </div>
          <div
            v-for="column in dataColumns"
            :key="column.key"
            :data-field="column.key"
            class="data-table__mobile-field"
          >
            <span class="data-table__mobile-label">{{ column.label }}</span>
            <div class="data-table__mobile-value">
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]" :expanded="actionsExpanded">
                {{ column.formatter ? column.formatter(row[column.key], row) : row[column.key] }}
              </slot>
            </div>
          </div>
          <div v-if="hasActionsColumn" class="data-table__mobile-actions">
            <slot name="cell-actions" :row="row" :value="row.actions" :expanded="actionsExpanded"></slot>
          </div>
        </div>
      </template>
    </div>

    <div
      v-else
      ref="tableWrapperRef"
      class="data-table__viewport table-wrapper"
      :class="{
        'actions-expanded': actionsExpanded,
        'is-scrollable': isScrollable,
      }"
    >
      <table class="data-table__table">
        <colgroup>
          <col v-if="selectable" class="data-table__selection-column" />
          <col
            v-for="column in columns"
            :key="column.key"
            :style="getColumnStyle(column)"
          />
        </colgroup>
        <thead class="data-table__header table-header">
          <tr>
            <th v-if="selectable" scope="col" class="data-table__header-cell data-table__selection-cell sticky-header-cell glass-sticky-cell">
              <input
                type="checkbox"
                class="data-table__checkbox"
                :checked="allVisibleSelected"
                :indeterminate="someVisibleSelected"
                :aria-label="t('common.selectAll')"
                data-test="select-all"
                @change="toggleAllVisible(($event.target as HTMLInputElement).checked)"
              />
            </th>
            <th
              v-for="(column, index) in columns"
              :key="column.key"
              scope="col"
              :aria-sort="column.sortable ? getColumnAriaSort(column.key) : undefined"
              :style="getColumnStyle(column)"
              :class="[
                'data-table__header-cell sticky-header-cell glass-sticky-cell',
                getAdaptivePaddingClass(),
                getColumnAlignmentClass(column),
                getStickyColumnClass(column, index),
                column.headerClass,
                {
                  'is-sortable': column.sortable,
                  'is-resized': hasCustomColumnWidth(column.key),
                  'is-being-resized': resizingColumnKey === column.key,
                },
              ]"
              @click="column.sortable && handleSort(column.key)"
            >
              <div class="data-table__header-content">
                <slot
                  :name="`header-${column.key}`"
                  :column="column"
                  :sort-key="sortKey"
                  :sort-order="sortOrder"
                >
                  <span>{{ column.label }}</span>
                </slot>
                <span v-if="column.sortable" class="data-table__sort-indicator" aria-hidden="true">
                  <svg
                    class="data-table__sort-arrow data-table__sort-arrow--ascending"
                    :class="getSortIndicatorClass(column.key, 'asc')"
                    fill="currentColor"
                    viewBox="0 0 10 10"
                  >
                    <path d="M5 2L1.5 6.5h7L5 2z" />
                  </svg>
                  <svg
                    class="data-table__sort-arrow data-table__sort-arrow--descending"
                    :class="getSortIndicatorClass(column.key, 'desc')"
                    fill="currentColor"
                    viewBox="0 0 10 10"
                  >
                    <path d="M5 8L1.5 3.5h7L5 8z" />
                  </svg>
                </span>
              </div>
              <button
                v-if="isColumnResizable(column)"
                type="button"
                class="data-table__resize-handle"
                :class="{ 'is-active': resizingColumnKey === column.key }"
                :aria-label="t('common.resizeColumn', { column: column.label })"
                :title="t('common.resizeColumnHint', { column: column.label })"
                data-test="column-resize-handle"
                @click.stop.prevent
                @dblclick.stop.prevent="resetColumnWidth(column.key)"
                @pointerdown.stop.prevent="startColumnResize($event, column)"
                @keydown="handleColumnResizeKeydown($event, column)"
              ></button>
            </th>
          </tr>
        </thead>
        <tbody class="data-table__body table-body">
          <tr v-if="isInitialLoading" v-for="i in 5" :key="i" class="data-table__skeleton-row">
            <td v-if="selectable" class="data-table__selection-cell data-table__body-cell">
              <span class="data-table__skeleton data-table__skeleton--checkbox"></span>
            </td>
            <td
              v-for="column in columns"
              :key="column.key"
              :style="getColumnStyle(column)"
              :class="[
                'data-table__body-cell',
                getAdaptivePaddingClass(),
                { 'is-resized': hasCustomColumnWidth(column.key) },
              ]"
            >
              <span class="data-table__skeleton data-table__skeleton--cell"></span>
            </td>
          </tr>

          <tr v-else-if="!data || data.length === 0">
            <td :colspan="tableColumnCount" :class="['data-table__empty-cell', getAdaptivePaddingClass()]">
              <slot name="empty">
                <div class="data-table__empty-content">
                  <Icon name="inbox" size="xl" class="data-table__empty-icon" />
                  <p class="data-table__empty-title">{{ t('empty.noData') }}</p>
                </div>
              </slot>
            </td>
          </tr>

          <template v-else>
            <tr v-if="virtualPaddingTop > 0" aria-hidden="true">
              <td :colspan="tableColumnCount" :style="{ height: `${virtualPaddingTop}px`, padding: 0, border: 'none' }"></td>
            </tr>
            <tr
              v-for="item in renderRows"
              :key="resolveRowKey(item.row, item.index)"
              :data-row-id="resolveRowKey(item.row, item.index)"
              :data-index="item.index"
              :ref="item.measure ? measureElement : undefined"
              class="data-table__row"
              :class="{
                'is-clickable': clickableRows,
                'is-selected': selectable && isRowSelected(item.row, item.index),
              }"
              @click="clickableRows && emit('rowClick', item.row)"
            >
              <td v-if="selectable" class="data-table__body-cell data-table__selection-cell">
                <input
                  type="checkbox"
                  class="data-table__checkbox"
                  :checked="isRowSelected(item.row, item.index)"
                  :aria-label="getRowSelectionLabel(item.row, item.index)"
                  data-test="select-row"
                  @click.stop
                  @change="toggleRowSelection(item.row, item.index, ($event.target as HTMLInputElement).checked)"
                />
              </td>
              <td
                v-for="(column, colIndex) in columns"
                :key="column.key"
                :style="getColumnStyle(column)"
                :class="[
                  'data-table__body-cell',
                  getAdaptivePaddingClass(),
                  getColumnAlignmentClass(column),
                  getStickyColumnClass(column, colIndex),
                  column.cellClass,
                  { 'is-resized': hasCustomColumnWidth(column.key) },
                ]"
              >
                <slot
                  :name="`cell-${column.key}`"
                  :row="item.row"
                  :value="item.row[column.key]"
                  :expanded="actionsExpanded"
                >
                  {{ column.formatter ? column.formatter(item.row[column.key], item.row) : item.row[column.key] }}
                </slot>
              </td>
            </tr>
            <tr v-if="virtualPaddingBottom > 0" aria-hidden="true">
              <td :colspan="tableColumnCount" :style="{ height: `${virtualPaddingBottom}px`, padding: 0, border: 'none' }"></td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <div
      v-if="isRefreshing"
      class="data-table__loading-overlay"
      :class="isDesktopViewport ? 'data-table__loading-overlay--desktop' : 'data-table__loading-overlay--mobile'"
      role="status"
      aria-live="polite"
      aria-atomic="true"
    >
      <div class="data-table__loading-state">
        <span class="data-table__spinner" aria-hidden="true"></span>
        <span>{{ t('common.loading') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useVirtualizer, observeElementRect as observeElementRectDefault } from '@tanstack/vue-virtual'
import { useI18n } from 'vue-i18n'
import type { Column } from './types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const desktopViewportQuery = '(min-width: 768px)'
const isDesktopViewport = ref(
  typeof window === 'undefined' ? true : window.matchMedia(desktopViewportQuery).matches
)

const emit = defineEmits<{
  sort: [key: string, order: 'asc' | 'desc']
  rowClick: [row: any]
  'update:selectedKeys': [keys: Array<string | number>]
  selectionChange: [keys: Array<string | number>]
}>()

// 表格容器引用
const tableWrapperRef = ref<HTMLElement | null>(null)
const isScrollable = ref(false)
const actionsColumnNeedsExpanding = ref(false)

// --- 虚拟滚动「整表空白」根治 ---
// 根因:本组件根 .table-wrapper 为 flex:1 / min-h-0,高度由父级 flex 链决定。@tanstack 虚拟化器
// 仅在 observeElementRect 回调里写 scrollRect;一旦该回调读到 0 高度(加载瞬间 flex 未结算,或
// 滚动中动态行高校正触发的 reflow),scrollRect 被钉死为 0 → calculateRange 返回 null → 整表空白。
// 对策(见下方 virtualizer 选项):
//   1) 覆写 observeElementRect,直接丢弃 height<=0 的读数,scrollRect 永不被钉成 0;
//   2) initialRect 给一屏兜底高度,首个有效读数到来前也有行可渲染,绝不空白。
// 兜底高度:表格区域大致 = 视口高度 - 顶栏/外边距/筛选/分页 ≈ 320px
const estimatedViewportHeight = () => {
  if (typeof window === 'undefined') return 600
  return Math.max(window.innerHeight - 320, 400)
}

// 覆写默认 observeElementRect:过滤掉 0 高度读数(根治整表空白的关键)
const observeElementRectNonZero = (
  instance: any,
  cb: (rect: { width: number; height: number }) => void
) => observeElementRectDefault(instance, (rect) => {
  if (rect.height > 0) cb(rect)
})

// 检查是否可滚动
const checkScrollable = () => {
  if (tableWrapperRef.value) {
    isScrollable.value = tableWrapperRef.value.scrollWidth > tableWrapperRef.value.clientWidth
  }
}

// 检查操作列是否需要展开
const checkActionsColumnWidth = () => {
  if (!props.expandableActions) {
    actionsColumnNeedsExpanding.value = false
    actionsExpanded.value = false
    return
  }
  if (!tableWrapperRef.value) return

  // 查找第一行的操作列单元格
  const firstActionCell = tableWrapperRef.value.querySelector('tbody tr:first-child td:last-child')
  if (!firstActionCell) return

  // 查找操作列内容的容器div
  const actionsContainer = firstActionCell.querySelector('div')
  if (!actionsContainer) return

  // 临时展开以测量完整宽度
  const wasExpanded = actionsExpanded.value
  actionsExpanded.value = true

  // 等待DOM更新
  nextTick(() => {
    // 测量所有按钮的总宽度
    const actionItems = actionsContainer.querySelectorAll('button, a, [role="button"]')
    if (actionItems.length <= 2) {
      actionsColumnNeedsExpanding.value = false
      actionsExpanded.value = wasExpanded
      return
    }

    // 计算所有按钮的总宽度（包括gap）
    let totalWidth = 0
    actionItems.forEach((item, index) => {
      totalWidth += (item as HTMLElement).offsetWidth
      if (index < actionItems.length - 1) {
        totalWidth += 4 // gap-1 = 4px
      }
    })

    // 获取单元格可用宽度（减去padding）
    const cellWidth = (firstActionCell as HTMLElement).clientWidth - 32 // 减去左右padding

    // 如果总宽度超过可用宽度，需要展开功能
    actionsColumnNeedsExpanding.value = totalWidth > cellWidth

    // 恢复原来的展开状态
    actionsExpanded.value = wasExpanded
  })
}

// 监听尺寸变化
let resizeObserver: ResizeObserver | null = null
let resizeHandler: (() => void) | null = null
let desktopViewportMediaQuery: MediaQueryList | null = null
let desktopViewportListener: ((event: MediaQueryListEvent) => void) | null = null

const detachDesktopTableTracking = () => {
  resizeObserver?.disconnect()
  resizeObserver = null
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
    resizeHandler = null
  }
}

const attachDesktopTableTracking = () => {
  checkScrollable()
  checkActionsColumnWidth()
  if (tableWrapperRef.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      checkScrollable()
      checkActionsColumnWidth()
    })
    resizeObserver.observe(tableWrapperRef.value)
  } else {
    // 降级方案：不支持 ResizeObserver 时使用 window resize
    resizeHandler = () => {
      checkScrollable()
      checkActionsColumnWidth()
    }
    window.addEventListener('resize', resizeHandler)
  }
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    desktopViewportMediaQuery = window.matchMedia(desktopViewportQuery)
    isDesktopViewport.value = desktopViewportMediaQuery.matches
    desktopViewportListener = (event: MediaQueryListEvent) => {
      isDesktopViewport.value = event.matches
    }
    if (typeof desktopViewportMediaQuery.addEventListener === 'function') {
      desktopViewportMediaQuery.addEventListener('change', desktopViewportListener)
    } else {
      desktopViewportMediaQuery.addListener(desktopViewportListener)
    }
  }
})

onUnmounted(() => {
  cleanupColumnResizeListeners()
  if (columnResizeFrame !== null && typeof window !== 'undefined') {
    window.cancelAnimationFrame(columnResizeFrame)
    columnResizeFrame = null
  }
  detachDesktopTableTracking()
  if (desktopViewportMediaQuery && desktopViewportListener) {
    if (typeof desktopViewportMediaQuery.removeEventListener === 'function') {
      desktopViewportMediaQuery.removeEventListener('change', desktopViewportListener)
    } else {
      desktopViewportMediaQuery.removeListener(desktopViewportListener)
    }
    desktopViewportListener = null
  }
  desktopViewportMediaQuery = null
})

interface Props {
  columns: Column[]
  data: any[]
  loading?: boolean
  stickyFirstColumn?: boolean
  stickyActionsColumn?: boolean
  expandableActions?: boolean
  actionsCount?: number // 操作按钮总数，用于判断是否需要展开功能
  rowKey?: string | ((row: any) => string | number)
  /**
   * Default sort configuration (only applied when there is no persisted sort state)
   */
  defaultSortKey?: string
  defaultSortOrder?: 'asc' | 'desc'
  /**
   * Persist sort state (key + order) to localStorage using this key.
   * If provided, DataTable will load the stored sort state on mount.
   */
  sortStorageKey?: string
  /** Enable desktop column resizing (default true). */
  resizableColumns?: boolean
  /**
   * Persist resized column widths to localStorage. When omitted, a key is
   * derived from sortStorageKey when available; otherwise widths last for the
   * lifetime of this component instance.
   */
  columnWidthStorageKey?: string
  /** Smallest user-selected column width in pixels (default 72). */
  minColumnWidth?: number
  /**
   * Enable server-side sorting mode. When true, clicking sort headers
   * will emit 'sort' events instead of performing client-side sorting.
   */
  serverSideSort?: boolean
  /** Emit 'rowClick' on row/card click and show pointer cursor (interactive cells should @click.stop) */
  clickableRows?: boolean
  /** Estimated row height in px for the virtualizer (default 56) */
  estimateRowHeight?: number
  /** Number of rows to render beyond the visible area (default 5) */
  overscan?: number
  /**
   * Only virtualize when the row count exceeds this threshold (default 100).
   * Smaller lists render in full, avoiding the scroll-compensation jank caused by
   * estimated-vs-actual row heights when rows have variable height.
   */
  virtualizeThreshold?: number
  /** Enable controlled row selection. Stable row keys are strongly recommended. */
  selectable?: boolean
  /** Selected row keys. Keys outside the current data page are preserved. */
  selectedKeys?: Array<string | number>
  /** Accessible label for a row selection checkbox. */
  selectionLabel?: string | ((row: any) => string)
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  stickyFirstColumn: true,
  stickyActionsColumn: true,
  expandableActions: true,
  defaultSortOrder: 'asc',
  serverSideSort: false,
  resizableColumns: true,
  minColumnWidth: 72,
  selectable: false,
  selectedKeys: () => []
})

const hasCompletedInitialLoad = ref(props.data.length > 0)
const hasObservedLoading = ref(props.loading)

watch(
  () => props.loading,
  (loading) => {
    if (loading) {
      hasObservedLoading.value = true
      return
    }
    if (hasObservedLoading.value) {
      hasCompletedInitialLoad.value = true
    }
  },
  { immediate: true }
)

const isInitialLoading = computed(() => props.loading && !hasCompletedInitialLoad.value)
const isRefreshing = computed(() => props.loading && hasCompletedInitialLoad.value)

const sortKey = ref<string>('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const actionsExpanded = ref(false)
const columnWidths = ref<Record<string, number>>({})
const resizingColumnKey = ref('')

type ActiveColumnResize = {
  key: string
  pointerId: number
  startX: number
  startWidth: number
  minWidth: number
  maxWidth: number
}

let activeColumnResize: ActiveColumnResize | null = null
let columnResizeFrame: number | null = null

const resolvedColumnWidthStorageKey = computed(() => {
  if (props.columnWidthStorageKey) return props.columnWidthStorageKey
  return props.sortStorageKey ? `${props.sortStorageKey}:column-widths` : ''
})

const hasCustomColumnWidth = (key: string) => Number.isFinite(columnWidths.value[key])

const isColumnResizable = (column: Column) =>
  props.resizableColumns && column.resizable !== false && column.key !== 'select'

const readPersistedColumnWidths = () => {
  const storageKey = resolvedColumnWidthStorageKey.value
  if (!storageKey || typeof window === 'undefined') return {}

  try {
    const raw = localStorage.getItem(storageKey)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as Record<string, unknown>
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}

    return Object.fromEntries(
      Object.entries(parsed).filter(([key, value]) =>
        key !== '__proto__'
        && key !== 'constructor'
        && key !== 'prototype'
        && typeof value === 'number'
        && Number.isFinite(value)
        && value >= 40
        && value <= 2400
      )
    ) as Record<string, number>
  } catch (e) {
    console.error('[DataTable] Failed to read persisted column widths:', e)
    return {}
  }
}

const writePersistedColumnWidths = () => {
  const storageKey = resolvedColumnWidthStorageKey.value
  if (!storageKey || typeof window === 'undefined') return

  try {
    localStorage.setItem(storageKey, JSON.stringify(columnWidths.value))
  } catch (e) {
    console.error('[DataTable] Failed to persist column widths:', e)
  }
}

const resolveCssLengthInPixels = (
  value: string | undefined,
  context: HTMLElement,
  fallback: number
) => {
  if (!value || typeof document === 'undefined') return fallback

  const probe = document.createElement('div')
  probe.style.position = 'absolute'
  probe.style.visibility = 'hidden'
  probe.style.pointerEvents = 'none'
  probe.style.width = value
  probe.style.padding = '0'
  probe.style.border = '0'
  context.appendChild(probe)
  const width = probe.getBoundingClientRect().width
  probe.remove()
  return Number.isFinite(width) && width > 0 ? width : fallback
}

const getColumnResizeBounds = (column: Column, headerCell: HTMLElement) => {
  const minWidth = Math.max(
    40,
    resolveCssLengthInPixels(column.minWidth, headerCell, props.minColumnWidth)
  )
  const configuredMax = resolveCssLengthInPixels(column.maxWidth, headerCell, Number.POSITIVE_INFINITY)
  return {
    minWidth,
    maxWidth: Math.max(minWidth, configuredMax),
  }
}

const clampColumnWidth = (width: number, minWidth: number, maxWidth: number) =>
  Math.round(Math.min(maxWidth, Math.max(minWidth, width)))

const refreshTableAfterColumnResize = () => {
  if (columnResizeFrame !== null || typeof window === 'undefined') return
  columnResizeFrame = window.requestAnimationFrame(() => {
    columnResizeFrame = null
    checkScrollable()
  })
}

const setColumnWidth = (key: string, width: number) => {
  columnWidths.value = { ...columnWidths.value, [key]: width }
  refreshTableAfterColumnResize()
}

const handleColumnResizeMove = (event: PointerEvent) => {
  const resize = activeColumnResize
  if (!resize || event.pointerId !== resize.pointerId) return
  event.preventDefault()
  setColumnWidth(
    resize.key,
    clampColumnWidth(
      resize.startWidth + event.clientX - resize.startX,
      resize.minWidth,
      resize.maxWidth
    )
  )
}

const finishColumnResize = () => {
  if (!activeColumnResize) return
  cleanupColumnResizeListeners()
  writePersistedColumnWidths()
  nextTick(() => {
    checkScrollable()
    checkActionsColumnWidth()
    rowVirtualizer.value.measure()
  })
}

const cleanupColumnResizeListeners = () => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('pointermove', handleColumnResizeMove)
    window.removeEventListener('pointerup', finishColumnResize)
    window.removeEventListener('pointercancel', finishColumnResize)
  }
  if (typeof document !== 'undefined') {
    document.documentElement.classList.remove('data-table-column-resizing')
  }
  activeColumnResize = null
  resizingColumnKey.value = ''
}

const startColumnResize = (event: PointerEvent, column: Column) => {
  if (!isColumnResizable(column) || event.button !== 0) return
  const headerCell = (event.currentTarget as HTMLElement).closest('th') as HTMLElement | null
  if (!headerCell) return

  cleanupColumnResizeListeners()
  const bounds = getColumnResizeBounds(column, headerCell)
  activeColumnResize = {
    key: column.key,
    pointerId: event.pointerId,
    startX: event.clientX,
    startWidth: headerCell.getBoundingClientRect().width,
    ...bounds,
  }
  resizingColumnKey.value = column.key
  document.documentElement.classList.add('data-table-column-resizing')
  window.addEventListener('pointermove', handleColumnResizeMove, { passive: false })
  window.addEventListener('pointerup', finishColumnResize)
  window.addEventListener('pointercancel', finishColumnResize)
}

const resetColumnWidth = (key: string) => {
  if (!hasCustomColumnWidth(key)) return
  const next = { ...columnWidths.value }
  delete next[key]
  columnWidths.value = next
  writePersistedColumnWidths()
  nextTick(() => {
    checkScrollable()
    checkActionsColumnWidth()
    rowVirtualizer.value.measure()
  })
}

const resetColumnWidths = () => {
  if (Object.keys(columnWidths.value).length === 0) return
  columnWidths.value = {}
  writePersistedColumnWidths()
  nextTick(() => {
    checkScrollable()
    checkActionsColumnWidth()
    rowVirtualizer.value.measure()
  })
}

const handleColumnResizeKeydown = (event: KeyboardEvent, column: Column) => {
  if (event.key === 'Enter' || event.key === 'Home') {
    event.preventDefault()
    event.stopPropagation()
    resetColumnWidth(column.key)
    return
  }
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return

  event.preventDefault()
  event.stopPropagation()

  const handle = event.currentTarget as HTMLElement
  const headerCell = handle.closest('th') as HTMLElement | null
  if (!headerCell) return
  const bounds = getColumnResizeBounds(column, headerCell)
  const currentWidth = columnWidths.value[column.key] ?? headerCell.getBoundingClientRect().width
  const direction = event.key === 'ArrowLeft' ? -1 : 1
  const step = event.shiftKey ? 32 : 8
  setColumnWidth(
    column.key,
    clampColumnWidth(currentWidth + direction * step, bounds.minWidth, bounds.maxWidth)
  )
  writePersistedColumnWidths()
  nextTick(() => rowVirtualizer.value.measure())
}

type PersistedSortState = {
  key: string
  order: 'asc' | 'desc'
}

const collator = new Intl.Collator(undefined, {
  numeric: true,
  sensitivity: 'base'
})

const getSortableKeys = () => {
  const keys = new Set<string>()
  for (const col of props.columns) {
    if (col.sortable) keys.add(col.key)
  }
  return keys
}

const normalizeSortKey = (candidate: string) => {
  if (!candidate) return ''
  const sortableKeys = getSortableKeys()
  return sortableKeys.has(candidate) ? candidate : ''
}

const normalizeSortOrder = (candidate: any): 'asc' | 'desc' => {
  return candidate === 'desc' ? 'desc' : 'asc'
}

const readPersistedSortState = (): PersistedSortState | null => {
  if (!props.sortStorageKey) return null
  try {
    const raw = localStorage.getItem(props.sortStorageKey)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<PersistedSortState>
    const key = normalizeSortKey(typeof parsed.key === 'string' ? parsed.key : '')
    if (!key) return null
    return { key, order: normalizeSortOrder(parsed.order) }
  } catch (e) {
    console.error('[DataTable] Failed to read persisted sort state:', e)
    return null
  }
}

const writePersistedSortState = (state: PersistedSortState) => {
  if (!props.sortStorageKey) return
  try {
    localStorage.setItem(props.sortStorageKey, JSON.stringify(state))
  } catch (e) {
    console.error('[DataTable] Failed to persist sort state:', e)
  }
}

const clearSort = () => {
  sortKey.value = ''
  sortOrder.value = 'asc'
  if (!props.sortStorageKey) return
  try {
    localStorage.removeItem(props.sortStorageKey)
  } catch (e) {
    console.error('[DataTable] Failed to clear persisted sort state:', e)
  }
}

const setSort = (key: string, order: 'asc' | 'desc') => {
  const normalizedKey = normalizeSortKey(key)
  if (!normalizedKey) return
  sortKey.value = normalizedKey
  sortOrder.value = normalizeSortOrder(order)
}

const resolveInitialSortState = (): PersistedSortState | null => {
  const persisted = readPersistedSortState()
  if (persisted) return persisted

  const key = normalizeSortKey(props.defaultSortKey || '')
  if (!key) return null
  return { key, order: normalizeSortOrder(props.defaultSortOrder) }
}

const applySortState = (state: PersistedSortState | null) => {
  if (!state) return
  sortKey.value = state.key
  sortOrder.value = state.order
}

const getSortIndicatorClass = (key: string, order: 'asc' | 'desc') => {
  return sortKey.value === key && sortOrder.value === order
    ? 'is-active'
    : ''
}

const getColumnAriaSort = (key: string) => {
  if (sortKey.value !== key) return 'none'
  return sortOrder.value === 'asc' ? 'ascending' : 'descending'
}

const getColumnAlignmentClass = (column: Column) =>
  `data-table__cell--align-${column.align || 'start'}`

const getColumnStyle = (column: Column) => {
  const resizedWidth = columnWidths.value[column.key]
  if (Number.isFinite(resizedWidth)) {
    const width = `${resizedWidth}px`
    return { width, minWidth: width, maxWidth: width }
  }

  return {
    width: column.width,
    minWidth: column.minWidth,
    maxWidth: column.maxWidth,
  }
}

const isNullishOrEmpty = (value: any) => value === null || value === undefined || value === ''

const toFiniteNumberOrNull = (value: any): number | null => {
  if (typeof value === 'number') return Number.isFinite(value) ? value : null
  if (typeof value === 'boolean') return value ? 1 : 0
  if (typeof value === 'string') {
    const trimmed = value.trim()
    if (!trimmed) return null
    const n = Number(trimmed)
    return Number.isFinite(n) ? n : null
  }
  return null
}

const toSortableString = (value: any): string => {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)
  if (value instanceof Date) return value.toISOString()
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

const compareSortValues = (a: any, b: any): number => {
  const aEmpty = isNullishOrEmpty(a)
  const bEmpty = isNullishOrEmpty(b)
  if (aEmpty && bEmpty) return 0
  if (aEmpty) return 1
  if (bEmpty) return -1

  const aNum = toFiniteNumberOrNull(a)
  const bNum = toFiniteNumberOrNull(b)
  if (aNum !== null && bNum !== null) {
    if (aNum === bNum) return 0
    return aNum < bNum ? -1 : 1
  }

  const aStr = toSortableString(a)
  const bStr = toSortableString(b)
  const res = collator.compare(aStr, bStr)
  if (res === 0) return 0
  return res < 0 ? -1 : 1
}
const resolveStableRowKey = (row: any): string | number | undefined => {
  if (typeof props.rowKey === 'function') {
    const key = props.rowKey(row)
    return key ?? undefined
  }
  if (typeof props.rowKey === 'string' && props.rowKey) {
    const key = row?.[props.rowKey]
    return key ?? undefined
  }
  const key = row?.id
  return key ?? undefined
}

const resolveRowKey = (row: any, index: number) => resolveStableRowKey(row) ?? index

const dataColumns = computed(() => props.columns.filter((column) => column.key !== 'actions'))
const columnsSignature = computed(() =>
  props.columns.map((column) => `${column.key}:${column.sortable ? '1' : '0'}`).join('|')
)

watch(
  isDesktopViewport,
  async (isDesktop) => {
    detachDesktopTableTracking()
    if (!isDesktop) return
    await nextTick()
    attachDesktopTableTracking()
  },
  { immediate: true, flush: 'post' }
)

// 数据/列变化时重新检查滚动状态
// 注意：不能监听 actionsExpanded，因为 checkActionsColumnWidth 会临时修改它，会导致无限循环
watch(
  [() => props.data.length, columnsSignature],
  async () => {
    await nextTick()
    checkScrollable()
    checkActionsColumnWidth()
  },
  { flush: 'post' }
)

// 单独监听展开状态变化，只更新滚动状态
watch(actionsExpanded, async () => {
  await nextTick()
  checkScrollable()
})

const handleSort = (key: string) => {
  let newOrder: 'asc' | 'desc' = 'asc'
  if (sortKey.value === key) {
    newOrder = sortOrder.value === 'asc' ? 'desc' : 'asc'
  }

  if (props.serverSideSort) {
    // Server-side sort mode: emit event and update internal state for UI feedback
    sortKey.value = key
    sortOrder.value = newOrder
    emit('sort', key, newOrder)
  } else {
    // Client-side sort mode: just update internal state
    sortKey.value = key
    sortOrder.value = newOrder
  }
}

const sortedData = computed(() => {
  // Server-side sort mode: return data as-is (server handles sorting)
  if (props.serverSideSort || !sortKey.value || !props.data) return props.data

  const key = sortKey.value
  const order = sortOrder.value

  // Stable sort (tie-break with original index) to avoid jitter when values are equal.
  return props.data
    .map((row, index) => ({ row, index }))
    .sort((a, b) => {
      const cmp = compareSortValues(a.row?.[key], b.row?.[key])
      if (cmp !== 0) return order === 'asc' ? cmp : -cmp
      return a.index - b.index
    })
    .map(item => item.row)
})

const tableColumnCount = computed(() => props.columns.length + (props.selectable ? 1 : 0))
const selectedKeySet = computed(() => new Set(props.selectedKeys))
const visibleRowKeys = computed(() =>
  (sortedData.value ?? []).map((row, index) => resolveRowKey(row, index))
)
const allVisibleSelected = computed(() =>
  visibleRowKeys.value.length > 0
  && visibleRowKeys.value.every((key) => selectedKeySet.value.has(key))
)
const someVisibleSelected = computed(() => {
  if (allVisibleSelected.value) return false
  return visibleRowKeys.value.some((key) => selectedKeySet.value.has(key))
})

const emitSelection = (next: Set<string | number>) => {
  const keys = Array.from(next)
  emit('update:selectedKeys', keys)
  emit('selectionChange', keys)
}

const isRowSelected = (row: any, index: number) =>
  selectedKeySet.value.has(resolveRowKey(row, index))

const getRowSelectionLabel = (row: any, index: number) => {
  if (typeof props.selectionLabel === 'function') return props.selectionLabel(row)
  if (props.selectionLabel) return props.selectionLabel
  return `${t('common.selectOption')} ${resolveRowKey(row, index)}`
}

const toggleRowSelection = (row: any, index: number, checked: boolean) => {
  const next = new Set(props.selectedKeys)
  const key = resolveRowKey(row, index)
  if (checked) next.add(key)
  else next.delete(key)
  emitSelection(next)
}

const toggleAllVisible = (checked: boolean) => {
  const next = new Set(props.selectedKeys)
  for (const key of visibleRowKeys.value) {
    if (checked) next.add(key)
    else next.delete(key)
  }
  emitSelection(next)
}

// --- Virtual scrolling ---
// 是否启用虚拟化:仅桌面端且行数超过阈值时开启。小列表全量渲染,彻底绕开虚拟器的
// 估算/测量/滚动补偿链路,消除可变行高导致的滚动抖动。
const shouldVirtualize = computed(() =>
  isDesktopViewport.value && (sortedData.value?.length ?? 0) > (props.virtualizeThreshold ?? 100)
)

const rowVirtualizer = useVirtualizer(computed(() => ({
  count: shouldVirtualize.value ? (sortedData.value?.length ?? 0) : 0,
  getScrollElement: () => tableWrapperRef.value,
  // 用行主键(与模板 :key 一致)而非默认的 index 作为 itemSizeCache 键,
  // 这样排序/筛选/跨阈值来回都能复用正确的已测行高,而不是残留的按 index 缓存 → 消除高度校正抖动。
  getItemKey: (index: number) => {
    const row = sortedData.value?.[index]
    return row != null ? resolveRowKey(row, index) : index
  },
  estimateSize: () => props.estimateRowHeight ?? 56,
  overscan: props.overscan ?? 5,
  // 兜底高度:首个有效高度读数到来前,先按一屏渲染,避免空白帧
  initialRect: { width: 0, height: estimatedViewportHeight() },
  // 关键:过滤 0 高度读数,杜绝 scrollRect 被钉成 0 → calculateRange 返回 null → 整表空白
  observeElementRect: observeElementRectNonZero,
  // 把测量类 ResizeObserver 回调批到 rAF,避免滚动中同步 reflow 风暴导致的校正抖动/空白
  useAnimationFrameWithResizeObserver: true,
})))

const virtualItems = computed(() => rowVirtualizer.value.getVirtualItems())

const virtualPaddingTop = computed(() => {
  const items = virtualItems.value
  return items.length > 0 ? items[0].start : 0
})

const virtualPaddingBottom = computed(() => {
  const items = virtualItems.value
  if (items.length === 0) return 0
  return rowVirtualizer.value.getTotalSize() - items[items.length - 1].end
})

const measureElement = (el: any) => {
  if (el) {
    rowVirtualizer.value.measureElement(el as Element)
  }
}

type RowIdentityToken = string | number | object | symbol

const rowIdentityKeys = computed<RowIdentityToken[]>(() =>
  (sortedData.value ?? []).map((row) => {
    const stableKey = resolveStableRowKey(row)
    if (stableKey !== undefined) return stableKey

    // Object references survive pure reordering but change across page/filter results.
    // Primitive rows have no stable identity, so force conservative invalidation.
    return row !== null && typeof row === 'object' ? row : Symbol('unstable-row')
  })
)

const hasSameRowIdentitySet = (
  current: RowIdentityToken[],
  previous: RowIdentityToken[]
) => {
  if (current.length !== previous.length) return false
  const currentKeys = new Set(current)
  const previousKeys = new Set(previous)
  // Duplicate keys make row-to-cache ownership ambiguous, even when the unique
  // key set looks unchanged (for example [1, 1, 2] -> [1, 2, 2]).
  if (currentKeys.size !== current.length || previousKeys.size !== previous.length) return false
  return [...currentKeys].every(key => previousKeys.has(key))
}

watch(
  rowIdentityKeys,
  (current, previous) => {
    if (hasSameRowIdentitySet(current, previous)) return

    // The virtualizer owns caches across option updates. A new page/filter result
    // must release detached rows and sizes, while pure reordering keeps them.
    rowVirtualizer.value.measureElement(null)
    rowVirtualizer.value.measure()
  },
  { flush: 'post' }
)

// 统一的渲染行列表:虚拟化开启时只取窗口内的行(需 measure 交给虚拟器测量),
// 关闭时取全部行(无需测量)。模板据此渲染,两种模式共用同一套单元格结构。
const renderRows = computed<Array<{ index: number; row: any; measure: boolean }>>(() => {
  const data = sortedData.value ?? []
  if (shouldVirtualize.value) {
    return virtualItems.value.map(vr => ({ index: vr.index, row: data[vr.index], measure: true }))
  }
  return data.map((row, index) => ({ index, row, measure: false }))
})

const hasActionsColumn = computed(() => {
  return props.columns.some(column => column.key === 'actions')
})

const hasSelectColumn = computed(() => {
  return props.columns.length > 0 && props.columns[0].key === 'select'
})

// 生成固定列的 CSS 类
const getStickyColumnClass = (column: Column, index: number) => {
  const classes: string[] = []

  if (props.stickyFirstColumn) {
    // 如果第一列是勾选列，固定前两列（勾选+名称）
    if (hasSelectColumn.value) {
      if (index === 0) {
        classes.push('sticky-col sticky-col-left-first')
      } else if (index === 1) {
        classes.push('sticky-col sticky-col-left-second')
      }
    } else {
      // 否则只固定第一列
      if (index === 0) {
        classes.push('sticky-col sticky-col-left')
      }
    }
  }

  // 操作列固定（最后一列）
  if (props.stickyActionsColumn && column.key === 'actions') {
    classes.push('sticky-col sticky-col-right')
  }

  return classes.join(' ')
}

// 根据列数自适应调整内边距
const getAdaptivePaddingClass = () => {
  const columnCount = props.columns.length

  if (columnCount >= 10) {
    return 'data-table__cell--padding-compact'
  } else if (columnCount >= 7) {
    return 'data-table__cell--padding-condensed'
  } else if (columnCount >= 5) {
    return 'data-table__cell--padding-default'
  } else {
    return 'data-table__cell--padding-comfortable'
  }
}

// Init + keep persisted sort state consistent with current columns
const didInitSort = ref(false)

onMounted(() => {
  const initial = resolveInitialSortState()
  applySortState(initial)
  didInitSort.value = true
})

onMounted(() => {
  columnWidths.value = readPersistedColumnWidths()
  nextTick(() => {
    checkScrollable()
    rowVirtualizer.value.measure()
  })
})

watch(resolvedColumnWidthStorageKey, () => {
  columnWidths.value = readPersistedColumnWidths()
  nextTick(() => {
    checkScrollable()
    rowVirtualizer.value.measure()
  })
})

watch(
  columnsSignature,
  () => {
    // If current sort key is no longer sortable/visible, fall back to default/persisted.
    const normalized = normalizeSortKey(sortKey.value)
    if (!sortKey.value) {
      const initial = resolveInitialSortState()
      applySortState(initial)
      return
    }

    if (!normalized) {
      const fallback = resolveInitialSortState()
      if (fallback) {
        applySortState(fallback)
      } else {
        sortKey.value = ''
        sortOrder.value = 'asc'
      }
    }
  },
  { flush: 'post' }
)

watch(
  [sortKey, sortOrder],
  ([nextKey, nextOrder]) => {
    if (!didInitSort.value) return
    if (!props.sortStorageKey) return
    const key = normalizeSortKey(nextKey)
    if (!key) return
    writePersistedSortState({ key, order: normalizeSortOrder(nextOrder) })
  },
  { flush: 'post' }
)

defineExpose({
  virtualizer: rowVirtualizer,
  shouldVirtualize,
  sortedData,
  resolveRowKey,
  tableWrapperEl: tableWrapperRef,
  resetColumnWidth,
  resetColumnWidths,
  clearSort,
  setSort,
})
</script>

<style scoped lang="scss">
.data-table {
  --data-table-select-width: 3.25rem;
  --data-table-sticky-blur: 140px;
  position: relative;
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-xl);
  background: var(--glass-layer-content-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate, 1.8));
  backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate, 1.8));
  box-shadow:
    var(--glass-shadow),
    0 1px 0 var(--glass-highlight) inset;
  color: var(--color-text-primary);

  &--refreshing {
    .data-table__viewport,
    .data-table__mobile-list {
      pointer-events: none;
      user-select: none;
    }

    .data-table__body,
    .data-table__mobile-card,
    .data-table__mobile-selection {
      opacity: 0.48;
    }
  }

  &__loading-overlay {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 250;
    display: grid;
    min-height: 6.5rem;
    place-items: center;
    border-radius: 0 0 var(--radius-xl) var(--radius-xl);
    background-color: var(--glass-layer-content-bg);
    -webkit-backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate));
    backdrop-filter: blur(var(--glass-layer-content-blur)) saturate(var(--glass-saturate));
    cursor: wait;
  }

  &__loading-overlay--desktop {
    top: 2.75rem;
  }

  &__loading-overlay--mobile {
    top: 0;
    border-radius: var(--radius-lg, 0.5rem);
  }

  &__loading-state {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2rem;
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md, 0.5rem);
    background: var(--glass-layer-inset-bg);
    box-shadow: var(--glass-shadow), 0 1px 0 var(--glass-highlight) inset;
    color: var(--color-text-secondary);
    font-size: var(--type-caption-size);
    line-height: 1;
    -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
    backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  }

  &__spinner {
    width: 0.875rem;
    height: 0.875rem;
    flex: 0 0 auto;
    border: 2px solid var(--color-border-strong);
    border-top-color: transparent;
    border-radius: 50%;
    animation: data-table-spin 0.7s linear infinite;
  }

  &__viewport {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow: auto;
    isolation: isolate;
    scrollbar-width: auto;

    &::-webkit-scrollbar {
      width: 0.75rem;
      height: 0.75rem;
      display: block;
      background: transparent;
    }

    &::-webkit-scrollbar-track {
      margin: 0 0.25rem;
      border-radius: 0.375rem;
      background: rgba(12, 12, 14, 0.04);
    }

    &::-webkit-scrollbar-thumb {
      border: 2px solid transparent;
      border-radius: 0.375rem;
      background: rgba(107, 114, 128, 0.7);
      background-clip: padding-box;

      &:hover {
        background-color: rgba(75, 85, 99, 0.9);
      }
    }
  }

  &__table {
    width: 100%;
    min-width: max-content;
    border-spacing: 0;
    border-collapse: separate;
    text-align: left;
  }

  &__selection-column {
    width: var(--data-table-select-width);
  }

  &__header {
    position: sticky;
    top: 0;
    z-index: 200;
    background: transparent;
  }

  // 玻璃表头:半透明底 + 背景模糊,滚动时下方内容微透(与导航壳同一材质语言)
  &__header-cell {
    position: sticky;
    top: 0;
    z-index: 210;
    height: 2.75rem;
    border-bottom: 1px solid var(--color-border);
    background: var(--glass-layer-shell-bg);
    -webkit-backdrop-filter: blur(var(--glass-layer-shell-blur)) saturate(var(--glass-saturate, 1.8));
    backdrop-filter: blur(var(--glass-layer-shell-blur)) saturate(var(--glass-saturate, 1.8));
    color: var(--color-text-secondary);
    font-size: var(--font-size-xs);
    font-weight: 600;
    line-height: 1rem;
    white-space: nowrap;
    letter-spacing: 0;
    transition: background-color 120ms ease, color 120ms ease;

    &.sticky-col {
      background: var(--glass-layer-modal-bg);
      -webkit-backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate, 1.8));
      backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate, 1.8));
    }

    &.is-being-resized {
      background: var(--color-primary-subtle);
    }

    &.is-sortable {
      cursor: pointer;

      &:hover {
        background: var(--glass-layer-shell-bg);
        -webkit-backdrop-filter: blur(var(--glass-layer-shell-blur-hover)) saturate(var(--glass-saturate-hover));
        backdrop-filter: blur(var(--glass-layer-shell-blur-hover)) saturate(var(--glass-saturate-hover));
        color: var(--color-text-primary);
      }

      &.sticky-col:hover {
        background: var(--glass-layer-modal-bg);
        -webkit-backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate-hover));
        backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate-hover));
      }
    }
  }

  &__header-content {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    min-width: 0;
  }

  &__resize-handle {
    position: absolute;
    top: 0;
    right: -0.3125rem;
    bottom: 0;
    z-index: 3;
    width: 0.625rem;
    padding: 0;
    border: 0;
    background: transparent;
    cursor: col-resize;
    touch-action: none;

    &::before {
      content: '';
      position: absolute;
      top: 0.5rem;
      right: calc(50% - 0.5px);
      bottom: 0.5rem;
      width: 1px;
      border-radius: 1px;
      background: var(--color-border-strong);
      opacity: 0;
      transition: opacity 120ms ease, background-color 120ms ease;
    }

    &:hover::before,
    &:focus-visible::before,
    &.is-active::before {
      background: color-mix(in srgb, var(--theme-accent) 72%, transparent);
      opacity: 1;
    }

    &:focus-visible {
      outline: 2px solid var(--color-primary);
      outline-offset: -2px;
    }
  }

  &__body {
    position: relative;
    z-index: 0;
    transition: opacity 160ms ease;
  }

  &__row {
    color: var(--color-text-primary);
    transition: background-color 120ms ease;

    &:hover {
      .data-table__body-cell {
        background: var(--glass-bg-interactive);
        -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
        backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
      }

      .data-table__body-cell.sticky-col {
        background: var(--glass-layer-modal-bg);
        -webkit-backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate-hover));
        backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate-hover));
      }
    }

    &.is-clickable {
      cursor: pointer;
    }

    &.is-selected {
      .data-table__body-cell {
        background: var(--glass-tint-brand);
      }

      .data-table__body-cell.sticky-col {
        background: color-mix(in srgb, var(--glass-layer-modal-bg) 82%, var(--theme-accent));
      }

      // 行首主色指示条(与侧边栏激活态同语言)
      .data-table__body-cell:first-child {
        box-shadow: inset 3px 0 0 var(--color-primary);
      }
    }
  }

  // 单元格默认透明，让玻璃卡底透出；sticky 列使用 L2 半透明材质遮挡滚动内容。
  &__body-cell {
    height: 3.25rem;
    border-bottom: 1px solid var(--color-border-subtle);
    background: transparent;
    color: inherit;
      font-size: var(--type-control-size);
    font-variant-numeric: tabular-nums;
    vertical-align: middle;
    white-space: nowrap;

    // sticky 列用近不透明磨砂挡住横向滚过的单元格
    &.sticky-col {
      background: var(--glass-layer-modal-bg);
      -webkit-backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate, 1.8));
      backdrop-filter: blur(var(--data-table-sticky-blur)) saturate(var(--glass-saturate, 1.8));
    }

    &.is-resized {
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  // 卡壳内最后一行不需要分隔线,收边更干净
  &__body tr:last-child .data-table__body-cell {
    border-bottom: 0;
  }

  &__selection-cell {
    width: var(--data-table-select-width);
    min-width: var(--data-table-select-width);
    padding: 0 1rem;
    text-align: center;
  }

  &__checkbox {
    width: 1rem;
    height: 1rem;
    margin: 0;
    border-radius: 0.25rem;
    accent-color: var(--color-primary);
    cursor: pointer;
  }

  &__cell {
    &--padding-compact {
      padding-right: 0.5rem;
      padding-left: 0.5rem;
    }

    &--padding-condensed {
      padding-right: 0.75rem;
      padding-left: 0.75rem;
    }

    &--padding-default {
      padding-right: 1rem;
      padding-left: 1rem;
    }

    &--padding-comfortable {
      padding-right: 1.5rem;
      padding-left: 1.5rem;
    }

    &--align-center {
      text-align: center;

      .data-table__header-content {
        justify-content: center;
      }
    }

    &--align-end {
      text-align: right;

      .data-table__header-content {
        justify-content: flex-end;
      }
    }
  }

  &__sort-indicator {
    display: inline-flex;
    flex-direction: column;
    color: var(--color-text-tertiary);
  }

  &__sort-arrow {
    width: 0.5rem;
    height: 0.5rem;

    &--descending {
      margin-top: -0.125rem;
    }

    &.is-active {
      color: var(--color-text-brand);
    }
  }

  &__empty-cell,
  &__mobile-empty {
    padding-top: 4rem;
    padding-bottom: 4rem;
    text-align: center;
  }

  &__empty-content {
    display: flex;
    align-items: center;
    flex-direction: column;
    gap: 0.75rem;
    color: var(--color-text-tertiary);
  }

  &__empty-icon {
    opacity: 0.7;
  }

  &__empty-title {
    margin: 0;
    font-size: var(--type-control-size);
  }

  &__skeleton-row .data-table__body-cell {
    pointer-events: none;
  }

  &__skeleton {
    display: inline-block;
    border-radius: 0.375rem;
    background: var(--color-skeleton);
    animation: data-table-pulse 1.4s ease-in-out infinite;

    &--cell {
      width: min(8rem, 80%);
      height: 0.875rem;
    }

    &--checkbox {
      width: 1rem;
      height: 1rem;
    }

    &--label {
      width: 4rem;
      height: 0.75rem;
    }

    &--value {
      width: 8rem;
      max-width: 45%;
      height: 0.875rem;
    }

    &--action {
      width: 5rem;
      height: 1.75rem;
    }
  }

  &__mobile-list {
    display: grid;
    gap: 0.75rem;
  }

  &__mobile-card,
  &__mobile-selection {
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg, 0.5rem);
    background: var(--glass-bg-thin);
    transition: opacity 160ms ease;
  }

  &__mobile-card {
    position: relative;
    display: grid;
    gap: 0.75rem;
    padding: 1rem;

    &.is-clickable {
      cursor: pointer;
    }

    &.is-selected {
      border-color: var(--color-primary-border);
      background: var(--color-primary-subtle);
      box-shadow: inset 3px 0 0 var(--color-primary);
    }
  }

  &__mobile-selection {
    padding: 0.75rem 1rem;
  }

  &__selection-label {
    display: inline-flex;
    align-items: center;
    gap: 0.625rem;
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
  }

  &__mobile-row-selection {
    position: absolute;
    top: 1rem;
    right: 1rem;
  }

  &__mobile-field {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  &__mobile-label {
    flex: 0 0 auto;
    color: var(--color-text-tertiary);
    font-size: var(--type-caption-size);
    font-weight: 500;
  }

  &__mobile-value {
    min-width: 0;
    color: var(--color-text-primary);
    font-size: var(--type-control-size);
    text-align: right;
    overflow-wrap: anywhere;
  }

  &__mobile-actions {
    display: flex;
    justify-content: flex-end;
    padding-top: 0.75rem;
    border-top: 1px solid var(--color-border-subtle);
  }
}

.sticky-col {
  position: sticky;
  z-index: 20;
}

.sticky-col-left,
.sticky-col-left-first {
  left: 0;
}

.sticky-col-left-second {
  left: var(--data-table-select-width);
}

.sticky-col-right {
  right: 0;
}

.sticky-header-cell.sticky-col {
  z-index: 220;
}

.is-scrollable {
  .sticky-col-left::after,
  .sticky-col-left-second::after,
  .sticky-col-right::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    width: 0.625rem;
    pointer-events: none;
  }

  .sticky-col-left::after,
  .sticky-col-left-second::after {
    right: 0;
    transform: translateX(100%);
    background: linear-gradient(to right, rgba(12, 12, 14, 0.08), transparent);
  }

  .sticky-col-right::before {
    left: 0;
    transform: translateX(-100%);
    background: linear-gradient(to left, rgba(12, 12, 14, 0.08), transparent);
  }
}

.dark .data-table {
  color: var(--color-text-primary);

  &__viewport::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.05);
  }

  &__viewport::-webkit-scrollbar-thumb {
    background-color: rgba(156, 163, 175, 0.7);

    &:hover {
      background-color: rgba(209, 213, 219, 0.9);
    }
  }

  &__header {
    background: transparent;
  }

  &__header-cell {
    border-color: var(--color-border);
  }

  &__row.is-selected .data-table__body-cell.sticky-col {
    background: color-mix(in srgb, var(--glass-layer-modal-bg) 82%, var(--theme-accent));
  }

  &__body-cell {
    border-color: var(--color-border);
  }

  &__mobile-actions {
    border-color: var(--color-border);
  }

  &__mobile-value {
    color: var(--color-text-primary);
  }

  &__skeleton {
    background: var(--color-skeleton);
  }

  .is-scrollable {
    .sticky-col-left::after,
    .sticky-col-left-second::after {
      background: linear-gradient(to right, rgba(0, 0, 0, 0.24), transparent);
    }

    .sticky-col-right::before {
      background: linear-gradient(to left, rgba(0, 0, 0, 0.24), transparent);
    }
  }
}

/* 位于 TablePageLayout 玻璃容器(.table-scroll-container)内时去壳,避免双层卡片。
   放在暗色规则之后,确保同时覆盖亮/暗两套卡壳。 */
.table-scroll-container .data-table,
.dark .table-scroll-container .data-table {
  border: 0;
  border-radius: 0;
  background: transparent;
  -webkit-backdrop-filter: none;
  backdrop-filter: none;
  box-shadow: none;
}

:global(html.data-table-column-resizing),
:global(html.data-table-column-resizing *) {
  cursor: col-resize !important;
  user-select: none !important;
}

@keyframes data-table-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes data-table-pulse {
  0%,
  100% {
    opacity: 0.55;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 767px) {
  /* 移动端为卡片列表形态,自带瓷片边框,外壳去掉 */
  .data-table,
  .dark .data-table {
    flex: none;
    min-height: auto;
    overflow: visible;
    border: 0;
    border-radius: 0;
    background: transparent;
    -webkit-backdrop-filter: none;
    backdrop-filter: none;
    box-shadow: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .data-table__spinner {
    animation: none;
  }

  .data-table__skeleton {
    animation-duration: 1ms;
    animation-iteration-count: 1;
  }

  .data-table__body,
  .data-table__mobile-card,
  .data-table__mobile-selection {
    transition-duration: 1ms;
  }
}
</style>
