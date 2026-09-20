import type { GroupDynamicRateRule } from '@/types'
import { formatDateValue, parsePickerValue } from '@/utils/datetime'

export type DynamicRateRuleStatus = 'legacy' | 'not_started' | 'active' | 'expired' | 'invalid'

export function toLocalDateTimeInput(value?: string): string {
  return formatDateValue(value, 'YYYY-MM-DDTHH:mm:ss')
}

export function localDateTimeToUTC(value: string): string {
  return parsePickerValue(value, 'datetime-local')?.toISOString() ?? ''
}

export function isLegacyDynamicRateRule(rule: GroupDynamicRateRule): boolean {
  return !rule.start_at && !rule.end_at && Boolean(
    rule.timezone || rule.start_time || rule.end_time || rule.quota_amount != null
  )
}

export function parseAbsoluteWindow(rule: GroupDynamicRateRule): { start: Date; end: Date } | null {
  if (!rule.start_at || !rule.end_at) return null
  const start = new Date(rule.start_at)
  const end = new Date(rule.end_at)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || start >= end) return null
  return { start, end }
}

export function getDynamicRateRuleStatus(
  rule: GroupDynamicRateRule,
  now: Date = new Date()
): DynamicRateRuleStatus {
  if (isLegacyDynamicRateRule(rule)) return 'legacy'
  const window = parseAbsoluteWindow(rule)
  if (!window) return 'invalid'
  if (now < window.start) return 'not_started'
  if (now < window.end) return 'active'
  return 'expired'
}
