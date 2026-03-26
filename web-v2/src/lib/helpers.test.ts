import { str, daysAgo } from './helpers'

describe('str', () => {
  it('returns the string for string input', () => {
    expect(str('hello')).toBe('hello')
  })

  it('returns empty string for null', () => {
    expect(str(null)).toBe('')
  })

  it('returns empty string for undefined', () => {
    expect(str(undefined)).toBe('')
  })

  it('returns empty string for numbers', () => {
    expect(str(42)).toBe('')
  })

  it('returns empty string for objects', () => {
    expect(str({})).toBe('')
  })

  it('returns empty string for boolean', () => {
    expect(str(true)).toBe('')
  })
})

describe('daysAgo', () => {
  it('returns 0 for today', () => {
    const today = new Date().toISOString()
    expect(daysAgo(today)).toBe(0)
  })

  it('returns positive number for past dates', () => {
    const past = new Date()
    past.setDate(past.getDate() - 5)
    expect(daysAgo(past.toISOString())).toBe(5)
  })

  it('handles date-only strings', () => {
    const past = new Date()
    past.setDate(past.getDate() - 3)
    const dateStr = past.toISOString().slice(0, 10)
    // Should be approximately 3 (may vary by a few hours depending on time of day)
    expect(daysAgo(dateStr)).toBeGreaterThanOrEqual(2)
    expect(daysAgo(dateStr)).toBeLessThanOrEqual(4)
  })
})
