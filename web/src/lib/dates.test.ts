import { getMonday, addDays, formatDate } from './dates'

describe('getMonday', () => {
  it('returns Monday for a Monday input', () => {
    const monday = new Date(2026, 2, 23) // March 23, 2026 is a Monday
    const result = getMonday(monday)
    expect(result.getDay()).toBe(1) // Monday
    expect(result.getDate()).toBe(23)
  })

  it('returns previous Monday for a Wednesday input', () => {
    const wed = new Date(2026, 2, 25) // March 25, 2026 is a Wednesday
    const result = getMonday(wed)
    expect(result.getDay()).toBe(1)
    expect(result.getDate()).toBe(23)
  })

  it('returns previous Monday for a Sunday input', () => {
    const sun = new Date(2026, 2, 29) // March 29, 2026 is a Sunday
    const result = getMonday(sun)
    expect(result.getDay()).toBe(1)
    expect(result.getDate()).toBe(23)
  })

  it('zeroes out time', () => {
    const d = new Date(2026, 2, 25, 14, 30, 0)
    const result = getMonday(d)
    expect(result.getHours()).toBe(0)
    expect(result.getMinutes()).toBe(0)
    expect(result.getSeconds()).toBe(0)
  })
})

describe('addDays', () => {
  it('adds positive days', () => {
    const d = new Date(2026, 2, 23)
    const result = addDays(d, 3)
    expect(result.getDate()).toBe(26)
  })

  it('subtracts with negative days', () => {
    const d = new Date(2026, 2, 23)
    const result = addDays(d, -7)
    expect(result.getDate()).toBe(16)
  })

  it('does not mutate original date', () => {
    const d = new Date(2026, 2, 23)
    addDays(d, 5)
    expect(d.getDate()).toBe(23)
  })
})

describe('formatDate', () => {
  it('formats a date as YYYY-MM-DD using local time', () => {
    const d = new Date(2026, 2, 5) // March 5
    expect(formatDate(d)).toBe('2026-03-05')
  })

  it('pads single-digit months and days', () => {
    const d = new Date(2026, 0, 3) // January 3
    expect(formatDate(d)).toBe('2026-01-03')
  })

  it('handles December correctly', () => {
    const d = new Date(2026, 11, 31) // December 31
    expect(formatDate(d)).toBe('2026-12-31')
  })
})
