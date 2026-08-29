<script setup lang="ts">
import { arc, pie, pointer, type PieArcDatum } from 'd3'
import { computed, ref, type CSSProperties } from 'vue'
import type {
  D3ArcTooltipItem,
  D3ChartData,
  D3ChartOptions,
} from './chartTypes'
import { useChartSize } from './useChartSize'

interface TooltipOptions {
  backgroundColor?: string
  titleColor?: string
  bodyColor?: string
  borderColor?: string
  borderWidth?: number
  displayColors?: boolean
  callbacks?: {
    title?: (items: D3ArcTooltipItem[]) => unknown
    label?: (item: D3ArcTooltipItem) => unknown
    footer?: (items: D3ArcTooltipItem[]) => unknown
  }
}

interface DonutOptions {
  cutout?: string | number
  plugins?: {
    legend?: {
      display?: boolean
      align?: string
      labels?: { color?: string; font?: { size?: number }; padding?: number }
    }
    tooltip?: TooltipOptions
  }
}

interface ArcValue {
  index: number
  label: string
  value: number
  color: string
}

interface ArcModel extends ArcValue {
  path: string
  hoverPath: string
}

const props = withDefaults(defineProps<{
  data: D3ChartData
  options?: D3ChartOptions
  ariaLabel?: string
}>(), {
  options: () => ({}),
  ariaLabel: 'Donut chart',
})

const palette = ['#2563eb', '#059669', '#d97706', '#dc2626', '#7c3aed', '#db2777', '#0891b2', '#65a30d']
const frameRef = ref<HTMLElement | null>(null)
const { width, height } = useChartSize(frameRef, 320, 240)
const hoverIndex = ref<number | null>(null)
const tooltipX = ref(0)
const tooltipY = ref(0)

const chartOptions = computed(() => props.options as DonutOptions)
const dataset = computed(() => props.data.datasets[0])
const values = computed<ArcValue[]>(() => (dataset.value?.data ?? []).map((raw, index) => ({
  index,
  label: String(props.data.labels[index] ?? ''),
  value: Math.max(0, numericValue(raw)),
  color: colorAt(index),
})))
const total = computed(() => values.value.reduce((sum, item) => sum + item.value, 0))
const radius = computed(() => Math.max(1, Math.min(width.value, height.value) / 2 - 10))
const innerRadius = computed(() => {
  const cutout = chartOptions.value.cutout ?? '50%'
  if (typeof cutout === 'number') return Math.max(0, Math.min(radius.value - 1, cutout))
  const ratio = Number.parseFloat(cutout) / 100
  return radius.value * (Number.isFinite(ratio) ? Math.max(0, Math.min(0.92, ratio)) : 0.5)
})

const arcs = computed<ArcModel[]>(() => {
  const layout = pie<ArcValue>()
    .sort(null)
    .value((item) => item.value)(values.value)
  const path = arc<PieArcDatum<ArcValue>>()
    .innerRadius(innerRadius.value)
    .outerRadius(radius.value)
    .cornerRadius(2)
    .padAngle(0.012)
  const hoverPath = arc<PieArcDatum<ArcValue>>()
    .innerRadius(Math.max(0, innerRadius.value - 1))
    .outerRadius(radius.value + 4)
    .cornerRadius(2)
    .padAngle(0.012)

  return layout.map((datum) => ({
    ...datum.data,
    path: path(datum) ?? '',
    hoverPath: hoverPath(datum) ?? '',
  }))
})

const showLegend = computed(() => chartOptions.value.plugins?.legend?.display !== false)
const legendStyle = computed<CSSProperties>(() => ({
  color: chartOptions.value.plugins?.legend?.labels?.color ?? '#64748b',
  justifyContent: chartOptions.value.plugins?.legend?.align === 'end' ? 'flex-end' : 'center',
  fontSize: `${chartOptions.value.plugins?.legend?.labels?.font?.size ?? 11}px`,
  gap: `${chartOptions.value.plugins?.legend?.labels?.padding ?? 12}px`,
}))

