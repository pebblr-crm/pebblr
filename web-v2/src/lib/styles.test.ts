import {
  statusVariant,
  statusDot,
  priorityStyle,
  priorityDot,
  priorityLabel,
  transitionColors,
  visitTypeBadge,
} from './styles'

describe('statusVariant', () => {
  it('maps planificat to primary', () => {
    expect(statusVariant.planificat).toBe('primary')
  })
  it('maps realizat to success', () => {
    expect(statusVariant.realizat).toBe('success')
  })
  it('maps anulat to danger', () => {
    expect(statusVariant.anulat).toBe('danger')
  })
})

describe('statusDot', () => {
  it('maps both Romanian and English status keys', () => {
    expect(statusDot.realizat).toBe('bg-emerald-500')
    expect(statusDot.completed).toBe('bg-emerald-500')
    expect(statusDot.planificat).toBe('bg-blue-500')
    expect(statusDot.planned).toBe('bg-blue-500')
    expect(statusDot.anulat).toBe('bg-red-500')
    expect(statusDot.cancelled).toBe('bg-red-500')
  })
})

describe('priorityStyle', () => {
  it('has styles for a, b, c', () => {
    expect(priorityStyle.a).toContain('red')
    expect(priorityStyle.b).toContain('amber')
    expect(priorityStyle.c).toContain('slate')
  })
})

describe('priorityDot', () => {
  it('has dot colors for a, b, c', () => {
    expect(priorityDot.a).toBe('bg-red-500')
    expect(priorityDot.b).toBe('bg-amber-500')
    expect(priorityDot.c).toBe('bg-slate-400')
  })
})

describe('priorityLabel', () => {
  it('maps to human-readable labels', () => {
    expect(priorityLabel.a).toBe('Priority A')
    expect(priorityLabel.b).toBe('Priority B')
    expect(priorityLabel.c).toBe('Priority C')
  })
})

describe('transitionColors', () => {
  it('has button colors for status transitions', () => {
    expect(transitionColors.realizat).toContain('emerald')
    expect(transitionColors.anulat).toContain('red')
  })
})

describe('visitTypeBadge', () => {
  it('returns amber styling for f2f', () => {
    const result = visitTypeBadge('f2f')
    expect(result.className).toContain('amber')
    expect(result.label).toBe('In person')
  })

  it('returns blue styling for remote', () => {
    const result = visitTypeBadge('remote')
    expect(result.className).toContain('blue')
    expect(result.label).toBe('Remote')
  })

  it('defaults to remote for unknown types', () => {
    const result = visitTypeBadge('unknown')
    expect(result.label).toBe('Remote')
  })
})
