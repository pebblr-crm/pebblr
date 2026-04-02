import { describe, it, expect } from 'vitest'
import { buildQuery } from './query'

describe('buildQuery', () => {
  it('returns base path when all params are undefined', () => {
    expect(buildQuery('/targets', { page: undefined, limit: undefined })).toBe('/targets')
  })

  it('returns base path when params object is empty', () => {
    expect(buildQuery('/audit', {})).toBe('/audit')
  })

  it('appends defined params as query string', () => {
    const result = buildQuery('/targets', { page: 1, limit: 20 })
    expect(result).toBe('/targets?page=1&limit=20')
  })

  it('omits undefined and null values', () => {
    const result = buildQuery('/activities', {
      page: 1,
      status: undefined,
      type: null,
      limit: 10,
    })
    expect(result).toBe('/activities?page=1&limit=10')
  })

  it('omits empty string values', () => {
    const result = buildQuery('/audit', { entityType: '', actorId: 'abc' })
    expect(result).toBe('/audit?actorId=abc')
  })

  it('converts numbers to strings', () => {
    const result = buildQuery('/targets', { page: 2 })
    expect(result).toBe('/targets?page=2')
  })
})
