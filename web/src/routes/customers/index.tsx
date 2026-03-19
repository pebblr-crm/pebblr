import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from '../__root'
import { DataTable } from '../../components/DataTable'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useCustomers } from '../../services/customers'
import type { Customer, CustomerType } from '../../types/customer'

const PAGE_SIZE = 20

const TYPE_OPTIONS: { value: CustomerType | ''; label: string }[] = [
  { value: '', label: 'All types' },
  { value: 'retail', label: 'Retail' },
  { value: 'wholesale', label: 'Wholesale' },
  { value: 'hospitality', label: 'Hospitality' },
  { value: 'institutional', label: 'Institutional' },
  { value: 'other', label: 'Other' },
]

const columnHelper = createColumnHelper<Customer>()

const columns = [
  columnHelper.accessor('name', { header: 'Name' }),
  columnHelper.accessor('type', {
    header: 'Type',
    cell: (info) => <TypeBadge type={info.getValue()} />,
  }),
  columnHelper.accessor('email', { header: 'Email' }),
  columnHelper.accessor('phone', { header: 'Phone' }),
  columnHelper.accessor('address', {
    header: 'City',
    cell: (info) => info.getValue()?.city ?? '—',
  }),
  columnHelper.accessor('createdAt', {
    header: 'Created',
    cell: (info) => formatDate(info.getValue()),
  }),
]

const paginationButtonStyle: React.CSSProperties = {
  padding: '4px 12px',
  fontSize: '13px',
  border: '1px solid var(--color-border, #e5e5e5)',
  borderRadius: '4px',
  backgroundColor: 'var(--color-surface, #fff)',
}

function formatDate(iso: string): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch {
    return iso
  }
}

const TYPE_COLORS: Record<CustomerType, string> = {
  retail: '#3b82f6',
  wholesale: '#8b5cf6',
  hospitality: '#f59e0b',
  institutional: '#10b981',
  other: '#6b7280',
}

function TypeBadge({ type }: { type: CustomerType }) {
  const color = TYPE_COLORS[type] ?? '#6b7280'
  return (
    <span
      style={{
        display: 'inline-block',
        padding: '2px 8px',
        borderRadius: '9999px',
        fontSize: '11px',
        fontWeight: 500,
        color: '#fff',
        backgroundColor: color,
        textTransform: 'capitalize',
        whiteSpace: 'nowrap',
      }}
    >
      {type}
    </span>
  )
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/customers',
  component: CustomersPage,
})

export function CustomersPage() {
  const [page, setPage] = useState(1)
  const [typeFilter, setTypeFilter] = useState<CustomerType | ''>('')

  const { data, isLoading, isError, error } = useCustomers({
    page,
    limit: PAGE_SIZE,
    ...(typeFilter ? { type: typeFilter } : {}),
  })

  const customers = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  return (
    <div>
      <div
        className="page-header"
        style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
      >
        <h1 className="page-title">Customers</h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <label
            htmlFor="type-filter"
            style={{ fontSize: '13px', color: 'var(--color-text-secondary, #666)' }}
          >
            Type:
          </label>
          <select
            id="type-filter"
            value={typeFilter}
            onChange={(e) => {
              setTypeFilter(e.target.value as CustomerType | '')
              setPage(1)
            }}
            style={{
              fontSize: '13px',
              padding: '4px 8px',
              border: '1px solid var(--color-border, #e5e5e5)',
              borderRadius: '4px',
              backgroundColor: 'var(--color-surface, #fff)',
              color: 'var(--color-text-primary, #181818)',
              cursor: 'pointer',
            }}
          >
            {TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="page-body">
        {isLoading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
            <LoadingSpinner size="lg" label="Loading customers..." />
          </div>
        ) : isError ? (
          <div
            data-testid="error-state"
            style={{
              padding: '32px 12px',
              textAlign: 'center',
              color: 'var(--color-text-danger, #ef4444)',
            }}
          >
            {error instanceof Error ? error.message : 'Failed to load customers. Please try again.'}
          </div>
        ) : (
          <>
            <DataTable columns={columns} data={customers} />

            {customers.length === 0 && (
              <div
                data-testid="empty-state"
                style={{
                  padding: '32px 12px',
                  textAlign: 'center',
                  color: 'var(--color-text-tertiary, #999)',
                }}
              >
                {typeFilter ? `No customers with type "${typeFilter}".` : 'No customers found.'}
              </div>
            )}

            <div
              data-testid="pagination"
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '12px',
                borderTop: '1px solid var(--color-border, #e5e5e5)',
                marginTop: '8px',
                fontSize: '13px',
                color: 'var(--color-text-secondary, #666)',
              }}
            >
              <span data-testid="result-count">
                {total > 0
                  ? `${(page - 1) * PAGE_SIZE + 1}–${Math.min(page * PAGE_SIZE, total)} of ${total}`
                  : '0 results'}
              </span>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <button
                  data-testid="prev-page"
                  onClick={() => setPage((p) => p - 1)}
                  disabled={!hasPrev}
                  style={{
                    ...paginationButtonStyle,
                    cursor: hasPrev ? 'pointer' : 'default',
                    opacity: hasPrev ? 1 : 0.4,
                  }}
                >
                  Previous
                </button>
                <span data-testid="page-indicator">
                  Page {page} of {totalPages}
                </span>
                <button
                  data-testid="next-page"
                  onClick={() => setPage((p) => p + 1)}
                  disabled={!hasNext}
                  style={{
                    ...paginationButtonStyle,
                    cursor: hasNext ? 'pointer' : 'default',
                    opacity: hasNext ? 1 : 0.4,
                  }}
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
