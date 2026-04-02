import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useActivities } from '@/hooks/useActivities'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { CreateActivityModal } from '@/components/activities/CreateActivityModal'
import { useRecoveryBalance } from '@/hooks/useDashboard'
import { useConfig } from '@/hooks/useConfig'
import { useUsers } from '@/hooks/useUsers'
import { useAuth } from '@/auth/context'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { statusVariant, priorityStyle } from '@/lib/styles'
import { str } from '@/lib/helpers'
import { Plus, Clock, AlertTriangle, Stethoscope, Building2, Briefcase } from 'lucide-react'
import type { Activity } from '@/types/activity'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities',
  component: ActivitiesPage,
})


function getWeekKey(dateStr: string): string {
  const d = new Date(dateStr)
  const day = d.getDay()
  const monday = new Date(d)
  monday.setDate(d.getDate() - (day === 0 ? 6 : day - 1))
  return monday.toISOString().slice(0, 10)
}

function groupByWeekThenDay(activities: Activity[]): Map<string, Map<string, Activity[]>> {
  const weeks = new Map<string, Map<string, Activity[]>>()
  for (const a of activities) {
    const dateStr = a.dueDate.slice(0, 10)
    const weekKey = getWeekKey(dateStr)
    let dayMap = weeks.get(weekKey)
    if (!dayMap) { dayMap = new Map(); weeks.set(weekKey, dayMap) }
    const list = dayMap.get(dateStr) ?? []
    list.push(a)
    dayMap.set(dateStr, list)
  }
  return weeks
}

function formatWeekLabel(mondayStr: string): string {
  const mon = new Date(mondayStr)
  const fri = new Date(mon)
  fri.setDate(mon.getDate() + 4)
  return `${mon.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} – ${fri.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}`
}

function getDayLabel(dateStr: string): string {
  const d = new Date(dateStr)
  const today = new Date().toISOString().slice(0, 10)
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  if (dateStr === today) return 'Today'
  if (dateStr === yesterday.toISOString().slice(0, 10)) return 'Yesterday'
  return d.toLocaleDateString('en-GB', { weekday: 'short', month: 'short', day: 'numeric' })
}

const activityIcon: Record<string, typeof Stethoscope> = {
  visit: Stethoscope,
  administrative: Briefcase,
  training: Briefcase,
  team_meeting: Briefcase,
  business_travel: Briefcase,
}

function getTargetPriority(a: Activity): string {
  return ((a.targetSummary?.fields?.potential as string) ?? '').toLowerCase()
}

// --- Main Page ---

const PAGE_SIZE = 20
const selectCls = 'w-full text-sm border border-slate-300 rounded-md py-2 px-3 bg-white focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none'

