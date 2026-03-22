import { useState } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Search, CheckCircle2, Clock, CalendarDays } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useLeads } from '../../services/leads'
import { useCalendarEvents } from '../../services/calendar'
import { useCurrentUser } from '../../services/me'
import type { Lead, LeadStatus } from '../../types/lead'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/my-leads',
  component: MyLeadsPage,
})

const PAGE_SIZE = 20

const STATUS_BADGE_STYLES: Record<LeadStatus, string> = {
  new: 'bg-primary-fixed text-primary',
  assigned: 'bg-secondary-container text-on-secondary-fixed-variant',
  in_progress: 'bg-secondary-container text-on-secondary-fixed-variant',
  visited: 'bg-tertiary-container text-tertiary-fixed',
  closed_won: 'bg-tertiary-container text-tertiary-fixed',
  closed_lost: 'bg-slate-200 text-slate-600',
}

const STATUS_LABELS: Record<LeadStatus, string> = {
  new: 'New',
  assigned: 'Scheduled',
  in_progress: 'In Progress',
  visited: 'Visited',
  closed_won: 'Done',
  closed_lost: 'Lost',
}

function priorityDots(lead: Lead) {
  const val = lead.valueCents ?? 0
  const level = val >= 500000 ? 3 : val >= 100000 ? 2 : 1
  return (
    <div className="flex gap-1">
      {[1, 2, 3].map((i) => (
        <span
          key={i}
          className={`w-1.5 h-1.5 rounded-full ${i <= level ? 'bg-error' : 'bg-slate-200'}`}
        />
      ))}
    </div>
  )
}

function formatCurrency(cents: number): string {
  if (!cents) return '$0'
  return `$${(cents / 100).toLocaleString(undefined, { maximumFractionDigits: 0 })}`
}

