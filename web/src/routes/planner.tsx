import { useState, useMemo, useCallback } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTargets, useTargetVisitStatus } from '@/hooks/useTargets'
import { useActivities, useCloneWeek, useBatchCreateActivities, usePatchActivity } from '@/hooks/useActivities'
import { useActivityStats, useCoverage } from '@/hooks/useDashboard'
import { WeekView } from '@/components/calendar/WeekView'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { TargetListPanel } from '@/components/planner/TargetListPanel'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { QueryError } from '@/components/ui/QueryError'
import { Modal } from '@/components/ui/Modal'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { useToast } from '@/components/ui/Toast'
import { getMonday, addDays, formatDate } from '@/lib/dates'
import { priorityDot } from '@/lib/styles'
import { getLat, getLng, getClassification } from '@/lib/target-fields'
import { ChevronLeft, ChevronRight, Copy, CalendarDays, MapIcon, X, Info, CalendarPlus } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

/* ── Main page ── */

function PlannerPage() {
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()))
  const [showMobileMap, setShowMobileMap] = useState(false)
  const [selectedTargetIds, setSelectedTargetIds] = useState<Set<string>>(new Set())
  const [targetSearch, setTargetSearch] = useState('')
  const [hoveredTargetId, setHoveredTargetId] = useState<string | null>(null)
  const [priorityFilter, setPriorityFilter] = useState('')
  const [detailActivityId, setDetailActivityId] = useState<string | null>(null)

  const { showToast, ToastContainer } = useToast()

  // Drag state
  const [dragTargetId, setDragTargetId] = useState<string | null>(null)
  const [dragActivityId, setDragActivityId] = useState<string | null>(null)
  const [dragPending, setDragPending] = useState<{ sourceDate: string; targetId: string } | null>(null)
  const isDragging = dragTargetId != null || dragActivityId != null || dragPending != null

  // Day assignments: dateStr → target IDs (pending creation)
  const [dayAssignments, setDayAssignments] = useState<Record<string, string[]>>({})

  const weekEnd = useMemo(() => addDays(weekStart, 4), [weekStart])
  const dateFrom = formatDate(weekStart)
  const dateTo = formatDate(weekEnd)

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

  // A-priority targets that haven't been visited in 21 days AND have no scheduled/pending visit
  const overdueA = useMemo(() => {
    const cadenceDays = 21
    const now = new Date()
    // Target IDs with an existing activity this week
    const scheduledIds = new Set(activities.map((a) => a.targetId).filter(Boolean))
    // Target IDs with a pending assignment
    for (const ids of Object.values(dayAssignments)) {
      for (const id of ids) scheduledIds.add(id)
    }
    return targets.filter((t) => {
      if (getClassification(t.fields) !== 'a') return false
      if (scheduledIds.has(t.id)) return false
      const lastVisit = visitMap.get(t.id)
      if (!lastVisit) return true // never visited
      const daysSince = Math.floor((now.getTime() - new Date(lastVisit).getTime()) / (1000 * 60 * 60 * 24))
      return daysSince >= cadenceDays
    }).length
  }, [targets, activities, dayAssignments, visitMap])

  const totalAssigned = useMemo(
    () => Object.values(dayAssignments).reduce((sum, arr) => sum + arr.length, 0),
    [dayAssignments],
  )

  // ── Navigation ──
  const prevWeek = useCallback(() => setWeekStart((w) => addDays(w, -7)), [])
  const nextWeek = useCallback(() => setWeekStart((w) => addDays(w, 7)), [])
  const goToday = useCallback(() => setWeekStart(getMonday(new Date())), [])

  const handleCloneWeek = useCallback(() => {
    const target = addDays(weekStart, 7)
    cloneWeek.mutate({ sourceWeekStart: dateFrom, targetWeekStart: formatDate(target) })
  }, [cloneWeek, weekStart, dateFrom])

  // ── Selection ──
  const toggleTarget = useCallback((id: string) => {
    setSelectedTargetIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const clearSelection = useCallback(() => setSelectedTargetIds(new Set()), [])

  // ── Drop sub-handlers ──

  const handleDropPending = useCallback((dateStr: string) => {
    if (!dragPending) return
    if (dragPending.sourceDate === dateStr) {
      setDragPending(null)
      return
    }
    const existingOnDay = activities.some(
      (a) => a.dueDate.slice(0, 10) === dateStr && a.targetId === dragPending.targetId,
    )
    if (existingOnDay) {
      showToast('Target already has a visit on this day', 'warning')
      setDragPending(null)
      return
    }
    setDayAssignments((prev) => {
      const srcArr = (prev[dragPending.sourceDate] ?? []).filter((id) => id !== dragPending.targetId)
      const dstArr = prev[dateStr] ?? []
      if (dstArr.includes(dragPending.targetId)) {
        showToast('Target is already on this day', 'warning')
        return prev
      }
      const next = { ...prev, [dateStr]: [...dstArr, dragPending.targetId] }
      if (srcArr.length === 0) delete next[dragPending.sourceDate]
      else next[dragPending.sourceDate] = srcArr
      return next
    })
    setDragPending(null)
  }, [dragPending, activities, showToast])

  const handleDropTargets = useCallback((dateStr: string) => {
    const ids = new Set(selectedTargetIds)
    if (dragTargetId) ids.add(dragTargetId)
    if (ids.size === 0) return

    const pendingOnDay = new Set(dayAssignments[dateStr] ?? [])
    const existingOnDay = new Set(
      activities.filter((a) => a.dueDate.slice(0, 10) === dateStr && a.targetId).map((a) => a.targetId!),
    )
    const toAdd = Array.from(ids).filter((id) => !pendingOnDay.has(id) && !existingOnDay.has(id))
    const dupCount = ids.size - toAdd.length

    if (toAdd.length > 0) {
      setDayAssignments((prev) => ({
        ...prev,
        [dateStr]: [...(prev[dateStr] ?? []), ...toAdd],
      }))
    }
    if (dupCount > 0) {
      let msg: string
      if (dupCount === ids.size) {
        const subject = dupCount === 1 ? 'target is' : `all ${dupCount} targets are`
        msg = `Already scheduled — ${subject} already on this day`
      } else {
        msg = `${toAdd.length} added, ${dupCount} already on this day`
      }
      showToast(msg, 'warning')
    }
    setDragTargetId(null)
    setSelectedTargetIds(new Set())
  }, [dragTargetId, selectedTargetIds, dayAssignments, activities, showToast])

  // ── Drop handler ──
  const handleDrop = useCallback((dateStr: string) => {
    if (dragPending) {
      handleDropPending(dateStr)
      return
    }
    if (dragActivityId) {
      patchActivity.mutate({ id: dragActivityId, dueDate: dateStr })
      setDragActivityId(null)
      return
    }
    if (dragTargetId || selectedTargetIds.size > 0) {
      handleDropTargets(dateStr)
    }
  }, [dragPending, dragActivityId, dragTargetId, selectedTargetIds, handleDropPending, handleDropTargets, patchActivity])

  const removeFromDay = useCallback((dateStr: string, targetId: string) => {
    setDayAssignments((prev) => {
      const arr = (prev[dateStr] ?? []).filter((id) => id !== targetId)
      const next = { ...prev }
      if (arr.length === 0) delete next[dateStr]
      else next[dateStr] = arr
      return next
    })
  }, [])

  // ── Batch create ──
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
      setSelectedTargetIds(new Set())
      if (errors > 0) {
        showToast(`${created} created, ${errors} failed`, 'warning')
      } else {
        showToast(`${created} visit${created !== 1 ? 's' : ''} created`, 'info')
      }
    } catch {
      showToast('Failed to create visits', 'error')
    }
  }, [dayAssignments, batchCreate, showToast])

  // ── Bulk schedule modal ──
  const [bulkModalOpen, setBulkModalOpen] = useState(false)
  const [bulkDate, setBulkDate] = useState(dateFrom)
  const [bulkVisitType, setBulkVisitType] = useState<'f2f' | 'remote'>('f2f')
  const handleBulkSchedule = useCallback(async () => {
    if (!bulkDate || selectedTargetIds.size === 0) return
    const items = Array.from(selectedTargetIds).map((targetId) => ({
      targetId,
      dueDate: bulkDate,
      fields: { visit_type: bulkVisitType },
    }))
    try {
      const result = await batchCreate.mutateAsync(items)
      const created = result.created?.length ?? items.length
      const errors = result.errors?.length ?? 0
      setSelectedTargetIds(new Set())
      setBulkModalOpen(false)
      if (errors > 0) {
        showToast(`${created} created, ${errors} failed`, 'warning')
      } else {
        showToast(`${created} visit${created !== 1 ? 's' : ''} created`, 'info')
      }
    } catch {
      showToast('Failed to create visits', 'error')
    }
  }, [selectedTargetIds, bulkDate, bulkVisitType, batchCreate, showToast])


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

        {/* ═══ LEFT PANEL: Map + Target List ═══ */}
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

        {/* ═══ RIGHT PANEL: Calendar ═══ */}
        <section className="flex-1 flex flex-col bg-slate-50">
          {/* Calendar header */}
          <div className="px-4 py-3 border-b border-slate-200 bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-2 shrink-0 md:px-6">
            <div className="flex flex-wrap items-center gap-2 md:gap-3">
              <div className="flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50">
                <button onClick={prevWeek} className="p-1.5 hover:bg-slate-100 rounded-l-lg" aria-label="Previous week">
                  <ChevronLeft size={16} />
                </button>
                <span className="px-2 text-xs font-medium text-slate-700 md:px-3 md:text-sm">
                  {weekStart.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} — {weekEnd.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}
                </span>
                <button onClick={nextWeek} className="p-1.5 hover:bg-slate-100 rounded-r-lg" aria-label="Next week">
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
              {totalAssigned > 0 && (
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleCreateActivities}
                  disabled={batchCreate.isPending}
                >
                  <CalendarPlus size={14} />
                  Create {totalAssigned} Visit{totalAssigned !== 1 ? 's' : ''}
                </Button>
              )}
            </div>
          </div>

          {/* Nudge / compliance banner */}
          {(overdueA > 0 || stats?.total) && (
            <div className="bg-indigo-50 border-b border-indigo-100 px-4 py-2 flex items-center justify-between shrink-0 md:px-6">
              <div className="flex items-center gap-3">
                <Info size={16} className="text-indigo-500 shrink-0" />
                <span className="text-sm text-indigo-900">
                  {overdueA > 0 ? (
                    <>You have <strong>{overdueA} A-priority targets</strong> that need visits.</>
                  ) : (
                    <>All A-priority targets are covered.</>
                  )}
                </span>
              </div>
              <div className="hidden sm:flex items-center gap-4 text-xs font-medium shrink-0">
                <span className="text-slate-600">
                  Completed: <span className={completionRate >= 80 ? 'text-emerald-600' : 'text-amber-600'}>{completionRate}%</span>
                </span>
                <span className="text-slate-600">
                  Coverage: <span className={coveragePct >= 80 ? 'text-emerald-600' : 'text-amber-600'}>{coveragePct}%</span>
                </span>
              </div>
            </div>
          )}

          {/* Mobile map toggle */}
          <button
            onClick={() => setShowMobileMap(true)}
            className="mx-4 mt-3 flex items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-50 lg:hidden"
          >
            <MapIcon size={16} />
            Show Map ({geoTargets.length} targets)
          </button>

          {/* Calendar grid */}
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

        {/* Mobile map overlay */}
        {showMobileMap && (
          <div className="fixed inset-0 z-50 flex flex-col bg-white lg:hidden">
            <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
              <h2 className="text-sm font-semibold text-slate-900">Target Map</h2>
              <button
                onClick={() => setShowMobileMap(false)}
                className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100"
                aria-label="Close map"
              >
                <X size={20} />
              </button>
            </div>
            <div className="flex-1">
              <MapContainer className="h-full">
                {geoTargets.map((t) => (
                  <TargetMarker
                    key={t.id}
                    lat={getLat(t.fields)!}
                    lng={getLng(t.fields)!}
                    name={t.name}
                    priority={getClassification(t.fields)}
                    selected={selectedTargetIds.has(t.id)}
                    onClick={() => toggleTarget(t.id)}
                  />
                ))}
              </MapContainer>
            </div>
          </div>
        )}
      </div>
    </div>

    {/* Activity detail modal */}
    <ActivityDetailModal activityId={detailActivityId} onClose={() => setDetailActivityId(null)} />

    {/* Bulk schedule modal */}
    <Modal
      open={bulkModalOpen}
      onClose={() => setBulkModalOpen(false)}
      title="Bulk Schedule"
      footer={
        <div className="flex items-center justify-between">
          <span className="text-[11px] text-slate-400">
            <Info size={12} className="inline text-slate-400 mr-1" />You can also drag and drop targets onto the calendar
          </span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" size="sm" onClick={() => setBulkModalOpen(false)}>
              Cancel
            </Button>
            <Button variant="primary" size="sm" onClick={handleBulkSchedule} disabled={batchCreate.isPending}>
              <CalendarPlus size={14} />
              {batchCreate.isPending ? 'Creating...' : `Schedule ${selectedTargetIds.size} Target${selectedTargetIds.size !== 1 ? 's' : ''}`}
            </Button>
          </div>
        </div>
      }
    >
      <div className="space-y-4">
        <p className="text-sm text-slate-600">
          Pick a date to schedule <strong>{selectedTargetIds.size}</strong> selected target{selectedTargetIds.size !== 1 ? 's' : ''}.
        </p>
        <input
          type="date"
          value={bulkDate}
          onChange={(e) => setBulkDate(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
        />
        {/* Visit type toggle */}
        <div>
          <fieldset>
          <legend className="mb-1.5 block text-sm font-medium text-slate-700">Visit type</legend>
          <div className="flex rounded-lg border border-slate-200 overflow-hidden">
            <button
              type="button"
              onClick={() => setBulkVisitType('f2f')}
              className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${
                bulkVisitType === 'f2f'
                  ? 'bg-teal-600 text-white'
                  : 'bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              Face to face
            </button>
            <button
              type="button"
              onClick={() => setBulkVisitType('remote')}
              className={`flex-1 px-3 py-2 text-sm font-medium border-l border-slate-200 transition-colors ${
                bulkVisitType === 'remote'
                  ? 'bg-teal-600 text-white'
                  : 'bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              Remote
            </button>
          </div>
          </fieldset>
        </div>
        <div className="max-h-48 overflow-y-auto rounded-lg border border-slate-200">
          <ul className="divide-y divide-slate-100">
            {Array.from(selectedTargetIds).map((id) => {
              const t = targetMap.get(id)
              if (!t) return null
              const p = getClassification(t.fields)
              return (
                <li key={id} className="flex items-center gap-2 px-3 py-2">
                  <span className={`w-2 h-2 rounded-full shrink-0 ${priorityDot[p] ?? priorityDot.c}`} />
                  <span className="text-sm text-slate-800 truncate flex-1">{t.name}</span>
                  <span className={`text-[9px] font-bold uppercase px-1.5 py-0.5 rounded shrink-0 ${
                    { a: 'bg-red-100 text-red-700', b: 'bg-amber-100 text-amber-700', c: 'bg-slate-100 text-slate-600' }[p] ?? 'bg-slate-100 text-slate-600'
                  }`}>{p.toUpperCase()}</span>
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </Modal>
    </>
  )
}
