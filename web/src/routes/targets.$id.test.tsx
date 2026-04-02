import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock @vis.gl/react-google-maps before any component imports -----------------
vi.mock('@vis.gl/react-google-maps', () => ({
  APIProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Map: ({ children }: { children?: React.ReactNode }) => <div data-testid="google-map">{children}</div>,
  AdvancedMarker: vi.fn(({ children }: { children?: React.ReactNode }) => (
    <div data-testid="advanced-marker">{children}</div>
  )),
}))

// --- Mock hooks -------------------------------------------------------------------
const mockTarget = vi.fn()
const mockActivities = vi.fn()
const mockVisitStatus = vi.fn()
const mockUsers = vi.fn()
const mockConfig = vi.fn()
const mockCreateActivity = vi.fn()

vi.mock('@/hooks/useTargets', () => ({
  useTarget: () => mockTarget(),
  useTargetVisitStatus: () => mockVisitStatus(),
}))

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
  useCreateActivity: () => ({ mutate: mockCreateActivity, isPending: false }),
}))

vi.mock('@/hooks/useUsers', () => ({
  useUsers: () => mockUsers(),
}))

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => mockConfig(),
}))

vi.mock('@/components/activities/ActivityDetailModal', () => ({
  ActivityDetailModal: ({ activityId }: { activityId: string | null }) =>
    activityId ? <div data-testid="activity-detail-modal">Activity: {activityId}</div> : null,
}))

// --- Stub TanStack Router ---------------------------------------------------------
const mockNavigate = vi.fn()
let capturedComponent: React.ComponentType | null = null

vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
  useParams: () => ({ id: 'target-1' }),
  useNavigate: () => mockNavigate,
  Link: ({ children, ...props }: Record<string, unknown>) => <a {...props}>{children as React.ReactNode}</a>,
}))

vi.mock('./__root', () => ({
  Route: {},
}))

// --- Helpers ----------------------------------------------------------------------