function ActivitiesPage() {
  const { data: config } = useConfig()
  const { role } = useAuth()
  const [typeFilter, setTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [repFilter, setRepFilter] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [detailId, setDetailId] = useState<string | null>(null)
  const [mobileCount, setMobileCount] = useState(PAGE_SIZE)

  const canFilterByRep = role === 'admin' || role === 'manager'
  const { data: usersData } = useUsers()
  const repUsers = useMemo(
    () => (usersData?.items ?? []).filter((u) => u.role === 'rep' || u.role === 'manager'),
    [usersData],
  )

  const { data, isLoading } = useActivities({
    activityType: typeFilter || undefined,
    status: statusFilter || undefined,
    creatorId: repFilter || undefined,
    limit: 200,
  })
  const { data: recovery } = useRecoveryBalance({})

  const allActivityTypes = useMemo(() => config?.activities.types ?? [], [config])
  const allStatuses = useMemo(() => config?.activities.statuses ?? [], [config])
  const userMap = useMemo(() => {
    const m = new Map<string, string>()
    usersData?.items?.forEach((u) => m.set(u.id, u.name || u.displayName))
    return m
  }, [usersData])

  const activities = useMemo(() => data?.items ?? [], [data])

  // Group by week then by day
  const weekGroups = useMemo(() => groupByWeekThenDay(activities), [activities])
  const sortedWeeks = useMemo(
    () => [...weekGroups.keys()].sort((a, b) => b.localeCompare(a)),
    [weekGroups],
  )

  // Mobile: limit total shown
  const mobileActivities = useMemo(() => activities.slice(0, mobileCount), [activities, mobileCount])
  const mobileWeekGroups = useMemo(() => groupByWeekThenDay(mobileActivities), [mobileActivities])
  const mobileSortedWeeks = useMemo(
    () => [...mobileWeekGroups.keys()].sort((a, b) => b.localeCompare(a)),
    [mobileWeekGroups],
  )
  const hasMore = mobileCount < activities.length

  if (isLoading) return <Spinner />

  const renderActivityCard = (activity: Activity) => {
    const Icon = activity.targetSummary?.targetType === 'pharmacy' ? Building2
      : activityIcon[activity.activityType] ?? Briefcase
    const priority = getTargetPriority(activity)
    const isCancelled = activity.status === 'anulat' || activity.status === 'cancelled'

    return (
      <button
        key={activity.id}
        type="button"
        onClick={() => setDetailId(activity.id)}
        className={`block w-full bg-white border border-slate-200 rounded-lg p-4 text-left shadow-sm hover:shadow-md transition-shadow ${isCancelled ? 'opacity-60' : ''}`}
      >
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
          <div className="flex gap-3">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center shrink-0 ${
              activity.activityType === 'visit' ? 'bg-teal-50 text-teal-600' : 'bg-slate-100 text-slate-500'
            }`}>
              <Icon size={18} />
            </div>
            <div>
              <div className="flex items-center gap-2 mb-0.5 flex-wrap">
                <h3 className={`text-sm font-semibold ${isCancelled ? 'text-slate-500 line-through' : 'text-slate-900'}`}>
                  {activity.targetName ?? activity.label ?? activity.activityType}
                </h3>
                {priority && (
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-semibold ${priorityStyle[priority] ?? ''}`}>
                    <span className={`w-1.5 h-1.5 rounded-full ${
                      priority === 'a' ? 'bg-red-500' : priority === 'b' ? 'bg-amber-500' : 'bg-slate-400'
                    }`} />
                    {priority.toUpperCase()}
                  </span>
                )}
              </div>
              <div className="text-xs text-slate-500 flex items-center gap-2 mb-1.5 flex-wrap">
                <span className="capitalize">{activity.activityType}</span>
                {str(activity.fields?.visit_type) && (
                  <>
                    <span className="text-slate-300">&middot;</span>
                    <span>{str(activity.fields.visit_type) === 'f2f' ? 'In-Person Visit' : 'Remote'}</span>
                  </>
                )}
                {activity.targetSummary && (
                  <>
                    <span className="text-slate-300">&middot;</span>
                    <span className="capitalize">{activity.targetSummary.targetType}</span>
                  </>
                )}
              </div>
              {str(activity.fields?.feedback) && (
                <p className="text-sm text-slate-700">{str(activity.fields.feedback)}</p>
              )}
              {str(activity.fields?.notes) && !str(activity.fields?.feedback) && (
                <p className={`text-sm ${isCancelled ? 'text-slate-500 italic' : 'text-slate-700'}`}>{str(activity.fields.notes)}</p>
              )}
            </div>
          </div>
          <div className="flex flex-col items-end gap-1.5 shrink-0">
            <Badge variant={statusVariant[activity.status] ?? 'default'}>
              {allStatuses.find((s) => s.key === activity.status)?.label ?? activity.status}
            </Badge>
            {activity.submittedAt && <Badge variant="success">Submitted</Badge>}
            {canFilterByRep && userMap.has(activity.creatorId) && (
              <span className="text-xs text-slate-500">{userMap.get(activity.creatorId)}</span>
            )}
          </div>
        </div>
        {Array.isArray(activity.fields?.tags) && (activity.fields.tags as string[]).length > 0 && (
          <div className="mt-3 pt-3 border-t border-slate-100 flex flex-wrap gap-2">
            {(activity.fields.tags as string[]).map((tag) => (
              <span key={tag} className="px-2 py-1 bg-slate-50 border border-slate-200 rounded text-xs text-slate-600">{tag}</span>
            ))}
          </div>
        )}
      </button>
    )
  }

  const renderWeekTimeline = (weeks: string[], groupMap: Map<string, Map<string, Activity[]>>) => (
    <div className="space-y-8">
      {weeks.map((weekKey) => {
        const dayMap = groupMap.get(weekKey)!
        const sortedDays = [...dayMap.keys()].sort((a, b) => b.localeCompare(a))
        return (
          <div key={weekKey} className="space-y-4">
            {/* Week header */}
            <div className="flex items-center gap-3 pb-2 border-b border-slate-200">
              <h2 className="text-lg font-bold text-slate-800">{formatWeekLabel(weekKey)}</h2>
              <span className="text-sm text-slate-500">{[...dayMap.values()].reduce((s, a) => s + a.length, 0)} activities</span>
            </div>

            {/* Days within the week */}
            <div className="relative pl-2">
              {sortedDays.map((dateStr) => {
                const dayActivities = dayMap.get(dateStr) ?? []
                return (
                  <div key={dateStr} className="mb-6 relative">
                    {/* Day separator */}
                    <div className="flex items-center gap-3 mb-4">
                      <div className="w-14 text-right">
                        <span className="text-xs font-bold text-slate-500 uppercase">{getDayLabel(dateStr)}</span>
                      </div>
                      <div className="w-4 h-4 rounded-full bg-slate-200 border-2 border-white shadow-sm z-10" />
                      <div className="flex-1 border-t border-slate-200" />
                    </div>

                    {/* Activity cards */}
                    <div className="space-y-3 pl-[76px]">
                      {dayActivities.map(renderActivityCard)}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </div>
  )

  return (
    <div className="flex h-full flex-col">
      {/* Header (sticky) */}
      <div className="shrink-0 px-6 py-5 bg-white border-b border-slate-200 shadow-sm sticky top-0 z-30">
        <div className="max-w-7xl mx-auto flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-slate-900">Activity Log</h1>
            <p className="text-sm text-slate-500 mt-1">Review your submitted visits and activities.</p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus size={14} />
            Log Activity
          </Button>
        </div>
      </div>

      {/* Scrollable content area */}
      <div className="flex-1 overflow-auto">
        <div className="p-6 max-w-7xl mx-auto space-y-6">

          {/* Filters + Recovery balance */}
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            {/* Filters card */}
            <div className="lg:col-span-3 bg-white rounded-xl shadow-sm border border-slate-200 p-4">
              <div className="flex flex-wrap items-end gap-4">
                <div className="flex-1 min-w-[160px]">
                  <label htmlFor="field-filter-type" className="block text-xs font-medium text-slate-500 mb-1">Activity Type</label>
                  <select id="field-filter-type" value={typeFilter} onChange={(e) => { setTypeFilter(e.target.value); setMobileCount(PAGE_SIZE) }} className={selectCls}>
                    <option value="">All Types</option>
                    {allActivityTypes.map((t) => (
                      <option key={t.key} value={t.key}>{t.label}</option>
                    ))}
                  </select>
                </div>
                <div className="w-40">
                  <label htmlFor="field-filter-status" className="block text-xs font-medium text-slate-500 mb-1">Status</label>
                  <select id="field-filter-status" value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setMobileCount(PAGE_SIZE) }} className={selectCls}>
                    <option value="">All</option>
                    {allStatuses.map((s) => (
                      <option key={s.key} value={s.key}>{s.label}</option>
                    ))}
                  </select>
                </div>
                {canFilterByRep && (
                  <div className="flex-1 min-w-[180px]">
                    <label htmlFor="field-filter-rep" className="block text-xs font-medium text-slate-500 mb-1">Rep</label>
                    <select id="field-filter-rep" value={repFilter} onChange={(e) => { setRepFilter(e.target.value); setMobileCount(PAGE_SIZE) }} className={selectCls}>
                      <option value="">All Reps</option>
                      {repUsers.map((u) => (
                        <option key={u.id} value={u.id}>{u.name || u.displayName}</option>
                      ))}
                    </select>
                  </div>
                )}
                <div className="text-sm text-slate-500 pb-1">{activities.length} results</div>
              </div>
            </div>

            {/* Recovery balance card */}
            {recovery && (
              <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-4 flex flex-col justify-center">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-semibold text-slate-700">Recovery Days</span>
                  <Clock size={16} className="text-slate-400" />
                </div>
                <div className="flex items-baseline gap-2">
                  <span className="text-3xl font-bold text-slate-900">{recovery.balance}</span>
                  <span className="text-sm text-slate-500">days available</span>
                </div>
                <div className="w-full h-1.5 bg-slate-100 rounded-full overflow-hidden mt-3">
                  <div
                    className="h-full bg-teal-500 rounded-full"
                    style={{ width: `${recovery.earned > 0 ? Math.round((recovery.balance / recovery.earned) * 100) : 0}%` }}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Nudge banner */}
          <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 flex items-start gap-3">
            <AlertTriangle size={16} className="text-amber-500 mt-0.5 shrink-0" />
            <div>
              <h4 className="text-sm font-medium text-amber-800">Action Required</h4>
              <p className="text-xs text-amber-700 mt-0.5">Review your submitted activities. Overdue A-priority targets need visits scheduled.</p>
            </div>
          </div>

          {/* Timeline content */}
          {activities.length === 0 ? (
            <p className="py-12 text-center text-sm text-slate-400">No activities found.</p>
          ) : (
            <>
              {/* Desktop */}
              <div className="hidden md:block">
                {renderWeekTimeline(sortedWeeks, weekGroups)}
              </div>

              {/* Mobile: load more */}
              <div className="md:hidden">
                {renderWeekTimeline(mobileSortedWeeks, mobileWeekGroups)}
                {hasMore && (
                  <div className="text-center pt-4">
                    <button
                      onClick={() => setMobileCount((c) => c + PAGE_SIZE)}
                      className="px-4 py-2 bg-white border border-slate-300 hover:bg-slate-50 text-slate-700 text-sm font-medium rounded-md shadow-sm transition-colors"
                    >
                      Load more ({activities.length - mobileCount} remaining)
                    </button>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Modals */}
      <CreateActivityModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <ActivityDetailModal activityId={detailId} onClose={() => setDetailId(null)} />
    </div>
  )
}
