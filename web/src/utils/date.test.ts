import { describe, it, expect } from 'vitest'
import { formatDate, extractDate, addDays, formatDateStr, formatPeriod, getMonday, displayDate } from './date'

describe('formatDate', () => {
  it('formats a date to YYYY-MM-DD', () => {
    expect(formatDate(new Date(2026, 2, 23))).toBe('2026-03-23')
  })

  it('pads single-digit month and day', () => {
    expect(formatDate(new Date(2026, 0, 5))).toBe('2026-01-05')
  })

  it('handles end of year', () => {
    expect(formatDate(new Date(2025, 11, 31))).toBe('2025-12-31')
  })
})

describe('extractDate', () => {
  it('extracts date portion from ISO string with time', () => {
    expect(extractDate('2026-03-23T14:30:00Z')).toBe('2026-03-23')
  })

  it('returns date unchanged when no time component', () => {
    expect(extractDate('2026-03-23')).toBe('2026-03-23')
  })
})

describe('addDays', () => {
  it('adds positive days', () => {
    const d = new Date(2026, 2, 20)
    expect(formatDate(addDays(d, 3))).toBe('2026-03-23')
  })

  it('adds negative days (subtracts)', () => {
    const d = new Date(2026, 2, 23)
    expect(formatDate(addDays(d, -1))).toBe('2026-03-22')
  })

  it('crosses month boundary', () => {
    const d = new Date(2026, 2, 31)
    expect(formatDate(addDays(d, 1))).toBe('2026-04-01')
  })

  it('does not mutate original date', () => {
    const d = new Date(2026, 2, 23)
    addDays(d, 5)
    expect(formatDate(d)).toBe('2026-03-23')
  })
})

describe('formatDateStr', () => {
  it('builds YYYY-MM-DD from year, month, day', () => {
    expect(formatDateStr(2026, 3, 23)).toBe('2026-03-23')
  })

  it('pads single-digit month and day', () => {
    expect(formatDateStr(2026, 1, 5)).toBe('2026-01-05')
  })
})

describe('formatPeriod', () => {
  it('returns YYYY-MM for a date', () => {
    expect(formatPeriod(new Date(2026, 2, 23))).toBe('2026-03')
  })

  it('pads single-digit month', () => {
    expect(formatPeriod(new Date(2026, 0, 15))).toBe('2026-01')
  })
})

describe('getMonday', () => {
  it('returns Monday for a Wednesday', () => {
    // 2026-03-25 is a Wednesday
    const result = getMonday(new Date(2026, 2, 25))
    expect(formatDate(result)).toBe('2026-03-23')
  })

  it('returns Monday itself for a Monday', () => {
    // 2026-03-23 is a Monday
    const result = getMonday(new Date(2026, 2, 23))
    expect(formatDate(result)).toBe('2026-03-23')
  })

  it('returns previous Monday for a Sunday', () => {
    // 2026-03-29 is a Sunday
    const result = getMonday(new Date(2026, 2, 29))
    expect(formatDate(result)).toBe('2026-03-23')
  })

  it('sets time to midnight', () => {
    const result = getMonday(new Date(2026, 2, 25, 15, 30, 0))
    expect(result.getHours()).toBe(0)
    expect(result.getMinutes()).toBe(0)
    expect(result.getSeconds()).toBe(0)
  })

  it('does not mutate the original date', () => {
    const d = new Date(2026, 2, 25)
    getMonday(d)
    expect(formatDate(d)).toBe('2026-03-25')
  })
})

describe('displayDate', () => {
  it('returns a non-empty string for a valid ISO date', () => {
    const result = displayDate('2026-03-23T10:00:00Z')
    expect(typeof result).toBe('string')
    expect(result.length).toBeGreaterThan(0)
  })

  it('returns a non-empty string for a date-only ISO string', () => {
    const result = displayDate('2026-03-23')
    expect(typeof result).toBe('string')
    expect(result.length).toBeGreaterThan(0)
  })
})
