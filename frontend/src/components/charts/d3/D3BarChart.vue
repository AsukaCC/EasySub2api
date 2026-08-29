<script setup lang="ts">
import { pointer, range, scaleBand, scaleLinear } from 'd3'
import { computed, ref, type CSSProperties } from 'vue'
import type {
  D3ChartData,
  D3ChartOptions,
  D3LineTooltipItem,
} from './chartTypes'
import { useChartSize } from './useChartSize'

interface AxisOptions {
  display?: boolean
  min?: number
  max?: number
  suggestedMax?: number
  beginAtZero?: boolean
  grid?: { display?: boolean; color?: string; borderDash?: number[] }
  ticks?: {
    color?: string
    maxTicksLimit?: number
    callback?: (value: string | number) => unknown
    font?: { size?: number }
  }
  title?: { display?: boolean; text?: string; color?: string; font?: { size?: number } }
}

interface BarOptions {
  plugins?: {
    legend?: {
      display?: boolean
      align?: string
      labels?: { color?: string; font?: { size?: number }; padding?: number }
    }
    tooltip?: {
      backgroundColor?: string
      titleColor?: string
      bodyColor?: string
      borderColor?: string
      borderWidth?: number
      displayColors?: boolean
      callbacks?: {
        title?: (items: D3LineTooltipItem[]) => unknown
        label?: (item: D3LineTooltipItem) => unknown
        footer?: (items: D3LineTooltipItem[]) => unknown
      }
    }
  }
  scales?: Record<string, AxisOptions>
}

interface BarModel {
  datasetIndex: number
  dataIndex: number
  x: number
  y: number
  width: number
  height: number
  value: number
  color: string
  radius: number
}

const props = withDefaults(defineProps<{
  data: D3ChartData
  options?: D3ChartOptions
  ariaLabel?: string
}>(), {
  options: () => ({}),
  ariaLabel: 'Bar chart',
})

const palette = ['#2563eb', '#059669', '#d97706', '#dc2626', '#7c3aed', '#0891b2']
const frameRef = ref<HTMLElement | null>(null)
const { width, height } = useChartSize(frameRef)
const hovered = ref<{ datasetIndex: number; dataIndex: number } | null>(null)
const tooltipX = ref(0)
const tooltipY = ref(0)
const chartOptions = computed(() => props.options as BarOptions)
const xOptions = computed(() => chartOptions.value.scales?.x ?? {})
const yOptions = computed(() => chartOptions.value.scales?.y ?? {})

const plot = computed(() => ({
  left: yOptions.value.title?.display ? 70 : 52,
  right: Math.max(53, width.value - 16),
  top: 10,
  bottom: Math.max(11, height.value - 36),
}))

const indices = computed(() => range(props.data.labels.length))
const xScale = computed(() => scaleBand<number>()
  .domain(indices.value)
  .range([plot.value.left, plot.value.right])
  .padding(0.18))
const groupScale = computed(() => scaleBand<number>()
  .domain(range(props.data.datasets.length))
  .range([0, xScale.value.bandwidth()])
  .padding(props.data.datasets.length > 1 ? 0.1 : 0))

const yScale = computed(() => {
  const values = props.data.datasets
    .flatMap((dataset) => dataset.data.map(numericValue))
    .filter((value): value is number => value !== null)
  const dataMin = values.length ? Math.min(...values) : 0
  const dataMax = values.length ? Math.max(...values) : 1
  let minimum = yOptions.value.min ?? ((yOptions.value.beginAtZero !== false && dataMin >= 0) ? 0 : dataMin)
  let maximum = yOptions.value.max ?? Math.max(dataMax, yOptions.value.suggestedMax ?? Number.NEGATIVE_INFINITY)
  if (!Number.isFinite(minimum)) minimum = 0
  if (!Number.isFinite(maximum)) maximum = 1
  if (minimum === maximum) maximum = minimum === 0 ? 1 : minimum + Math.abs(minimum * 0.1)
  const scale = scaleLinear().domain([minimum, maximum]).range([plot.value.bottom, plot.value.top])
  if (yOptions.value.min === undefined && yOptions.value.max === undefined) scale.nice(5)
  return scale
})

