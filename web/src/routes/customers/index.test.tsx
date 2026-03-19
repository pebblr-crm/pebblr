import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { CustomersPage } from './index'
import type { Customer, CustomerType } from '../../types/customer'
import type { PaginatedResponse } from '../../types/api'

// Mock the customers service so tests don't hit the network
vi.mock('../../services/customers', () => ({
  useCustomers: vi.fn(),
}))

import { useCustomers } from '../../services/customers'
const mockUseCustomers = vi.mocked(useCustomers)

function makeCustomer(overrides: Partial<Customer> = {}): Customer {
  return {
    id: 'cust-1',
    name: 'Acme Corp',
    type: 'retail',
    address: { street: '123 Main St', city: 'Springfield', state: 'IL', country: 'US', zip: '62701' },
    phone: '555-0100',
    email: 'contact@acme.com',
    notes: '',
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    ...overrides,
  }
}

function makePage(
  items: Customer[],
  overrides: Partial<PaginatedResponse<Customer>> = {},
): PaginatedResponse<Customer> {
  return {
    items,
    total: items.length,
    page: 1,
    limit: 20,
    ...overrides,
  }
}

describe('CustomersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading spinner while fetching', () => {
    mockUseCustomers.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByLabelText('Loading customers...')).toBeInTheDocument()
  })

  it('shows error state when fetch fails', () => {
    mockUseCustomers.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Network error'),
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('error-state')).toBeInTheDocument()
    expect(screen.getByText('Network error')).toBeInTheDocument()
  })

  it('shows empty state when no customers', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    expect(screen.getByText('No customers found.')).toBeInTheDocument()
  })

  it('renders customer rows with correct columns', () => {
    const customers = [
      makeCustomer({ id: 'c1', name: 'Acme Corp', type: 'retail' }),
      makeCustomer({ id: 'c2', name: 'Globex Inc', type: 'wholesale' }),
    ]
    mockUseCustomers.mockReturnValue({
      data: makePage(customers, { total: 2 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByText('Acme Corp')).toBeInTheDocument()
    expect(screen.getByText('Globex Inc')).toBeInTheDocument()
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Type')).toBeInTheDocument()
    expect(screen.getByText('Email')).toBeInTheDocument()
  })

  it('shows pagination controls', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([], { total: 0, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('pagination')).toBeInTheDocument()
    expect(screen.getByTestId('prev-page')).toBeInTheDocument()
    expect(screen.getByTestId('next-page')).toBeInTheDocument()
  })

  it('disables Previous button on first page', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([], { total: 0, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('prev-page')).toBeDisabled()
  })

  it('enables Next button when more pages exist', () => {
    const customers = Array.from({ length: 20 }, (_, i) =>
      makeCustomer({ id: `c${i}`, name: `Customer ${i}` }),
    )
    mockUseCustomers.mockReturnValue({
      data: makePage(customers, { total: 50, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('next-page')).not.toBeDisabled()
  })

  it('shows page indicator', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([], { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('page-indicator')).toHaveTextContent('Page 1 of 2')
  })

  it('advances to next page when Next is clicked', async () => {
    const user = userEvent.setup()
    const customers = Array.from({ length: 20 }, (_, i) =>
      makeCustomer({ id: `c${i}`, name: `Customer ${i}` }),
    )
    mockUseCustomers.mockReturnValue({
      data: makePage(customers, { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    await user.click(screen.getByTestId('next-page'))

    await waitFor(() => {
      expect(mockUseCustomers).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      )
    })
  })

  it('shows type filter dropdown', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByLabelText('Type:')).toBeInTheDocument()
    expect(screen.getByText('All types')).toBeInTheDocument()
  })

  it('passes type filter to useCustomers when changed', async () => {
    const user = userEvent.setup()
    mockUseCustomers.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    await user.selectOptions(screen.getByLabelText('Type:'), 'retail')

    await waitFor(() => {
      expect(mockUseCustomers).toHaveBeenCalledWith(
        expect.objectContaining({ type: 'retail' }),
      )
    })
  })

  it('resets to page 1 when type filter changes', async () => {
    const user = userEvent.setup()
    const customers = Array.from({ length: 20 }, (_, i) =>
      makeCustomer({ id: `c${i}`, name: `Customer ${i}` }),
    )
    mockUseCustomers.mockReturnValue({
      data: makePage(customers, { total: 40, page: 1, limit: 20 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    // Advance to page 2
    await user.click(screen.getByTestId('next-page'))
    await waitFor(() => {
      expect(mockUseCustomers).toHaveBeenCalledWith(expect.objectContaining({ page: 2 }))
    })

    // Change filter — should reset to page 1
    await user.selectOptions(screen.getByLabelText('Type:'), 'wholesale')

    await waitFor(() => {
      expect(mockUseCustomers).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, type: 'wholesale' }),
      )
    })
  })

  it('shows filtered empty state message when type filter is active', async () => {
    const user = userEvent.setup()
    mockUseCustomers.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    await user.selectOptions(screen.getByLabelText('Type:'), 'hospitality')

    await waitFor(() => {
      expect(screen.getByText('No customers with type "hospitality".')).toBeInTheDocument()
    })
  })

  it('shows result count', () => {
    const customers = [makeCustomer()]
    mockUseCustomers.mockReturnValue({
      data: makePage(customers, { total: 1 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('1–1 of 1')
  })

  it('shows 0 results when total is 0', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([], { total: 0 }),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    expect(screen.getByTestId('result-count')).toHaveTextContent('0 results')
  })

  it('shows type filter options for all customer types', () => {
    mockUseCustomers.mockReturnValue({
      data: makePage([]),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useCustomers>)

    render(<CustomersPage />)

    const typeOptions: CustomerType[] = ['retail', 'wholesale', 'hospitality', 'institutional', 'other']
    for (const type of typeOptions) {
      expect(screen.getByRole('option', { name: new RegExp(type, 'i') })).toBeInTheDocument()
    }
  })
})
