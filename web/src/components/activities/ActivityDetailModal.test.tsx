import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks -------------------------------------------------------------------
const mockUseActivity = vi.fn()
const mockPatchActivity = vi.fn()
const mockPatchStatus = vi.fn()
const mockSubmitActivity = vi.fn()
const mockUseConfig = vi.fn()

vi.mock('@/hooks/useActivities', () => ({
  useActivity: () => mockUseActivity(),
  usePatchActivity: () => mockPatchActivity(),
  usePatchActivityStatus: () => mockPatchStatus(),
  useSubmitActivity: () => mockSubmitActivity(),
}))

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => mockUseConfig(),
}))

// --- Helpers ----------------------------------------------------------------------

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function makeActivity(overrides: Record<string, unknown> = {}) {
  return {
    id: 'a1',
    activityType: 'visit',
    status: 'planificat',
    dueDate: '2026-04-10',
    duration: '30min',
    fields: {},
    creatorId: 'u1',
    createdAt: '2026-04-01T10:00:00Z',
    updatedAt: '2026-04-01T10:00:00Z',
    ...overrides,
  }
}

function makeConfig(overrides: Record<string, unknown> = {}) {
  return {
    tenant: { name: 'Test', locale: 'en' },
    accounts: { types: [] },
    activities: {
      statuses: [
        { key: 'planificat', label: 'Planned', initial: true },
        { key: 'realizat', label: 'Completed', submittable: true },
        { key: 'anulat', label: 'Cancelled' },
      ],
      status_transitions: {
        planificat: ['realizat', 'anulat'],
        realizat: [],
      },
      durations: [{ key: '30min', label: '30 minutes' }],
      types: [
        {
          key: 'visit',
          label: 'Visit',
          category: 'field',
          fields: [
            { key: 'feedback', label: 'Feedback', type: 'text', required: false },
            { key: 'promoted_products', label: 'Promoted Products', type: 'multi_select', required: false, options_ref: 'products' },
          ],
          submit_required: ['feedback', 'promoted_products'],
        },
      ],
      routing_options: [],
    },
    options: {
      products: [
        { key: 'prod_a', label: 'Product A' },
        { key: 'prod_b', label: 'Product B' },
      ],
    },
    rules: { frequency: {}, max_activities_per_day: 10, default_visit_duration_minutes: {}, visit_duration_step_minutes: 15 },
    ...overrides,
  }
}

const defaultMutationReturn = { mutate: vi.fn(), isPending: false }

// --- Import component after mocks -------------------------------------------------
import { ActivityDetailModal } from './ActivityDetailModal'

// --- Tests ------------------------------------------------------------------------

