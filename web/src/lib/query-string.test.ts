import { buildQueryString } from './query-string'

describe('buildQueryString', () => {
  it('appends params to the base path', () => {
    const result = buildQueryString('/items', { page: 1, limit: 20 })
    expect(result).toBe('/items?page=1&limit=20')
  })

  it('omits undefined values', () => {
    const result = buildQueryString('/items', { page: 1, status: undefined })
    expect(result).toBe('/items?page=1')
  })

  it('omits null values', () => {
    const result = buildQueryString('/items', { page: 1, status: null })
    expect(result).toBe('/items?page=1')
  })

  it('omits empty-string values', () => {
    const result = buildQueryString('/items', { page: 1, status: '' })
    expect(result).toBe('/items?page=1')
  })

  it('returns the bare path when all params are empty', () => {
    const result = buildQueryString('/items', { status: undefined, q: '' })
    expect(result).toBe('/items')
  })

  it('returns the bare path when params object is empty', () => {
    const result = buildQueryString('/items', {})
    expect(result).toBe('/items')
  })

  it('converts numbers to strings', () => {
    const result = buildQueryString('/items', { page: 2 })
    expect(result).toBe('/items?page=2')
  })
})
