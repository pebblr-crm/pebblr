import { useState, useMemo, useCallback } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTargets, useTargetVisitStatus } from '@/hooks/useTargets'
import { useActivities, useCloneWeek, useBatchCreateActivities, usePatchActivity } from '@/hooks/useActivities'
import { useActivityStats, useCoverage } from '@/hooks/useDashboard'
import { useWeekNav } from '@/hooks/useWeekNav'
import { useDragDropPlanner } from '@/hooks/useDragDropPlanner'
import { WeekView } from '@/components/calendar/WeekView'
import { TargetListPanel } from '@/components/planner/TargetListPanel'
import { PlannerHeader } from '@/components/planner/PlannerHeader'
import { PlannerNudgeBanner } from '@/components/planner/PlannerNudgeBanner'
import { PlannerMobileMap } from '@/components/planner/PlannerMobileMap'
import { BulkScheduleModal } from '@/components/planner/BulkScheduleModal'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { Spinner } from '@/components/ui/Spinner'
import { QueryError } from '@/components/ui/QueryError'
import { useToast } from '@/components/ui/Toast'
import { addDays, formatDate } from '@/lib/dates'
import { getLat, getLng, getClassification } from '@/lib/target-fields'
import { MapIcon } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

function PlannerPage() {
  const { weekStart, weekEnd, dateFrom, dateTo, prevWeek, nextWeek, goToday } = useWeekNav()
  const [showMobileMap, setShowMobileMap] = useState(false)
  const [selectedTargetIds, setSelectedTargetIds] = useState<Set<string>>(new Set())
  const [targetSearch, setTargetSearch] = useState('')
  const [hoveredTargetId, setHoveredTargetId] = useState<string | null>(null)
  const [priorityFilter, setPriorityFilter] = useState('')
  const [detailActivityId, setDetailActivityId] = useState<string | null>(null)
  const [bulkModalOpen, setBulkModalOpen] = useState(false)

  const { showToast, ToastContainer } = useToast()

  const { data: targetData, isLoading: targetsLoading, isError: targetsError, refetch: refetchTargets } = useTargets({ limit: 500 })
  const { data: activityData, isLoading: activitiesLoading, isError: activitiesError, refetch: refetchActivities } = useActivities({ dateFrom, dateTo, limit: 200 })
  const { data: stats } = useActivityStats({ dateFrom, dateTo })
  const { data: coverage } = useCoverage({ dateFrom, dateTo })
  const { data: visitStatusData } = useTargetVisitStatus()
  const cloneWeek = useCloneWeek()
  const batchCreate = useBatchCreateActivities()
  const patchActivity = usePatchActivity()

  const targets = useMemo(() => targetData?.items ?? [], [targetData])
  const activities = useMemo(() => activityData?.items ?? [], [activityData])
  const targetMap = useMemo(() => new Map(targets.map((t) => [t.id, t])), [targets])

  const toggleTarget = useCallback((id: string) => {
    setSelectedTargetIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const clearSelection = useCallback(() => setSelectedTargetIds(new Set()), [])

  const {
    dragActivityId, dragPending, isDragging, dayAssignments,
    setDragTargetId, setDragActivityId, setDragPending,
    handleDrop, removeFromDay, setDayAssignments,
  } = useDragDropPlanner({
    activities,
    selectedTargetIds,
    clearSelection,
    patchActivity,
    showToast,
  })

  const visitMap = useMemo(() => {
    const m = new Map<string, string>()
    visitStatusData?.items?.forEach((v) => m.set(v.targetId, v.lastVisitDate))
    return m
  }, [visitStatusData])

  const geoTargets = useMemo(
    () => targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [targets],
  )

  const filteredTargets = useMemo(() => {
    let result = geoTargets
    if (targetSearch) {
      const q = targetSearch.toLowerCase()
      result = result.filter((t) => t.name.toLowerCase().includes(q))
    }
    if (priorityFilter) {
      result = result.filter((t) => getClassification(t.fields) === priorityFilter)
    }
    return result
  }, [geoTargets, targetSearch, priorityFilter])

  const scheduledTargetIds = useMemo(() => {
    const set = new Set(activities.map((a) => a.targetId).filter(Boolean))
    for (const ids of Object.values(dayAssignments)) {
      for (const id of ids) set.add(id)
    }
    return set
  }, [activities, dayAssignments])

  const overdueA = useMemo(() => {
    const cadenceDays = 21
    const now = new Date()
    return targets.filter((t) => {
      if (getClassification(t.fields) !== 'a') return false
      if (scheduledTargetIds.has(t.id)) return false
      const lastVisit = visitMap.get(t.id)
      if (!lastVisit) return true
      const daysSince = Math.floor((now.getTime() - new Date(lastVisit).getTime()) / (1000 * 60 * 60 * 24))
      return daysSince >= cadenceDays
    }).length
  }, [targets, scheduledTargetIds, visitMap])

  const totalAssigned = useMemo(
    () => Object.values(dayAssignments).reduce((sum, arr) => sum + arr.length, 0),
    [dayAssignments],
  )

  const handleCloneWeek = useCallback(() => {
    const target = addDays(weekStart, 7)
    cloneWeek.mutate({ sourceWeekStart: dateFrom, targetWeekStart: formatDate(target) })
  }, [cloneWeek, weekStart, dateFrom])

  const handleCreateActivities = useCallback(async () => {
    const items: Array<{ targetId: string; dueDate: string; fields: Record<string, unknown> }> = []
    for (const [date, ids] of Object.entries(dayAssignments)) {
      for (const id of ids) {
        items.push({ targetId: id, dueDate: date, fields: { visit_type: 'f2f' } })
      }
    }
    if (items.length === 0) return
    try {
      const result = await batchCreate.mutateAsync(items)
      const created = result.created?.length ?? items.length
      const errors = result.errors?.length ?? 0
      setDayAssignments({})
      clearSelection()
      if (errors > 0) {
        showToast(`${created} created, ${errors} failed`, 'warning')
      } else {
        showToast(`${created} visit${created === 1 ? '' : 's'} created`, 'info')
      }
    } catch {
      showToast('Failed to create visits', 'error')
    }
  }, [dayAssignments, batchCreate, showToast, setDayAssignments, clearSelection])

  const handleBulkSchedule = useCallback(async (date: string, visitType: 'f2f' | 'remote') => {
    if (!date || selectedTargetIds.size === 0) return
    const items = Array.from(selectedTargetIds).map((targetId) => ({
      targetId,
      dueDate: date,
      fields: { visit_type: visitType },
    }))
    try {
      const result = await batchCreate.mutateAsync(items)
      const created = result.created?.length ?? items.length
      const errors = result.errors?.length ?? 0
      clearSelection()
      setBulkModalOpen(false)
      if (errors > 0) {
        showToast(`${created} created, ${errors} failed`, 'warning')
      } else {
        showToast(`${created} visit${created === 1 ? '' : 's'} created`, 'info')
      }
    } catch {
      showToast('Failed to create visits', 'error')
    }
  }, [selectedTargetIds, batchCreate, showToast, clearSelection])

  if (targetsLoading || activitiesLoading) return <Spinner />
  if (targetsError || activitiesError) return <QueryError message="Failed to load planner data" onRetry={() => { refetchTargets(); refetchActivities() }} />

  const completedCount = stats?.byStatus?.realizat ?? 0
  const completionRate = stats?.total ? Math.round((completedCount / stats.total) * 100) : 0
  const coveragePct = coverage ? Math.round(coverage.percentage) : 0

  return (
    <>
    <ToastContainer />
    <div className="flex h-full flex-col">
      <div className="flex flex-1 min-h-0 flex-col lg:flex-row">

        <TargetListPanel
          filteredTargets={filteredTargets}
          selectedTargetIds={selectedTargetIds}
          hoveredTargetId={hoveredTargetId}
          targetSearch={targetSearch}
          priorityFilter={priorityFilter}
          visitMap={visitMap}
          onTargetSearchChange={setTargetSearch}
          onPriorityFilterChange={setPriorityFilter}
          onToggleTarget={toggleTarget}
          onClearSelection={clearSelection}
          onHoverTarget={setHoveredTargetId}
          onDragTargetStart={setDragTargetId}
          onDragTargetEnd={() => setDragTargetId(null)}
          onBulkSchedule={() => setBulkModalOpen(true)}
        />

        <section className="flex-1 flex flex-col bg-slate-50">
          <PlannerHeader
            weekStart={weekStart}
            weekEnd={weekEnd}
            totalAssigned={totalAssigned}
            onPrevWeek={prevWeek}
            onNextWeek={nextWeek}
            onGoToday={goToday}
            onCloneWeek={handleCloneWeek}
            onCreateActivities={handleCreateActivities}
            cloneWeekPending={cloneWeek.isPending}
            batchCreatePending={batchCreate.isPending}
          />

          {(overdueA > 0 || stats?.total) && (
            <PlannerNudgeBanner
              overdueA={overdueA}
              completionRate={completionRate}
              coveragePct={coveragePct}
            />
          )}

          <button
            onClick={() => setShowMobileMap(true)}
            className="mx-4 mt-3 flex items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-50 lg:hidden"
          >
            <MapIcon size={16} />
            Show Map ({geoTargets.length} targets)
          </button>

          <div className="flex-1 overflow-auto p-4">
            <WeekView
              weekStart={weekStart}
              activities={activities}
              dayAssignments={dayAssignments}
              targetMap={targetMap}
              isDragging={isDragging}
              draggingActivityId={dragActivityId}
              draggingPending={dragPending}
              onDrop={handleDrop}
              onRemoveAssignment={removeFromDay}
              onActivityDragStart={(id) => setDragActivityId(id)}
              onActivityDragEnd={() => setDragActivityId(null)}
              onPendingDragStart={(sourceDate, targetId) => setDragPending({ sourceDate, targetId })}
              onPendingDragEnd={() => setDragPending(null)}
              onActivityClick={(a) => setDetailActivityId(a.id)}
            />
          </div>
        </section>

        {showMobileMap && (
          <PlannerMobileMap
            geoTargets={geoTargets}
            selectedTargetIds={selectedTargetIds}
            onToggleTarget={toggleTarget}
            onClose={() => setShowMobileMap(false)}
          />
        )}
      </div>
    </div>

    <ActivityDetailModal activityId={detailActivityId} onClose={() => setDetailActivityId(null)} />

    <BulkScheduleModal
      open={bulkModalOpen}
      onClose={() => setBulkModalOpen(false)}
      selectedTargetIds={selectedTargetIds}
      targetMap={targetMap}
      initialDate={dateFrom}
      onSchedule={handleBulkSchedule}
      isPending={batchCreate.isPending}
    />
    </>
  )
}
