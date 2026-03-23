import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { Activity } from '../../types/activity'
import type { TenantConfig } from '../../types/config'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className }: { children?: React.ReactNode; className?: string }) => (
      <a className={className}>{children}</a>
    ),
    useNavigate: () => vi.fn(),
  }
})

vi.mock('../../services/activities', () => ({
  useActivity: vi.fn(),
  usePatchActivity: () => ({ mutate: vi.fn(), isPending: false }),
  usePatchActivityStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useSubmitActivity: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('../../services/config', () => ({
  useConfig: vi.fn(),
}))

vi.mock('../../services/targets', () => ({
  useTarget: () => ({ data: undefined }),
  useTargets: () => ({ data: { items: [] } }),
}))

vi.mock('../../services/teams', () => ({
  useTeamMembers: () => ({ data: { items: [], total: 0, page: 1, limit: 20 } }),
}))

vi.mock('../../hooks/useToast', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}))

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
      <div {...props}>{children}</div>
    ),
  },
}))

import { useActivity } from '../../services/activities'
import { ActivityDetailInner } from './$activityId'

vi.mocked(useActivity)

const testConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: { types: [] },
  activities: {
    statuses: [
      { key: 'planificat', label: 'Planned', initial: true },
      { key: 'realizat', label: 'Completed' },
    ],
    status_transitions: { planificat: ['realizat'] },
    durations: [{ key: 'full_day', label: 'Full Day' }],
    types: [
      {
        key: 'visit',
        label: 'Visit',
        category: 'field',
        fields: [],
      },
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

const testActivity: Activity = {
  id: 'act-1',
  activityType: 'visit',
  status: 'planificat',
  dueDate: '2026-03-23T00:00:00Z',
  duration: 'full_day',
  fields: {},
  creatorId: 'user-1',
  createdAt: '2026-03-23T00:00:00Z',
  updatedAt: '2026-03-23T00:00:00Z',
}

describe('ActivityDetailInner', () => {
  beforeEach(() => {
    vi.mocked(useActivity).mockReturnValue({
      isLoading: false,
      isError: false,
      data: testActivity,
      error: null,
    } as ReturnType<typeof useActivity>)
  })

  it('renders status badge', async () => {
    render(<ActivityDetailInner activityId="act-1" config={testConfig} />)
    await waitFor(() => {
      expect(screen.getByTestId('status-badge')).toBeInTheDocument()
    })
  })

  it('shows submit button when not submitted', async () => {
    render(<ActivityDetailInner activityId="act-1" config={testConfig} />)
    await waitFor(() => {
      const buttons = screen.getAllByTestId('submit-report-button')
      expect(buttons.length).toBeGreaterThan(0)
    })
  })

  it('shows Submitted lock badge and hides submit button when submittedAt is set', async () => {
    const submitted: Activity = { ...testActivity, submittedAt: '2026-03-23T12:00:00Z' }
    vi.mocked(useActivity).mockReturnValue({
      isLoading: false,
      isError: false,
      data: submitted,
      error: null,
    } as ReturnType<typeof useActivity>)
    render(<ActivityDetailInner activityId="act-1" config={testConfig} />)
    await waitFor(() => {
      expect(screen.getByText(/Submitted/i)).toBeInTheDocument()
      expect(screen.queryByTestId('submit-report-button')).toBeNull()
    })
  })

  it('renders duration select with correct value', async () => {
    render(<ActivityDetailInner activityId="act-1" config={testConfig} />)
    await waitFor(() => {
      const select = screen.getByTestId('duration-select') as HTMLSelectElement
      expect(select.value).toBe('full_day')
    })
  })

  it('renders due date input', async () => {
    render(<ActivityDetailInner activityId="act-1" config={testConfig} />)
    await waitFor(() => {
      expect(screen.getByTestId('due-date-input')).toBeInTheDocument()
    })
  })
})
