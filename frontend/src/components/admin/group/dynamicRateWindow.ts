import type { GroupDynamicRateRule } from '@/types'

export type DynamicRateRuleStatus = 'legacy' | 'not_started' | 'active' | 'expired' | 'invalid'

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

export function toLocalDateTimeInput(value?: string): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return [
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`,
    `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
  ].join('T')
}

export function localDateTimeToUTC(value: string): string {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/.exec(value)
  if (!match) return ''
  const [, year, month, day, hour, minute, second] = match
  const date = new Date(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hour),
    Number(minute),
    Number(second ?? 0),
    0
  )
  if (Number.isNaN(date.getTime())) return ''
  if (
    date.getFullYear() !== Number(year) ||
    date.getMonth() !== Number(month) - 1 ||
    date.getDate() !== Number(day) ||
    date.getHours() !== Number(hour) ||
    date.getMinutes() !== Number(minute) ||
    date.getSeconds() !== Number(second ?? 0)
  ) {
    return ''
  }
  return date.toISOString()
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
