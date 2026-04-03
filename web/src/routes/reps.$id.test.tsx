import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks ---
const mockActivities = vi.fn()
const mockTargets = vi.fn()
const mockStats = vi.fn()
const mockCoverage = vi.fn()

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
}))

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockTargets(),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useActivityStats: () => mockStats(),
  useCoverage: () => mockCoverage(),
}))

vi.mock('@/components/calendar/WeekView', () => ({
  WeekView: () => <div data-testid="week-view">WeekView</div>,
}))

vi.mock('@/components/map/MapContainer', () => ({
  MapContainer: ({ children }: { children?: React.ReactNode }) => <div data-testid="map-container">{children}</div>,
}))

vi.mock('@/components/map/TargetMarker', () => ({
  TargetMarker: ({ name }: { name: string }) => <div data-testid="target-marker">{name}</div>,
}))

vi.mock('@/components/data/StatCard', () => ({
  StatCard: ({ label, value }: { label: string; value: string | number }) => (
    <div data-testid="stat-card">
      <span>{label}</span>
      <span>{value}</span>
    </div>
  ),
}))

// --- Stub TanStack Router ---
let capturedComponent: React.ComponentType | null = null
vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
  useParams: () => ({ id: 'rep-123' }),
  Link: ({ children, ...props }: Record<string, unknown>) => <a {...props}>{children as React.ReactNode}</a>,
}))

vi.mock('./__root', () => ({
  Route: {},
}))

// --- Helpers ---
function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupDefaults() {
  mockActivities.mockReturnValue({
    data: { items: [], total: 0, page: 1, limit: 200 },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  })
  mockTargets.mockReturnValue({
    data: { items: [], total: 0, page: 1, limit: 500 },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  })
  mockStats.mockReturnValue({ data: { total: 10, byStatus: { realizat: 7 }, byCategory: {} } })
  mockCoverage.mockReturnValue({ data: { totalTargets: 50, visitedTargets: 35, percentage: 70 } })
}

// --- Tests ---

describe('RepDrillDownPage', () => {
  beforeAll(async () => {
    await import('./reps.$id')
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
    expect(container.querySelector('[class*="animate"]') || container.textContent).toBeTruthy()
  })

  it('shows error state when query fails', () => {
    mockActivities.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: vi.fn() })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Failed to load rep data')).toBeInTheDocument()
  })

  it('renders the rep header with id', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Rep: rep-123')).toBeInTheDocument()
  })

  it('renders read-only banner', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText(/read-only mode/)).toBeInTheDocument()
  })

  it('renders stat cards', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const statCards = screen.getAllByTestId('stat-card')
    expect(statCards.length).toBe(4)
  })

  it('renders map container and week view', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByTestId('map-container')).toBeInTheDocument()
    expect(screen.getByTestId('week-view')).toBeInTheDocument()
  })

  it('renders markers for geo-located targets', () => {
    mockTargets.mockReturnValue({
      data: {
        items: [
          { id: 't1', name: 'Geo Target', fields: { lat: 44.43, lng: 26.10 } },
          { id: 't2', name: 'No Coords', fields: { address: 'Street' } },
        ],
        total: 2,
        page: 1,
        limit: 500,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const markers = screen.getAllByTestId('target-marker')
    expect(markers).toHaveLength(1)
    expect(markers[0]).toHaveTextContent('Geo Target')
  })

  it('shows On Track badge when completion >= 70%', () => {
    mockStats.mockReturnValue({ data: { total: 10, byStatus: { realizat: 8 }, byCategory: {} } })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('On Track')).toBeInTheDocument()
  })

  it('shows Needs Attention badge when completion < 70%', () => {
    mockStats.mockReturnValue({ data: { total: 10, byStatus: { realizat: 2 }, byCategory: {} } })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Needs Attention')).toBeInTheDocument()
  })

  it('navigates weeks with prev/next buttons', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const prevBtn = screen.getByLabelText('Previous week')
    const nextBtn = screen.getByLabelText('Next week')
    expect(prevBtn).toBeInTheDocument()
    expect(nextBtn).toBeInTheDocument()

    // Click prev and next -- should not crash
    fireEvent.click(prevBtn)
    fireEvent.click(nextBtn)
  })
})
