import dayjs, { type ConfigType, type Dayjs } from 'dayjs'
import customParseFormat from 'dayjs/plugin/customParseFormat'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/en-gb'

dayjs.extend(customParseFormat)
dayjs.extend(utc)
dayjs.extend(timezone)

export { dayjs }
export type { Dayjs }
export type DateInput = ConfigType
export type PickerType = 'date' | 'datetime-local' | 'time'

export function dateLocale(locale = 'zh'): string {
  return locale.toLowerCase().startsWith('zh') ? 'zh-cn' : locale.toLowerCase().startsWith('en-gb') ? 'en-gb' : 'en'
}

/** Display dates in a stable, 24-hour format; numeric inputs are milliseconds. */
export function formatDateValue(value: DateInput, format = 'YYYY-MM-DD HH:mm:ss', locale = 'zh', zone?: string): string {
  if (value == null || value === '') return ''
  let date = dayjs(value)
  if (!date.isValid()) return ''
  if (zone) date = date.tz(zone)
  return date.locale(dateLocale(locale)).format(format)
}

/** Compatibility for callers requesting a specific precision or timezone. */
export function formatDateWithOptions(value: DateInput, options?: Intl.DateTimeFormatOptions, locale?: string): string {
  if (!options || Object.keys(options).length === 0) return formatDateValue(value, undefined, locale)
  const dateParts = [options.year && 'YYYY', options.month && 'MM', options.day && 'DD'].filter(Boolean)
  if (options.dateStyle) dateParts.splice(0, dateParts.length, 'YYYY', 'MM', 'DD')
  const hour = options.hour12 ? 'hh' : 'HH'
  const timeParts = [options.hour && hour, options.minute && 'mm', options.second && 'ss'].filter(Boolean)
  if (options.timeStyle) timeParts.splice(0, timeParts.length, hour, 'mm', ...(options.timeStyle === 'short' ? [] : ['ss']))
  let format = [dateParts.join('-'), timeParts.join(':')].filter(Boolean).join(' ')
  if (options.weekday) format = `${options.weekday === 'long' ? 'dddd' : 'ddd'} ${format}`.trim()
  if (options.hour12 && timeParts.length) format += ' A'
  if (options.timeZoneName) format += ' [UTC]Z'
  return formatDateValue(value, format || 'YYYY-MM-DD', locale, options.timeZone)
}

export function pickerFormat(type: PickerType, seconds = false): string {
  return type === 'date' ? 'YYYY-MM-DD' : type === 'time' ? `HH:mm${seconds ? ':ss' : ''}` : `YYYY-MM-DDTHH:mm${seconds ? ':ss' : ''}`
}

/** Strict local parsing prevents invalid days and DST gaps from silently rolling over. */
export function parsePickerValue(value: string, type: PickerType): Dayjs | null {
  const formats = type === 'date' ? ['YYYY-MM-DD'] : type === 'time'
    ? ['HH:mm', 'HH:mm:ss'] : ['YYYY-MM-DDTHH:mm', 'YYYY-MM-DDTHH:mm:ss', 'YYYY-MM-DDTHH:mm:ss.SSS']
  for (const format of formats) {
    // Recurring clock values must not depend on today's daylight-saving transition.
    const parsed = type === 'time' ? dayjs(`2000-01-01T${value}`, `YYYY-MM-DDT${format}`, true) : dayjs(value, format, true)
    if (parsed.isValid() && parsed.year() > 0) return parsed
  }
  return null
}
