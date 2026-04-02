import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useAuditLog, useUpdateAuditStatus } from '@/hooks/useAudit'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { QueryError } from '@/components/ui/QueryError'
import { FileText, CheckCircle, XCircle } from 'lucide-react'
import { Select } from '@/components/ui/Select'
import type { AuditEntry, AuditStatus } from '@/types/audit'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/audit',
  component: AuditPage,
})

const auditStatusVariant: Record<string, 'warning' | 'success' | 'default'> = {
  pending: 'warning',
  accepted: 'success',
  false_positive: 'default',
}

const columnHelper = createColumnHelper<AuditEntry>()

function TimestampCell({ getValue }: { getValue: () => string }) {
  return (
    <span className="text-xs text-slate-500 whitespace-nowrap">
      {new Date(getValue()).toLocaleString('en-GB', {
        day: '2-digit', month: 'short', year: 'numeric',
        hour: '2-digit', minute: '2-digit',
      })}
    </span>
  )
}

function ActorCell({ getValue }: { getValue: () => string }) {
  return (
    <span className="text-sm text-slate-700 font-mono">{getValue().slice(0, 8)}...</span>
  )
}

function EntityTypeCell({ getValue }: { getValue: () => string }) {
  return <Badge>{getValue()}</Badge>
}

function EventTypeCell({ getValue }: { getValue: () => string }) {
  return <span className="text-sm capitalize">{getValue().replace('_', ' ')}</span>
}

function StatusCell({ getValue }: { getValue: () => AuditStatus }) {
  return (
    <Badge variant={auditStatusVariant[getValue()] ?? 'default'}>
      {getValue().replace('_', ' ')}
    </Badge>
  )
}

function ReviewActionsCell({ entry, onAccept, onFalsePositive }: {
  entry: AuditEntry
  onAccept: (id: string) => void
  onFalsePositive: (id: string) => void
}) {
  if (entry.status !== 'pending') return null
  return (
    <div className="flex gap-1">
      <button
        onClick={() => onAccept(entry.id)}
        className="rounded p-1 text-emerald-600 hover:bg-emerald-50"
        title="Accept"
        aria-label="Accept audit entry"
      >
        <CheckCircle size={16} />
      </button>
      <button
        onClick={() => onFalsePositive(entry.id)}
        className="rounded p-1 text-slate-400 hover:bg-slate-100"
        title="False positive"
        aria-label="Mark as false positive"
      >
        <XCircle size={16} />
      </button>
    </div>
  )
}

function AuditPage() {
  const [entityTypeFilter, setEntityTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, isError, refetch } = useAuditLog({
    entityType: entityTypeFilter || undefined,
    status: (statusFilter as AuditStatus) || undefined,
    page,
    limit: 50,
  })
  const updateStatus = useUpdateAuditStatus()

  const entries = useMemo(() => data?.items ?? [], [data])
  const pendingCount = useMemo(() => entries.filter((e) => e.status === 'pending').length, [entries])

  const columns = useMemo(
    () => [
      columnHelper.accessor('createdAt', {
        header: 'Timestamp',
        cell: (info) => <TimestampCell getValue={info.getValue} />,
      }),
      columnHelper.accessor('actorId', {
        header: 'Actor',
        cell: (info) => <ActorCell getValue={info.getValue} />,
      }),
      columnHelper.accessor('entityType', {
        header: 'Entity',
        cell: (info) => <EntityTypeCell getValue={info.getValue} />,
      }),
      columnHelper.accessor('eventType', {
        header: 'Action',
        cell: (info) => <EventTypeCell getValue={info.getValue} />,
      }),
      columnHelper.accessor('status', {
        header: 'Status',
        cell: (info) => <StatusCell getValue={info.getValue} />,
      }),
      columnHelper.display({
        id: 'actions',
        header: 'Review',
        cell: ({ row }) => (
          <ReviewActionsCell
            entry={row.original}
            onAccept={(id) => updateStatus.mutate({ id, status: 'accepted' })}
            onFalsePositive={(id) => updateStatus.mutate({ id, status: 'false_positive' })}
          />
        ),
      }),
    ],
    [updateStatus],
  )

  if (isLoading) return <Spinner />
  if (isError) return <QueryError message="Failed to load audit logs" onRetry={() => { refetch() }} />

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="mb-4 flex flex-col gap-3 sm:mb-6 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 sm:text-2xl">Audit Logs</h1>
          <p className="mt-1 text-sm text-slate-500">Immutable change history and review workflow.</p>
        </div>
        <div className="flex items-center gap-3">
          {pendingCount > 0 && (
            <Badge variant="warning">
              <FileText size={12} className="mr-1" />
              {pendingCount} pending review
            </Badge>
          )}
          <Button variant="secondary" size="sm" disabled>
            Export Logs
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <Select
          value={entityTypeFilter}
          onChange={(e) => { setEntityTypeFilter(e.target.value); setPage(1) }}
          className="w-auto"
        >
          <option value="">All entities</option>
          <option value="activity">Activity</option>
          <option value="target">Target</option>
          <option value="user">User</option>
        </Select>
        <Select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }}
          className="w-auto"
        >
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="accepted">Accepted</option>
          <option value="false_positive">False Positive</option>
        </Select>
        <span className="text-sm text-slate-500">{data?.total ?? 0} entries</span>
      </div>

      {/* Table */}
      <DataTable data={entries} columns={columns} pageSize={50} />

      {/* Pagination */}
      {data && data.total > 50 && (
        <div className="mt-4 flex items-center justify-center gap-2">
          <Button variant="ghost" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1}>
            Previous
          </Button>
          <span className="text-sm text-slate-500">Page {page}</span>
          <Button variant="ghost" size="sm" onClick={() => setPage((p) => p + 1)} disabled={entries.length < 50}>
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
