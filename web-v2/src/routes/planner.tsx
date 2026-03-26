import { useState, useMemo, useCallback, useRef } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTargets } from '@/hooks/useTargets'
import { useActivities, useCloneWeek } from '@/hooks/useActivities'
import { useActivityStats, useCoverage } from '@/hooks/useDashboard'
import { WeekView } from '@/components/calendar/WeekView'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { StatCard } from '@/components/data/StatCard'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { ChevronLeft, ChevronRight, Copy, CalendarDays } from 'lucide-react'
import type maplibregl from 'maplibre-gl'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

function getMonday(d: Date): Date {
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1)
  const monday = new Date(d)
  monday.setDate(diff)
  monday.setHours(0, 0, 0, 0)
  return monday
}

function addDays(d: Date, n: number): Date {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

function getLat(fields: Record<string, unknown>): number | null {
  const v = fields.latitude
  return typeof v === 'number' ? v : null
}

function getLng(fields: Record<string, unknown>): number | null {
  const v = fields.longitude
  return typeof v === 'number' ? v : null
}

function getClassification(fields: Record<string, unknown>): string {
  return (fields.classification as string) ?? 'C'
}

function PlannerPage() {
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()))
  const mapInstanceRef = useRef<maplibregl.Map | null>(null)

  const weekEnd = useMemo(() => addDays(weekStart, 4), [weekStart])
  const dateFrom = formatDate(weekStart)
  const dateTo = formatDate(weekEnd)

  const { data: targetData, isLoading: targetsLoading } = useTargets({ limit: 500 })
  const { data: activityData, isLoading: activitiesLoading } = useActivities({ dateFrom, dateTo, limit: 200 })
  const { data: stats } = useActivityStats({ dateFrom, dateTo })
  const { data: coverage } = useCoverage({ dateFrom, dateTo })
  const cloneWeek = useCloneWeek()

  const targets = useMemo(() => targetData?.items ?? [], [targetData])
  const activities = useMemo(() => activityData?.items ?? [], [activityData])

  const geoTargets = useMemo(
    () => targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [targets],
  )

  const overdueA = useMemo(
    () => targets.filter((t) => getClassification(t.fields) === 'A').length,
    [targets],
  )

  const prevWeek = useCallback(() => setWeekStart((w) => addDays(w, -7)), [])
  const nextWeek = useCallback(() => setWeekStart((w) => addDays(w, 7)), [])
  const goToday = useCallback(() => setWeekStart(getMonday(new Date())), [])

  const handleCloneWeek = useCallback(() => {
    const target = addDays(weekStart, 7)
    cloneWeek.mutate({ sourceWeekStart: dateFrom, targetWeekStart: formatDate(target) })
  }, [cloneWeek, weekStart, dateFrom])

  const handleMapReady = useCallback((map: maplibregl.Map) => {
    mapInstanceRef.current = map
  }, [])

  if (targetsLoading || activitiesLoading) return <Spinner />

  return (
    <div className="flex h-full flex-col">
      {/* Nudge banner */}
      {overdueA > 0 && (
        <div className="border-b border-amber-200 bg-amber-50 px-6 py-2">
          <span className="text-sm text-amber-800">
            <Badge variant="danger" className="mr-2">{overdueA}</Badge>
            A-priority targets need attention. Schedule visits soon.
          </span>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-3">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-slate-900">Planning Workspace</h1>
          <div className="flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50">
            <button onClick={prevWeek} className="p-1.5 hover:bg-slate-100 rounded-l-lg">
              <ChevronLeft size={16} />
            </button>
            <span className="px-3 text-sm font-medium text-slate-700">
              {weekStart.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} — {weekEnd.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}
            </span>
            <button onClick={nextWeek} className="p-1.5 hover:bg-slate-100 rounded-r-lg">
              <ChevronRight size={16} />
            </button>
          </div>
          <Button variant="ghost" size="sm" onClick={goToday}>
            <CalendarDays size={14} />
            Today
          </Button>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={handleCloneWeek} disabled={cloneWeek.isPending}>
            <Copy size={14} />
            Clone Week
          </Button>
        </div>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-4 gap-4 border-b border-slate-200 bg-white px-6 py-4">
        <StatCard label="Planned" value={stats?.total ?? 0} />
        <StatCard
          label="Completed"
          value={stats?.byStatus?.realizat ?? 0}
          subtitle={stats?.total ? `${Math.round(((stats.byStatus?.realizat ?? 0) / stats.total) * 100)}%` : undefined}
          trend="up"
        />
        <StatCard label="Coverage" value={coverage ? `${Math.round(coverage.percentage)}%` : '-'} />
        <StatCard label="Overdue A" value={overdueA} trend={overdueA > 0 ? 'down' : 'neutral'} />
      </div>

      {/* Main content: map + calendar */}
      <div className="flex flex-1 min-h-0">
        {/* Map */}
        <div className="w-1/2 border-r border-slate-200">
          <MapContainer className="h-full" onMapReady={handleMapReady}>
            {geoTargets.map((t) => (
              <TargetMarker
                key={t.id}
                map={mapInstanceRef.current}
                lat={getLat(t.fields)!}
                lng={getLng(t.fields)!}
                name={t.name}
                priority={getClassification(t.fields)}
              />
            ))}
          </MapContainer>
        </div>

        {/* Calendar */}
        <div className="w-1/2 overflow-auto p-4">
          <WeekView
            weekStart={weekStart}
            activities={activities}
          />
        </div>
      </div>
    </div>
  )
}
