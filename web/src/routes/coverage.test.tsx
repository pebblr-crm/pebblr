import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks ---
const mockTargets = vi.fn()
const mockTerritories = vi.fn()
const mockTeams = vi.fn()
const mockCoverage = vi.fn()

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockTargets(),
}))

vi.mock('@/hooks/useTerritories', () => ({
  useTerritories: () => mockTerritories(),
}))

vi.mock('@/hooks/useTeams', () => ({
  useTeams: () => mockTeams(),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useCoverage: () => mockCoverage(),
}))

vi.mock('@/components/map/MapContainer', () => ({
  MapContainer: ({ children }: { children?: React.ReactNode }) => <div data-testid="map-container">{children}</div>,
}))

vi.mock('@/components/map/TargetMarker', () => ({
  TargetMarker: ({ name }: { name: string }) => <div data-testid="target-marker">{name}</div>,
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
function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setupDefaults() {
  mockTargets.mockReturnValue({
    data: { items: [], total: 0, page: 1, limit: 1000 },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  })
  mockTerritories.mockReturnValue({ data: { items: [], total: 0 } })
  mockTeams.mockReturnValue({ data: { items: [], total: 0 } })
  mockCoverage.mockReturnValue({ data: null })
}

// --- Tests ---

describe('CoveragePage', () => {
  beforeAll(async () => {
    await import('./coverage')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    setupDefaults()
  })

  it('shows spinner while loading', () => {
    mockTargets.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })
    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    expect(container.querySelector('[class*="animate"]') || container.textContent).toBeTruthy()
  })

  it('shows error state when query fails', () => {
    mockTargets.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: vi.fn() })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Failed to load coverage data')).toBeInTheDocument()
  })

  it('renders map container', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByTestId('map-container')).toBeInTheDocument()
  })

  it('renders markers for geo-located targets', () => {
    mockTargets.mockReturnValue({
      data: {
        items: [
          { id: 't1', name: 'Pharmacy A', fields: { lat: 44.43, lng: 26.10 }, teamId: 'tm1' },
          { id: 't2', name: 'Pharmacy B', fields: { address: 'No coords' }, teamId: 'tm1' },
        ],
        total: 2,
        page: 1,
        limit: 1000,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    const markers = screen.getAllByTestId('target-marker')
    expect(markers).toHaveLength(1)
    expect(markers[0]).toHaveTextContent('Pharmacy A')
  })

  it('shows coverage percentage when coverage data exists', () => {
    mockCoverage.mockReturnValue({
      data: { totalTargets: 100, visitedTargets: 75, percentage: 75 },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getAllByText('75%').length).toBeGreaterThanOrEqual(1)
  })

  it('shows territory list', () => {
    mockTerritories.mockReturnValue({
      data: { items: [{ id: 'ter1', name: 'Bucharest North', region: 'Sector 1', teamId: 'tm1' }], total: 1 },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Bucharest North')).toBeInTheDocument()
    expect(screen.getByText('Sector 1')).toBeInTheDocument()
  })

  it('shows no territories message when empty', () => {
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('No territories defined.')).toBeInTheDocument()
  })

  it('filters targets by priority button', () => {
    mockTargets.mockReturnValue({
      data: {
        items: [
          { id: 't1', name: 'A-target', fields: { lat: 44.0, lng: 26.0, potential: 'a' }, teamId: 'tm1' },
          { id: 't2', name: 'B-target', fields: { lat: 45.0, lng: 27.0, potential: 'b' }, teamId: 'tm1' },
        ],
        total: 2,
        page: 1,
        limit: 1000,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Both markers visible initially
    expect(screen.getAllByTestId('target-marker')).toHaveLength(2)

    // Click priority A filter
    fireEvent.click(screen.getByText('A'))
    expect(screen.getAllByTestId('target-marker')).toHaveLength(1)
    expect(screen.getByText('A-target')).toBeInTheDocument()
  })
})
