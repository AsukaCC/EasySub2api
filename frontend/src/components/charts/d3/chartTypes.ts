export type D3ChartValue = number | null | undefined

export interface D3ChartDataset {
  label?: string
  data: readonly D3ChartValue[]
  borderColor?: string
  backgroundColor?: string | readonly string[]
  borderWidth?: number
  borderDash?: readonly number[]
  borderRadius?: number
  fill?: boolean | string
  tension?: number
  yAxisID?: string
  pointRadius?: number
  pointHoverRadius?: number
  pointBackgroundColor?: string
  pointBorderColor?: string
  pointBorderWidth?: number
  spanGaps?: boolean
  barPercentage?: number
  [key: string]: unknown
}

export interface D3ChartData {
  labels: readonly (string | number)[]
  datasets: readonly D3ChartDataset[]
}

export type D3ChartOptions = Record<string, unknown>

export interface D3LineTooltipItem {
  dataset: D3ChartDataset
  datasetIndex: number
  dataIndex: number
  label: string
  raw: D3ChartValue
  parsed: {
    x: number
    y: number | null
  }
}

export interface D3ArcTooltipItem {
  dataset: D3ChartDataset
  datasetIndex: number
  dataIndex: number
  label: string
  raw: number
  parsed: number
}

export interface D3LineChartHandle {
  resetZoom: () => void
  toDataUrl: () => string | null
}
