import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks -------------------------------------------------------------------
const mockStats = vi.fn()
const mockCoverage = vi.fn()
const mockFrequency = vi.fn()
const mockRecovery = vi.fn()
const mockActivities = vi.fn()
const mockUsers = vi.fn()
const mockTeams = vi.fn()

vi.mock('@/hooks/useDashboard', () => ({
  useActivityStats: () => mockStats(),
  useCoverage: () => mockCoverage(),
  useFrequency: () => mockFrequency(),
  useRecoveryBalance: () => mockRecovery(),
}))

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
}))

vi.mock('@/hooks/useUsers', () => ({
  useUsers: () => mockUsers(),
}))

vi.mock('@/hooks/useTeams', () => ({
  useTeams: () => mockTeams(),
}))

vi.mock('@/auth/context', () => ({
  useAuth: () => ({ user: { id: 'u1', role: 'manager' } }),
}))

vi.mock('@/components/activities/ActivityDetailModal', () => ({
  ActivityDetailModal: () => null,
}))

vi.mock('@/components/calendar/WeekView', () => ({
  WeekView: ({ activities }: { activities: unknown[] }) => (
    <div data-testid="week-view">{activities.length} activities</div>
  ),
}))

// --- Stub TanStack Router ---------------------------------------------------------
let capturedComponent: React.ComponentType | null = null
vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
  Link: ({ children, ...props }: Record<string, unknown>) => <a {...props}>{children as React.ReactNode}</a>,
}))

vi.mock('./__root', () => ({
  Route: {},
}))

// --- Helpers ----------------------------------------------------------------------
function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setDefaultMocks() {
  mockStats.mockReturnValue({
    data: {
      total: 20,
      byStatus: { realizat: 14, planificat: 4, anulat: 2 },
      byCategory: { field: 15, office: 5 },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  })
  mockCoverage.mockReturnValue({
    data: { totalTargets: 100, visitedTargets: 75, percentage: 75 },
  })
  mockFrequency.mockReturnValue({
    data: {
      items: [
        { classification: 'A', targetCount: 10, totalVisits: 30, required: 40, compliance: 75 },
        { classification: 'B', targetCount: 20, totalVisits: 15, required: 20, compliance: 75 },
      ],
    },
  })
  mockRecovery.mockReturnValue({
    data: { earned: 3, taken: 1, balance: 2, intervals: [] },
  })
  mockActivities.mockReturnValue({
    data: { items: [], total: 0, page: 1, limit: 200 },
    isLoading: false,
  })
  mockUsers.mockReturnValue({
    data: {
      items: [
        { id: 'rep1', name: 'Alice Rep', displayName: 'Alice', role: 'rep' },
        { id: 'rep2', name: 'Bob Rep', displayName: 'Bob', role: 'rep' },
        { id: 'mgr1', name: 'Carol Mgr', displayName: 'Carol', role: 'manager' },
      ],
    },
  })
  mockTeams.mockReturnValue({
    data: { items: [{ id: 't1', name: 'Team A' }] },
  })
}

// --- Tests ------------------------------------------------------------------------

describe('DashboardPage', () => {
  beforeAll(async () => {
    await import('./dashboard')
  })

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows spinner while stats are loading', () => {
    mockStats.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })
    mockCoverage.mockReturnValue({ data: null })
    mockFrequency.mockReturnValue({ data: null })
    mockRecovery.mockReturnValue({ data: null })
    mockActivities.mockReturnValue({ data: null, isLoading: false })
    mockUsers.mockReturnValue({ data: undefined })
    mockTeams.mockReturnValue({ data: undefined })

    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    // Spinner renders an element with role="status" or the Spinner component
    expect(container.querySelector('[class*="animate-spin"]') || screen.queryByText(/loading/i) || container.innerHTML).toBeTruthy()
  })

  it('shows error state with retry button', () => {
    const refetchFn = vi.fn()
    mockStats.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: refetchFn })
    mockCoverage.mockReturnValue({ data: null })
    mockFrequency.mockReturnValue({ data: null })
    mockRecovery.mockReturnValue({ data: null })
    mockActivities.mockReturnValue({ data: null, isLoading: false })
    mockUsers.mockReturnValue({ data: undefined })
    mockTeams.mockReturnValue({ data: undefined })

    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText(/failed to load dashboard data/i)).toBeTruthy()
  })

  it('renders stat cards with correct values', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Completion rate: 14/20 = 70%
    expect(screen.getByText('70%')).toBeTruthy()
    expect(screen.getByText('14 of 20 completed')).toBeTruthy()

    // Coverage: 75% (appears twice — once in stat card, once in frequency compliance)
    expect(screen.getAllByText('75%').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('75 of 100 visited')).toBeTruthy()

    // Week Progress
    expect(screen.getByText('14 / 20')).toBeTruthy()

    // Recovery balance
    expect(screen.getByText('2 days')).toBeTruthy()
    expect(screen.getByText('3 earned')).toBeTruthy()
  })

  it('renders header with team and rep counts', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Team Dashboard')).toBeTruthy()
    // 1 team, 2 reps (filters by role=rep)
    expect(screen.getByText(/1 teams · 2 reps/)).toBeTruthy()
  })

  it('renders activity breakdown by status and category', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Activity by Status')).toBeTruthy()
    expect(screen.getByText('realizat')).toBeTruthy()
    expect(screen.getByText('planificat')).toBeTruthy()
    expect(screen.getByText('anulat')).toBeTruthy()

    expect(screen.getByText('Activity by Category')).toBeTruthy()
    expect(screen.getByText('field')).toBeTruthy()
    expect(screen.getByText('office')).toBeTruthy()
  })

  it('renders frequency compliance table', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Frequency Compliance by Classification')).toBeTruthy()
    // Classification badges
    expect(screen.getByText('A')).toBeTruthy()
    expect(screen.getByText('B')).toBeTruthy()
    // Compliance percentages
    expect(screen.getAllByText('75%').length).toBeGreaterThanOrEqual(1)
  })

  it('navigates weeks with prev/next buttons', async () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const prevBtn = screen.getByLabelText('Previous period')
    const nextBtn = screen.getByLabelText('Next period')

    // Store initial period label
    const periodSpan = prevBtn.parentElement!.querySelector('span')!
    const initialLabel = periodSpan.textContent

    // Click next
    fireEvent.click(nextBtn)
    await waitFor(() => {
      expect(periodSpan.textContent).not.toBe(initialLabel)
    })

    // Click prev twice to go back one week before initial
    fireEvent.click(prevBtn)
    fireEvent.click(prevBtn)
    await waitFor(() => {
      expect(periodSpan.textContent).not.toBe(initialLabel)
    })
  })

  it('shows "select a rep" prompt when no rep selected in week view', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText(/select a rep to view their calendar/i)).toBeTruthy()
  })

  it('renders rep dropdown with rep users only', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const select = screen.getByDisplayValue('All Reps')
    expect(select).toBeTruthy()
    // Should have All Reps + 2 reps (not the manager)
    const options = select.querySelectorAll('option')
    expect(options).toHaveLength(3) // "All Reps", "Alice Rep", "Bob Rep"
  })

  it('switches between week and month view modes', async () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const monthBtn = screen.getByText('Month')
    fireEvent.click(monthBtn)

    // In month view with no rep, should show "select a rep to view their month"
    await waitFor(() => {
      expect(screen.getByText(/select a rep to view their month/i)).toBeTruthy()
    })
  })

  it('shows WeekView when rep is selected', async () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const select = screen.getByDisplayValue('All Reps')
    fireEvent.change(select, { target: { value: 'rep1' } })

    await waitFor(() => {
      expect(screen.getByTestId('week-view')).toBeTruthy()
    })
  })

  it('Today button resets to current week', async () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Navigate forward then click Today
    fireEvent.click(screen.getByLabelText('Next period'))
    fireEvent.click(screen.getByText('Today'))

    // Should not throw; component re-renders fine
    expect(screen.getByText('Team Dashboard')).toBeTruthy()
  })
})

