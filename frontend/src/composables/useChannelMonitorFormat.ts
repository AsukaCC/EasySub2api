/**
 * Shared formatting helpers for channel monitor views (admin + user).
 *
 * Centralises:
 *  - status / provider label + badge class lookups
 *  - latency / availability / percent number formatting
 *  - dashboard-style helpers (HSL for availability, provider gradient, relative time)
 *
 * i18n keys live under `monitorCommon.*` so admin and user views share the
 * same translation source.
 */

import { useI18n } from 'vue-i18n'
import type { CheckMode, MonitorStatus, Provider } from '@/api/admin/channelMonitor'
import {
  PROVIDER_OPENAI,
  PROVIDER_ANTHROPIC,
  PROVIDER_GEMINI,
  PROVIDER_GROK,
  PROVIDER_ANTIGRAVITY,
  PROVIDER_KIMI,
  PROVIDER_ZHIPU,
  PROVIDER_DEEPSEEK,
  PROVIDERS,
  STATUS_OPERATIONAL,
  STATUS_DEGRADED,
  STATUS_FAILED,
  STATUS_ERROR,
} from '@/constants/channelMonitor'

const NEUTRAL_BADGE = 'composables-use-channel-monitor-format__state'

/** Availability HSL hue multiplier: 0%=red(0) / 50%=yellow(60) / 100%=green(120). */
const HSL_HUE_PER_PERCENT = 1.2
const HSL_SATURATION = 72
const HSL_LIGHTNESS = 42

export interface AvailabilityRow {
  primary_status: MonitorStatus | ''
  availability_7d: number | null | undefined
}

