import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from '../__root'
import { DataTable } from '../../components/DataTable'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useLeads } from '../../services/leads'
import type { Lead, LeadStatus } from '../../types/lead'

const PAGE_SIZE = 20

const STATUS_OPTIONS: { value: LeadStatus | ''; label: string }[] = [
  { value: '', label: 'All statuses' },
  { value: 'new', label: 'New' },
  { value: 'assigned', label: 'Assigned' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'visited', label: 'Visited' },
  { value: 'closed_won', label: 'Closed Won' },
  { value: 'closed_lost', label: 'Closed Lost' },
]

const columnHelper = createColumnHelper<Lead>()

const columns = [
  columnHelper.accessor('title', { header: 'Title' }),
  columnHelper.accessor('status', {
    header: 'Status',
    cell: (info) => <StatusBadge status={info.getValue()} />,
  }),
  columnHelper.accessor('customerType', { header: 'Customer Type' }),
  columnHelper.accessor('assigneeId', { header: 'Assignee' }),
  columnHelper.accessor('teamId', { header: 'Team' }),
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

const STATUS_COLORS: Record<LeadStatus, string> = {
  new: '#6b7280',
  assigned: '#3b82f6',
  in_progress: '#f59e0b',
  visited: '#8b5cf6',
  closed_won: '#10b981',
  closed_lost: '#ef4444',
}

function StatusBadge({ status }: { status: LeadStatus }) {
  const color = STATUS_COLORS[status] ?? '#6b7280'
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
      {status.replace(/_/g, ' ')}
    </span>
  )
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/leads',
  component: LeadsPage,
})

export function LeadsPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<LeadStatus | ''>('')

  const { data, isLoading, isError, error } = useLeads({
    page,
    limit: PAGE_SIZE,
    ...(statusFilter ? { status: statusFilter } : {}),
  })

  const leads = data?.items ?? []
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
        <h1 className="page-title">Leads</h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <label
            htmlFor="status-filter"
            style={{ fontSize: '13px', color: 'var(--color-text-secondary, #666)' }}
          >
            Status:
          </label>
          <select
            id="status-filter"
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value as LeadStatus | '')
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
            {STATUS_OPTIONS.map((opt) => (
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
            <LoadingSpinner size="lg" label="Loading leads..." />
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
            {error instanceof Error ? error.message : 'Failed to load leads. Please try again.'}
          </div>
        ) : (
          <>
            <DataTable columns={columns} data={leads} />

            {leads.length === 0 && (
              <div
                data-testid="empty-state"
                style={{
                  padding: '32px 12px',
                  textAlign: 'center',
                  color: 'var(--color-text-tertiary, #999)',
                }}
              >
                {statusFilter ? `No leads with status "${statusFilter}".` : 'No leads found.'}
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
