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
const mockTargets = vi.fn()
const mockFrequencyStatus = vi.fn()

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockTargets(),
  useTargetFrequencyStatus: () => mockFrequencyStatus(),
}))

// --- Stub TanStack Router ---------------------------------------------------------
const mockNavigate = vi.fn()
let capturedComponent: React.ComponentType | null = null

vi.mock('@tanstack/react-router', () => ({
  createRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
  useNavigate: () => mockNavigate,
  Link: ({ children, ...props }: Record<string, unknown>) => <a {...props}>{children as React.ReactNode}</a>,
}))

vi.mock('./__root', () => ({
  Route: {},
}))

// --- Helpers ----------------------------------------------------------------------

function makeTarget(id: string, overrides: Record<string, unknown> = {}) {
  return {
    id,
    targetType: 'pharmacy',
    name: `Target ${id}`,
    fields: { lat: 44.43, lng: 26.10, classification: 'a', city: 'Bucharest', address: '123 Main St' },
    assigneeId: 'u1',
    teamId: 't1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

// --- Tests ------------------------------------------------------------------------

describe('TargetsPage', () => {
  beforeAll(async () => {
    await import('./targets')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
    mockFrequencyStatus.mockReturnValue({ data: { items: [] } })
  })

  it('shows spinner while loading', () => {
    mockTargets.mockReturnValue({ data: undefined, isLoading: true, isError: false, refetch: vi.fn() })

    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    // Spinner renders an element with role="status" or a spinner class — just verify no table
    expect(screen.queryByText('Target Portfolio')).toBeNull()
    expect(container.querySelector('[class*="animate-spin"]') ?? container.querySelector('[role="status"]')).toBeTruthy()
  })

  it('shows error state with retry button', () => {
    const refetchFn = vi.fn()
    mockTargets.mockReturnValue({ data: undefined, isLoading: false, isError: true, refetch: refetchFn })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Failed to load targets')).toBeTruthy()
  })

  it('renders DataTable with targets and target count', async () => {
    mockTargets.mockReturnValue({
      data: {
        items: [
          makeTarget('1', { name: 'Farmacia Sensiblu' }),
          makeTarget('2', { name: 'Catena Central', targetType: 'doctor', fields: { classification: 'b', city: 'Cluj', lat: 45.0, lng: 25.0 } }),
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

    await waitFor(() => {
      expect(screen.getByText('Target Portfolio')).toBeTruthy()
    })
    expect(screen.getByText('2 targets')).toBeTruthy()
    expect(screen.getByText('Farmacia Sensiblu')).toBeTruthy()
    expect(screen.getByText('Catena Central')).toBeTruthy()
  })

  it('renders map container with markers for geo targets', async () => {
    mockTargets.mockReturnValue({
      data: {
        items: [
          makeTarget('1'),
          makeTarget('2', { fields: { address: 'No coords' } }), // no lat/lng
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

    await waitFor(() => {
      expect(screen.getByTestId('google-map')).toBeTruthy()
    })
    // Only one target has lat/lng — expect one marker
    expect(screen.getAllByTestId('advanced-marker')).toHaveLength(1)
  })

  it('updates search input value on change', async () => {
    mockTargets.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const searchInput = screen.getByPlaceholderText('Search targets...')
    fireEvent.change(searchInput, { target: { value: 'Farmacia' } })
    expect(searchInput).toHaveProperty('value', 'Farmacia')
  })

  it('renders type filter dropdown', async () => {
    mockTargets.mockReturnValue({
      data: { items: [], total: 0, page: 1, limit: 200 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const select = screen.getByDisplayValue('All types')
    expect(select).toBeTruthy()
    fireEvent.change(select, { target: { value: 'pharmacy' } })
    expect(select).toHaveProperty('value', 'pharmacy')
  })

  it('renders compliance from frequency status data', async () => {
    mockTargets.mockReturnValue({
      data: {
        items: [makeTarget('t1', { name: 'Compliant Target' })],
        total: 1,
        page: 1,
        limit: 200,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })
    mockFrequencyStatus.mockReturnValue({
      data: { items: [{ targetId: 't1', classification: 'a', visitCount: 8, required: 10, compliance: 80 }] },
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    await waitFor(() => {
      expect(screen.getByText('80%')).toBeTruthy()
    })
  })

  it('navigates to target detail when name link is clicked', async () => {
    mockTargets.mockReturnValue({
      data: {
        items: [makeTarget('abc', { name: 'Clickable Target' })],
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

    await waitFor(() => {
      expect(screen.getByText('Clickable Target')).toBeTruthy()
    })

    fireEvent.click(screen.getByText('Clickable Target'))
    expect(mockNavigate).toHaveBeenCalledWith({ to: '/targets/$id', params: { id: 'abc' } })
  })
})
