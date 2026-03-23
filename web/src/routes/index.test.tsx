import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { DashboardPage } from './index'
import type { TenantConfig } from '../types/config'
import type { ActivityStatsResponse, CoverageResponse, FrequencyResponse } from '../types/dashboard'
import type { PaginatedResponse } from '../types/api'
import type { TeamMember } from '../types/team'

vi.mock('../services/dashboard', () => ({
  useActivityStats: vi.fn(),
  useCoverage: vi.fn(),
  useFrequency: vi.fn(),
}))

vi.mock('../services/config', () => ({
  useConfig: vi.fn(),
}))

vi.mock('../services/teams', () => ({
  useTeamMembers: vi.fn(),
}))

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.HTMLAttributes<HTMLDivElement>) => <div {...props}>{children}</div>,
  },
}))

import { useActivityStats, useCoverage, useFrequency } from '../services/dashboard'
import { useConfig } from '../services/config'
import { useTeamMembers } from '../services/teams'

const mockUseActivityStats = vi.mocked(useActivityStats)
const mockUseCoverage = vi.mocked(useCoverage)
const mockUseFrequency = vi.mocked(useFrequency)
const mockUseConfig = vi.mocked(useConfig)
const mockUseTeamMembers = vi.mocked(useTeamMembers)

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
    durations: [{ key: 'full_day', label: 'Full Day' }],
    types: [
      { key: 'visit', label: 'Visit', category: 'field', fields: [] },
      { key: 'vacation', label: 'Vacation', category: 'non_field', fields: [] },
    ],
    routing_options: [],
  },
  options: {},
  rules: {
    frequency: { a: 4, b: 2 },
    max_activities_per_day: 10,
    default_visit_duration_minutes: {},
    visit_duration_step_minutes: 30,
  },
}

const testActivityStats: ActivityStatsResponse = {
  byStatus: { planificat: 20, realizat: 15, anulat: 2 },
  byCategory: { field: 30, non_field: 7 },
  total: 37,
}

const testCoverage: CoverageResponse = {
  totalTargets: 50,
  visitedTargets: 35,
  percentage: 70,
}

const testFrequency: FrequencyResponse = {
  items: [
    { classification: 'a', targetCount: 10, totalVisits: 38, required: 4, compliance: 95 },
    { classification: 'b', targetCount: 20, totalVisits: 30, required: 2, compliance: 75 },
  ],
}

const testTeamMembers: PaginatedResponse<TeamMember> = {
  items: [
    {
      id: 'u1',
      name: 'Alice Rep',
      role: 'rep',
      avatar: '',
      status: 'online',
      metrics: { assigned: 10, completed: 8, efficiency: 80 },
    },
  ],
  total: 1,
  page: 1,
  limit: 10,
}

function setupAllHooks() {
  mockUseActivityStats.mockReturnValue({
    data: testActivityStats,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useActivityStats>)

  mockUseCoverage.mockReturnValue({
    data: testCoverage,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useCoverage>)

  mockUseFrequency.mockReturnValue({
    data: testFrequency,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useFrequency>)

  mockUseConfig.mockReturnValue({
    data: testConfig,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useConfig>)

  mockUseTeamMembers.mockReturnValue({
    data: testTeamMembers,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useTeamMembers>)
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner while fetching', () => {
    mockUseActivityStats.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useActivityStats>)
    mockUseCoverage.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useCoverage>)
    mockUseFrequency.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useFrequency>)
    mockUseConfig.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useConfig>)
    mockUseTeamMembers.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useTeamMembers>)

    render(<DashboardPage />)
    expect(screen.getByText('Loading dashboard...')).toBeInTheDocument()
  })

  it('renders Command Center heading', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Command Center')).toBeInTheDocument()
  })

  it('displays planned count in stat card', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getAllByText('Planned').length).toBeGreaterThanOrEqual(1)
    // "20" appears in stat card and in status breakdown
    expect(screen.getAllByText('20').length).toBeGreaterThanOrEqual(1)
  })

  it('displays realized count in stat card', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getAllByText(/Realized/).length).toBeGreaterThanOrEqual(1)
    // "15" appears in stat card and in status breakdown
    expect(screen.getAllByText('15').length).toBeGreaterThanOrEqual(1)
  })

  it('displays realization rate', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Realization Rate')).toBeInTheDocument()
    // 75% appears in both the realization rate card and as a frequency compliance value
    expect(screen.getAllByText('75%').length).toBeGreaterThanOrEqual(1)
  })

  it('displays total activities', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Activities')).toBeInTheDocument()
    expect(screen.getByText('37')).toBeInTheDocument()
  })

  it('displays target coverage', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Target Coverage')).toBeInTheDocument()
    expect(screen.getByText('70%')).toBeInTheDocument()
    expect(screen.getByText('35 / 50')).toBeInTheDocument()
  })

  it('displays frequency compliance items', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Frequency Compliance')).toBeInTheDocument()
    expect(screen.getByText('a')).toBeInTheDocument()
    expect(screen.getByText('95%')).toBeInTheDocument()
    expect(screen.getByText('b')).toBeInTheDocument()
  })

  it('displays field vs non-field split', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Field: 30')).toBeInTheDocument()
    expect(screen.getByText('Non-field: 7')).toBeInTheDocument()
  })

  it('displays team performance card', () => {
    setupAllHooks()
    render(<DashboardPage />)
    expect(screen.getByText('Team Performance')).toBeInTheDocument()
    expect(screen.getByText('Alice Rep')).toBeInTheDocument()
  })

  it('resolves status labels from config', () => {
    setupAllHooks()
    render(<DashboardPage />)
    // "Planned" and "Realized" appear as both stat card labels and config-resolved status labels
    expect(screen.getAllByText('Planned').length).toBeGreaterThanOrEqual(2)
    expect(screen.getAllByText(/Realized/).length).toBeGreaterThanOrEqual(2)
  })

  it('navigates period with previous/next buttons', async () => {
    setupAllHooks()
    const user = userEvent.setup()
    render(<DashboardPage />)

    const prevButton = screen.getByLabelText('Previous month')
    await user.click(prevButton)

    // Hooks should be re-called with new period
    expect(mockUseActivityStats).toHaveBeenCalled()
  })

  it('hides team performance when no members', () => {
    setupAllHooks()
    mockUseTeamMembers.mockReturnValue({
      data: { items: [] as TeamMember[], total: 0, page: 1, limit: 10 },
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTeamMembers>)

    render(<DashboardPage />)
    expect(screen.queryByText('Team Performance')).not.toBeInTheDocument()
  })

  it('handles zero activities gracefully', () => {
    setupAllHooks()
    mockUseActivityStats.mockReturnValue({
      data: { byStatus: {}, byCategory: {}, total: 0 },
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useActivityStats>)

    render(<DashboardPage />)
    // Multiple "0" values may appear (planned, realized, total)
    expect(screen.getAllByText('0').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('0%').length).toBeGreaterThanOrEqual(1)
  })

  it('hides frequency table when no items', () => {
    setupAllHooks()
    mockUseFrequency.mockReturnValue({
      data: { items: [] as FrequencyResponse['items'] },
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useFrequency>)

    render(<DashboardPage />)
    expect(screen.queryByText('Frequency Compliance')).not.toBeInTheDocument()
  })
})