export function useChannelMonitorFormat() {
  const { t } = useI18n()

  function statusLabel(s: MonitorStatus | ''): string {
    if (!s) return t('monitorCommon.status.unknown')
    return t(`monitorCommon.status.${s}`)
  }

  function statusBadgeClass(s: MonitorStatus | ''): string {
    switch (s) {
      case STATUS_OPERATIONAL:
        return 'composables-use-channel-monitor-format__state-2'
      case STATUS_DEGRADED:
        return 'composables-use-channel-monitor-format__state-3'
      case STATUS_FAILED:
        return 'composables-use-channel-monitor-format__state-4'
      case STATUS_ERROR:
      default:
        return NEUTRAL_BADGE
    }
  }

  function providerLabel(p: Provider | string): string {
    if (PROVIDERS.includes(p as Provider)) {
      return t(`monitorCommon.providers.${p}`)
    }
    return p || '-'
  }

  function checkModeLabel(m: CheckMode | string): string {
    if (m === 'probe' || m === 'quota' || m === 'quota_probe') {
      return t(`monitorCommon.checkMode.${m}`)
    }
    return m || '-'
  }

  function providerBadgeClass(p: Provider | string): string {
    switch (p) {
      case PROVIDER_OPENAI:
        return 'composables-use-channel-monitor-format__state-2'
      case PROVIDER_ANTHROPIC:
        return 'composables-use-channel-monitor-format__state-5'
      case PROVIDER_GEMINI:
        return 'composables-use-channel-monitor-format__state-6'
      case PROVIDER_GROK:
        return 'composables-use-channel-monitor-format__state-7'
      case PROVIDER_ANTIGRAVITY:
        return 'composables-use-channel-monitor-format__state-8'
      // 配色与 utils/platformColors.ts 的平台色对齐。
      case PROVIDER_KIMI:
        return 'composables-use-channel-monitor-format__state-9'
      case PROVIDER_ZHIPU:
        return 'composables-use-channel-monitor-format__state-10'
      case PROVIDER_DEEPSEEK:
        return 'composables-use-channel-monitor-format__state-11'
      default:
        return NEUTRAL_BADGE
    }
  }

  /**
   * Semantic class for a provider radio-button-style picker state.
   * Reuses the same emerald/orange/sky palette as providerBadgeClass to keep
   * visual semantics consistent across badges and pickers.
   */
  function providerPickerClass(p: Provider | string, active: boolean): string {
    switch (p) {
      case PROVIDER_OPENAI:
        return active
          ? 'composables-use-channel-monitor-format__state-12'
          : 'composables-use-channel-monitor-format__state-13'
      case PROVIDER_ANTHROPIC:
        return active
          ? 'composables-use-channel-monitor-format__state-14'
          : 'composables-use-channel-monitor-format__state-15'
      case PROVIDER_GEMINI:
        return active
          ? 'composables-use-channel-monitor-format__state-16'
          : 'composables-use-channel-monitor-format__state-17'
      case PROVIDER_GROK:
        return active
          ? 'composables-use-channel-monitor-format__state-18'
          : 'composables-use-channel-monitor-format__state-19'
      case PROVIDER_ANTIGRAVITY:
        return active
          ? 'composables-use-channel-monitor-format__state-20'
          : 'composables-use-channel-monitor-format__state-21'
      case PROVIDER_KIMI:
        return active
          ? 'composables-use-channel-monitor-format__state-22'
          : 'composables-use-channel-monitor-format__state-23'
      case PROVIDER_ZHIPU:
        return active
          ? 'composables-use-channel-monitor-format__state-24'
          : 'composables-use-channel-monitor-format__state-25'
      case PROVIDER_DEEPSEEK:
        return active
          ? 'composables-use-channel-monitor-format__state-26'
          : 'composables-use-channel-monitor-format__state-27'
      default:
        return active
          ? 'composables-use-channel-monitor-format__state-28'
          : 'composables-use-channel-monitor-format__state-29'
    }
  }

  function formatLatency(ms: number | null | undefined): string {
    if (ms == null) return t('monitorCommon.latencyEmpty')
    return String(Math.round(ms))
  }

  function formatPercent(v: number | null | undefined): string {
    if (v == null || Number.isNaN(v)) return '-'
    return `${v.toFixed(2)}%`
  }

  function formatAvailability(row: AvailabilityRow): string {
    if (!row.primary_status) return '-'
    return formatPercent(row.availability_7d)
  }

  function formatRelativeTime(iso: string | null | undefined): string {
    if (!iso) return t('monitorCommon.latencyEmpty')
    const ts = Date.parse(iso)
    if (Number.isNaN(ts)) return t('monitorCommon.latencyEmpty')
    const diffSec = Math.max(0, Math.floor((Date.now() - ts) / 1000))
    if (diffSec < 60) return t('monitorCommon.relativeSecondsAgo', { n: diffSec })
    const diffMin = Math.floor(diffSec / 60)
    if (diffMin < 60) return t('monitorCommon.relativeMinutesAgo', { n: diffMin })
    const diffHour = Math.floor(diffMin / 60)
    if (diffHour < 24) return t('monitorCommon.relativeHoursAgo', { n: diffHour })
    const diffDay = Math.floor(diffHour / 24)
    return t('monitorCommon.relativeDaysAgo', { n: diffDay })
  }

  return {
    statusLabel,
    statusBadgeClass,
    providerLabel,
    checkModeLabel,
    providerBadgeClass,
    providerPickerClass,
    formatLatency,
    formatPercent,
    formatAvailability,
    formatRelativeTime,
  }
}

/**
 * Map availability percent to an HSL colour (red -> yellow -> green).
 * Returns undefined for null/NaN so callers can fall back to a neutral colour.
 */
export function hslForPct(pct: number | null | undefined): string | undefined {
  if (pct === null || pct === undefined || Number.isNaN(pct)) return undefined
  const clamped = Math.max(0, Math.min(100, pct))
  const hue = clamped * HSL_HUE_PER_PERCENT
  return `hsl(${hue} ${HSL_SATURATION}% ${HSL_LIGHTNESS}%)`
}

/**
 * Semantic gradient class for the provider icon tile background.
 */
export function providerGradient(provider: string): string {
  switch (provider) {
    case PROVIDER_OPENAI:
      return 'composables-use-channel-monitor-format__state-30'
    case PROVIDER_ANTHROPIC:
      return 'composables-use-channel-monitor-format__state-31'
    case PROVIDER_GEMINI:
      return 'composables-use-channel-monitor-format__state-32'
    case PROVIDER_GROK:
      return 'composables-use-channel-monitor-format__state-33'
    case PROVIDER_ANTIGRAVITY:
      return 'composables-use-channel-monitor-format__state-34'
    case PROVIDER_KIMI:
      return 'composables-use-channel-monitor-format__state-35'
    case PROVIDER_ZHIPU:
      return 'composables-use-channel-monitor-format__state-36'
    case PROVIDER_DEEPSEEK:
      return 'composables-use-channel-monitor-format__state-37'
    default:
      return 'composables-use-channel-monitor-format__state-38'
  }
}
