import { describe, expect, it } from 'vitest'

import { formatDateLocalInput } from '../format'

describe('formatDateLocalInput', () => {
  it('formats the calendar date in local time', () => {
    const localDate = new Date(2026, 6, 13, 0, 30)

    expect(formatDateLocalInput(localDate)).toBe('2026-07-13')
  })

  it('returns an empty string for an invalid date', () => {
    expect(formatDateLocalInput(new Date('invalid'))).toBe('')
  })
})