const bars = computed<BarModel[]>(() => props.data.datasets.flatMap((dataset, datasetIndex) => {
  const zeroY = yScale.value(Math.max(yScale.value.domain()[0], Math.min(0, yScale.value.domain()[1])))
  return indices.value.flatMap((dataIndex) => {
    const value = numericValue(dataset.data[dataIndex])
    if (value === null) return []
    const valueY = yScale.value(value)
    const baseX = xScale.value(dataIndex)
    const groupX = groupScale.value(datasetIndex)
    if (baseX === undefined || groupX === undefined) return []
    return [{
      datasetIndex,
      dataIndex,
      x: baseX + groupX,
      y: Math.min(valueY, zeroY),
      width: groupScale.value.bandwidth(),
      height: Math.max(1, Math.abs(zeroY - valueY)),
      value,
      color: colorAt(datasetIndex, dataIndex),
      radius: Number(dataset.borderRadius ?? 3),
    }]
  })
}))

const yTicks = computed(() => yScale.value.ticks(5))
const xTicks = computed(() => {
  const maximum = Math.max(2, xOptions.value.ticks?.maxTicksLimit ?? Math.floor((plot.value.right - plot.value.left) / 64))
  const step = Math.max(1, Math.ceil(indices.value.length / maximum))
  const ticks = indices.value.filter((_, index) => index % step === 0)
  const last = indices.value[indices.value.length - 1]
  if (last !== undefined && ticks[ticks.length - 1] !== last) ticks.push(last)
  return ticks
})

const showLegend = computed(() => chartOptions.value.plugins?.legend?.display !== false)
const legendStyle = computed<CSSProperties>(() => ({
  color: chartOptions.value.plugins?.legend?.labels?.color ?? '#64748b',
  justifyContent: chartOptions.value.plugins?.legend?.align === 'end' ? 'flex-end' : 'center',
  fontSize: `${chartOptions.value.plugins?.legend?.labels?.font?.size ?? 11}px`,
  gap: `${chartOptions.value.plugins?.legend?.labels?.padding ?? 12}px`,
}))

const tooltipItem = computed<D3LineTooltipItem | null>(() => {
  if (!hovered.value) return null
  const dataset = props.data.datasets[hovered.value.datasetIndex]
  const raw = dataset?.data[hovered.value.dataIndex]
  const value = numericValue(raw)
  if (!dataset || value === null) return null
  return {
    dataset,
    datasetIndex: hovered.value.datasetIndex,
    dataIndex: hovered.value.dataIndex,
    label: String(props.data.labels[hovered.value.dataIndex] ?? ''),
    raw,
    parsed: { x: hovered.value.dataIndex, y: value },
  }
})

const tooltipContent = computed(() => {
  const item = tooltipItem.value
  if (!item) return null
  const callbacks = chartOptions.value.plugins?.tooltip?.callbacks
  const customTitle = normalizeText(callbacks?.title?.([item]))
  const customLabel = normalizeText(callbacks?.label?.(item))
  return {
    title: customTitle.length ? customTitle : [item.label],
    body: customLabel.length ? customLabel : [`${item.dataset.label || 'Value'}: ${formatDefault(item.parsed.y ?? 0)}`],
    footer: normalizeText(callbacks?.footer?.([item])),
    color: colorAt(item.datasetIndex, item.dataIndex),
    displayColors: chartOptions.value.plugins?.tooltip?.displayColors !== false,
  }
})