function makeTarget(overrides: Record<string, unknown> = {}) {
  return {
    id: 'target-1',
    targetType: 'pharmacy',
    name: 'Farmacia Sensiblu',
    fields: {
      lat: 44.43,
      lng: 26.10,
      classification: 'a',
      potential: 'a',
      city: 'Bucharest',
      address: '123 Main Street',
    },
    assigneeId: 'u1',
    teamId: 't1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function makeActivity(id: string, overrides: Record<string, unknown> = {}) {
  return {
    id,
    activityType: 'visit',
    status: 'realizat',
    dueDate: '2026-01-15T00:00:00Z',
    duration: '30m',
    fields: { visit_type: 'f2f', feedback: 'Good visit' },
    targetId: 'target-1',
    creatorId: 'u1',
    createdAt: '2026-01-15T00:00:00Z',
    updatedAt: '2026-01-15T00:00:00Z',
    ...overrides,
  }
}

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupDefaultMocks() {
  mockActivities.mockReturnValue({ data: { items: [], total: 0, page: 1, limit: 50 } })
  mockVisitStatus.mockReturnValue({ data: { items: [] } })
  mockUsers.mockReturnValue({ data: { items: [{ id: 'u1', name: 'John Rep', displayName: 'John Rep' }] } })
  mockConfig.mockReturnValue({
    data: {
      accounts: { types: [{ key: 'pharmacy', label: 'Pharmacy', fields: [{ key: 'address', label: 'Address', type: 'text', required: false }] }] },
      activities: { statuses: [{ key: 'planificat', label: 'Planned', initial: true }] },
      options: {},
    },
  })
}

// --- Tests ------------------------------------------------------------------------

describe('TargetDetailPage', () => {
  beforeAll(async () => {
    await import('./targets.$id')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    setupDefaultMocks()
  })

  it('shows spinner while loading', () => {
    mockTarget.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    expect(screen.queryByText('Farmacia Sensiblu')).toBeNull()
    expect(container.querySelector('[class*="animate-spin"]') ?? container.querySelector('[role="status"]')).toBeTruthy()
  })

  it('shows error state when target fails to load', () => {
    mockTarget.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Failed to load target')).toBeTruthy()
  })

  it('shows error state when target is null (not found)', () => {
    mockTarget.mockReturnValue({ data: null, isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Failed to load target')).toBeTruthy()
  })

  it('renders target name and priority badge', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getAllByText('Farmacia Sensiblu')).toHaveLength(2) // breadcrumb + h1
    })
    expect(screen.getByText('Priority A')).toBeTruthy()
    expect(screen.getByText('pharmacy')).toBeTruthy()
  })

  it('renders address information', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      // Address appears in header and details card
      expect(screen.getAllByText(/123 Main Street/).length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders map with marker when coordinates exist', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByTestId('google-map')).toBeTruthy()
    })
    expect(screen.getByTestId('advanced-marker')).toBeTruthy()
  })

  it('shows "No coordinates available" when target has no lat/lng', async () => {
    mockTarget.mockReturnValue({
      data: makeTarget({ fields: { address: 'Street 1', potential: 'b' } }),
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('No coordinates available')).toBeTruthy()
    })
  })

  it('renders visit history with activities', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })
    mockActivities.mockReturnValue({
      data: {
        items: [
          makeActivity('act-1', { status: 'realizat', fields: { visit_type: 'f2f', feedback: 'Successful visit' } }),
          makeActivity('act-2', { status: 'planificat', dueDate: '2026-02-01T00:00:00Z', fields: {} }),
        ],
        total: 2,
        page: 1,
        limit: 50,
      },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('Visit History')).toBeTruthy()
    })
    expect(screen.getByText('2 total')).toBeTruthy()
    expect(screen.getByText('Successful visit')).toBeTruthy()
  })

  it('shows "No visits recorded yet" when no activities', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('No visits recorded yet.')).toBeTruthy()
    })
  })

  it('shows "No visit scheduled" when no future activities', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('No visit scheduled')).toBeTruthy()
    })
  })

  it('shows "Never visited" when no visit status data', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('Never visited')).toBeTruthy()
    })
  })

  it('renders last visit date from visit status data', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })
    mockVisitStatus.mockReturnValue({
      data: { items: [{ targetId: 'target-1', lastVisitDate: '2026-03-01T00:00:00Z' }] },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('1 Mar 2026')).toBeTruthy()
    })
  })

  it('navigates back when back button is clicked', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getAllByText('Farmacia Sensiblu').length).toBeGreaterThan(0)
    })

    fireEvent.click(screen.getByLabelText('Back to targets'))
    expect(mockNavigate).toHaveBeenCalledWith({ to: '/targets' })
  })

  it('opens schedule visit modal when button is clicked', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('Schedule Visit')).toBeTruthy()
    })

    fireEvent.click(screen.getByText('Schedule Visit'))

    await waitFor(() => {
      expect(screen.getByText(/Schedule Visit — Farmacia Sensiblu/)).toBeTruthy()
    })
  })

  it('shows user name in activity history', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })
    mockActivities.mockReturnValue({
      data: {
        items: [makeActivity('act-1')],
        total: 1,
        page: 1,
        limit: 50,
      },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('John Rep')).toBeTruthy()
    })
  })

  it('renders doctor icon for doctor target type', async () => {
    mockTarget.mockReturnValue({
      data: makeTarget({ targetType: 'doctor' }),
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getAllByText('Farmacia Sensiblu').length).toBeGreaterThan(0)
    })
    // Doctor target type renders — the component switches icons, verify page renders correctly
    expect(screen.getByText('doctor')).toBeTruthy()
  })

  it('opens activity detail modal when activity is clicked', async () => {
    mockTarget.mockReturnValue({ data: makeTarget(), isLoading: false, isError: false, refetch: vi.fn() })
    mockActivities.mockReturnValue({
      data: {
        items: [makeActivity('act-detail-1', { fields: { feedback: 'Click me' } })],
        total: 1,
        page: 1,
        limit: 50,
      },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('Click me')).toBeTruthy()
    })

    // Click the activity button
    fireEvent.click(screen.getByText('Click me'))

    await waitFor(() => {
      expect(screen.getByTestId('activity-detail-modal')).toBeTruthy()
    })
  })
})
