import { render, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock maplibre-gl before any component imports --------------------------------
const mockAddTo = vi.fn()
const mockMarkerRemove = vi.fn()
const MockMarker = vi.fn().mockImplementation(function () {
  const instance = {
    setLngLat: vi.fn().mockReturnValue(null as unknown),
    addTo: mockAddTo.mockReturnValue(null as unknown),
    remove: mockMarkerRemove,
  }
  // Make chaining work: setLngLat returns instance, addTo returns instance
  instance.setLngLat.mockReturnValue(instance)
  instance.addTo.mockReturnValue(instance)
  return instance
})

const mockMapRemove = vi.fn()
const mockMapOn = vi.fn()
const mockAddControl = vi.fn()
const MockMap = vi.fn().mockImplementation(function (this: Record<string, unknown>) {
  this.remove = mockMapRemove
  this.on = mockMapOn
  this.addControl = mockAddControl
  // Immediately fire 'load' so children are rendered
  setTimeout(() => {
    const loadCb = mockMapOn.mock.calls.find((c: unknown[]) => c[0] === 'load')
    if (loadCb) (loadCb[1] as () => void)()
  }, 0)
})
const MockNavigationControl = vi.fn()

vi.mock('maplibre-gl', () => ({
  default: {
    Map: MockMap,
    Marker: MockMarker,
    NavigationControl: MockNavigationControl,
  },
  Map: MockMap,
  Marker: MockMarker,
  NavigationControl: MockNavigationControl,
}))

// --- Mock hooks -------------------------------------------------------------------
const mockTargets = vi.fn()
const mockActivities = vi.fn()
const mockStats = vi.fn()
const mockCoverage = vi.fn()
const mockCloneWeek = vi.fn()

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockTargets(),
  useTargetVisitStatus: () => ({ data: { items: [] } }),
}))

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
  useCloneWeek: () => ({ mutate: mockCloneWeek, isPending: false }),
  useBatchCreateActivities: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  usePatchActivity: () => ({ mutate: vi.fn(), isPending: false }),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useActivityStats: () => mockStats(),
  useCoverage: () => mockCoverage(),
}))

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => ({ data: null }),
}))

vi.mock('@/components/activities/ActivityDetailModal', () => ({
  ActivityDetailModal: () => null,
}))

vi.mock('@/components/ui/Toast', () => ({
  useToast: () => ({ showToast: vi.fn(), ToastContainer: () => null }),
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

function makeTarget(id: string, fields: Record<string, unknown>) {
  return {
    id,
    targetType: 'pharmacy',
    name: `Target ${id}`,
    fields,
    assigneeId: 'u1',
    teamId: 't1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  }
}

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

// ---------- Unit tests for the field-name contract ---------------------------------

describe('Planner geo-target filtering', () => {
  /*
   * The backend stores coordinates as fields.lat / fields.lng
   * (see internal/service/target_service.go:211-212).
   * The planner must recognise these field names when filtering
   * geo-located targets and passing coords to TargetMarker.
   */

  it('recognises targets with lat/lng fields (backend contract)', () => {
    const targets = [
      makeTarget('1', { lat: 44.43, lng: 26.10, classification: 'A' }),
      makeTarget('2', { lat: 45.0, lng: 25.0 }),
      makeTarget('3', { address: 'No coords' }),
    ]

    const getLat = (fields: Record<string, unknown>): number | null => {
      const v = fields.lat
      return typeof v === 'number' ? v : null
    }
    const getLng = (fields: Record<string, unknown>): number | null => {
      const v = fields.lng
      return typeof v === 'number' ? v : null
    }

    const geo = targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null)
    expect(geo).toHaveLength(2)
    expect(geo.map((t) => t.id)).toEqual(['1', '2'])
  })

  it('rejects targets that only have latitude/longitude (wrong field names)', () => {
    const targets = [
      makeTarget('1', { latitude: 44.43, longitude: 26.10 }),
    ]

    const getLat = (fields: Record<string, unknown>): number | null => {
      const v = fields.lat
      return typeof v === 'number' ? v : null
    }
    const getLng = (fields: Record<string, unknown>): number | null => {
      const v = fields.lng
      return typeof v === 'number' ? v : null
    }

    const geo = targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null)
    expect(geo).toHaveLength(0)
  })
})

// ---------- Integration: TargetMarker receives coords from planner -----------------

describe('Planner map markers', () => {
  beforeAll(async () => {
    await import('./planner')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('creates map markers for targets with lat/lng fields', async () => {
    const PlannerPage = capturedComponent!
    expect(PlannerPage).toBeTruthy()

    mockTargets.mockReturnValue({
      data: {
        items: [
          makeTarget('t1', { lat: 44.43, lng: 26.10, classification: 'A' }),
          makeTarget('t2', { lat: 45.0, lng: 25.0, classification: 'B' }),
          makeTarget('t3', { address: 'No coords' }),
        ],
        total: 3,
        page: 1,
        limit: 500,
      },
      isLoading: false,
    })
    mockActivities.mockReturnValue({ data: { items: [], total: 0, page: 1, limit: 200 }, isLoading: false })
    mockStats.mockReturnValue({ data: { total: 0, byStatus: {} } })
    mockCoverage.mockReturnValue({ data: null })

    renderWithProviders(<PlannerPage />)

    // Wait for the map 'load' callback (setTimeout 0)
    await waitFor(() => {
      // Two markers should have been created (t1 and t2 have coords, t3 does not)
      expect(MockMarker).toHaveBeenCalledTimes(2)
    })

    expect(mockAddTo).toHaveBeenCalledTimes(2)
  })

  it('shows no markers when targets lack lat/lng', async () => {
    const PlannerPage = capturedComponent!

    mockTargets.mockReturnValue({
      data: {
        items: [makeTarget('t1', { address: 'Street 1' })],
        total: 1,
        page: 1,
        limit: 500,
      },
      isLoading: false,
    })
    mockActivities.mockReturnValue({ data: { items: [], total: 0, page: 1, limit: 200 }, isLoading: false })
    mockStats.mockReturnValue({ data: { total: 0, byStatus: {} } })
    mockCoverage.mockReturnValue({ data: null })

    renderWithProviders(<PlannerPage />)

    await waitFor(() => {
      expect(MockMap).toHaveBeenCalled()
    })

    expect(MockMarker).not.toHaveBeenCalled()
  })
})