const tooltipStyle = computed<CSSProperties>(() => {
  const options = chartOptions.value.plugins?.tooltip
  return {
    left: `${tooltipX.value}px`,
    top: `${tooltipY.value}px`,
    color: 'var(--color-text-secondary)',
    borderColor: options?.borderColor ?? 'var(--glass-border-hover)',
    borderWidth: `${options?.borderWidth ?? 1}px`,
  }
})

const tooltipTitleStyle = computed<CSSProperties>(() => ({
  color: 'var(--color-text-primary)',
}))

function numericValue(value: unknown): number | null {
  if (value === null || value === undefined || value === '') return null
  const result = Number(value)
  return Number.isFinite(result) ? result : null
}

function colorAt(datasetIndex: number, dataIndex: number) {
  const background = props.data.datasets[datasetIndex]?.backgroundColor
  if (Array.isArray(background)) return String(background[dataIndex] ?? palette[datasetIndex % palette.length])
  if (typeof background === 'string') return background
  return palette[datasetIndex % palette.length]
}

function normalizeText(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String).filter(Boolean)
  if (value === null || value === undefined || value === '') return []
  return [String(value)]
}

function formatDefault(value: number) {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function formatXTick(index: number) {
  const label = String(props.data.labels[index] ?? '')
  const callback = xOptions.value.ticks?.callback
  return callback ? String(callback(label)) : label
}

function formatYTick(value: number) {
  const callback = yOptions.value.ticks?.callback
  return callback ? String(callback(value)) : formatDefault(value)
}

function showTooltip(bar: BarModel, event: MouseEvent) {
  hovered.value = { datasetIndex: bar.datasetIndex, dataIndex: bar.dataIndex }
  if (!frameRef.value) return
  const [x, y] = pointer(event, frameRef.value)
  tooltipX.value = Math.max(8, Math.min(width.value - 190, x + 10))
  tooltipY.value = Math.max(8, Math.min(height.value - 70, y - 16))
}
</script>

<template>
  <div class="d3-bar-chart">
    <div v-if="showLegend" class="d3-bar-chart__legend" :style="legendStyle">
      <span v-for="(dataset, index) in data.datasets" :key="index" class="d3-bar-chart__legend-item">
        <span class="d3-bar-chart__legend-mark" :style="{ backgroundColor: colorAt(index, 0) }"></span>
        <span>{{ dataset.label }}</span>
      </span>
    </div>
    <div ref="frameRef" class="d3-bar-chart__frame">
      <svg
        class="d3-bar-chart__svg"
        :viewBox="`0 0 ${width} ${height}`"
        preserveAspectRatio="none"
        role="img"
        :aria-label="ariaLabel"
        @mouseleave="hovered = null"
      >
        <title>{{ ariaLabel }}</title>
        <g v-if="yOptions.grid?.display !== false">
          <line
            v-for="tick in yTicks"
            :key="`grid-${tick}`"
            :x1="plot.left"
            :x2="plot.right"
            :y1="yScale(tick)"
            :y2="yScale(tick)"
            :stroke="yOptions.grid?.color ?? '#e5e7eb'"
            :stroke-dasharray="yOptions.grid?.borderDash?.join(' ')"
          />
        </g>

        <rect
          v-for="bar in bars"
          :key="`${bar.datasetIndex}-${bar.dataIndex}`"
          class="d3-bar-chart__bar"
          :x="bar.x"
          :y="bar.y"
          :width="bar.width"
          :height="bar.height"
          :rx="bar.radius"
          :fill="bar.color"
          :opacity="hovered && (hovered.datasetIndex !== bar.datasetIndex || hovered.dataIndex !== bar.dataIndex) ? 0.58 : 0.9"
          @mousemove="showTooltip(bar, $event)"
        />

        <g v-if="xOptions.display !== false">
          <line :x1="plot.left" :x2="plot.right" :y1="plot.bottom" :y2="plot.bottom" stroke="#cbd5e1" />
          <text
            v-for="index in xTicks"
            :key="`x-${index}`"
            :x="(xScale(index) ?? plot.left) + xScale.bandwidth() / 2"
            :y="plot.bottom + 18"
            text-anchor="middle"
            :fill="xOptions.ticks?.color ?? '#64748b'"
            :font-size="xOptions.ticks?.font?.size ?? 10"
          >{{ formatXTick(index) }}</text>
        </g>

        <g v-if="yOptions.display !== false">
          <line :x1="plot.left" :x2="plot.left" :y1="plot.top" :y2="plot.bottom" stroke="#cbd5e1" />
          <g v-for="tick in yTicks" :key="`y-${tick}`">
            <line :x1="plot.left - 4" :x2="plot.left" :y1="yScale(tick)" :y2="yScale(tick)" stroke="#94a3b8" />
            <text
              :x="plot.left - 8"
              :y="yScale(tick) + 3"
              text-anchor="end"
              :fill="yOptions.ticks?.color ?? '#64748b'"
              :font-size="yOptions.ticks?.font?.size ?? 10"
            >{{ formatYTick(tick) }}</text>
          </g>
          <text
            v-if="yOptions.title?.display"
            x="12"
            :y="(plot.top + plot.bottom) / 2"
            :transform="`rotate(-90 12 ${(plot.top + plot.bottom) / 2})`"
            text-anchor="middle"
            :fill="yOptions.title.color ?? yOptions.ticks?.color ?? '#64748b'"
            :font-size="yOptions.title.font?.size ?? 11"
          >{{ yOptions.title.text }}</text>
        </g>
      </svg>

      <div v-if="tooltipContent" class="d3-bar-chart__tooltip" :style="tooltipStyle">
        <div v-for="title in tooltipContent.title" :key="title" class="d3-bar-chart__tooltip-title" :style="tooltipTitleStyle">
          {{ title }}
        </div>
        <div class="d3-bar-chart__tooltip-line">
          <span v-if="tooltipContent.displayColors" class="d3-bar-chart__tooltip-swatch" :style="{ backgroundColor: tooltipContent.color }"></span>
          <span>
            <span v-for="body in tooltipContent.body" :key="body" class="d3-bar-chart__tooltip-text">{{ body }}</span>
          </span>
        </div>
        <div v-for="footer in tooltipContent.footer" :key="footer" class="d3-bar-chart__tooltip-footer">{{ footer }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.d3-bar-chart {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 140px;
  flex-direction: column;
}

.d3-bar-chart__legend {
  display: flex;
  min-height: 24px;
  flex-wrap: wrap;
  align-items: center;
  padding: 0 8px 5px;
}

.d3-bar-chart__legend-item {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}

.d3-bar-chart__legend-mark,
.d3-bar-chart__tooltip-swatch {
  width: 8px;
  height: 8px;
  flex: 0 0 auto;
  border-radius: 2px;
}

.d3-bar-chart__frame {
  position: relative;
  min-height: 0;
  flex: 1;
}

.d3-bar-chart__svg {
  display: block;
  width: 100%;
  height: 100%;
}

.d3-bar-chart__bar {
  transition: opacity 140ms ease;
}

.d3-bar-chart__tooltip {
  position: absolute;
  z-index: 5;
  max-width: 220px;
  padding: 8px 10px;
  border-style: solid;
  border-radius: 6px;
  background: var(--glass-layer-floating-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  font-size: var(--font-size-2xs);
  line-height: 1.45;
  pointer-events: none;
  box-shadow: 0 8px 20px rgb(12 12 14 / 16%);
}

.d3-bar-chart__tooltip-title {
  margin-bottom: 4px;
  font-weight: 600;
}

.d3-bar-chart__tooltip-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.d3-bar-chart__tooltip-swatch {
  margin-top: 4px;
}

.d3-bar-chart__tooltip-text {
  display: block;
}

.d3-bar-chart__tooltip-footer {
  margin-top: 5px;
  padding-top: 5px;
  border-top: 1px solid rgb(148 163 184 / 28%);
}
</style>
