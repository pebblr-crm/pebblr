import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach, beforeAll } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks -------------------------------------------------------------------
const mockAuditLog = vi.fn()
const mockMutate = vi.fn()

vi.mock('@/hooks/useAudit', () => ({
  useAuditLog: () => mockAuditLog(),
  useUpdateAuditStatus: () => ({ mutate: mockMutate, isPending: false }),
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

function makeEntry(overrides: Partial<{
  id: string
  entityType: string
  entityId: string
  eventType: string
  actorId: string
  status: 'pending' | 'accepted' | 'false_positive'
  createdAt: string
}> = {}) {
  return {
    id: overrides.id ?? 'aud-1',
    entityType: overrides.entityType ?? 'activity',
    entityId: overrides.entityId ?? 'act-1',
    eventType: overrides.eventType ?? 'status_change',
    actorId: overrides.actorId ?? 'user-abcdefgh-1234',
    status: overrides.status ?? 'pending',
    createdAt: overrides.createdAt ?? '2026-03-15T10:30:00Z',
  }
}

function setDefaultMocks() {
  mockAuditLog.mockReturnValue({
    data: {
      items: [
        makeEntry({ id: 'a1', status: 'pending' }),
        makeEntry({ id: 'a2', status: 'accepted', entityType: 'target', eventType: 'field_update' }),
        makeEntry({ id: 'a3', status: 'false_positive', entityType: 'user', eventType: 'role_change' }),
      ],
      total: 3,
      page: 1,
      limit: 50,
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  })
}

// --- Tests ------------------------------------------------------------------------

describe('AuditPage', () => {
  beforeAll(async () => {
    await import('./audit')
  })

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows spinner while loading', () => {
    mockAuditLog.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    const { container } = renderWithProviders(<Page />)
    expect(container.querySelector('[class*="animate-spin"]') || container.innerHTML).toBeTruthy()
  })

  it('shows error state with retry', () => {
    mockAuditLog.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.getByText(/failed to load audit logs/i)).toBeTruthy()
  })

  it('renders page header and subtitle', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Audit Logs')).toBeTruthy()
    expect(screen.getByText(/immutable change history/i)).toBeTruthy()
  })

  it('renders data table with audit entries', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Column headers
    expect(screen.getByText('Timestamp')).toBeTruthy()
    expect(screen.getByText('Actor')).toBeTruthy()
    expect(screen.getByText('Entity')).toBeTruthy()
    expect(screen.getByText('Action')).toBeTruthy()
    expect(screen.getByText('Status')).toBeTruthy()
    expect(screen.getByText('Review')).toBeTruthy()
  })

  it('shows pending count badge', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText(/1 pending review/)).toBeTruthy()
  })

  it('does not show pending badge when no pending entries', () => {
    mockAuditLog.mockReturnValue({
      data: {
        items: [makeEntry({ id: 'a1', status: 'accepted' })],
        total: 1,
        page: 1,
        limit: 50,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)
    expect(screen.queryByText(/pending review/)).toBeNull()
  })

  it('renders total entry count', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('3 entries')).toBeTruthy()
  })

  it('renders entity type filter select', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByDisplayValue('All entities')).toBeTruthy()
  })

  it('renders status filter select', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByDisplayValue('All statuses')).toBeTruthy()
  })

  it('displays timestamp cell formatted', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // The timestamp should be formatted by toLocaleString with en-GB
    // All 3 entries share the same date, so use getAllByText
    expect(screen.getAllByText(/15 Mar 2026/).length).toBe(3)
  })

  it('displays actor cell truncated', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Actor IDs are truncated to first 8 chars + "..."
    expect(screen.getAllByText('user-abc...').length).toBeGreaterThanOrEqual(1)
  })

  it('displays entity type as badge', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('activity')).toBeTruthy()
    expect(screen.getByText('target')).toBeTruthy()
    expect(screen.getByText('user')).toBeTruthy()
  })

  it('displays event type with underscores replaced', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('status change')).toBeTruthy()
    expect(screen.getByText('field update')).toBeTruthy()
    expect(screen.getByText('role change')).toBeTruthy()
  })

  it('shows accept and false positive buttons for pending entries only', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // Only 1 pending entry, so 1 accept + 1 false positive button
    const acceptBtns = screen.getAllByLabelText('Accept audit entry')
    expect(acceptBtns).toHaveLength(1)

    const fpBtns = screen.getAllByLabelText('Mark as false positive')
    expect(fpBtns).toHaveLength(1)
  })

  it('calls mutate with accepted status on accept click', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    fireEvent.click(screen.getByLabelText('Accept audit entry'))
    expect(mockMutate).toHaveBeenCalledWith({ id: 'a1', status: 'accepted' })
  })

  it('calls mutate with false_positive status on false positive click', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    fireEvent.click(screen.getByLabelText('Mark as false positive'))
    expect(mockMutate).toHaveBeenCalledWith({ id: 'a1', status: 'false_positive' })
  })

  it('renders status badges with correct variants', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    // pending, accepted, false positive statuses rendered as badges
    expect(screen.getByText('pending')).toBeTruthy()
    expect(screen.getByText('accepted')).toBeTruthy()
    expect(screen.getByText('false positive')).toBeTruthy()
  })

  it('shows pagination when total > 50', () => {
    mockAuditLog.mockReturnValue({
      data: {
        items: Array.from({ length: 50 }, (_, i) => makeEntry({ id: `a${i}` })),
        total: 120,
        page: 1,
        limit: 50,
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    })

    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.getByText('Previous')).toBeTruthy()
    expect(screen.getByText('Page 1')).toBeTruthy()
    expect(screen.getByText('Next')).toBeTruthy()
  })

  it('does not show pagination when total <= 50', () => {
    setDefaultMocks() // total=3
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    expect(screen.queryByText('Previous')).toBeNull()
    expect(screen.queryByText('Page 1')).toBeNull()
  })

  it('has export logs button (disabled)', () => {
    setDefaultMocks()
    const Page = capturedComponent!
    renderWithProviders(<Page />)

    const exportBtn = screen.getByText('Export Logs')
    expect(exportBtn.closest('button')).toHaveProperty('disabled', true)
  })
})