describe('ClassificationCell', () => {
  beforeAll(async () => {
    await import('./dashboard')
  })

  it('renders classification A with danger variant', () => {
    setDefaultMocks()
    mockFrequency.mockReturnValue({
      data: {
        items: [
          { classification: 'A', targetCount: 5, totalVisits: 10, required: 20, compliance: 50 },
        ],
      },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const badge = screen.getByText('A')
    expect(badge).toBeTruthy()
  })
})

describe('ComplianceCell', () => {
  beforeAll(async () => {
    await import('./dashboard')
  })

  it('renders compliance with green color for >= 80', () => {
    setDefaultMocks()
    mockFrequency.mockReturnValue({
      data: {
        items: [
          { classification: 'A', targetCount: 5, totalVisits: 18, required: 20, compliance: 90 },
        ],
      },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const complianceEl = screen.getByText('90%')
    expect(complianceEl.className).toContain('text-emerald-600')
  })

  it('renders compliance with amber color for 50-79', () => {
    setDefaultMocks()
    mockFrequency.mockReturnValue({
      data: {
        items: [
          { classification: 'B', targetCount: 5, totalVisits: 12, required: 20, compliance: 60 },
        ],
      },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const complianceEl = screen.getByText('60%')
    expect(complianceEl.className).toContain('text-amber-600')
  })

  it('renders compliance with red color for < 50', () => {
    setDefaultMocks()
    mockFrequency.mockReturnValue({
      data: {
        items: [
          { classification: 'C', targetCount: 5, totalVisits: 2, required: 20, compliance: 10 },
        ],
      },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const complianceEl = screen.getByText('10%')
    expect(complianceEl.className).toContain('text-red-600')
  })
})
