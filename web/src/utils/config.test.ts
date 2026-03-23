import { describe, it, expect } from 'vitest'
import {
  getTypeConfig,
  getTypeLabel,
  getTypeCategory,
  getStatusLabel,
  getDurationLabel,
  CATEGORY_COLORS,
  STATUS_BADGE_COLORS,
  MONTH_NAMES,
} from './config'
import type { ActivitiesConfig } from '@/types/config'

const mockConfig: ActivitiesConfig = {
  types: [
    { key: 'visit', label: 'Vizită', category: 'field', fields: [] },
    { key: 'training', label: 'Training', category: 'non_field', fields: [] },
  ],
  statuses: [
    { key: 'planificat', label: 'Planificat', initial: true },
    { key: 'realizat', label: 'Realizat' },
    { key: 'anulat', label: 'Anulat' },
  ],
  durations: [
    { key: '30m', label: '30 minutes' },
    { key: '1h', label: '1 hour' },
  ],
  status_transitions: { planificat: ['realizat', 'anulat'] },
  routing_options: [],
}

describe('getTypeConfig', () => {
  it('returns matching type config', () => {
    expect(getTypeConfig(mockConfig, 'visit')).toEqual(mockConfig.types[0])
  })

  it('returns undefined for unknown key', () => {
    expect(getTypeConfig(mockConfig, 'unknown')).toBeUndefined()
  })

  it('returns undefined when config is undefined', () => {
    expect(getTypeConfig(undefined, 'visit')).toBeUndefined()
  })
})

describe('getTypeLabel', () => {
  it('returns label for known type', () => {
    expect(getTypeLabel(mockConfig, 'visit')).toBe('Vizită')
  })

  it('returns key as fallback for unknown type', () => {
    expect(getTypeLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns key as fallback when config is undefined', () => {
    expect(getTypeLabel(undefined, 'visit')).toBe('visit')
  })
})

describe('getTypeCategory', () => {
  it('returns field category', () => {
    expect(getTypeCategory(mockConfig, 'visit')).toBe('field')
  })

  it('returns non_field category', () => {
    expect(getTypeCategory(mockConfig, 'training')).toBe('non_field')
  })

  it('returns field as fallback for unknown type', () => {
    expect(getTypeCategory(mockConfig, 'unknown')).toBe('field')
  })

  it('returns field as fallback when config is undefined', () => {
    expect(getTypeCategory(undefined, 'visit')).toBe('field')
  })
})

describe('getStatusLabel', () => {
  it('returns label for known status', () => {
    expect(getStatusLabel(mockConfig, 'planificat')).toBe('Planificat')
    expect(getStatusLabel(mockConfig, 'realizat')).toBe('Realizat')
  })

  it('returns key as fallback for unknown status', () => {
    expect(getStatusLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns key as fallback when config is undefined', () => {
    expect(getStatusLabel(undefined, 'planificat')).toBe('planificat')
  })
})

describe('getDurationLabel', () => {
  it('returns label for known duration', () => {
    expect(getDurationLabel(mockConfig, '30m')).toBe('30 minutes')
    expect(getDurationLabel(mockConfig, '1h')).toBe('1 hour')
  })

  it('returns key as fallback for unknown duration', () => {
    expect(getDurationLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns key as fallback when config is undefined', () => {
    expect(getDurationLabel(undefined, '30m')).toBe('30m')
  })
})

describe('CATEGORY_COLORS', () => {
  it('has field and non_field entries', () => {
    expect(CATEGORY_COLORS).toHaveProperty('field')
    expect(CATEGORY_COLORS).toHaveProperty('non_field')
  })

  it('field includes amber colors', () => {
    expect(CATEGORY_COLORS.field).toContain('amber')
  })

  it('non_field includes blue colors', () => {
    expect(CATEGORY_COLORS.non_field).toContain('blue')
  })
})

describe('STATUS_BADGE_COLORS', () => {
  it('has planificat, realizat, anulat entries', () => {
    expect(STATUS_BADGE_COLORS).toHaveProperty('planificat')
    expect(STATUS_BADGE_COLORS).toHaveProperty('realizat')
    expect(STATUS_BADGE_COLORS).toHaveProperty('anulat')
  })
})

describe('MONTH_NAMES', () => {
  it('has 12 entries', () => {
    expect(MONTH_NAMES).toHaveLength(12)
  })

  it('starts with January and ends with December', () => {
    expect(MONTH_NAMES[0]).toBe('January')
    expect(MONTH_NAMES[11]).toBe('December')
  })
})
