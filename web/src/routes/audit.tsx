import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useAuditLog, useUpdateAuditStatus } from '@/hooks/useAudit'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
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

function AuditPage() {
  const [entityTypeFilter, setEntityTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useAuditLog({
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
        cell: (info) => (
          <span className="text-xs text-slate-500 whitespace-nowrap">
            {new Date(info.getValue()).toLocaleString('en-GB', {
              day: '2-digit', month: 'short', year: 'numeric',
              hour: '2-digit', minute: '2-digit',
            })}
          </span>
        ),
      }),
      columnHelper.accessor('actorId', {
        header: 'Actor',
        cell: (info) => (
          <span className="text-sm text-slate-700 font-mono">{info.getValue().slice(0, 8)}...</span>
        ),
      }),
      columnHelper.accessor('entityType', {
        header: 'Entity',
        cell: (info) => <Badge>{info.getValue()}</Badge>,
      }),
      columnHelper.accessor('eventType', {
        header: 'Action',
        cell: (info) => <span className="text-sm capitalize">{info.getValue().replace('_', ' ')}</span>,
      }),
      columnHelper.accessor('status', {
        header: 'Status',
        cell: (info) => (
          <Badge variant={auditStatusVariant[info.getValue()] ?? 'default'}>
            {info.getValue().replace('_', ' ')}
          </Badge>
        ),
      }),
      columnHelper.display({
        id: 'actions',
        header: 'Review',
        cell: ({ row }) => {
          if (row.original.status !== 'pending') return null
          return (
            <div className="flex gap-1">
              <button
                onClick={() => updateStatus.mutate({ id: row.original.id, status: 'accepted' })}
                className="rounded p-1 text-emerald-600 hover:bg-emerald-50"
                title="Accept"
              >
                <CheckCircle size={16} />
              </button>
              <button
                onClick={() => updateStatus.mutate({ id: row.original.id, status: 'false_positive' })}
                className="rounded p-1 text-slate-400 hover:bg-slate-100"
                title="False positive"
              >
                <XCircle size={16} />
              </button>
            </div>
          )
        },
      }),
    ],
    [updateStatus],
  )

  if (isLoading) return <Spinner />

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
