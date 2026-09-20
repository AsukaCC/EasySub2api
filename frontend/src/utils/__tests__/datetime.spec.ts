import { describe, expect, it } from 'vitest'
import { dayjs, formatDateValue, formatDateWithOptions, parsePickerValue } from '../datetime'

describe('shared dayjs formatting', () => {
  it('uses the same numeric date and 24-hour time for both languages', () => {
    const date = new Date(2026, 8, 20, 0, 5, 9)
    expect(formatDateValue(date, undefined, 'zh')).toBe('2026-09-20 00:05:09')
    expect(formatDateValue(date, undefined, 'en')).toBe('2026-09-20 00:05:09')
    expect(formatDateWithOptions(date, { hour: '2-digit', minute: '2-digit' })).toBe('00:05')
    expect(formatDateWithOptions(date, { dateStyle: 'medium' })).toBe('2026-09-20')
  })
  it('handles explicit timezones without shifting date-only inputs', () => {
    expect(formatDateValue('2026-09-20T00:05:09Z', undefined, 'zh', 'Asia/Shanghai')).toBe('2026-09-20 08:05:09')
    expect(formatDateValue('2026-09-20', 'YYYY-MM-DD')).toBe('2026-09-20')
    expect(formatDateValue(0, undefined, 'en', 'UTC')).toBe('1970-01-01 00:00:00')
    expect(formatDateValue('2026-07-20T00:00:00Z', '[UTC]Z', 'en', 'Europe/London')).toBe('UTC+01:00')
  })
  it('returns an empty placeholder for missing and invalid dates', () => {
    for (const value of [undefined, null, '', 'invalid', new Date(NaN)]) expect(formatDateValue(value)).toBe('')
  })
  it('strictly validates leap days, clock values, and local timezone semantics', () => {
    expect(parsePickerValue('2024-02-29', 'date')?.format('YYYY-MM-DD')).toBe('2024-02-29')
    for (const value of ['2026-02-29', '2026-04-31', '2026-13-01']) expect(parsePickerValue(value, 'date')).toBeNull()
    for (const value of ['24:00', '12:60', '01:02:60']) expect(parsePickerValue(value, 'time')).toBeNull()
    expect(parsePickerValue('2026-09-20T08:15:32', 'datetime-local')?.unix()).toBe(dayjs(new Date(2026, 8, 20, 8, 15, 32)).unix())
    expect(parsePickerValue('2026-09-20T08:15:32Z', 'datetime-local')).toBeNull()
  })
})
