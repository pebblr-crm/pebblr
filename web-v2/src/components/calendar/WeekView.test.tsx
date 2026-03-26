import { render, screen } from '@testing-library/react'
import { WeekView } from './WeekView'
import type { Activity } from '@/types/activity'

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: 'act-1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-03-23', // Monday
    duration: 'full_day',
    fields: {},
    creatorId: 'rep-1',
    createdAt: '2026-03-23T00:00:00Z',
    updatedAt: '2026-03-23T00:00:00Z',
    targetName: 'Apex Medical',
    ...overrides,
  }
}

describe('WeekView', () => {
  const monday = new Date(2026, 2, 23) // March 23, 2026 is a Monday

  it('renders 5 day columns', () => {
    render(<WeekView weekStart={monday} activities={[]} />)
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Tue')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Thu')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
  })

  it('renders activity in the correct day', () => {
    const activity = makeActivity({ dueDate: '2026-03-23' })
    render(<WeekView weekStart={monday} activities={[activity]} />)
    expect(screen.getByText('Apex Medical')).toBeInTheDocument()
  })

  it('shows visit count per day', () => {
    const activities = [
      makeActivity({ id: 'a1', dueDate: '2026-03-23' }),
      makeActivity({ id: 'a2', dueDate: '2026-03-23', targetName: 'Beta Clinic' }),
    ]
    render(<WeekView weekStart={monday} activities={activities} />)
    expect(screen.getByText('2 visits')).toBeInTheDocument()
  })

  it('shows "No visits" for empty days', () => {
    render(<WeekView weekStart={monday} activities={[]} />)
    const emptyLabels = screen.getAllByText('No visits')
    expect(emptyLabels.length).toBe(5)
  })

  it('renders status badge', () => {
    const activity = makeActivity({ status: 'planificat' })
    render(<WeekView weekStart={monday} activities={[activity]} />)
    expect(screen.getByText('planificat')).toBeInTheDocument()
  })
})
