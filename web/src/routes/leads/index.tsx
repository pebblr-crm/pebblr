import { useState } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { SlidersHorizontal, Download, MapPin, MoreVertical, ChevronLeft, ChevronRight } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useLeads } from '../../services/leads'
import type { LeadStatus } from '../../types/lead'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/leads',
  component: LeadsPage,
})

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

const STATUS_BADGE_STYLES: Record<LeadStatus, string> = {
  new: 'bg-primary-fixed text-primary',
  assigned: 'bg-secondary-container text-on-secondary-fixed-variant',
  in_progress: 'bg-amber-100 text-amber-700',
  visited: 'bg-violet-100 text-violet-700',
  closed_won: 'bg-tertiary-container/10 text-tertiary-container',
  closed_lost: 'bg-slate-200 text-slate-600',
}

const INITIALS_COLORS = [
  'bg-primary/5 text-primary',
  'bg-amber-50 text-amber-700',
  'bg-emerald-50 text-tertiary-container',
  'bg-slate-100 text-slate-600',
  'bg-violet-50 text-violet-700',
  'bg-rose-50 text-rose-600',
]

function colorForInitials(s: string): string {
  let hash = 0
  for (let i = 0; i < s.length; i++) hash = s.charCodeAt(i) + ((hash << 5) - hash)
  return INITIALS_COLORS[Math.abs(hash) % INITIALS_COLORS.length]
}