const tooltipItem = computed<D3ArcTooltipItem | null>(() => {
  if (hoverIndex.value === null || !dataset.value) return null
  const item = values.value[hoverIndex.value]
  if (!item) return null
  return {
    dataset: dataset.value,
    datasetIndex: 0,
    dataIndex: item.index,
    label: item.label,
    raw: item.value,
    parsed: item.value,
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
    body: customLabel.length ? customLabel : [`${formatDefault(item.raw)} (${formatPercent(item.raw)})`],
    footer: normalizeText(callbacks?.footer?.([item])),
    color: colorAt(item.dataIndex),
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

function numericValue(value: unknown) {
  const result = Number(value ?? 0)
  return Number.isFinite(result) ? result : 0
}

function colorAt(index: number) {
  const background = dataset.value?.backgroundColor
  if (Array.isArray(background)) return String(background[index] ?? palette[index % palette.length])
  if (typeof background === 'string') return background
  return palette[index % palette.length]
}

function formatDefault(value: number) {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function formatPercent(value: number) {
  return total.value > 0 ? `${((value / total.value) * 100).toFixed(1)}%` : '0%'
}

function normalizeText(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String).filter(Boolean)
  if (value === null || value === undefined || value === '') return []
  return [String(value)]
}

function showTooltip(index: number, event: MouseEvent) {
  hoverIndex.value = index
  if (!frameRef.value) return
  const [x, y] = pointer(event, frameRef.value)
  tooltipX.value = Math.max(8, Math.min(width.value - 190, x + 10))
  tooltipY.value = Math.max(8, Math.min(height.value - 70, y - 16))
}
</script>

<template>
  <div class="d3-donut-chart">
    <div v-if="showLegend" class="d3-donut-chart__legend" :style="legendStyle">
      <span v-for="item in values" :key="item.index" class="d3-donut-chart__legend-item">
        <span class="d3-donut-chart__legend-mark" :style="{ backgroundColor: item.color }"></span>
        <span>{{ item.label }}</span>
      </span>
    </div>
    <div ref="frameRef" class="d3-donut-chart__frame">
      <svg
        class="d3-donut-chart__svg"
        :viewBox="`0 0 ${width} ${height}`"
        preserveAspectRatio="xMidYMid meet"
        role="img"
        :aria-label="ariaLabel"
        @mouseleave="hoverIndex = null"
      >
        <title>{{ ariaLabel }}</title>
        <g :transform="`translate(${width / 2}, ${height / 2})`">
          <path
            v-for="item in arcs"
            :key="item.index"
            class="d3-donut-chart__arc"
            :d="hoverIndex === item.index ? item.hoverPath : item.path"
            :fill="item.color"
            :aria-label="`${item.label}: ${item.value}`"
            @mousemove="showTooltip(item.index, $event)"
          />
        </g>
      </svg>

      <div v-if="tooltipContent" class="d3-donut-chart__tooltip" :style="tooltipStyle">
        <div v-for="title in tooltipContent.title" :key="title" class="d3-donut-chart__tooltip-title" :style="tooltipTitleStyle">
          {{ title }}
        </div>
        <div class="d3-donut-chart__tooltip-line">
          <span v-if="tooltipContent.displayColors" class="d3-donut-chart__tooltip-swatch" :style="{ backgroundColor: tooltipContent.color }"></span>
          <span>
            <span v-for="body in tooltipContent.body" :key="body" class="d3-donut-chart__tooltip-text">{{ body }}</span>
          </span>
        </div>
        <div v-for="footer in tooltipContent.footer" :key="footer" class="d3-donut-chart__tooltip-footer">{{ footer }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.d3-donut-chart {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 140px;
  flex-direction: column;
}

.d3-donut-chart__legend {
  display: flex;
  min-height: 24px;
  flex-wrap: wrap;
  align-items: center;
  padding: 0 8px 5px;
}

.d3-donut-chart__legend-item {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}

.d3-donut-chart__legend-mark,
.d3-donut-chart__tooltip-swatch {
  width: 8px;
  height: 8px;
  flex: 0 0 auto;
  border-radius: 50%;
}

.d3-donut-chart__frame {
  position: relative;
  min-height: 0;
  flex: 1;
}

.d3-donut-chart__svg {
  display: block;
  width: 100%;
  height: 100%;
}

.d3-donut-chart__arc {
  cursor: default;
  transition: opacity 140ms ease;
}

.d3-donut-chart__arc:hover {
  opacity: 0.9;
}

.d3-donut-chart__tooltip {
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

.d3-donut-chart__tooltip-title {
  margin-bottom: 4px;
  font-weight: 600;
}

.d3-donut-chart__tooltip-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
}

.d3-donut-chart__tooltip-swatch {
  margin-top: 4px;
}

.d3-donut-chart__tooltip-text {
  display: block;
}

.d3-donut-chart__tooltip-footer {
  margin-top: 5px;
  padding-top: 5px;
  border-top: 1px solid rgb(148 163 184 / 28%);
}
</style>
