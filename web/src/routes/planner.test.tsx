import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock @vis.gl/react-google-maps before any component imports -----------------
vi.mock('@vis.gl/react-google-maps', () => ({
  APIProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Map: ({ children }: { children?: React.ReactNode }) => <div data-testid="google-map">{children}</div>,
  AdvancedMarker: ({ children }: { children?: React.ReactNode }) => <div data-testid="advanced-marker">{children}</div>,
}))

// --- Mock hooks -------------------------------------------------------------------
const mockTargets = vi.fn()
const mockTargetVisitStatus = vi.fn()
const mockActivities = vi.fn()
const mockStats = vi.fn()
const mockCoverage = vi.fn()
const mockCloneWeek = vi.fn()
const mockBatchCreate = vi.fn()
const mockPatchActivity = vi.fn()
const mockWeekNav = vi.fn()
const mockDragDrop = vi.fn()
const mockShowToast = vi.fn()

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockTargets(),
  useTargetVisitStatus: () => mockTargetVisitStatus(),
}))

vi.mock('@/hooks/useActivities', () => ({
  useActivities: () => mockActivities(),
  useCloneWeek: () => mockCloneWeek(),
  useBatchCreateActivities: () => mockBatchCreate(),
  usePatchActivity: () => mockPatchActivity(),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useActivityStats: () => mockStats(),
  useCoverage: () => mockCoverage(),
}))

vi.mock('@/hooks/useWeekNav', () => ({
  useWeekNav: () => mockWeekNav(),
}))

vi.mock('@/hooks/useDragDropPlanner', () => ({
  useDragDropPlanner: () => mockDragDrop(),
}))

vi.mock('@/components/ui/Toast', () => ({
  useToast: () => ({ showToast: mockShowToast, ToastContainer: () => null }),
}))

vi.mock('@/components/activities/ActivityDetailModal', () => ({
  ActivityDetailModal: () => null,
}))

// --- Mock child components as simple stubs with data-testid ----------------------
vi.mock('@/components/calendar/WeekView', () => ({
  WeekView: (props: { activities: unknown[] }) => (
    <div data-testid="week-view">{props.activities.length} activities</div>
  ),
}))

vi.mock('@/components/planner/TargetListPanel', () => ({
  TargetListPanel: (props: { filteredTargets: unknown[] }) => (
    <div data-testid="target-list-panel">{props.filteredTargets.length} targets</div>
  ),
}))

vi.mock('@/components/planner/PlannerHeader', () => ({
  PlannerHeader: (props: { totalAssigned: number }) => (
    <div data-testid="planner-header">{props.totalAssigned} assigned</div>
  ),
}))

vi.mock('@/components/planner/PlannerNudgeBanner', () => ({
  PlannerNudgeBanner: (props: { overdueA: number; completionRate: number; coveragePct: number }) => (
    <div data-testid="nudge-banner">
      overdueA={props.overdueA} rate={props.completionRate} coverage={props.coveragePct}
    </div>
  ),
}))

vi.mock('@/components/planner/PlannerMobileMap', () => ({
  PlannerMobileMap: (props: { geoTargets: unknown[] }) => (
    <div data-testid="mobile-map">{props.geoTargets.length} geo targets</div>
  ),
}))

