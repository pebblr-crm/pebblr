import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { TenantConfig } from '../types/config'
import type {
  DashboardStatsResponse,
  CoverageStats,
  UserStatsResponse,
} from '../types/dashboard'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    createRoute: () => ({ component: null }),
    Link: ({ children, className, ...rest }: { children?: React.ReactNode; className?: string; to?: string }) => (
      <a className={className} href={rest.to}>{children}</a>
    ),
  }
})

vi.mock('../services/dashboard', () => ({
  useDashboardStats: vi.fn(),
  useCoverageStats: vi.fn(),
  useUserStats: vi.fn(),
}))

vi.mock('../services/config', () => ({
  useConfig: vi.fn(),
}))

import { useDashboardStats, useCoverageStats, useUserStats } from '../services/dashboard'
import { useConfig } from '../services/config'

const mockUseDashboardStats = vi.mocked(useDashboardStats)
const mockUseCoverageStats = vi.mocked(useCoverageStats)
const mockUseUserStats = vi.mocked(useUserStats)
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

const testStats: DashboardStatsResponse = {
  period: '2026-03',
  dateFrom: '2026-03-01',
  dateTo: '2026-03-31',
  stats: {
    total: 50,
    submittedCount: 30,
    byStatus: [
      { status: 'realizat', count: 30 },
      { status: 'planificat', count: 15 },
      { status: 'anulat', count: 5 },
    ],
    byType: [
      { activityType: 'visit', count: 40 },
      { activityType: 'vacation', count: 10 },
    ],
  },
  byCategory: [
    { category: 'field', count: 40 },
    { category: 'non_field', count: 10 },
  ],
}

const testCoverage: CoverageStats = {
  totalTargets: 100,
  visitedTargets: 75,
  coveragePercent: 75.0,
}

const testUserStats: UserStatsResponse = {
  users: [
    {
      userId: 'rep-1',
      userName: 'Alice Rep',
      total: 25,
      byStatus: [
        { status: 'realizat', count: 20 },
        { status: 'planificat', count: 5 },
      ],
    },
    {
      userId: 'rep-2',
      userName: 'Bob Rep',
      total: 25,
      byStatus: [
        { status: 'realizat', count: 10 },
        { status: 'planificat', count: 10 },
        { status: 'anulat', count: 5 },
      ],
    },
  ],
}

beforeEach(() => {
  vi.clearAllMocks()

  mockUseConfig.mockReturnValue({
    data: testConfig,
    isLoading: false,
  } as ReturnType<typeof useConfig>)

  mockUseDashboardStats.mockReturnValue({
    data: testStats,
    isLoading: false,
  } as ReturnType<typeof useDashboardStats>)

  mockUseCoverageStats.mockReturnValue({
    data: testCoverage,
    isLoading: false,
  } as ReturnType<typeof useCoverageStats>)

  mockUseUserStats.mockReturnValue({
    data: testUserStats,
    isLoading: false,
  } as ReturnType<typeof useUserStats>)
})

describe('PeriodSelector', () => {
  it('renders month and year', async () => {
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2026-03" onPeriodChange={onChange} />)

    expect(screen.getByText('March 2026')).toBeTruthy()
  })

  it('navigates to previous month', async () => {
    const user = userEvent.setup()
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2026-03" onPeriodChange={onChange} />)

    await user.click(screen.getByLabelText('Previous month'))
    expect(onChange).toHaveBeenCalledWith('2026-02')
  })

  it('navigates to next month', async () => {
    const user = userEvent.setup()
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2026-03" onPeriodChange={onChange} />)

    await user.click(screen.getByLabelText('Next month'))
    expect(onChange).toHaveBeenCalledWith('2026-04')
  })

  it('wraps year boundary going backward', async () => {
    const user = userEvent.setup()
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2026-01" onPeriodChange={onChange} />)

    await user.click(screen.getByLabelText('Previous month'))
    expect(onChange).toHaveBeenCalledWith('2025-12')
  })

  it('wraps year boundary going forward', async () => {
    const user = userEvent.setup()
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2026-12" onPeriodChange={onChange} />)

    await user.click(screen.getByLabelText('Next month'))
    expect(onChange).toHaveBeenCalledWith('2027-01')
  })

  it('Today button navigates to current month', async () => {
    const user = userEvent.setup()
    const { PeriodSelector } = await import('../components/dashboard/PeriodSelector')
    const onChange = vi.fn()
    render(<PeriodSelector period="2025-06" onPeriodChange={onChange} />)

    await user.click(screen.getByText('Today'))
    // Should be called with current month
    expect(onChange).toHaveBeenCalledTimes(1)
    const called = onChange.mock.calls[0][0] as string
    expect(called).toMatch(/^\d{4}-\d{2}$/)
  })
})

describe('UserStatsTable', () => {
  it('renders user rows', async () => {
    const { UserStatsTable } = await import('../components/dashboard/UserStatsTable')
    render(<UserStatsTable users={testUserStats.users} config={testConfig} />)

    expect(screen.getByText('Alice Rep')).toBeTruthy()
    expect(screen.getByText('Bob Rep')).toBeTruthy()
  })

  it('resolves status labels from config', async () => {
    const { UserStatsTable } = await import('../components/dashboard/UserStatsTable')
    render(<UserStatsTable users={testUserStats.users} config={testConfig} />)

    expect(screen.getByText('Realized')).toBeTruthy()
    expect(screen.getByText('Planned')).toBeTruthy()
  })

  it('shows totals per user', async () => {
    const { UserStatsTable } = await import('../components/dashboard/UserStatsTable')
    render(<UserStatsTable users={testUserStats.users} config={testConfig} />)

    // Both users have total 25
    const cells = screen.getAllByText('25')
    expect(cells.length).toBe(2)
  })

  it('shows empty state when no users', async () => {
    const { UserStatsTable } = await import('../components/dashboard/UserStatsTable')
    render(<UserStatsTable users={[]} config={testConfig} />)

    expect(screen.getByText('No activity data for this period.')).toBeTruthy()
  })

  it('shows userId as fallback when name is empty', async () => {
    const { UserStatsTable } = await import('../components/dashboard/UserStatsTable')
    render(
      <UserStatsTable
        users={[{ userId: 'user-99', userName: '', total: 5, byStatus: [] }]}
        config={testConfig}
      />,
    )

    expect(screen.getByText('user-99')).toBeTruthy()
  })
})

describe('StatCard', () => {
  it('renders label and value', async () => {
    const { StatCard } = await import('../components/dashboard/StatCard')
    render(<StatCard label="Total" value="42" />)

    expect(screen.getByText('Total')).toBeTruthy()
    expect(screen.getByText('42')).toBeTruthy()
  })

  it('renders primary variant', async () => {
    const { StatCard } = await import('../components/dashboard/StatCard')
    render(<StatCard label="KPI" value="100" variant="primary" />)

    expect(screen.getByText('KPI')).toBeTruthy()
    expect(screen.getByText('100')).toBeTruthy()
  })

  it('renders change indicator', async () => {
    const { StatCard } = await import('../components/dashboard/StatCard')
    render(<StatCard label="Coverage" value="75%" change="+10%" />)

    expect(screen.getByText('+10%')).toBeTruthy()
  })
})
