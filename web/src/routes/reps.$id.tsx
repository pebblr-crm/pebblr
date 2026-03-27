import { useState, useMemo, useCallback } from 'react'
import { createRoute, useParams } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useActivities } from '@/hooks/useActivities'
import { useTargets } from '@/hooks/useTargets'
import { useActivityStats, useCoverage } from '@/hooks/useDashboard'
import { WeekView } from '@/components/calendar/WeekView'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { StatCard } from '@/components/data/StatCard'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { getMonday, addDays, formatDate } from '@/lib/dates'
import { ArrowLeft, ChevronLeft, ChevronRight, Info } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/reps/$id',
  component: RepDrillDownPage,
})

function getLat(fields: Record<string, unknown>): number | null {
  const v = fields.lat
  return typeof v === 'number' ? v : null
}

function getLng(fields: Record<string, unknown>): number | null {
  const v = fields.lng
  return typeof v === 'number' ? v : null
}

function getClassification(fields: Record<string, unknown>): string {
  return ((fields.potential as string) ?? 'c').toLowerCase()
}

function RepDrillDownPage() {
  const { id: repId } = useParams({ from: '/reps/$id' })
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()))

  const weekEnd = useMemo(() => addDays(weekStart, 4), [weekStart])
  const dateFrom = formatDate(weekStart)
  const dateTo = formatDate(weekEnd)

  const { data: activityData, isLoading: actLoading } = useActivities({
    creatorId: repId,
    dateFrom,
    dateTo,
    limit: 200,
  })
  const { data: targetData, isLoading: targetLoading } = useTargets({ assignee: repId, limit: 500 })
  const { data: stats } = useActivityStats({ userId: repId, dateFrom, dateTo })
  const { data: coverage } = useCoverage({ userId: repId, dateFrom, dateTo })

  const activities = useMemo(() => activityData?.items ?? [], [activityData])
  const targets = useMemo(() => targetData?.items ?? [], [targetData])
  const geoTargets = useMemo(
    () => targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [targets],
  )

  const prevWeek = useCallback(() => setWeekStart((w) => addDays(w, -7)), [])
  const nextWeek = useCallback(() => setWeekStart((w) => addDays(w, 7)), [])

  const completedCount = stats?.byStatus?.realizat ?? 0
  const completionRate = stats?.total ? Math.round((completedCount / stats.total) * 100) : 0

  if (actLoading || targetLoading) return <Spinner />

  return (
    <div className="flex h-full flex-col">
      {/* Read-only banner */}
      <div className="flex items-center gap-2 border-b border-blue-200 bg-blue-50 px-4 py-2 md:px-6">
        <Info size={16} className="text-blue-600" />
        <span className="text-sm text-blue-800">
          You are viewing this rep&apos;s schedule in read-only mode.
        </span>
      </div>

      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 bg-white px-4 py-3 md:px-6">
        <div className="flex items-center gap-3">
          <a href="/dashboard" className="text-slate-400 hover:text-slate-600">
            <ArrowLeft size={20} />
          </a>
          <div>
            <h1 className="text-base font-semibold text-slate-900 md:text-lg">Rep: {repId}</h1>
            <Badge variant={completionRate >= 70 ? 'success' : 'warning'}>
              {completionRate >= 70 ? 'On Track' : 'Needs Attention'}
            </Badge>
          </div>
        </div>

        <div className="flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50">
          <button onClick={prevWeek} className="p-1.5 hover:bg-slate-100 rounded-l-lg">
            <ChevronLeft size={16} />
          </button>
          <span className="px-2 text-xs font-medium text-slate-700 md:px-3 md:text-sm">
            {weekStart.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} — {weekEnd.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}
          </span>
          <button onClick={nextWeek} className="p-1.5 hover:bg-slate-100 rounded-r-lg">
            <ChevronRight size={16} />
          </button>
        </div>

        <Button variant="secondary" size="sm" disabled>
          Message Rep
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-3 border-b border-slate-200 bg-white px-4 py-3 md:grid-cols-4 md:gap-4 md:px-6 md:py-4">
        <StatCard label="Compliance" value={`${completionRate}%`} trend={completionRate >= 70 ? 'up' : 'down'} />
        <StatCard label="Completed" value={completedCount} subtitle={`of ${stats?.total ?? 0}`} />
        <StatCard label="Coverage" value={coverage ? `${Math.round(coverage.percentage)}%` : '-'} />
        <StatCard label="Targets" value={targets.length} />
      </div>

      {/* Map + Calendar */}
      <div className="flex flex-1 min-h-0 flex-col md:flex-row">
        <div className="h-56 border-b border-slate-200 md:h-auto md:w-1/2 md:border-b-0 md:border-r">
          <MapContainer className="h-full">
            {geoTargets.map((t) => (
              <TargetMarker
                key={t.id}
                lat={getLat(t.fields)!}
                lng={getLng(t.fields)!}
                name={t.name}
                priority={getClassification(t.fields)}
              />
            ))}
          </MapContainer>
        </div>
        <div className="flex-1 overflow-auto p-4 md:w-1/2">
          <WeekView weekStart={weekStart} activities={activities} />
        </div>
      </div>
    </div>
  )
}