vi.mock('@/components/planner/BulkScheduleModal', () => ({
  BulkScheduleModal: () => <div data-testid="bulk-schedule-modal" />,
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

function makeTarget(id: string, fields: Record<string, unknown> = {}, overrides: Record<string, unknown> = {}) {
  return {
    id,
    targetType: 'pharmacy',
    name: `Target ${id}`,
    fields,
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
    status: 'planificat',
    dueDate: '2026-03-30T09:00:00Z',
    duration: '30m',
    fields: {},
    creatorId: 'u1',
    createdAt: '2026-03-30T00:00:00Z',
    updatedAt: '2026-03-30T00:00:00Z',
    ...overrides,
  }
}

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function setDefaultMocks(overrides: Partial<{
  targets: unknown[]
  activities: unknown[]
  stats: unknown
  coverage: unknown
  visitStatus: unknown[]
  dayAssignments: Record<string, string[]>
  targetsLoading: boolean
  activitiesLoading: boolean
  targetsError: boolean
  activitiesError: boolean
}> = {}) {
  const refetchTargets = vi.fn()
  const refetchActivities = vi.fn()

  mockTargets.mockReturnValue({
    data: overrides.targets !== undefined
      ? { items: overrides.targets, total: (overrides.targets as unknown[]).length, page: 1, limit: 500 }
      : { items: [], total: 0, page: 1, limit: 500 },
    isLoading: overrides.targetsLoading ?? false,
    isError: overrides.targetsError ?? false,
    refetch: refetchTargets,
  })

  mockActivities.mockReturnValue({
    data: overrides.activities !== undefined
      ? { items: overrides.activities, total: (overrides.activities as unknown[]).length, page: 1, limit: 200 }
      : { items: [], total: 0, page: 1, limit: 200 },
    isLoading: overrides.activitiesLoading ?? false,
    isError: overrides.activitiesError ?? false,
    refetch: refetchActivities,
  })

  mockStats.mockReturnValue({
    data: overrides.stats ?? { total: 0, byStatus: {} },
  })

  mockCoverage.mockReturnValue({
    data: overrides.coverage ?? null,
  })

  mockTargetVisitStatus.mockReturnValue({
    data: { items: overrides.visitStatus ?? [] },
  })

  mockWeekNav.mockReturnValue({
    weekStart: new Date('2026-03-30'),
    weekEnd: new Date('2026-04-03'),
    dateFrom: '2026-03-30',
    dateTo: '2026-04-03',
    prevWeek: vi.fn(),
    nextWeek: vi.fn(),
    goToday: vi.fn(),
  })

  mockDragDrop.mockReturnValue({
    dragActivityId: null,
    dragPending: null,
    isDragging: false,
    dayAssignments: overrides.dayAssignments ?? {},
    setDragTargetId: vi.fn(),
    setDragActivityId: vi.fn(),
    setDragPending: vi.fn(),
    handleDrop: vi.fn(),
    removeFromDay: vi.fn(),
    setDayAssignments: vi.fn(),
  })

  mockCloneWeek.mockReturnValue({ mutate: vi.fn(), isPending: false })
  mockBatchCreate.mockReturnValue({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false })
  mockPatchActivity.mockReturnValue({ mutate: vi.fn(), isPending: false })

  return { refetchTargets, refetchActivities }
}

// --- Tests ------------------------------------------------------------------------

describe('PlannerPage', () => {
  beforeAll(async () => {
    await import('./planner')
  })

  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  // 1. Loading state: spinner when targets loading
  it('shows spinner when targets are loading', () => {
    setDefaultMocks({ targetsLoading: true })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  // 2. Loading state: spinner when activities loading
  it('shows spinner when activities are loading', () => {
    setDefaultMocks({ activitiesLoading: true })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  // 3. Error state: shows error message with retry button
  it('shows error message with retry button when targets fail', () => {
    setDefaultMocks({ targetsError: true })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Failed to load planner data')).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  // 4. Error state: retry calls refetch on both queries
  it('retry button calls refetch on both targets and activities', () => {
    const { refetchTargets, refetchActivities } = setDefaultMocks({ targetsError: true })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    fireEvent.click(screen.getByText('Retry'))
    expect(refetchTargets).toHaveBeenCalledOnce()
    expect(refetchActivities).toHaveBeenCalledOnce()
  })

  // 5. Empty state: renders with no targets and no activities
  it('renders with no targets and no activities', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByTestId('target-list-panel')).toHaveTextContent('0 targets')
    expect(screen.getByTestId('week-view')).toHaveTextContent('0 activities')
  })

  // 6. Renders with targets: passes filteredTargets to TargetListPanel
  it('passes filteredTargets (geo-only) to TargetListPanel', () => {
    setDefaultMocks({
      targets: [
        makeTarget('t1', { lat: 44.43, lng: 26.10 }),
        makeTarget('t2', { lat: 45.0, lng: 25.0 }),
        makeTarget('t3', { address: 'No coords' }),
      ],
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // Only t1 and t2 have lat/lng, t3 is excluded
    expect(screen.getByTestId('target-list-panel')).toHaveTextContent('2 targets')
  })

  // 7. Renders with activities: passes activities to WeekView
  it('passes activities to WeekView', () => {
    setDefaultMocks({
      activities: [
        makeActivity('a1'),
        makeActivity('a2'),
        makeActivity('a3'),
      ],
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByTestId('week-view')).toHaveTextContent('3 activities')
  })

  // 8. Geo filtering: only targets with lat/lng counted as geoTargets
  it('only targets with lat/lng are counted as geoTargets', () => {
    setDefaultMocks({
      targets: [
        makeTarget('t1', { lat: 44.43, lng: 26.10 }),
        makeTarget('t2', { latitude: 45.0, longitude: 25.0 }), // wrong field names
        makeTarget('t3', {}),
        makeTarget('t4', { lat: 43.0 }), // missing lng
      ],
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // Only t1 has both lat and lng
    expect(screen.getByTestId('target-list-panel')).toHaveTextContent('1 targets')
    // Mobile map button shows count of geoTargets
    expect(screen.getByText('Show Map (1 targets)')).toBeInTheDocument()
  })

  // 9. Overdued A calculation: counts A-priority targets without recent visits
  it('counts overdue A-priority targets without recent visits', () => {
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    setDefaultMocks({
      targets: [
        makeTarget('t1', { lat: 44.0, lng: 26.0, potential: 'a' }),
        makeTarget('t2', { lat: 45.0, lng: 25.0, potential: 'a' }),
        makeTarget('t3', { lat: 46.0, lng: 24.0, potential: 'b' }),
      ],
      visitStatus: [
        { targetId: 't1', lastVisitDate: thirtyDaysAgo.toISOString() }, // overdue (>21 days)
      ],
      // t2 has no visit at all, so it's overdue
      // t3 is potential B, not counted
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // overdueA = 2 (t1 overdue + t2 never visited), should show nudge banner
    expect(screen.getByTestId('nudge-banner')).toHaveTextContent('overdueA=2')
  })

  // 10. Overdued A: excludes targets with activities this week
  it('excludes targets with activities this week from overdueA', () => {
    setDefaultMocks({
      targets: [
        makeTarget('t1', { lat: 44.0, lng: 26.0, potential: 'a' }),
        makeTarget('t2', { lat: 45.0, lng: 25.0, potential: 'a' }),
      ],
      activities: [
        makeActivity('a1', { targetId: 't1' }), // t1 has an activity this week
      ],
      // No visit status at all -> both would be overdue, but t1 is scheduled
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // t1 is scheduled (has activity), so only t2 is overdue
    expect(screen.getByTestId('nudge-banner')).toHaveTextContent('overdueA=1')
  })

  // 11. Stats display: shows NudgeBanner when stats.total > 0
  it('shows NudgeBanner when stats.total > 0', () => {
    setDefaultMocks({
      stats: { total: 10, byStatus: { realizat: 5, planificat: 3, anulat: 2 } },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByTestId('nudge-banner')).toBeInTheDocument()
  })

  // 12. Stats display: hides NudgeBanner when no stats and overdueA = 0
  it('hides NudgeBanner when no stats and overdueA is 0', () => {
    setDefaultMocks({
      stats: { total: 0, byStatus: {} },
      targets: [], // no targets, so overdueA = 0
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.queryByTestId('nudge-banner')).not.toBeInTheDocument()
  })

  // 13. Completion rate: calculates correctly from stats
  it('calculates completion rate correctly from stats', () => {
    setDefaultMocks({
      stats: { total: 20, byStatus: { realizat: 14, planificat: 4, anulat: 2 } },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // completionRate = Math.round(14/20 * 100) = 70
    expect(screen.getByTestId('nudge-banner')).toHaveTextContent('rate=70')
  })

  // 14. Coverage: rounds percentage
  it('rounds coverage percentage', () => {
    setDefaultMocks({
      stats: { total: 1, byStatus: { realizat: 1 } },
      coverage: { percentage: 73.7 },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // coveragePct = Math.round(73.7) = 74
    expect(screen.getByTestId('nudge-banner')).toHaveTextContent('coverage=74')
  })

  // 15. Day assignments: totalAssigned counts pending assignments
  it('totalAssigned counts pending day assignments', () => {
    setDefaultMocks({
      dayAssignments: {
        '2026-03-30': ['t1', 't2'],
        '2026-03-31': ['t3'],
      },
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    // totalAssigned = 2 + 1 = 3
    expect(screen.getByTestId('planner-header')).toHaveTextContent('3 assigned')
  })

  // 16. Mobile map toggle: shows map button with target count
  it('shows mobile map button with geo target count', () => {
    setDefaultMocks({
      targets: [
        makeTarget('t1', { lat: 44.0, lng: 26.0 }),
        makeTarget('t2', { lat: 45.0, lng: 25.0 }),
        makeTarget('t3', { lat: 46.0, lng: 24.0 }),
      ],
    })
    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText('Show Map (3 targets)')).toBeInTheDocument()
  })
})
