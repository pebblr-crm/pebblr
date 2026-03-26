import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { PlannerDailyPage } from './daily'
import { formatDate } from '@/utils/date'
import type { Activity } from '../../types/activity'
import type { PaginatedResponse } from '../../types/api'
import type { TenantConfig } from '../../types/config'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className, to, ...rest }: Record<string, unknown>) => (
      <a className={className as string} href={to as string} data-testid={rest['data-testid'] as string}>{children as React.ReactNode}</a>
    ),
  }
})

vi.mock('../../services/activities', () => ({
  useActivities: vi.fn(),
}))

vi.mock('../../services/config', () => ({
  useConfig: vi.fn(),
}))

import { useActivities } from '../../services/activities'
import { useConfig } from '../../services/config'
const mockUseActivities = vi.mocked(useActivities)
const mockUseConfig = vi.mocked(useConfig)

const testConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: { types: [] },
  activities: {
    statuses: [
      { key: 'planificat', label: 'Planned', initial: true },
      { key: 'realizat', label: 'Realized' },
    ],
    status_transitions: {},
    durations: [{ key: 'full_day', label: 'Full Day' }],
    types: [
      { key: 'visit', label: 'Visit', category: 'field', fields: [] },
    ],
    routing_options: [],
  },
  options: {},
  rules: {
    frequency: {},
    max_activities_per_day: 10,
    default_visit_duration_minutes: {},
    visit_duration_step_minutes: 30,
  },
}

function todayStr(): string {
  return formatDate(new Date())
}

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: 'act-1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: todayStr(),
    duration: 'full_day',
    fields: {},
    creatorId: 'user-1',
    createdAt: '2026-03-20T10:00:00Z',
    updatedAt: '2026-03-20T10:00:00Z',
    ...overrides,
  }
}

function makePage(items: Activity[]): PaginatedResponse<Activity> {
  return { items, total: items.length, page: 1, limit: 50 }
}

function setupConfig() {
  mockUseConfig.mockReturnValue({
    data: testConfig,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useConfig>)
}

function setupActivities(items: Activity[] = []) {
  mockUseActivities.mockReturnValue({
    data: makePage(items),
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useActivities>)
}

describe('PlannerDailyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupConfig()
  })

  it('shows loading spinner while fetching', () => {
    mockUseActivities.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useActivities>)

    render(<PlannerDailyPage />)

    expect(screen.getByLabelText('Loading activities...')).toBeInTheDocument()
  })

  it('shows empty state when no activities', () => {
    setupActivities([])

    render(<PlannerDailyPage />)

    expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    expect(screen.getByText('No activities scheduled for this day.')).toBeInTheDocument()
  })

  it('renders activity rows when data present', () => {
    setupActivities([
      makeActivity({ id: 'a1' }),
      makeActivity({ id: 'a2', activityType: 'visit' }),
    ])

    render(<PlannerDailyPage />)

    expect(screen.getByTestId('daily-activities')).toBeInTheDocument()
    expect(screen.getAllByTestId('daily-activity-row')).toHaveLength(2)
  })

  it('shows activity type label from config', () => {
    setupActivities([makeActivity()])

    render(<PlannerDailyPage />)

    expect(screen.getByText('Visit')).toBeInTheDocument()
  })

  it('shows status badge from config', () => {
    setupActivities([makeActivity()])

    render(<PlannerDailyPage />)

    expect(screen.getByText('Planned')).toBeInTheDocument()
  })

  it('shows target name in compact layout', () => {
    setupActivities([makeActivity({ targetName: 'Dr. Popescu' })])

    render(<PlannerDailyPage />)

    expect(screen.getByText('Visit — Dr. Popescu')).toBeInTheDocument()
  })

  it('shows back to planner link', () => {
    setupActivities([])

    render(<PlannerDailyPage />)

    expect(screen.getByText('Back to planner')).toBeInTheDocument()
  })

  it('shows daily summary count', () => {
    setupActivities([makeActivity()])

    render(<PlannerDailyPage />)

    expect(screen.getByTestId('daily-summary')).toHaveTextContent('1 activity scheduled')
  })

  it('navigates days with prev/next buttons', async () => {
    const user = userEvent.setup()
    setupActivities([])

    render(<PlannerDailyPage />)

    const dateBefore = screen.getByTestId('daily-date').textContent
    await user.click(screen.getByLabelText('Next day'))
    const dateAfter = screen.getByTestId('daily-date').textContent

    expect(dateAfter).not.toBe(dateBefore)
  })

  it('shows submitted badge when activity is submitted', () => {
    setupActivities([makeActivity({ submittedAt: '2026-03-23T10:00:00Z' })])

    render(<PlannerDailyPage />)

    expect(screen.getByText('Submitted')).toBeInTheDocument()
  })

  it('shows (Today) indicator for current date', () => {
    setupActivities([])

    render(<PlannerDailyPage />)

    expect(screen.getByText('(Today)')).toBeInTheDocument()
  })

  it('passes correct date to useActivities', () => {
    setupActivities([])

    render(<PlannerDailyPage />)

    const today = todayStr()
    expect(mockUseActivities).toHaveBeenCalledWith(
      expect.objectContaining({
        dateFrom: today,
        dateTo: today,
        limit: 50,
      }),
    )
  })
})