function formatDate(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const initialsColors = [
  'bg-blue-100 text-blue-600',
  'bg-emerald-100 text-emerald-600',
  'bg-slate-100 text-slate-600',
  'bg-amber-100 text-amber-600',
  'bg-violet-100 text-violet-600',
  'bg-rose-100 text-rose-600',
]

function colorForInitials(initials: string): string {
  let hash = 0
  for (let i = 0; i < initials.length; i++) {
    hash = initials.charCodeAt(i) + ((hash << 5) - hash)
  }
  return initialsColors[Math.abs(hash) % initialsColors.length]
}

export function MyLeadsPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<LeadStatus | ''>('')
  const { data: me, isLoading: meLoading } = useCurrentUser()

  const { data, isLoading: leadsLoading } = useLeads({
    page,
    limit: PAGE_SIZE,
    ...(me?.id ? { assigneeId: me.id } : {}),
    ...(statusFilter ? { status: statusFilter } : {}),
  })

  const isLoading = meLoading || leadsLoading

  const now = new Date()
  const { data: upcomingEvents = [] } = useCalendarEvents({
    year: now.getFullYear(),
    month: now.getMonth() + 1,
  })

  const leads = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const scheduledCount = leads.filter(
    (l) => l.status === 'assigned' || l.status === 'in_progress',
  ).length
  const completedCount = leads.filter(
    (l) => l.status === 'closed_won' || l.status === 'visited',
  ).length

  // Next upcoming events (sorted, future only)
  const nextEvents = upcomingEvents
    .filter((e) => new Date(e.startTime) > now)
    .sort((a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime())
    .slice(0, 2)

  const nextEventDate = nextEvents[0]
    ? new Date(nextEvents[0].startTime).toLocaleDateString(undefined, {
        weekday: 'long',
        month: 'short',
        day: 'numeric',
      })
    : ''

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 space-y-8 max-w-7xl mx-auto w-full"
    >
      {/* Header & Metrics */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1">
          <h1 className="font-headline text-3xl font-extrabold text-primary tracking-tight">
            My Leads
          </h1>
          <p className="text-on-surface-variant mt-1 text-sm">Your assigned pipeline</p>
        </div>
        <div className="lg:col-span-3 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-surface-container-lowest p-5 rounded-xl border border-slate-100 flex items-center gap-4">
            <div className="w-12 h-12 bg-primary-fixed rounded-full flex items-center justify-center text-primary">
              <Search className="w-5 h-5" />
            </div>
            <div>
              <p className="text-on-surface-variant text-xs font-semibold uppercase tracking-wider">
                Total Leads
              </p>
              <p className="font-headline text-2xl font-bold text-primary">{total}</p>
            </div>
          </div>
          <div className="bg-surface-container-lowest p-5 rounded-xl border border-slate-100 flex items-center gap-4">
            <div className="w-12 h-12 bg-secondary-container rounded-full flex items-center justify-center text-secondary">
              <Clock className="w-5 h-5" />
            </div>
            <div>
              <p className="text-on-surface-variant text-xs font-semibold uppercase tracking-wider">
                Scheduled
              </p>
              <p className="font-headline text-2xl font-bold text-primary">{scheduledCount}</p>
            </div>
          </div>
          <div className="bg-surface-container-lowest p-5 rounded-xl border border-slate-100 flex items-center gap-4">
            <div className="w-12 h-12 bg-tertiary-container/10 rounded-full flex items-center justify-center text-tertiary-container">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div>
              <p className="text-on-surface-variant text-xs font-semibold uppercase tracking-wider">
                Completed
              </p>
              <p className="font-headline text-2xl font-bold text-primary">{completedCount}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Bento Layout */}
      <div className="grid grid-cols-1 xl:grid-cols-12 gap-8">
        {/* Main Table */}
        <div className="xl:col-span-8 space-y-6">
          <div className="bg-surface-container-lowest rounded-xl p-1 shadow-sm border border-slate-50">
            <div className="p-6 flex justify-between items-center">
              <h2 className="font-headline text-lg font-bold text-primary">Recent Activity</h2>
              <div className="flex gap-2">
                <select
                  value={statusFilter}
                  onChange={(e) => {
                    setStatusFilter(e.target.value as LeadStatus | '')
                    setPage(1)
                  }}
                  className="px-3 py-1.5 bg-surface-container-low rounded-lg text-xs font-semibold text-secondary hover:bg-surface-container-high transition-colors border-none cursor-pointer"
                >
                  <option value="">All Statuses</option>
                  <option value="new">New</option>
                  <option value="assigned">Assigned</option>
                  <option value="in_progress">In Progress</option>
                  <option value="visited">Visited</option>
                  <option value="closed_won">Closed Won</option>
                  <option value="closed_lost">Closed Lost</option>
                </select>
              </div>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <LoadingSpinner size="lg" label="Loading leads..." />
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead className="bg-surface-container-low/50 text-on-surface-variant text-[10px] uppercase tracking-widest font-bold">
                    <tr>
                      <th className="px-6 py-4">Lead Name</th>
                      <th className="px-6 py-4">Company</th>
                      <th className="px-6 py-4">Current Status</th>
                      <th className="px-6 py-4">Priority</th>
                      <th className="px-6 py-4 text-right">Value</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-50">
                    {leads.map((lead) => (
                      <tr
                        key={lead.id}
                        className="group hover:bg-surface-container-low transition-colors"
                      >
                        <td className="px-6 py-5">
                          <Link
                            to="/leads/$leadId"
                            params={{ leadId: lead.id }}
                            className="flex items-center gap-3 no-underline"
                          >
                            <div
                              className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs ${colorForInitials(lead.initials || lead.title)}`}
                            >
                              {lead.initials || lead.title.slice(0, 2).toUpperCase()}
                            </div>
                            <div>
                              <p className="text-sm font-semibold text-primary">{lead.title}</p>
                              <p className="text-[10px] text-slate-400">
                                {lead.customerType} &middot; {formatDate(lead.createdAt)}
                              </p>
                            </div>
                          </Link>
                        </td>
                        <td className="px-6 py-5">
                          <span className="text-xs text-secondary font-medium">
                            {lead.company || '—'}
                          </span>
                        </td>
                        <td className="px-6 py-5">
                          <span
                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[10px] font-bold ${STATUS_BADGE_STYLES[lead.status]}`}
                          >
                            {STATUS_LABELS[lead.status]}
                          </span>
                        </td>
                        <td className="px-6 py-5">{priorityDots(lead)}</td>
                        <td className="px-6 py-5 text-right">
                          <span className="text-xs font-semibold text-primary">
                            {formatCurrency(lead.valueCents)}
                          </span>
                        </td>
                      </tr>
                    ))}
                    {leads.length === 0 && (
                      <tr>
                        <td colSpan={5} className="px-6 py-12 text-center text-on-surface-variant">
                          {statusFilter
                            ? `No leads with status "${statusFilter}".`
                            : 'No leads assigned to you yet.'}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            )}

            {/* Pagination */}
            {total > PAGE_SIZE && (
              <div className="p-4 border-t border-slate-50 flex items-center justify-between text-xs text-on-surface-variant">
                <span>
                  {(page - 1) * PAGE_SIZE + 1}&ndash;{Math.min(page * PAGE_SIZE, total)} of {total}
                </span>
                <div className="flex gap-2">
                  <button
                    onClick={() => setPage((p) => p - 1)}
                    disabled={page <= 1}
                    className="px-3 py-1.5 bg-surface-container-low rounded-lg font-semibold text-secondary hover:bg-surface-container-high transition-colors disabled:opacity-40"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setPage((p) => p + 1)}
                    disabled={page >= totalPages}
                    className="px-3 py-1.5 bg-surface-container-low rounded-lg font-semibold text-secondary hover:bg-surface-container-high transition-colors disabled:opacity-40"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}

            {total > 0 && total <= PAGE_SIZE && (
              <div className="p-4 border-t border-slate-50 flex justify-center">
                <Link
                  to="/leads"
                  className="text-xs font-bold text-primary hover:underline no-underline"
                >
                  View All Leads
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Widgets Column */}
        <div className="xl:col-span-4 space-y-6">
          {/* Next Up Calendar Widget */}
          <div className="bg-primary text-white rounded-xl p-6 shadow-xl relative overflow-hidden">
            <div className="absolute -right-10 -top-10 w-40 h-40 bg-white/5 rounded-full blur-3xl" />
            <div className="absolute -left-10 -bottom-10 w-40 h-40 bg-primary-container/20 rounded-full blur-2xl" />
            <div className="relative z-10">
              <div className="flex justify-between items-start mb-8">
                <div>
                  <h3 className="font-headline text-lg font-bold">Next Up</h3>
                  <p className="text-primary-fixed text-xs">{nextEventDate || 'No upcoming events'}</p>
                </div>
                <CalendarDays className="w-5 h-5 text-primary-fixed" />
              </div>
              <div className="space-y-4">
                {nextEvents.length > 0 ? (
                  nextEvents.map((event, i) => (
                    <div
                      key={event.id}
                      className={`${i === 0 ? 'bg-white/10 border-white/10' : 'bg-white/5 border-white/5'} p-4 rounded-xl border`}
                    >
                      <div className="flex items-center gap-3 mb-2">
                        <Clock className="w-3.5 h-3.5" />
                        <span className="text-xs font-semibold tracking-wider opacity-80 uppercase">
                          {formatTime(event.startTime)}
                        </span>
                      </div>
                      <p className="text-sm font-bold">{event.title}</p>
                      {event.client && (
                        <p className="text-[11px] opacity-70">Client: {event.client}</p>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="bg-white/5 p-4 rounded-xl border border-white/5">
                    <p className="text-sm opacity-70">No upcoming events scheduled.</p>
                  </div>
                )}
                <Link
                  to="/calendar"
                  className="block w-full py-2 text-xs font-bold text-center border border-white/20 rounded-lg hover:bg-white hover:text-primary transition-all no-underline text-white"
                >
                  Open Full Calendar
                </Link>
              </div>
            </div>
          </div>

          {/* Pipeline Conversion Widget */}
          <div className="bg-surface-container-low rounded-xl p-6 border border-slate-100">
            <h3 className="font-headline text-sm font-bold text-primary mb-6">
              Pipeline Conversion
            </h3>
            <div className="relative pt-1">
              <div className="flex mb-2 items-center justify-between">
                <span className="text-xs font-semibold inline-block py-1 px-2 uppercase rounded-full text-tertiary-container bg-tertiary-fixed">
                  {completedCount > 0 ? 'Active' : 'Starting'}
                </span>
                <span className="text-xs font-bold inline-block text-primary">
                  {total > 0 ? Math.round((completedCount / total) * 100) : 0}%
                </span>
              </div>
              <div className="overflow-hidden h-2 mb-4 text-xs flex rounded-full bg-slate-200">
                <div
                  className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center primary-gradient"
                  style={{ width: `${total > 0 ? (completedCount / total) * 100 : 0}%` }}
                />
              </div>
            </div>
            <p className="text-[11px] text-on-surface-variant leading-relaxed">
              <span className="font-bold">{completedCount}</span> of{' '}
              <span className="font-bold">{total}</span> leads have been completed or visited.
            </p>
          </div>

          {/* Value Summary */}
          <div className="bg-surface-container-lowest rounded-xl p-6 border border-slate-100">
            <h3 className="font-headline text-sm font-bold text-primary mb-4">Pipeline Value</h3>
            <p className="font-headline text-2xl font-extrabold text-primary">
              {formatCurrency(leads.reduce((sum, l) => sum + (l.valueCents || 0), 0))}
            </p>
            <p className="text-[11px] text-on-surface-variant mt-1">
              Across {leads.length} leads on this page
            </p>
          </div>
        </div>
      </div>
    </motion.div>
  )
}
