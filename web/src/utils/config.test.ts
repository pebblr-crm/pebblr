import { describe, it, expect } from 'vitest'
import {
  getTypeConfig,
  getTypeLabel,
  getTypeCategory,
  getStatusLabel,
  getDurationLabel,
  getActivityTitle,
  getActivityDisplayName,
  CATEGORY_COLORS,
  MONTH_NAMES,
} from './config'
import type { ActivitiesConfig, TenantConfig } from '@/types/config'
import type { Activity } from '@/types/activity'

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
  // i18n defaults to English in tests, so configLabels.type.visit = 'Visit'
  it('returns i18n label for known type', () => {
    expect(getTypeLabel(mockConfig, 'visit')).toBe('Visit')
  })

  it('returns key as fallback for unknown type', () => {
    expect(getTypeLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns i18n label when config is undefined but translation exists', () => {
    expect(getTypeLabel(undefined, 'visit')).toBe('Visit')
  })

  it('returns raw key when no config and no translation', () => {
    expect(getTypeLabel(undefined, 'custom_unknown')).toBe('custom_unknown')
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
  // 'planificat' has no English i18n key, so falls back to config label
  it('returns config label for status without i18n key', () => {
    expect(getStatusLabel(mockConfig, 'planificat')).toBe('Planificat')
    expect(getStatusLabel(mockConfig, 'realizat')).toBe('Realizat')
  })

  it('returns key as fallback for unknown status', () => {
    expect(getStatusLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns key as fallback when config is undefined', () => {
    expect(getStatusLabel(undefined, 'planificat')).toBe('planificat')
  })

  // 'planned' has an English i18n key
  it('returns i18n label for known status key', () => {
    const enConfig: ActivitiesConfig = {
      ...mockConfig,
      statuses: [{ key: 'planned', label: 'Planned', initial: true }],
    }
    expect(getStatusLabel(enConfig, 'planned')).toBe('Planned')
  })
})

describe('getDurationLabel', () => {
  it('returns config label for duration without i18n key', () => {
    expect(getDurationLabel(mockConfig, '30m')).toBe('30 minutes')
    expect(getDurationLabel(mockConfig, '1h')).toBe('1 hour')
  })

  it('returns key as fallback for unknown duration', () => {
    expect(getDurationLabel(mockConfig, 'unknown')).toBe('unknown')
  })

  it('returns key as fallback when config is undefined', () => {
    expect(getDurationLabel(undefined, '30m')).toBe('30m')
  })

  it('returns i18n label for known duration key', () => {
    const enConfig: ActivitiesConfig = {
      ...mockConfig,
      durations: [{ key: 'full_day', label: 'Full Day' }],
    }
    expect(getDurationLabel(enConfig, 'full_day')).toBe('Full Day')
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

describe('MONTH_NAMES', () => {
  it('has 12 entries', () => {
    expect(MONTH_NAMES).toHaveLength(12)
  })

  it('starts with January and ends with December', () => {
    expect(MONTH_NAMES[0]).toBe('January')
    expect(MONTH_NAMES[11]).toBe('December')
  })
})

// ── getActivityTitle / getActivityDisplayName ─────────────────────────────

const fullConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: { types: [] },
  activities: mockConfig,
  options: {},
  rules: {
    frequency: {},
    max_activities_per_day: 10,
    default_visit_duration_minutes: {},
    visit_duration_step_minutes: 30,
  },
}

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: 'a1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-03-24',
    duration: '1h',
    fields: {},
    creatorId: 'u1',
    createdAt: '2026-03-24T00:00:00Z',
    updatedAt: '2026-03-24T00:00:00Z',
    ...overrides,
  }
}

describe('getActivityTitle', () => {
  it('returns i18n type label with target name for field activities', () => {
    const a = makeActivity({ targetName: 'Dr. Popescu' })
    expect(getActivityTitle(fullConfig, a)).toBe('Visit — Dr. Popescu')
  })

  it('returns just type label when no target name', () => {
    const a = makeActivity({ activityType: 'training' })
    expect(getActivityTitle(fullConfig, a)).toBe('Training')
  })

  it('returns label override when set', () => {
    const a = makeActivity({ label: 'Custom Label', targetName: 'Dr. Popescu' })
    expect(getActivityTitle(fullConfig, a)).toBe('Custom Label')
  })

  it('falls back to i18n label when config is undefined', () => {
    const a = makeActivity({ targetName: 'Dr. Popescu' })
    expect(getActivityTitle(undefined, a)).toBe('Visit — Dr. Popescu')
  })

  it('returns i18n label for non-field activity without config', () => {
    const a = makeActivity({ activityType: 'training' })
    expect(getActivityTitle(undefined, a)).toBe('Training')
  })
})

describe('getActivityDisplayName', () => {
  it('appends short date to title', () => {
    const a = makeActivity({ targetName: 'Dr. Popescu', dueDate: '2026-03-24' })
    const result = getActivityDisplayName(fullConfig, a)
    expect(result).toBe('Visit — Dr. Popescu — 24 Mar')
  })

  it('works for non-field activities without target', () => {
    const a = makeActivity({ activityType: 'training', dueDate: '2026-01-15' })
    const result = getActivityDisplayName(fullConfig, a)
    expect(result).toBe('Training — 15 Jan')
  })

  it('uses label override with date', () => {
    const a = makeActivity({ label: 'Custom', dueDate: '2026-12-25' })
    const result = getActivityDisplayName(fullConfig, a)
    expect(result).toBe('Custom — 25 Dec')
  })
})
