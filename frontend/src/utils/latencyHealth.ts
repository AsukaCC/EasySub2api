import type { UsageLog } from '@/types'

export function formatUsageOutputRate(row: Pick<UsageLog, 'output_tokens' | 'duration_ms' | 'image_count' | 'image_output_tokens' | 'billing_mode' | 'request_type'>): string {
  if (row.image_count > 0 || row.image_output_tokens > 0 || row.billing_mode === 'image'
    || (row.request_type && !['sync', 'stream', 'ws_v2', 'cyber'].includes(row.request_type))
    || !Number.isFinite(row.output_tokens) || row.output_tokens <= 0
    || row.duration_ms == null || !Number.isFinite(row.duration_ms) || row.duration_ms <= 0) return '-'
  return `${(row.output_tokens * 1000 / row.duration_ms).toFixed(1)} tok/s`
}

/**
 * 请求延迟健康度分档（用于用量明细"延迟"列的纵向健康扫视）。
 *
 * 首 Token（TTFT）：10s 内正常，10-30s 偏慢，30-60s 缓慢，60s 及以上严重。
 * 总耗时：流式请求整体时长天然更长，阈值放宽为 1min / 3min / 5min。
 */
export type LatencySeverity = 'good' | 'warn' | 'slow' | 'critical'

export const FIRST_TOKEN_THRESHOLDS_MS = {
  warn: 10_000,
  slow: 30_000,
  critical: 60_000,
} as const

export const DURATION_THRESHOLDS_MS = {
  warn: 60_000,
  slow: 180_000,
  critical: 300_000,
} as const

interface Thresholds {
  warn: number
  slow: number
  critical: number
}

const classify = (ms: number, thresholds: Thresholds): LatencySeverity => {
  if (ms >= thresholds.critical) return 'critical'
  if (ms >= thresholds.slow) return 'slow'
  if (ms >= thresholds.warn) return 'warn'
  return 'good'
}

export const firstTokenSeverity = (ms: number): LatencySeverity =>
  classify(ms, FIRST_TOKEN_THRESHOLDS_MS)

export const durationSeverity = (ms: number): LatencySeverity =>
  classify(ms, DURATION_THRESHOLDS_MS)

export const LATENCY_TEXT_CLASSES: Record<LatencySeverity, string> = {
  good: 'utils-latency-health__state',
  warn: 'utils-latency-health__state-2',
  slow: 'utils-latency-health__state-3',
  critical: 'utils-latency-health__state-4',
}

/** 无首字数据时的纯色色条（仅按总耗时档着色）。 */
export const LATENCY_BAR_CLASSES: Record<LatencySeverity, string> = {
  good: 'status-fill--success',
  warn: 'status-fill--warning',
  slow: 'status-fill--warning',
  critical: 'status-fill--danger',
}

/** 渐变色条上端（首字档）；与 LATENCY_BAR_TO_CLASSES 组合成上下渐变，避免两段硬切割裂感。 */
export const LATENCY_BAR_FROM_CLASSES: Record<LatencySeverity, string> = {
  good: 'from-emerald-500',
  warn: 'from-amber-400',
  slow: 'from-orange-500',
  critical: 'from-red-500',
}

/** 渐变色条下端（总耗时档）。 */
export const LATENCY_BAR_TO_CLASSES: Record<LatencySeverity, string> = {
  good: 'to-emerald-500',
  warn: 'to-amber-400',
  slow: 'to-orange-500',
  critical: 'to-red-500',
}
