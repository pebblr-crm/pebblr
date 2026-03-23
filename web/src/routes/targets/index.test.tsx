import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { TargetsPage } from './index'
import type { Target } from '../../types/target'
import type { PaginatedResponse } from '../../types/api'
import type { TenantConfig } from '../../types/config'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className, ...rest }: { children?: React.ReactNode; className?: string; to?: string }) => (
      <a className={className} href={rest.to}>{children}</a>
    ),
  }
})

vi.mock('../../services/targets', () => ({
  useTargets: vi.fn(),
}))

vi.mock('../../services/config', () => ({
  useConfig: vi.fn(),
}))

import { useTargets } from '../../services/targets'
import { useConfig } from '../../services/config'
const mockUseTargets = vi.mocked(useTargets)
const mockUseConfig = vi.mocked(useConfig)

const testConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: {
    types: [
      {
        key: 'doctor',
        label: 'Doctor',
        fields: [
          { key: 'name', type: 'text', required: true },
          { key: 'specialty', type: 'select', required: false, options_ref: 'specialties' },
          { key: 'city', type: 'text', required: false },
        ],
      },
      {
        key: 'pharmacy',
        label: 'Pharmacy',
        fields: [
          { key: 'name', type: 'text', required: true },
          { key: 'city', type: 'text', required: false },
        ],
      },
    ],
  },
  activities: {
    statuses: [{ key: 'planificat', label: 'Planned', initial: true }],
    status_transitions: {},
    durations: [{ key: 'full_day', label: 'Full Day' }],
    types: [],
    routing_options: [],
  },
  options: {
    specialties: [
      { key: 'cardiology', label: 'Cardiology' },
      { key: 'neurology', label: 'Neurology' },
    ],
  },
  rules: {
    frequency: {},
    max_activities_per_day: 10,
    default_visit_duration_minutes: {},
    visit_duration_step_minutes: 30,
  },
}

function makeTarget(overrides: Partial<Target> = {}): Target {
  return {
    id: 'target-1',
    targetType: 'doctor',
    name: 'Dr. Smith',
    fields: { specialty: 'cardiology', city: 'Bucharest', county: 'Ilfov' },
    assigneeId: 'user-1',
    teamId: 'team-1',
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    ...overrides,
  }
}

function makePage(
  items: Target[],
  overrides: Partial<PaginatedResponse<Target>> = {},
): PaginatedResponse<Target> {
  return {
    items,
    total: items.length,
    page: 1,
    limit: 20,
    ...overrides,
  }
}

function setupConfig() {
  mockUseConfig.mockReturnValue({
    data: testConfig,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useConfig>)
}

describe('TargetsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setupConfig()
  })

  it('shows loading spinner while fetching', () => {
    mockUseTargets.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByLabelText('Loading targets...')).toBeInTheDocument()
  })

  it('shows error state when fetch fails', () => {
    mockUseTargets.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Network error'),
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('error-state')).toBeInTheDocument()
    expect(screen.getByText('Network error')).toBeInTheDocument()
  })

  it('shows empty state when no targets', () => {
    mockUseTargets.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    expect(screen.getByText('No targets found.')).toBeInTheDocument()
  })

  it('renders target rows with name and type', () => {
    const targets = [
      makeTarget({ id: 't1', name: 'Dr. Smith', targetType: 'doctor' }),
      makeTarget({ id: 't2', name: 'PharmaCo', targetType: 'pharmacy', fields: { city: 'Cluj' } }),
    ]
    mockUseTargets.mockReturnValue({
      data: makePage(targets, { total: 2 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByText('Dr. Smith')).toBeInTheDocument()
    expect(screen.getByText('PharmaCo')).toBeInTheDocument()
    // "Doctor" and "Pharmacy" appear in both the type filter and the badges/stats
    expect(screen.getAllByText('Doctor').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Pharmacy').length).toBeGreaterThanOrEqual(1)
  })

  it('shows type filter dropdown with config-driven options', () => {
    mockUseTargets.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByLabelText('Type:')).toBeInTheDocument()
    expect(screen.getByText('All types')).toBeInTheDocument()
    expect(screen.getByText('Doctor')).toBeInTheDocument()
    expect(screen.getByText('Pharmacy')).toBeInTheDocument()
  })

  it('passes type filter to useTargets when changed', async () => {
    const user = userEvent.setup()
    mockUseTargets.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    await user.selectOptions(screen.getByLabelText('Type:'), 'doctor')

    await waitFor(() => {
      expect(mockUseTargets).toHaveBeenCalledWith(
        expect.objectContaining({ type: 'doctor' }),
      )
    })
  })

  it('resets to page 1 when type filter changes', async () => {
    const user = userEvent.setup()
    const targets = Array.from({ length: 20 }, (_, i) =>
      makeTarget({ id: `t${i}`, name: `Target ${i}` }),
    )
    mockUseTargets.mockReturnValue({
      data: makePage(targets, { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    await user.click(screen.getByTestId('next-page'))
    await waitFor(() => {
      expect(mockUseTargets).toHaveBeenCalledWith(expect.objectContaining({ page: 2 }))
    })

    await user.selectOptions(screen.getByLabelText('Type:'), 'pharmacy')

    await waitFor(() => {
      expect(mockUseTargets).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, type: 'pharmacy' }),
      )
    })
  })

  it('shows pagination controls', () => {
    mockUseTargets.mockReturnValue({
      data: makePage([], { total: 0, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('pagination')).toBeInTheDocument()
    expect(screen.getByTestId('prev-page')).toBeDisabled()
  })

  it('enables Next button when more pages exist', () => {
    const targets = Array.from({ length: 20 }, (_, i) =>
      makeTarget({ id: `t${i}`, name: `Target ${i}` }),
    )
    mockUseTargets.mockReturnValue({
      data: makePage(targets, { total: 50, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('next-page')).not.toBeDisabled()
  })

  it('shows page indicator', () => {
    mockUseTargets.mockReturnValue({
      data: makePage([], { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('page-indicator')).toHaveTextContent('Page 1 of 2')
  })

  it('shows result count', () => {
    const targets = [makeTarget()]
    mockUseTargets.mockReturnValue({
      data: makePage(targets, { total: 1 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('1–1 of 1')
  })

  it('shows 0 results when total is 0', () => {
    mockUseTargets.mockReturnValue({
      data: makePage([], { total: 0 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('0 results')
  })

  it('shows filtered empty state message when type filter is active', async () => {
    const user = userEvent.setup()
    mockUseTargets.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    await user.selectOptions(screen.getByLabelText('Type:'), 'doctor')

    await waitFor(() => {
      expect(screen.getByText('No targets of type "doctor".')).toBeInTheDocument()
    })
  })

  it('shows location from fields when no type filter', () => {
    const targets = [makeTarget({ fields: { city: 'Bucharest', county: 'Ilfov' } })]
    mockUseTargets.mockReturnValue({
      data: makePage(targets, { total: 1 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTargets>)

    render(<TargetsPage />)

    expect(screen.getByText('Bucharest, Ilfov')).toBeInTheDocument()
  })
})