function StatusBadge({ status }: { status: LeadStatus }) {
  const style = STATUS_BADGE_STYLES[status] ?? 'bg-slate-200 text-slate-600'
  return (
    <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${style}`}>
      {status.replace(/_/g, ' ')}
    </span>
  )
}

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

  const newCount = leads.filter((l) => l.status === 'new').length
  const scheduledCount = leads.filter((l) => l.status === 'assigned' || l.status === 'in_progress').length
  const wonCount = leads.filter((l) => l.status === 'closed_won').length

  // Visible page numbers for numbered pagination
  const visiblePages: number[] = []
  const maxVisible = 3
  let startPage = Math.max(1, page - Math.floor(maxVisible / 2))
  const endPage = Math.min(totalPages, startPage + maxVisible - 1)
  startPage = Math.max(1, endPage - maxVisible + 1)
  for (let i = startPage; i <= endPage; i++) visiblePages.push(i)

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 max-w-7xl mx-auto w-full space-y-8"
    >
      {/* Page Header */}
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-extrabold tracking-tight text-primary leading-tight font-headline">
            Lead Directory
          </h1>
          <p className="text-on-surface-variant mt-1 font-medium">
            Managing {total.toLocaleString()} leads across all regions.
          </p>
        </div>
        <div className="flex gap-3 items-center">
          {/* Hidden label for test compatibility */}
          <label htmlFor="status-filter" className="sr-only">Status:</label>
          <select
            id="status-filter"
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value as LeadStatus | '')
              setPage(1)
            }}
            className="px-4 py-2.5 bg-surface-container-high text-on-surface rounded-xl text-sm font-semibold border-none cursor-pointer hover:bg-surface-dim transition-colors"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
          <button className="px-5 py-2.5 bg-primary text-white rounded-xl text-sm font-semibold flex items-center gap-2 shadow-sm hover:opacity-90 transition-opacity">
            <Download className="w-4 h-4" />
            Export CSV
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <div className="flex justify-between items-start">
            <div className="p-2 bg-blue-50 text-primary rounded-lg">
              <SlidersHorizontal className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm font-medium text-slate-500">Total Leads</div>
            <div className="text-3xl font-extrabold font-headline mt-1">{total}</div>
          </div>
        </div>
        <div className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <div className="flex justify-between items-start">
            <div className="p-2 bg-primary-fixed text-primary rounded-lg">
              <SlidersHorizontal className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm font-medium text-slate-500">New Leads</div>
            <div className="text-3xl font-extrabold font-headline mt-1">{newCount}</div>
          </div>
        </div>
        <div className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <div className="flex justify-between items-start">
            <div className="p-2 bg-amber-50 text-amber-600 rounded-lg">
              <SlidersHorizontal className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm font-medium text-slate-500">Scheduled</div>
            <div className="text-3xl font-extrabold font-headline mt-1">{scheduledCount}</div>
          </div>
        </div>
        <div className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <div className="flex justify-between items-start">
            <div className="p-2 bg-emerald-50 text-tertiary-container rounded-lg">
              <SlidersHorizontal className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm font-medium text-slate-500">Closed Wins</div>
            <div className="text-3xl font-extrabold font-headline mt-1">{wonCount}</div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label="Loading leads..." />
        </div>
      ) : isError ? (
        <div
          data-testid="error-state"
          className="p-8 text-center text-error"
        >
          {error instanceof Error ? error.message : 'Failed to load leads. Please try again.'}
        </div>
      ) : (
        <>
          {/* Data Table */}
          <div className="bg-surface-container-low rounded-xl p-1 overflow-hidden">
            <div className="bg-surface-container-lowest rounded-lg shadow-[0px_24px_48px_rgba(25,28,30,0.06)] overflow-hidden">
              <table className="w-full text-left border-separate border-spacing-0">
                <thead>
                  <tr className="bg-surface-container-low/50">
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      Title
                    </th>
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      Location
                    </th>
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      Status
                    </th>
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      Customer Type
                    </th>
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400 text-right">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {leads.map((lead) => (
                    <tr key={lead.id} className="hover:bg-slate-50/50 transition-colors">
                      <td className="px-6 py-5">
                        <Link
                          to="/leads/$leadId"
                          params={{ leadId: lead.id }}
                          className="flex items-center gap-3 no-underline"
                        >
                          <div
                            className={`w-10 h-10 rounded-lg flex items-center justify-center font-bold text-sm ${colorForInitials(lead.initials || lead.company || lead.title)}`}
                          >
                            {lead.initials || (lead.company || lead.title).slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="text-sm font-bold text-primary">{lead.title}</div>
                            <div className="text-xs text-slate-400">{lead.industry || lead.customerType}</div>
                          </div>
                        </Link>
                      </td>
                      <td className="px-6 py-5">
                        {lead.location ? (
                          <div className="flex items-center gap-2">
                            <MapPin className="w-4 h-4 text-slate-300" />
                            <span className="text-sm text-slate-600">{lead.location}</span>
                          </div>
                        ) : (
                          <span className="text-sm text-slate-300">—</span>
                        )}
                      </td>
                      <td className="px-6 py-5">
                        <StatusBadge status={lead.status} />
                      </td>
                      <td className="px-6 py-5">
                        <span className="text-sm font-medium text-slate-600 capitalize">
                          {lead.customerType}
                        </span>
                      </td>
                      <td className="px-6 py-5 text-right">
                        <button className="p-2 text-slate-300 hover:text-primary transition-colors">
                          <MoreVertical className="w-5 h-5" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {leads.length === 0 && (
                <div
                  data-testid="empty-state"
                  className="px-6 py-12 text-center text-on-surface-variant"
                >
                  {statusFilter ? `No leads with status "${statusFilter}".` : 'No leads found.'}
                </div>
              )}

              {/* Pagination */}
              <div
                data-testid="pagination"
                className="px-6 py-4 bg-surface-container-lowest border-t border-slate-50 flex items-center justify-between"
              >
                <span data-testid="result-count" className="text-xs font-medium text-slate-400">
                  {total > 0
                    ? `${(page - 1) * PAGE_SIZE + 1}–${Math.min(page * PAGE_SIZE, total)} of ${total}`
                    : '0 results'}
                </span>
                <div className="flex gap-2">
                  <button
                    data-testid="prev-page"
                    onClick={() => setPage((p) => p - 1)}
                    disabled={!hasPrev}
                    className="w-8 h-8 flex items-center justify-center border border-slate-100 rounded-lg text-slate-400 hover:bg-slate-50 transition-colors disabled:opacity-40"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </button>
                  {visiblePages.map((p) => (
                    <button
                      key={p}
                      onClick={() => setPage(p)}
                      className={`w-8 h-8 flex items-center justify-center rounded-lg text-xs font-bold transition-colors ${
                        p === page
                          ? 'bg-primary text-white'
                          : 'border border-slate-100 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      {p}
                    </button>
                  ))}
                  <button
                    data-testid="next-page"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={!hasNext}
                    className="w-8 h-8 flex items-center justify-center border border-slate-100 rounded-lg text-slate-400 hover:bg-slate-50 transition-colors disabled:opacity-40"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
                <span data-testid="page-indicator" className="text-xs font-medium text-slate-400">
                  Page {page} of {totalPages}
                </span>
              </div>
            </div>
          </div>
        </>
      )}
    </motion.div>
  )
}
