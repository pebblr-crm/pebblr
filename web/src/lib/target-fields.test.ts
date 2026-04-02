import { describe, it, expect } from 'vitest'
import { getLat, getLng, getClassification, getCity, hasGeoCoords } from './target-fields'

describe('getLat', () => {
  it('returns number when lat is a number', () => {
    expect(getLat({ lat: 44.4268 })).toBe(44.4268)
  })

  it('returns null when lat is missing', () => {
    expect(getLat({})).toBeNull()
  })

  it('returns null when lat is not a number', () => {
    expect(getLat({ lat: '44.4268' })).toBeNull()
  })
})

describe('getLng', () => {
  it('returns number when lng is a number', () => {
    expect(getLng({ lng: 26.1025 })).toBe(26.1025)
  })

  it('returns null when lng is missing', () => {
    expect(getLng({})).toBeNull()
  })

  it('returns null when lng is not a number', () => {
    expect(getLng({ lng: null })).toBeNull()
  })
})

describe('getClassification', () => {
  it('returns lowercased potential field', () => {
    expect(getClassification({ potential: 'A' })).toBe('a')
  })

  it('defaults to c when potential is missing', () => {
    expect(getClassification({})).toBe('c')
  })

  it('handles already-lowercase values', () => {
    expect(getClassification({ potential: 'b' })).toBe('b')
  })
})

describe('getCity', () => {
  it('returns city string', () => {
    expect(getCity({ city: 'Bucharest' })).toBe('Bucharest')
  })

  it('returns empty string when city is missing', () => {
    expect(getCity({})).toBe('')
  })
})

describe('hasGeoCoords', () => {
  it('returns true when both lat and lng are present', () => {
    expect(hasGeoCoords({ lat: 44.4, lng: 26.1 })).toBe(true)
  })

  it('returns false when lat is missing', () => {
    expect(hasGeoCoords({ lng: 26.1 })).toBe(false)
  })

  it('returns false when lng is missing', () => {
    expect(hasGeoCoords({ lat: 44.4 })).toBe(false)
  })

  it('returns false when both are missing', () => {
    expect(hasGeoCoords({})).toBe(false)
  })
})
