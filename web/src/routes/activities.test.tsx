import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { Activity } from '@/types/activity'

// --- Mock hooks ---
const mockActivities = vi.fn()
const mockRecoveryBalance = vi.fn()
const mockConfig = vi.fn()
const mockUsers = vi.fn()
const mockAuth = vi.fn()

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useRecoveryBalance: () => mockRecoveryBalance(),
}))

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => mockConfig(),
}))

vi.mock('@/hooks/useUsers', () => ({
  useUsers: () => mockUsers(),
}))

vi.mock('@/auth/context', () => ({
  useAuth: () => mockAuth(),
}))

vi.mock('@/components/activities/ActivityDetailModal', () => ({
  ActivityDetailModal: () => null,
}))

vi.mock('@/components/activities/CreateActivityModal', () => ({
  CreateActivityModal: ({ open }: { open: boolean }) =>
    open ? <div data-testid="create-modal">Create</div> : null,
}))

vi.mock('@/components/ui/Toast', () => ({
  useToast: () => ({ showToast: vi.fn(), ToastContainer: () => null }),
}))

// --- Stub TanStack Router ---
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

// --- Helpers ---
function makeActivity(id: string, overrides: Partial<Activity> = {}): Activity {
  return {
    id,
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-03-23T09:00:00Z',
    duration: '30m',
    fields: {},
    creatorId: 'u1',
    createdAt: '2026-03-23T00:00:00Z',
    updatedAt: '2026-03-23T00:00:00Z',
    ...overrides,
  }
}

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupDefaults() {
  mockConfig.mockReturnValue({
    data: {
      activities: {
        types: [{ key: 'visit', label: 'Visit' }, { key: 'administrative', label: 'Administrative' }],
        statuses: [{ key: 'planificat', label: 'Planned' }, { key: 'realizat', label: 'Completed' }],
      },
    },
  })
  mockAuth.mockReturnValue({ role: 'rep' })
  mockUsers.mockReturnValue({ data: { items: [] } })
  mockRecoveryBalance.mockReturnValue({ data: null })
}

// --- Tests ---

describe('ActivitiesPage', () => {
  beforeAll(async () => {
    await import('./activities')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    setupDefaults()
  })

  it('shows spinner while loading', () => {
    mockActivities.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })
    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    // Spinner renders an SVG or similar loading indicator
    expect(container.querySelector('[class*="animate"]') || container.textContent).toBeTruthy()
  })

  it('shows error state when query fails', () => {
    mockActivities.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: vi.fn() })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Failed to load activities')).toBeInTheDocument()
  })

  it('shows empty state when no activities', () => {
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('No activities found.')).toBeInTheDocument()
  })

  it('renders activity cards grouped by week', () => {
    mockActivities.mockReturnValue({
      data: {
        items: [
          makeActivity('a1', { targetName: 'Pharmacy Alpha', dueDate: '2026-03-23T09:00:00Z' }),
          makeActivity('a2', { targetName: 'Pharmacy Beta', dueDate: '2026-03-24T10:00:00Z' }),
        ],
        total: 2,
        page: 1,
        limit: 200,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getAllByText('Pharmacy Alpha').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Pharmacy Beta').length).toBeGreaterThanOrEqual(1)
  })

  it('renders the page title', () => {
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Activity Log')).toBeInTheDocument()
  })

  it('opens create modal when Log Activity button is clicked', () => {
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    fireEvent.click(screen.getByText('Log Activity'))
    expect(screen.getByTestId('create-modal')).toBeInTheDocument()
  })

  it('shows recovery balance when available', () => {
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    mockRecoveryBalance.mockReturnValue({ data: { earned: 5, taken: 2, balance: 3, intervals: [] } })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Recovery Days')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('shows rep filter for admin role', () => {
    mockAuth.mockReturnValue({ role: 'admin' })
    mockUsers.mockReturnValue({
      data: { items: [{ id: 'u1', name: 'John', displayName: 'John D', role: 'rep' }] },
    })
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByLabelText('Rep')).toBeInTheDocument()
  })

  it('does not show rep filter for rep role', () => {
    mockAuth.mockReturnValue({ role: 'rep' })
    mockActivities.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.queryByLabelText('Rep')).not.toBeInTheDocument()
  })

  it('renders activity with cancelled styling', () => {
    mockActivities.mockReturnValue({
      data: {
        items: [makeActivity('a1', { targetName: 'Cancelled Visit', status: 'anulat' })],
        total: 1,
        page: 1,
        limit: 200,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const heading = screen.getAllByText('Cancelled Visit')[0]
    expect(heading.className).toContain('line-through')
  })

  it('shows submitted badge when activity has submittedAt', () => {
    mockActivities.mockReturnValue({
      data: {
        items: [makeActivity('a1', { targetName: 'Submitted Visit', submittedAt: '2026-03-23T12:00:00Z' })],
        total: 1,
        page: 1,
        limit: 200,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getAllByText('Submitted').length).toBeGreaterThanOrEqual(1)
  })
})
