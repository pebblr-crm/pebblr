import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// --- Mock hooks -------------------------------------------------------------------
const mockUseConfig = vi.fn()
const mockCreateActivity = vi.fn()
const mockUseTargets = vi.fn()
const mockUseRecoveryBalance = vi.fn()

vi.mock('@/hooks/useConfig', () => ({
  useConfig: () => mockUseConfig(),
}))

vi.mock('@/hooks/useActivities', () => ({
  useCreateActivity: () => mockCreateActivity(),
}))

vi.mock('@/hooks/useTargets', () => ({
  useTargets: () => mockUseTargets(),
}))

vi.mock('@/hooks/useDashboard', () => ({
  useRecoveryBalance: () => mockUseRecoveryBalance(),
}))

// --- Helpers ----------------------------------------------------------------------

function renderWithProviders(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

function makeConfig(overrides: Record<string, unknown> = {}) {
  return {
    tenant: { name: 'Test', locale: 'en' },
    accounts: { types: [] },
    activities: {
      statuses: [
        { key: 'planificat', label: 'Planned', initial: true },
        { key: 'realizat', label: 'Completed', submittable: true },
      ],
      status_transitions: { planificat: ['realizat'] },
      durations: [
        { key: '15min', label: '15 minutes' },
        { key: '30min', label: '30 minutes' },
      ],
      types: [
        {
          key: 'visit',
          label: 'Visit',
          category: 'field',
          fields: [
            { key: 'duration', type: 'select', required: false },
            { key: 'routing', label: 'Routing', type: 'select', required: false, options: ['north', 'south'] },
            { key: 'feedback', label: 'Feedback', type: 'text', required: false },
          ],
          submit_required: [],
        },
        {
          key: 'office',
          label: 'Office Work',
          category: 'non_field',
          fields: [],
          submit_required: [],
        },
      ],
      routing_options: [{ key: 'north', label: 'North' }, { key: 'south', label: 'South' }],
    },
    options: {
      products: [{ key: 'prod_a', label: 'Product A' }],
    },
    rules: { frequency: {}, max_activities_per_day: 10, default_visit_duration_minutes: {}, visit_duration_step_minutes: 15 },
    recovery: { weekend_activity_flag: true, recovery_window_days: 5, recovery_type: 'recovery' },
    ...overrides,
  }
}

// --- Import component after mocks -------------------------------------------------
import { CreateActivityModal } from './CreateActivityModal'

// --- Tests ------------------------------------------------------------------------

describe('CreateActivityModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseConfig.mockReturnValue({ data: makeConfig() })
    mockCreateActivity.mockReturnValue({ mutate: vi.fn(), isPending: false })
    mockUseTargets.mockReturnValue({ data: { items: [], total: 0 } })
    mockUseRecoveryBalance.mockReturnValue({ data: null })
  })

  it('renders nothing when closed', () => {
    renderWithProviders(<CreateActivityModal open={false} onClose={vi.fn()} />)
    expect(screen.queryByText('Log Activity')).not.toBeInTheDocument()
  })

  it('renders modal title and form when open', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)
    expect(screen.getByText('Log Activity')).toBeInTheDocument()
    expect(screen.getByText('Activity Type')).toBeInTheDocument()
    expect(screen.getByText('Date')).toBeInTheDocument()
  })

  it('renders activity type options from config', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)
    const select = screen.getByLabelText('Activity Type')
    expect(select).toBeInTheDocument()
    expect(screen.getByText('Visit')).toBeInTheDocument()
    expect(screen.getByText('Office Work')).toBeInTheDocument()
  })

  it('shows target search when field activity type is selected', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })

    expect(screen.getByLabelText('Target')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Search targets...')).toBeInTheDocument()
  })

  it('does not show target search for non-field activity type', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'office' } })

    expect(screen.queryByLabelText('Target')).not.toBeInTheDocument()
  })

  it('shows duration selector when type has duration field', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })

    expect(screen.getByLabelText('Duration')).toBeInTheDocument()
    expect(screen.getByText('15 minutes')).toBeInTheDocument()
    expect(screen.getByText('30 minutes')).toBeInTheDocument()
  })

  it('shows target dropdown and allows selection', async () => {
    mockUseTargets.mockReturnValue({
      data: {
        items: [
          { id: 't1', name: 'Pharmacy Alpha', targetType: 'pharmacy' },
          { id: 't2', name: 'Pharmacy Beta', targetType: 'pharmacy' },
        ],
        total: 2,
      },
    })

    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })
    fireEvent.change(screen.getByPlaceholderText('Search targets...'), {
      target: { value: 'Pharmacy' },
    })

    await waitFor(() => {
      expect(screen.getByText('Pharmacy Alpha')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('Pharmacy Alpha'))

    expect(screen.getByText(/Selected:.*Pharmacy Alpha/)).toBeInTheDocument()
  })

  it('disables Create button when no activity type is selected', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)
    const createBtn = screen.getByText('Create Activity')
    expect(createBtn).toBeDisabled()
  })

  it('enables Create button for non-field type with date selected', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'office' } })

    const createBtn = screen.getByText('Create Activity')
    expect(createBtn).not.toBeDisabled()
  })

  it('disables Create button for field type without target', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })

    const createBtn = screen.getByText('Create Activity')
    expect(createBtn).toBeDisabled()
  })

  it('calls createActivity.mutate on form submit', async () => {
    const mutateFn = vi.fn((_args: unknown, opts?: { onSuccess?: () => void }) => {
      opts?.onSuccess?.()
    })
    mockCreateActivity.mockReturnValue({ mutate: mutateFn, isPending: false })

    mockUseTargets.mockReturnValue({
      data: {
        items: [{ id: 't1', name: 'Pharmacy Alpha', targetType: 'pharmacy' }],
        total: 1,
      },
    })

    const onClose = vi.fn()
    renderWithProviders(<CreateActivityModal open={true} onClose={onClose} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })
    fireEvent.change(screen.getByPlaceholderText('Search targets...'), { target: { value: 'Pharmacy' } })

    await waitFor(() => {
      expect(screen.getByText('Pharmacy Alpha')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByText('Pharmacy Alpha'))

    fireEvent.click(screen.getByText('Create Activity'))

    expect(mutateFn).toHaveBeenCalledWith(
      expect.objectContaining({
        activityType: 'visit',
        status: 'planificat',
        targetId: 't1',
      }),
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    )
  })

  it('shows "Creating..." text when mutation is pending', () => {
    mockCreateActivity.mockReturnValue({ mutate: vi.fn(), isPending: true })
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)
    expect(screen.getByText('Creating...')).toBeInTheDocument()
  })

  it('shows recovery balance info for recovery activity type', () => {
    const configWithRecovery = makeConfig()
    ;(configWithRecovery.activities as Record<string, unknown[]>).types = [
      ...(configWithRecovery.activities as { types: unknown[] }).types,
      { key: 'recovery', label: 'Recovery', category: 'non_field', fields: [], submit_required: [] },
    ]
    mockUseConfig.mockReturnValue({ data: configWithRecovery })
    mockUseRecoveryBalance.mockReturnValue({
      data: { earned: 3, taken: 1, balance: 2, intervals: [] },
    })

    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'recovery' } })

    expect(screen.getByText('Recovery Balance')).toBeInTheDocument()
    expect(screen.getByText(/2 day/)).toBeInTheDocument()
    expect(screen.getByText(/3 earned, 1 taken/)).toBeInTheDocument()
  })

  it('disables Create when recovery balance is 0', () => {
    const configWithRecovery = makeConfig()
    ;(configWithRecovery.activities as Record<string, unknown[]>).types = [
      ...(configWithRecovery.activities as { types: unknown[] }).types,
      { key: 'recovery', label: 'Recovery', category: 'non_field', fields: [], submit_required: [] },
    ]
    mockUseConfig.mockReturnValue({ data: configWithRecovery })
    mockUseRecoveryBalance.mockReturnValue({
      data: { earned: 1, taken: 1, balance: 0, intervals: [] },
    })

    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'recovery' } })

    expect(screen.getByText('Create Activity')).toBeDisabled()
    expect(screen.getByText('No recovery days available to claim.')).toBeInTheDocument()
  })

  it('renders dynamic fields (routing select) for activity type', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    fireEvent.change(screen.getByLabelText('Activity Type'), { target: { value: 'visit' } })

    // Dynamic field section should appear with routing and feedback
    expect(screen.getByText('Details')).toBeInTheDocument()
    expect(screen.getByText('Routing')).toBeInTheDocument()
    expect(screen.getByText('Feedback')).toBeInTheDocument()
  })

  it('renders date picker with today as default', () => {
    renderWithProviders(<CreateActivityModal open={true} onClose={vi.fn()} />)

    const dateInput = screen.getByLabelText('Date') as HTMLInputElement
    const today = new Date().toISOString().slice(0, 10)
    expect(dateInput.value).toBe(today)
  })
})
