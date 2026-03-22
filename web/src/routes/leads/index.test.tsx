import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { LeadsPage } from './index'
import type { Lead } from '../../types/lead'
import type { PaginatedResponse } from '../../types/api'

// Mock the router's Link component so tests don't need a full RouterProvider
vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className, ...rest }: { children?: React.ReactNode; className?: string; to?: string }) => (
      <a className={className} href={rest.to}>{children}</a>
    ),
  }
})

// Mock the leads service so tests don't hit the network
vi.mock('../../services/leads', () => ({
  useLeads: vi.fn(),
}))

import { useLeads } from '../../services/leads'
const mockUseLeads = vi.mocked(useLeads)

function makeLead(overrides: Partial<Lead> = {}): Lead {
  return {
    id: 'lead-1',
    title: 'Acme Corp',
    description: 'A big account',
    status: 'new',
    assigneeId: 'user-1',
    teamId: 'team-1',
    customerId: 'cust-1',
    customerType: 'retail',
    company: '',
    industry: '',
    location: '',
    valueCents: 0,
    initials: '',
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    ...overrides,
  }
}

function makePage(
  items: Lead[],
  overrides: Partial<PaginatedResponse<Lead>> = {},
): PaginatedResponse<Lead> {
  return {
    items,
    total: items.length,
    page: 1,
    limit: 20,
    ...overrides,
  }
}

describe('LeadsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner while fetching', () => {
    mockUseLeads.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByLabelText('Loading leads...')).toBeInTheDocument()
  })

  it('shows error state when fetch fails', () => {
    mockUseLeads.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Network error'),
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('error-state')).toBeInTheDocument()
    expect(screen.getByText('Network error')).toBeInTheDocument()
  })

  it('shows empty state when no leads', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    expect(screen.getByText('No leads found.')).toBeInTheDocument()
  })

  it('renders lead rows with correct columns', () => {
    const leads = [
      makeLead({ id: 'l1', title: 'Acme Corp', status: 'new', customerType: 'retail' }),
      makeLead({ id: 'l2', title: 'Globex Inc', status: 'assigned', customerType: 'wholesale' }),
    ]
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 2 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByText('Acme Corp')).toBeInTheDocument()
    expect(screen.getByText('Globex Inc')).toBeInTheDocument()
    expect(screen.getByText('Title')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Customer Type')).toBeInTheDocument()
  })

  it('renders status badge with underscores replaced', () => {
    const leads = [makeLead({ status: 'closed_won' })]
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 1 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByText('closed won')).toBeInTheDocument()
  })

  it('shows pagination controls', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([], { total: 0, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('pagination')).toBeInTheDocument()
    expect(screen.getByTestId('prev-page')).toBeInTheDocument()
    expect(screen.getByTestId('next-page')).toBeInTheDocument()
  })

  it('disables Previous button on first page', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([], { total: 0, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('prev-page')).toBeDisabled()
  })

  it('enables Next button when more pages exist', () => {
    const leads = Array.from({ length: 20 }, (_, i) =>
      makeLead({ id: `l${i}`, title: `Lead ${i}` }),
    )
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 50, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('next-page')).not.toBeDisabled()
  })

  it('shows page indicator', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([], { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('page-indicator')).toHaveTextContent('Page 1 of 2')
  })

  it('advances to next page when Next is clicked', async () => {
    const user = userEvent.setup()
    const leads = Array.from({ length: 20 }, (_, i) =>
      makeLead({ id: `l${i}`, title: `Lead ${i}` }),
    )
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    await user.click(screen.getByTestId('next-page'))

    await waitFor(() => {
      expect(mockUseLeads).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      )
    })
  })

  it('shows status filter dropdown', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByLabelText('Status:')).toBeInTheDocument()
    expect(screen.getByText('All statuses')).toBeInTheDocument()
  })

  it('passes status filter to useLeads when changed', async () => {
    const user = userEvent.setup()
    mockUseLeads.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    await user.selectOptions(screen.getByLabelText('Status:'), 'new')

    await waitFor(() => {
      expect(mockUseLeads).toHaveBeenCalledWith(
        expect.objectContaining({ status: 'new' }),
      )
    })
  })

  it('resets to page 1 when status filter changes', async () => {
    const user = userEvent.setup()
    const leads = Array.from({ length: 20 }, (_, i) =>
      makeLead({ id: `l${i}`, title: `Lead ${i}` }),
    )
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    // Advance to page 2
    await user.click(screen.getByTestId('next-page'))
    await waitFor(() => {
      expect(mockUseLeads).toHaveBeenCalledWith(expect.objectContaining({ page: 2 }))
    })

    // Change filter — should reset to page 1
    await user.selectOptions(screen.getByLabelText('Status:'), 'assigned')

    await waitFor(() => {
      expect(mockUseLeads).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, status: 'assigned' }),
      )
    })
  })

  it('shows filtered empty state message when status filter is active', async () => {
    const user = userEvent.setup()
    mockUseLeads.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    await user.selectOptions(screen.getByLabelText('Status:'), 'visited')

    await waitFor(() => {
      expect(screen.getByText('No leads with status "visited".')).toBeInTheDocument()
    })
  })

  it('shows result count', () => {
    const leads = [makeLead()]
    mockUseLeads.mockReturnValue({
      data: makePage(leads, { total: 1 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('1–1 of 1')
  })

  it('shows 0 results when total is 0', () => {
    mockUseLeads.mockReturnValue({
      data: makePage([], { total: 0 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useLeads>)

    render(<LeadsPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('0 results')
  })
})
