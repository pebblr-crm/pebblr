import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks -------------------------------------------------------------------
const mockUseUsers = vi.fn()
const mockUseTeams = vi.fn()
const mockUseTerritories = vi.fn()
const mockUseConfig = vi.fn()

vi.mock('@/hooks/useUsers', () => ({
  useUsers: () => mockUseUsers(),
}))

vi.mock('@/hooks/useTeams', () => ({
  useTeams: () => mockUseTeams(),
}))

vi.mock('@/hooks/useTerritories', () => ({
  useTerritories: () => mockUseTerritories(),
}))

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => mockUseConfig(),
}))

// --- Stub TanStack Router ---------------------------------------------------------
let capturedComponent: React.ComponentType | null = null
vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
}))

vi.mock('./__root', () => ({
  Route: {},
}))

// --- Helpers ----------------------------------------------------------------------

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function makeConfig() {
  return {
    tenant: { name: 'Test', locale: 'en' },
    accounts: { types: [] },
    activities: {
      statuses: [
        { key: 'planificat', label: 'Planned', initial: true, submittable: false },
        { key: 'realizat', label: 'Completed', submittable: true },
      ],
      status_transitions: { planificat: ['realizat'] },
      durations: [],
      types: [
        { key: 'visit', label: 'Visit', category: 'field', fields: [], submit_required: [], blocks_field_activities: false },
        { key: 'office', label: 'Office', category: 'non_field', fields: [], submit_required: [], blocks_field_activities: true },
      ],
      routing_options: [],
    },
    options: {},
    rules: {
      frequency: { A: 4, B: 2 },
      max_activities_per_day: 8,
      visit_cadence_days: 14,
      default_visit_duration_minutes: {},
      visit_duration_step_minutes: 15,
    },
  }
}

// --- Tests ------------------------------------------------------------------------

describe('ConsolePage', () => {
  beforeAll(async () => {
    await import('./console')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseConfig.mockReturnValue({ data: makeConfig() })
    mockUseTeams.mockReturnValue({ data: { items: [], total: 0 } })
    mockUseTerritories.mockReturnValue({ data: { items: [], total: 0 } })
  })

  it('shows loading spinner when users are loading', () => {
    mockUseUsers.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('shows error state when users fail to load', () => {
    mockUseUsers.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Failed to load configuration')).toBeInTheDocument()
  })

  it('renders section tabs with correct labels', () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Desktop nav has the labels; mobile nav also has them
    expect(screen.getAllByText('Users & Roles').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Teams').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Territories').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Business Rules').length).toBeGreaterThanOrEqual(1)
  })

  it('renders users data table with role badges', () => {
    mockUseUsers.mockReturnValue({
      data: {
        items: [
          { id: 'u1', name: 'Alice', displayName: 'Alice Smith', email: 'alice@test.com', role: 'admin' },
          { id: 'u2', name: 'Bob', displayName: 'Bob Jones', email: 'bob@test.com', role: 'rep' },
        ],
        total: 2,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // User names in table
    expect(screen.getByText('Alice Smith')).toBeInTheDocument()
    expect(screen.getByText('Bob Jones')).toBeInTheDocument()
    // Emails
    expect(screen.getByText('alice@test.com')).toBeInTheDocument()
    expect(screen.getByText('bob@test.com')).toBeInTheDocument()
    // Role badges
    expect(screen.getByText('admin')).toBeInTheDocument()
    expect(screen.getByText('rep')).toBeInTheDocument()
  })

  it('UserNameCell renders avatar initial and name', () => {
    mockUseUsers.mockReturnValue({
      data: {
        items: [
          { id: 'u1', name: 'Charlie', displayName: 'Charlie Brown', email: 'c@test.com', role: 'manager' },
        ],
        total: 1,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('C')).toBeInTheDocument()
    expect(screen.getByText('Charlie Brown')).toBeInTheDocument()
  })

  it('switches to Teams section on tab click', async () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    mockUseTeams.mockReturnValue({
      data: {
        items: [
          { id: 'team1', name: 'North Team', managerId: 'u1' },
        ],
        total: 1,
      },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Click the desktop "Teams" button
    const teamButtons = screen.getAllByText('Teams')
    fireEvent.click(teamButtons[0])

    await waitFor(() => {
      expect(screen.getByText('North Team')).toBeInTheDocument()
      expect(screen.getByText('Manager: u1')).toBeInTheDocument()
    })
  })

  it('switches to Territories section on tab click', async () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    mockUseTerritories.mockReturnValue({
      data: {
        items: [
          { id: 'terr1', name: 'Bucharest Zone', teamId: 'team1', region: 'South', createdAt: '2026-01-01', updatedAt: '2026-01-01' },
        ],
        total: 1,
      },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const terrButtons = screen.getAllByText('Territories')
    fireEvent.click(terrButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Bucharest Zone')).toBeInTheDocument()
      expect(screen.getByText('South')).toBeInTheDocument()
      expect(screen.getByText('Team: team1')).toBeInTheDocument()
    })
  })

  it('switches to Business Rules section and shows config', async () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const rulesButtons = screen.getAllByText('Business Rules')
    fireEvent.click(rulesButtons[0])

    await waitFor(() => {
      expect(screen.getByText('Visit Frequency Requirements')).toBeInTheDocument()
      expect(screen.getByText('Class A')).toBeInTheDocument()
      expect(screen.getByText('4 visits/period')).toBeInTheDocument()
      expect(screen.getByText('Activity Rules')).toBeInTheDocument()
      expect(screen.getByText('8')).toBeInTheDocument() // max_activities_per_day
      expect(screen.getByText('Activity Types')).toBeInTheDocument()
      expect(screen.getByText('Status Workflow')).toBeInTheDocument()
    })
  })

  it('shows empty state for teams when none exist', async () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const teamButtons = screen.getAllByText('Teams')
    fireEvent.click(teamButtons[0])

    await waitFor(() => {
      expect(screen.getByText('No teams configured.')).toBeInTheDocument()
    })
  })

  it('shows empty state for territories when none exist', async () => {
    mockUseUsers.mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const terrButtons = screen.getAllByText('Territories')
    fireEvent.click(terrButtons[0])

    await waitFor(() => {
      expect(screen.getByText('No territories configured.')).toBeInTheDocument()
    })
  })
})
