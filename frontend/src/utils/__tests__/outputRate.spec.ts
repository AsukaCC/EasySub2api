import { describe, expect, it } from 'vitest'
import { formatUsageOutputRate } from '../latencyHealth'

describe('usage output rate', () => {
  const row = { output_tokens: 100, duration_ms: 2000, image_count: 0, image_output_tokens: 0, billing_mode: 'token', request_type: 'stream' } as const
  it('uses the duration of the individual request', () => {
    expect(formatUsageOutputRate(row)).toBe('50.0 tok/s')
  })
  it('excludes media and missing or invalid measurements', () => {
    for (const patch of [{ image_count: 1 }, { image_output_tokens: 1 }, { billing_mode: 'image' }, { duration_ms: 0 }, { duration_ms: null }, { output_tokens: NaN }]) {
      expect(formatUsageOutputRate({ ...row, ...patch })).toBe('-')
    }
  })
})
