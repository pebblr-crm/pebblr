import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { PlannerPage } from './index'
import type { Activity } from '../../types/activity'
import type { PaginatedResponse } from '../../types/api'
import type { TenantConfig } from '../../types/config'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className, ...rest }: { children?: React.ReactNode; className?: string; to?: string }) => (
      <a className={className} href={rest.to}>{children}</a>
    ),
  }
})

vi.mock('../../services/activities', () => ({
  useActivities: vi.fn(),
}))

vi.mock('../../services/config', () => ({
  useConfig: vi.fn(),
}))

vi.mock('../../services/dashboard', () => ({
  useRecoveryBalance: vi.fn(() => ({ data: undefined })),
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
      { key: 'anulat', label: 'Cancelled' },
    ],
    status_transitions: { planificat: ['realizat', 'anulat'] },
    durations: [
      { key: 'full_day', label: 'Full Day' },
      { key: 'half_day', label: 'Half Day' },
    ],
    types: [
      { key: 'visit', label: 'Visit', category: 'field', fields: [] },
      { key: 'vacation', label: 'Vacation', category: 'non_field', fields: [], blocks_field_activities: true },
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

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: 'act-1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-03-23',
    duration: 'full_day',
    fields: {},
    creatorId: 'user-1',
    createdAt: '2026-03-20T10:00:00Z',
    updatedAt: '2026-03-20T10:00:00Z',
    ...overrides,
  }
}

function makePage(items: Activity[]): PaginatedResponse<Activity> {
  return { items, total: items.length, page: 1, limit: 200 }
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

describe('PlannerPage', () => {
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

    render(<PlannerPage />)

    expect(screen.getByLabelText('Loading planner...')).toBeInTheDocument()
  })

  it('renders page header', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByText('Planner')).toBeInTheDocument()
    expect(screen.getByText('Plan and track field activities.')).toBeInTheDocument()
  })

  it('defaults to week view', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByTestId('week-grid')).toBeInTheDocument()
  })

  it('switches to month view', async () => {
    const user = userEvent.setup()
    setupActivities()

    render(<PlannerPage />)

    await user.click(screen.getByText('Month'))

    expect(screen.getByTestId('month-grid')).toBeInTheDocument()
  })

  it('has view toggle with Week and Month buttons', () => {
    setupActivities()

    render(<PlannerPage />)

    const toggle = screen.getByTestId('view-toggle')
    expect(toggle).toBeInTheDocument()
    expect(screen.getByText('Week')).toBeInTheDocument()
    expect(screen.getByText('Month')).toBeInTheDocument()
  })

  it('shows period label', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByTestId('period-label')).toBeInTheDocument()
  })

  it('navigates to previous period', async () => {
    const user = userEvent.setup()
    setupActivities()

    render(<PlannerPage />)

    const labelBefore = screen.getByTestId('period-label').textContent
    await user.click(screen.getByLabelText('Previous period'))
    const labelAfter = screen.getByTestId('period-label').textContent

    expect(labelAfter).not.toBe(labelBefore)
  })

  it('navigates to next period', async () => {
    const user = userEvent.setup()
    setupActivities()

    render(<PlannerPage />)

    const labelBefore = screen.getByTestId('period-label').textContent
    await user.click(screen.getByLabelText('Next period'))
    const labelAfter = screen.getByTestId('period-label').textContent

    expect(labelAfter).not.toBe(labelBefore)
  })

  it('shows status legend from config', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByText('Planned')).toBeInTheDocument()
    expect(screen.getByText('Realized')).toBeInTheDocument()
    expect(screen.getByText('Cancelled')).toBeInTheDocument()
  })

  it('shows category legend', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByText('Field activities')).toBeInTheDocument()
    expect(screen.getByText('Non-field activities')).toBeInTheDocument()
  })

  it('shows daily pulse stats', () => {
    setupActivities([makeActivity()])

    render(<PlannerPage />)

    expect(screen.getByText('1 In View')).toBeInTheDocument()
  })

  it('passes date range to useActivities', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(mockUseActivities).toHaveBeenCalledWith(
      expect.objectContaining({
        dateFrom: expect.any(String),
        dateTo: expect.any(String),
        limit: 200,
      }),
    )
  })

  it('shows New Activity link', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByText('New Activity')).toBeInTheDocument()
  })

  it('has Today button', () => {
    setupActivities()

    render(<PlannerPage />)

    expect(screen.getByText('Today')).toBeInTheDocument()
  })
})
