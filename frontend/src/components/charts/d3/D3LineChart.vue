<script setup lang="ts">
import {
  area,
  curveLinear,
  curveMonotoneX,
  line,
  pointer,
  range,
  scaleLinear,
  scalePoint,
} from 'd3'
import { computed, ref, watch, type CSSProperties } from 'vue'
import type { ScaleLinear } from 'd3'
import type {
  D3ChartData,
  D3ChartDataset,
  D3ChartOptions,
  D3LineTooltipItem,
} from './chartTypes'
import { useChartSize } from './useChartSize'

interface TickOptions {
  color?: string
  maxTicksLimit?: number
  callback?: (value: string | number) => unknown
  font?: { size?: number }
}

interface AxisOptions {
  display?: boolean
  position?: 'left' | 'right'
  min?: number
  max?: number
  suggestedMax?: number
  beginAtZero?: boolean
  grid?: { display?: boolean; color?: string; drawOnChartArea?: boolean; borderDash?: number[] }
  ticks?: TickOptions
  title?: { display?: boolean; text?: string; color?: string; font?: { size?: number } }
}

interface TooltipOptions {
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

interface LineOptions {
  plugins?: {
    legend?: {
      display?: boolean
      position?: string
      align?: string
      labels?: { color?: string; font?: { size?: number }; padding?: number }
    }
    tooltip?: TooltipOptions
    zoom?: { zoom?: { wheel?: { enabled?: boolean } } }
  }
  scales?: Record<string, AxisOptions>
}

interface AxisModel {
  key: string
  options: AxisOptions
  position: 'left' | 'right'
  scale: ScaleLinear<number, number>
  ticks: number[]
}

interface RenderPoint {
  index: number
  value: number | null
  x: number
  y: number
}

interface SeriesModel {
  dataset: D3ChartDataset
  datasetIndex: number
  color: string
  fillColor: string
  linePath: string
  areaPath: string
  dash: string | undefined
  lineWidth: number
  pointRadius: number
  points: RenderPoint[]
}

const props = withDefaults(defineProps<{
  data: D3ChartData
  options?: D3ChartOptions
  ariaLabel?: string
}>(), {
  options: () => ({}),
  ariaLabel: 'Line chart',
})

const palette = ['#2563eb', '#059669', '#d97706', '#dc2626', '#7c3aed', '#0891b2', '#db2777']
let chartSequence = 0
const clipId = `d3-line-clip-${++chartSequence}`
const frameRef = ref<HTMLElement | null>(null)
const svgRef = ref<SVGSVGElement | null>(null)
const { width, height } = useChartSize(frameRef)
const hoverIndex = ref<number | null>(null)
const tooltipX = ref(0)
const tooltipY = ref(0)
const zoomStart = ref(0)
const zoomEnd = ref(Number.MAX_SAFE_INTEGER)

const chartOptions = computed(() => props.options as LineOptions)
const scalesOptions = computed(() => chartOptions.value.scales ?? {})
const rightAxisKeys = computed(() => axisKeys.value.filter((key) => scalesOptions.value[key]?.position === 'right'))
const leftAxisKeys = computed(() => axisKeys.value.filter((key) => scalesOptions.value[key]?.position !== 'right'))

const plot = computed(() => {
  const leftHasTitle = leftAxisKeys.value.some((key) => scalesOptions.value[key]?.title?.display)
  const rightHasTitle = rightAxisKeys.value.some((key) => scalesOptions.value[key]?.title?.display)
  const left = leftHasTitle ? 72 : 54
  const right = rightAxisKeys.value.length ? (rightHasTitle ? 72 : 56) : 18
  const top = 10
  const bottom = 34
  return {
    left,
    right: Math.max(left + 1, width.value - right),
    top,
    bottom: Math.max(top + 1, height.value - bottom),
  }
})

const axisKeys = computed(() => {
  const keys: string[] = []
  for (const dataset of props.data.datasets) {
    const key = dataset.yAxisID || 'y'
    if (!keys.includes(key)) keys.push(key)
  }
  return keys.length ? keys : ['y']
})

const zoomBounds = computed(() => {
  const last = Math.max(0, props.data.labels.length - 1)
  return {
    start: Math.min(Math.max(0, zoomStart.value), last),
    end: Math.min(Math.max(zoomStart.value, zoomEnd.value), last),
  }
})

const visibleIndices = computed(() => {
  if (!props.data.labels.length) return []
  return range(zoomBounds.value.start, zoomBounds.value.end + 1)
})

const xScale = computed(() => scalePoint<number>()
  .domain(visibleIndices.value)
  .range([plot.value.left, plot.value.right])
  .padding(0.16))

const axisModels = computed<AxisModel[]>(() => axisKeys.value.map((key) => {
  const options = scalesOptions.value[key] ?? {}
  const values = props.data.datasets
    .filter((dataset) => (dataset.yAxisID || 'y') === key)
    .flatMap((dataset) => visibleIndices.value.map((index) => numericValue(dataset.data[index])))
    .filter((value): value is number => value !== null)

  const dataMin = values.length ? Math.min(...values) : 0
  const dataMax = values.length ? Math.max(...values) : 1
  let minimum = options.min ?? ((options.beginAtZero !== false && dataMin >= 0) ? 0 : dataMin)
  let maximum = options.max ?? Math.max(dataMax, options.suggestedMax ?? Number.NEGATIVE_INFINITY)

  if (!Number.isFinite(minimum)) minimum = 0
  if (!Number.isFinite(maximum)) maximum = 1
  if (minimum === maximum) {
    maximum = minimum === 0 ? 1 : minimum + Math.abs(minimum * 0.1)
    minimum = minimum > 0 ? 0 : minimum - 1
  }

  const scale = scaleLinear()
    .domain([minimum, maximum])
    .range([plot.value.bottom, plot.value.top])

  if (options.min === undefined && options.max === undefined) scale.nice(5)

  return {
    key,
    options,
    position: options.position === 'right' ? 'right' : 'left',
    scale,
    ticks: scale.ticks(5),
  }
}))

const xTicks = computed(() => {
  const indices = visibleIndices.value
  if (indices.length <= 1) return indices
  const configured = scalesOptions.value.x?.ticks?.maxTicksLimit
  const maximum = Math.max(2, configured ?? Math.floor((plot.value.right - plot.value.left) / 72))
  const step = Math.max(1, Math.ceil(indices.length / maximum))
  const ticks = indices.filter((_, index) => index % step === 0)
  const last = indices[indices.length - 1]
  if (ticks[ticks.length - 1] !== last) ticks.push(last)
  return ticks
})

const seriesModels = computed<SeriesModel[]>(() => props.data.datasets.map((dataset, datasetIndex) => {
  const axis = axisModels.value.find((item) => item.key === (dataset.yAxisID || 'y')) ?? axisModels.value[0]
  const points = visibleIndices.value.map((index) => {
    const value = numericValue(dataset.data[index])
    return {
      index,
      value,
      x: xScale.value(index) ?? plot.value.left,
      y: value === null ? plot.value.bottom : axis.scale(value),
    }
  })
  const curve = Number(dataset.tension ?? 0) > 0 ? curveMonotoneX : curveLinear
  const defined = (point: RenderPoint) => point.value !== null
  const path = line<RenderPoint>()
    .defined(defined)
    .x((point) => point.x)
    .y((point) => point.y)
    .curve(curve)(points) ?? ''
  const baselineValue = Math.max(axis.scale.domain()[0], Math.min(0, axis.scale.domain()[1]))
  const fillPath = dataset.fill
    ? area<RenderPoint>()
      .defined(defined)
      .x((point) => point.x)
      .y0(axis.scale(baselineValue))
      .y1((point) => point.y)
      .curve(curve)(points) ?? ''
    : ''
  const color = dataset.borderColor || palette[datasetIndex % palette.length]
  const background = typeof dataset.backgroundColor === 'string' ? dataset.backgroundColor : `${color}20`

  return {
    dataset,
    datasetIndex,
    color,
    fillColor: background,
    linePath: path,
    areaPath: fillPath,
    dash: dataset.borderDash?.join(' '),
    lineWidth: Number(dataset.borderWidth ?? 2),
    pointRadius: Number(dataset.pointRadius ?? 2.5),
    points,
  }
}))

const primaryAxis = computed(() => axisModels.value.find((axis) => axis.position === 'left') ?? axisModels.value[0])
const showLegend = computed(() => chartOptions.value.plugins?.legend?.display !== false)
const legendStyle = computed<CSSProperties>(() => ({
  color: chartOptions.value.plugins?.legend?.labels?.color ?? '#64748b',
  justifyContent: chartOptions.value.plugins?.legend?.align === 'end' ? 'flex-end' : 'center',
  fontSize: `${chartOptions.value.plugins?.legend?.labels?.font?.size ?? 11}px`,
  gap: `${chartOptions.value.plugins?.legend?.labels?.padding ?? 14}px`,
}))

const hoverItems = computed<D3LineTooltipItem[]>(() => {
  if (hoverIndex.value === null) return []
  return props.data.datasets.flatMap((dataset, datasetIndex) => {
    const raw = dataset.data[hoverIndex.value as number]
    const value = numericValue(raw)
    if (value === null) return []
    return [{
      dataset,
      datasetIndex,
      dataIndex: hoverIndex.value as number,
      label: String(props.data.labels[hoverIndex.value as number] ?? ''),
      raw,
      parsed: { x: hoverIndex.value as number, y: value },
    }]
  })
})

const tooltipContent = computed(() => {
  const tooltip = chartOptions.value.plugins?.tooltip
  const callbacks = tooltip?.callbacks
  const items = hoverItems.value
  if (!items.length) return null
  const customTitle = normalizeText(callbacks?.title?.(items))
  const lines = items.map((item) => {
    const custom = normalizeText(callbacks?.label?.(item))
    return {
      color: item.dataset.borderColor || palette[item.datasetIndex % palette.length],
      text: custom.length ? custom : [`${item.dataset.label || 'Value'}: ${formatDefault(item.parsed.y ?? 0)}`],
    }
  })
  return {
    title: customTitle.length ? customTitle : [items[0].label],
    lines,
    footer: normalizeText(callbacks?.footer?.(items)),
    displayColors: tooltip?.displayColors !== false,
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

const hoverX = computed(() => hoverIndex.value === null ? null : xScale.value(hoverIndex.value))
const xAxisOptions = computed(() => scalesOptions.value.x ?? {})

watch(
  () => [
    props.data.labels.length,
    props.data.labels[0],
    props.data.labels[props.data.labels.length - 1],
  ],
  () => resetZoom(),
)

function numericValue(value: unknown): number | null {
  if (value === null || value === undefined || value === '') return null
  const result = Number(value)
  return Number.isFinite(result) ? result : null
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
  const callback = xAxisOptions.value.ticks?.callback
  return callback ? String(callback(label)) : label
}

function formatYTick(axis: AxisModel, value: number) {
  const callback = axis.options.ticks?.callback
  return callback ? String(callback(value)) : formatDefault(value)
}

function onPointerMove(event: MouseEvent) {
  if (!svgRef.value || !visibleIndices.value.length) return
  const [x, y] = pointer(event, svgRef.value)
  if (x < plot.value.left || x > plot.value.right || y < plot.value.top || y > plot.value.bottom) {
    hoverIndex.value = null
    return
  }
  let nearest = visibleIndices.value[0]
  let distance = Number.POSITIVE_INFINITY
  for (const index of visibleIndices.value) {
    const position = xScale.value(index)
    if (position === undefined) continue
    const nextDistance = Math.abs(position - x)
    if (nextDistance < distance) {
      distance = nextDistance
      nearest = index
    }
  }
  hoverIndex.value = nearest
  tooltipX.value = Math.max(8, Math.min(width.value - 210, x + 12))
  tooltipY.value = Math.max(8, Math.min(height.value - 80, y - 18))
}

function onWheel(event: WheelEvent) {
  if (!chartOptions.value.plugins?.zoom?.zoom?.wheel?.enabled || props.data.labels.length < 4) return
  event.preventDefault()
  const current = zoomBounds.value
  const currentCount = current.end - current.start + 1
  const nextCount = Math.max(3, Math.min(props.data.labels.length, Math.round(currentCount * (event.deltaY < 0 ? 0.8 : 1.25))))
  const rect = svgRef.value?.getBoundingClientRect()
  const ratio = rect?.width ? Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width)) : 0.5
  const anchor = current.start + Math.round((currentCount - 1) * ratio)
  let start = Math.round(anchor - (nextCount - 1) * ratio)
  start = Math.max(0, Math.min(props.data.labels.length - nextCount, start))
  zoomStart.value = start
  zoomEnd.value = start + nextCount - 1
}

function resetZoom() {
  zoomStart.value = 0
  zoomEnd.value = Number.MAX_SAFE_INTEGER
  hoverIndex.value = null
}

function toDataUrl() {
  if (!svgRef.value || typeof XMLSerializer === 'undefined') return null
  const clone = svgRef.value.cloneNode(true) as SVGSVGElement
  clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg')
  clone.setAttribute('width', String(width.value))
  clone.setAttribute('height', String(height.value))
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(new XMLSerializer().serializeToString(clone))}`
}

defineExpose({ resetZoom, toDataUrl })
</script>

<template>
  <div class="d3-line-chart">
    <div v-if="showLegend" class="d3-line-chart__legend" :style="legendStyle">
      <span v-for="series in seriesModels" :key="series.datasetIndex" class="d3-line-chart__legend-item">
        <span class="d3-line-chart__legend-mark" :style="{ backgroundColor: series.color }"></span>
        <span>{{ series.dataset.label }}</span>
      </span>
    </div>
    <div ref="frameRef" class="d3-line-chart__frame">
      <svg
        ref="svgRef"
        class="d3-line-chart__svg"
        :viewBox="`0 0 ${width} ${height}`"
        preserveAspectRatio="none"
        role="img"
        :aria-label="ariaLabel"
        @mousemove="onPointerMove"
        @mouseleave="hoverIndex = null"
        @wheel="onWheel"
      >
        <title>{{ ariaLabel }}</title>
        <defs>
          <clipPath :id="clipId">
            <rect :x="plot.left" :y="plot.top" :width="plot.right - plot.left" :height="plot.bottom - plot.top" />
          </clipPath>
        </defs>

        <g v-if="primaryAxis && primaryAxis.options.grid?.display !== false && primaryAxis.options.grid?.drawOnChartArea !== false">
          <line
            v-for="tick in primaryAxis.ticks"
            :key="`grid-${tick}`"
            :x1="plot.left"
            :x2="plot.right"
            :y1="primaryAxis.scale(tick)"
            :y2="primaryAxis.scale(tick)"
            :stroke="primaryAxis.options.grid?.color ?? '#e5e7eb'"
            stroke-width="1"
            :stroke-dasharray="primaryAxis.options.grid?.borderDash?.join(' ')"
          />
        </g>

        <g :clip-path="`url(#${clipId})`">
          <path
            v-for="series in seriesModels.filter((item) => item.areaPath)"
            :key="`area-${series.datasetIndex}`"
            :d="series.areaPath"
            :fill="series.fillColor"
          />
          <path
            v-for="series in seriesModels"
            :key="`line-${series.datasetIndex}`"
            :d="series.linePath"
            fill="none"
            :stroke="series.color"
            :stroke-width="series.lineWidth"
            :stroke-dasharray="series.dash"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
          <template v-for="series in seriesModels" :key="`points-${series.datasetIndex}`">
            <circle
              v-for="point in series.points.filter((item) => item.value !== null && series.pointRadius > 0)"
              :key="point.index"
              :cx="point.x"
              :cy="point.y"
              :r="series.pointRadius"
              :fill="String(series.dataset.pointBackgroundColor ?? series.color)"
              :stroke="String(series.dataset.pointBorderColor ?? '#ffffff')"
              :stroke-width="Number(series.dataset.pointBorderWidth ?? 0)"
            />
          </template>
          <line
            v-if="hoverX !== null && hoverX !== undefined"
            :x1="hoverX"
            :x2="hoverX"
            :y1="plot.top"
            :y2="plot.bottom"
            stroke="#94a3b8"
            stroke-width="1"
            stroke-dasharray="3 3"
            pointer-events="none"
          />
          <circle
            v-for="item in hoverItems"
            :key="`hover-${item.datasetIndex}`"
            :cx="hoverX ?? 0"
            :cy="axisModels.find((axis) => axis.key === (item.dataset.yAxisID || 'y'))?.scale(item.parsed.y ?? 0)"
            r="4"
            :fill="item.dataset.borderColor || palette[item.datasetIndex % palette.length]"
            stroke="#ffffff"
            stroke-width="2"
            pointer-events="none"
          />
        </g>

        <g v-if="xAxisOptions.display !== false">
          <line :x1="plot.left" :x2="plot.right" :y1="plot.bottom" :y2="plot.bottom" stroke="#cbd5e1" />
          <g v-for="index in xTicks" :key="`x-${index}`">
            <line :x1="xScale(index)" :x2="xScale(index)" :y1="plot.bottom" :y2="plot.bottom + 4" stroke="#94a3b8" />
            <text
              :x="xScale(index)"
              :y="plot.bottom + 17"
              text-anchor="middle"
              :fill="xAxisOptions.ticks?.color ?? '#64748b'"
              :font-size="xAxisOptions.ticks?.font?.size ?? 10"
            >{{ formatXTick(index) }}</text>
          </g>
        </g>

        <g v-for="axis in axisModels.filter((item) => item.options.display !== false)" :key="axis.key">
          <line
            :x1="axis.position === 'right' ? plot.right : plot.left"
            :x2="axis.position === 'right' ? plot.right : plot.left"
            :y1="plot.top"
            :y2="plot.bottom"
            stroke="#cbd5e1"
          />
          <g v-for="tick in axis.ticks" :key="`${axis.key}-${tick}`">
            <line
              :x1="axis.position === 'right' ? plot.right : plot.left - 4"
              :x2="axis.position === 'right' ? plot.right + 4 : plot.left"
              :y1="axis.scale(tick)"
              :y2="axis.scale(tick)"
              stroke="#94a3b8"
            />
            <text
              :x="axis.position === 'right' ? plot.right + 8 : plot.left - 8"
              :y="axis.scale(tick) + 3"
              :text-anchor="axis.position === 'right' ? 'start' : 'end'"
              :fill="axis.options.ticks?.color ?? '#64748b'"
              :font-size="axis.options.ticks?.font?.size ?? 10"
            >{{ formatYTick(axis, tick) }}</text>
          </g>
          <text
            v-if="axis.options.title?.display"
            :x="axis.position === 'right' ? width - 12 : 12"
            :y="(plot.top + plot.bottom) / 2"
            :transform="`rotate(${axis.position === 'right' ? 90 : -90} ${axis.position === 'right' ? width - 12 : 12} ${(plot.top + plot.bottom) / 2})`"
            text-anchor="middle"
            :fill="axis.options.title.color ?? axis.options.ticks?.color ?? '#64748b'"
            :font-size="axis.options.title.font?.size ?? 11"
          >{{ axis.options.title.text }}</text>
        </g>
      </svg>

      <div v-if="tooltipContent" class="d3-line-chart__tooltip" :style="tooltipStyle">
        <div v-for="title in tooltipContent.title" :key="title" class="d3-line-chart__tooltip-title" :style="tooltipTitleStyle">
          {{ title }}
        </div>
        <div v-for="(lineItem, lineIndex) in tooltipContent.lines" :key="lineIndex" class="d3-line-chart__tooltip-line">
          <span v-if="tooltipContent.displayColors" class="d3-line-chart__tooltip-swatch" :style="{ backgroundColor: lineItem.color }"></span>
          <span>
            <span v-for="text in lineItem.text" :key="text" class="d3-line-chart__tooltip-text">{{ text }}</span>
          </span>
        </div>
        <div v-for="footer in tooltipContent.footer" :key="footer" class="d3-line-chart__tooltip-footer">{{ footer }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.d3-line-chart {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 140px;
  flex-direction: column;
}

.d3-line-chart__legend {
  display: flex;
  min-height: 24px;
  flex-wrap: wrap;
  align-items: center;
  padding: 0 8px 5px;
}

.d3-line-chart__legend-item {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}

.d3-line-chart__legend-mark,
.d3-line-chart__tooltip-swatch {
  width: 8px;
  height: 8px;
  flex: 0 0 auto;
  border-radius: 50%;
}

.d3-line-chart__frame {
  position: relative;
  min-height: 0;
  flex: 1;
}

.d3-line-chart__svg {
  display: block;
  width: 100%;
  height: 100%;
  overflow: visible;
  touch-action: pan-y;
}

.d3-line-chart__tooltip {
  position: absolute;
  z-index: 5;
  max-width: 240px;
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

.d3-line-chart__tooltip-title {
  margin-bottom: 4px;
  font-weight: 600;
}

.d3-line-chart__tooltip-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.d3-line-chart__tooltip-swatch {
  margin-top: 4px;
}

.d3-line-chart__tooltip-text {
  display: block;
}

.d3-line-chart__tooltip-footer {
  margin-top: 5px;
  padding-top: 5px;
  border-top: 1px solid rgb(148 163 184 / 28%);
}
</style>
