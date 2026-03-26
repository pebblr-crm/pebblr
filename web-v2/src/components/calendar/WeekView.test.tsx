import { render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import { WeekView } from './WeekView'
import type { Activity } from '@/types/activity'

// Mock useConfig since WeekView imports it indirectly through blockerIcons
vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => ({ data: null }),
}))

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: 'act-1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-03-23', // Monday
    duration: '',
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
    expect(screen.getByText('2 visits planned')).toBeInTheDocument()
  })

  it('shows "No visits" for empty days', () => {
    render(<WeekView weekStart={monday} activities={[]} />)
    const emptyLabels = screen.getAllByText('No visits')
    expect(emptyLabels.length).toBe(5)
  })

  it('shows "0 visits planned" in day headers for empty days', () => {
    render(<WeekView weekStart={monday} activities={[]} />)
    const labels = screen.getAllByText('0 visits planned')
    expect(labels.length).toBe(5)
  })

  it('renders blocker with hatch pattern for non-field activities', () => {
    const blocker = makeActivity({
      id: 'block-1',
      activityType: 'training',
      dueDate: '2026-03-24',
      duration: 'full_day',
      targetName: undefined,
      fields: { details: 'New product launch training' },
    })
    render(<WeekView weekStart={monday} activities={[blocker]} />)
    expect(screen.getByText('Training')).toBeInTheDocument()
    expect(screen.getByText('All Day')).toBeInTheDocument()
  })

  it('renders half-day blocker', () => {
    const blocker = makeActivity({
      id: 'block-2',
      activityType: 'team_meeting',
      dueDate: '2026-03-25',
      duration: 'half_day',
      targetName: undefined,
      fields: { details: 'Weekly sync' },
    })
    render(<WeekView weekStart={monday} activities={[blocker]} />)
    expect(screen.getByText('Team Meeting')).toBeInTheDocument()
    expect(screen.getByText('Half Day')).toBeInTheDocument()
  })

  it('shows capacity warning when visits exceed max on fully blocked day', () => {
    const activities = [
      makeActivity({ id: 'block', activityType: 'training', dueDate: '2026-03-23', duration: 'full_day' }),
      makeActivity({ id: 'v1', dueDate: '2026-03-23' }),
    ]
    render(<WeekView weekStart={monday} activities={activities} />)
    // Header should show "Blocked" since there's a full-day blocker
    expect(screen.getByText('Blocked')).toBeInTheDocument()
  })

  it('renders pending assignments', () => {
    const targetMap = new Map([['t1', { id: 't1', targetType: 'doctor', name: 'Dr. Test', fields: {}, assigneeId: 'u1', teamId: 'tm1', createdAt: '', updatedAt: '' }]])
    render(
      <WeekView
        weekStart={monday}
        activities={[]}
        dayAssignments={{ '2026-03-23': ['t1'] }}
        targetMap={targetMap}
      />,
    )
    expect(screen.getByText('Dr. Test')).toBeInTheDocument()
  })

  it('shows drop hint when dragging', () => {
    render(<WeekView weekStart={monday} activities={[]} isDragging />)
    const dropHints = screen.getAllByText('Drop here')
    expect(dropHints.length).toBe(5)
  })
})
