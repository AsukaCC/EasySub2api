import { describe, expect, it } from 'vitest'
import type { GroupDynamicRateRule } from '@/types'
import {
  getDynamicRateRuleStatus,
  isLegacyDynamicRateRule,
  localDateTimeToUTC,
  parseAbsoluteWindow,
  toLocalDateTimeInput
} from '../dynamicRateWindow'

function absoluteRule(start_at: string, end_at: string): GroupDynamicRateRule {
  return {
    id: '00000000-0000-0000-0000-000000000001',
    name: 'window',
    enabled: true,
    start_at,
    end_at,
    levels: [],
    multiplier: 0.8,
    activation_spend: 0,
    shared_quota_amount: 0,
    personal_quota_amount: 0
  }
}

describe('dynamicRateWindow', () => {
  it('round-trips browser-local seconds through UTC', () => {
    const local = '2026-09-03T12:34:56'
    const utc = localDateTimeToUTC(local)
    expect(utc).toMatch(/Z$/)
    expect(toLocalDateTimeInput(utc)).toBe(local)
  })

  it('accepts a datetime-local value that omits zero seconds', () => {
    const utc = localDateTimeToUTC('2026-09-03T12:34')
    expect(toLocalDateTimeInput(utc)).toBe('2026-09-03T12:34:00')
  })

  it('rejects invalid local dates', () => {
    expect(localDateTimeToUTC('2026-02-30T12:00:00')).toBe('')
    expect(localDateTimeToUTC('not-a-date')).toBe('')
  })

  it('uses a half-open absolute interval', () => {
    const rule = absoluteRule('2026-09-03T00:00:00Z', '2026-09-03T01:00:00Z')
    expect(getDynamicRateRuleStatus(rule, new Date('2026-09-02T23:59:59.999Z'))).toBe('not_started')
    expect(getDynamicRateRuleStatus(rule, new Date('2026-09-03T00:00:00Z'))).toBe('active')
    expect(getDynamicRateRuleStatus(rule, new Date('2026-09-03T01:00:00Z'))).toBe('expired')
    expect(parseAbsoluteWindow(rule)).not.toBeNull()
  })

  it('marks old daily-clock rules as legacy', () => {
    const rule: GroupDynamicRateRule = {
      id: '00000000-0000-0000-0000-000000000002',
      name: 'legacy',
      enabled: true,
      timezone: 'Asia/Shanghai',
      start_time: '09:00',
      end_time: '10:00',
      quota_amount: 2,
      levels: [],
      multiplier: 0.8,
      activation_spend: 0
    }
    expect(isLegacyDynamicRateRule(rule)).toBe(true)
    expect(getDynamicRateRuleStatus(rule)).toBe('legacy')
  })
})
