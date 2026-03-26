import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useActivities } from '@/hooks/useActivities'
import { useRecoveryBalance } from '@/hooks/useDashboard'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Plus, Clock, AlertTriangle } from 'lucide-react'
import type { Activity } from '@/types/activity'

function str(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities',
  component: ActivitiesPage,
})

const statusVariant: Record<string, 'primary' | 'success' | 'danger' | 'default'> = {
  planificat: 'primary',
  realizat: 'success',
  anulat: 'danger',
}

const statusColor: Record<string, string> = {
  realizat: 'bg-emerald-500',
  planificat: 'bg-blue-500',
  anulat: 'bg-red-500',
}

function groupByDate(activities: Activity[]): Map<string, Activity[]> {
  const map = new Map<string, Activity[]>()
  for (const a of activities) {
    const list = map.get(a.dueDate) ?? []
    list.push(a)
    map.set(a.dueDate, list)
  }
  return map
}

function ActivitiesPage() {
  const [typeFilter, setTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const { data, isLoading } = useActivities({
    activityType: typeFilter || undefined,
    status: statusFilter || undefined,
    limit: 100,
  })
  const { data: recovery } = useRecoveryBalance({})

  const activities = useMemo(() => data?.items ?? [], [data])
  const grouped = useMemo(() => groupByDate(activities), [activities])
  const sortedDates = useMemo(
    () => [...grouped.keys()].sort((a, b) => b.localeCompare(a)),
    [grouped],
  )

  if (isLoading) return <Spinner />

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Activity Log</h1>
        <a href="/activities/new">
          <Button size="sm">
            <Plus size={14} />
            Log Activity
          </Button>
        </a>
      </div>

      {/* Recovery + nudge cards */}
      <div className="mb-6 grid grid-cols-3 gap-4">
        {recovery && (
          <Card>
            <div className="flex items-center gap-2">
              <Clock size={16} className="text-teal-600" />
              <span className="text-sm font-medium text-slate-700">Recovery Balance</span>
            </div>
            <p className="mt-2 text-2xl font-semibold text-slate-900">{recovery.balance} days</p>
            <p className="text-xs text-slate-500">{recovery.earned} earned, {recovery.taken} taken</p>
          </Card>
        )}
        <Card className="col-span-2 flex items-center gap-3 bg-amber-50 border-amber-200">
          <AlertTriangle size={18} className="text-amber-600 shrink-0" />
          <span className="text-sm text-amber-800">
            Review your submitted activities. Overdue targets need visits scheduled.
          </span>
        </Card>
      </div>

      {/* Filters */}
      <div className="mb-4 flex items-center gap-3">
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">All types</option>
          <option value="visit">Visit</option>
          <option value="administrative">Administrative</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">All statuses</option>
          <option value="planificat">Planned</option>
          <option value="realizat">Completed</option>
          <option value="anulat">Cancelled</option>
        </select>
        <span className="text-sm text-slate-500">{activities.length} activities</span>
      </div>

      {/* Timeline */}
      {sortedDates.length === 0 ? (
        <p className="py-12 text-center text-sm text-slate-400">No activities found.</p>
      ) : (
        <div className="space-y-6">
          {sortedDates.map((date) => {
            const dayActivities = grouped.get(date) ?? []
            const d = new Date(date)
            return (
              <div key={date}>
                <h3 className="mb-3 text-sm font-semibold text-slate-500">
                  {d.toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
                </h3>
                <div className="relative pl-6 space-y-3">
                  <div className="absolute left-2 top-2 bottom-2 w-px bg-slate-200" />
                  {dayActivities.map((activity) => (
                    <div key={activity.id} className="relative rounded-lg border border-slate-200 bg-white p-4">
                      <div className={`absolute -left-4 top-5 h-3 w-3 rounded-full border-2 border-white ${statusColor[activity.status] ?? 'bg-slate-400'}`} />
                      <div className="flex items-start justify-between">
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-slate-900">
                              {activity.targetName ?? activity.label ?? activity.activityType}
                            </span>
                            <Badge variant={statusVariant[activity.status] ?? 'default'}>
                              {activity.status}
                            </Badge>
                          </div>
                          <p className="mt-1 text-xs text-slate-500 capitalize">{activity.activityType} &middot; {activity.duration}</p>
                        </div>
                        {activity.submittedAt && (
                          <Badge variant="success">Submitted</Badge>
                        )}
                      </div>
                      {str(activity.fields?.notes) && (
                        <p className="mt-2 text-sm text-slate-600">{str(activity.fields.notes)}</p>
                      )}
                      {Array.isArray(activity.fields?.tags) && (
                        <div className="mt-2 flex flex-wrap gap-1">
                          {(activity.fields.tags as string[]).map((tag) => (
                            <Badge key={tag}>{tag}</Badge>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