describe('ActivityDetailModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockPatchActivity.mockReturnValue(defaultMutationReturn)
    mockPatchStatus.mockReturnValue(defaultMutationReturn)
    mockSubmitActivity.mockReturnValue(defaultMutationReturn)
    mockUseConfig.mockReturnValue({ data: makeConfig() })
  })

  it('renders nothing when activityId is null', () => {
    mockUseActivity.mockReturnValue({ data: undefined, isLoading: false, refetch: vi.fn() })
    const { container } = renderWithProviders(
      <ActivityDetailModal activityId={null} onClose={vi.fn()} />,
    )
    expect(container.innerHTML).toBe('')
  })

  it('shows loading spinner when activity is loading', () => {
    mockUseActivity.mockReturnValue({ data: undefined, isLoading: true, refetch: vi.fn() })
    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders activity details with status badge and metadata', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({
        targetName: 'Pharmacy Central',
        fields: { notes: 'Good visit', tags: ['vip', 'new'] },
      }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    // Title should be targetName
    expect(screen.getByText('Pharmacy Central')).toBeInTheDocument()
    // Status badge
    expect(screen.getByText('Planned')).toBeInTheDocument()
    // Notes
    expect(screen.getByText('Good visit')).toBeInTheDocument()
    // Tags
    expect(screen.getByText('vip')).toBeInTheDocument()
    expect(screen.getByText('new')).toBeInTheDocument()
    // Metadata
    expect(screen.getByText('Created')).toBeInTheDocument()
  })

  it('shows target summary link when present', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({
        targetSummary: { id: 't1', targetType: 'pharmacy', name: 'DrugStore X', fields: {} },
      }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('DrugStore X')).toBeInTheDocument()
    expect(screen.getByText('pharmacy')).toBeInTheDocument()
  })

  it('renders status transition buttons', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'planificat' }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('Completed')).toBeInTheDocument()
    expect(screen.getByText('Cancelled')).toBeInTheDocument()
  })

  it('calls patchStatus.mutate when transition button is clicked', () => {
    const mutatePatch = vi.fn()
    const mutateStatus = vi.fn((_args: unknown, opts?: { onSuccess?: () => void }) => {
      opts?.onSuccess?.()
    })
    mockPatchActivity.mockReturnValue({ mutate: mutatePatch, isPending: false })
    mockPatchStatus.mockReturnValue({ mutate: mutateStatus, isPending: false })

    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'planificat' }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    fireEvent.click(screen.getByText('Completed'))
    expect(mutateStatus).toHaveBeenCalledWith(
      { id: 'a1', status: 'realizat' },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    )
  })

  it('shows Submit button when status is submittable and not yet submitted', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'realizat' }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('Submit')).toBeInTheDocument()
  })

  it('does not show Submit button when already submitted', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'realizat', submittedAt: '2026-04-05T12:00:00Z' }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.queryByText('Submit')).not.toBeInTheDocument()
    // Should show Submitted badge (there are two: the badge and metadata row label)
    expect(screen.getAllByText('Submitted').length).toBeGreaterThanOrEqual(1)
  })

  it('renders editable feedback textarea for non-locked activity', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'planificat', fields: { feedback: 'Draft notes' } }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    const textarea = screen.getByPlaceholderText('How did the visit go?')
    expect(textarea).toBeInTheDocument()
    expect(textarea).toHaveValue('Draft notes')
  })

  it('renders read-only feedback for locked (submitted) activity', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({
        status: 'realizat',
        submittedAt: '2026-04-05T12:00:00Z',
        fields: { feedback: 'Final feedback' },
      }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('Final feedback')).toBeInTheDocument()
    expect(screen.queryByPlaceholderText('How did the visit go?')).not.toBeInTheDocument()
  })

  it('toggles product selection when product button is clicked', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'planificat', fields: {} }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    const productBtn = screen.getByText('Product A')
    expect(productBtn).toBeInTheDocument()
    fireEvent.click(productBtn)
    // After click, it should have the teal selected style
    expect(productBtn.closest('button')).toHaveClass('border-teal-500')
  })

  it('shows routing in metadata when present', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ routing: 'north-route' }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('Routing')).toBeInTheDocument()
    expect(screen.getByText('north-route')).toBeInTheDocument()
  })

  it('shows visit type badge when visit_type field is present', () => {
    mockUseActivity.mockReturnValue({
      data: makeActivity({ fields: { visit_type: 'f2f' } }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    expect(screen.getByText('In person')).toBeInTheDocument()
  })

  it('shows validation errors on submit when required fields missing', async () => {
    const mutateSubmit = vi.fn()
    mockSubmitActivity.mockReturnValue({ mutate: mutateSubmit, isPending: false })

    mockUseActivity.mockReturnValue({
      data: makeActivity({ status: 'realizat', fields: {} }),
      isLoading: false,
      refetch: vi.fn(),
    })

    renderWithProviders(
      <ActivityDetailModal activityId="a1" onClose={vi.fn()} />,
    )

    fireEvent.click(screen.getByText('Submit'))

    // Validation errors for missing required fields
    await waitFor(() => {
      expect(screen.getByText('Feedback is required before submission')).toBeInTheDocument()
      expect(screen.getByText('Select at least one product')).toBeInTheDocument()
    })

    // submitActivity.mutate should NOT have been called
    expect(mutateSubmit).not.toHaveBeenCalled()
  })
})
