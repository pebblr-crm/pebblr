import { useState, useMemo, useCallback } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTargets, useTargetVisitStatus } from '@/hooks/useTargets'
import { useActivities, useCloneWeek, useBatchCreateActivities, usePatchActivity } from '@/hooks/useActivities'
import { useActivityStats, useCoverage } from '@/hooks/useDashboard'
import { WeekView } from '@/components/calendar/WeekView'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Modal } from '@/components/ui/Modal'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { useToast } from '@/components/ui/Toast'
import { getMonday, addDays, formatDate } from '@/lib/dates'
import { priorityDot } from '@/lib/styles'
import { daysAgo } from '@/lib/helpers'
import {
  ChevronLeft,
  ChevronRight,
  Copy,
  CalendarDays,
  MapIcon,
  X,
  Search,
  Info,
  GripVertical,
  CalendarPlus,
} from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

/* ── Helpers ── */

import { getLat, getLng, getClassification } from '@/lib/target-fields'


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

  const { data: targetData, isLoading: targetsLoading } = useTargets({ limit: 500 })
  const { data: activityData, isLoading: activitiesLoading } = useActivities({ dateFrom, dateTo, limit: 200 })
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

  // ── Drop handler ──
  const handleDrop = useCallback((dateStr: string) => {
    // Moving a pending assignment between days
    if (dragPending) {
      if (dragPending.sourceDate !== dateStr) {
        // Check both pending and existing activities on target day
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
      }
      setDragPending(null)
      return
    }
    // Rescheduling existing activity
    if (dragActivityId) {
      patchActivity.mutate({ id: dragActivityId, dueDate: dateStr })
      setDragActivityId(null)
      return
    }
    // Dropping targets: always drop all selected (the dragged one is included in selection)
    if (dragTargetId || selectedTargetIds.size > 0) {
      const ids = new Set(selectedTargetIds)
      if (dragTargetId) ids.add(dragTargetId)
      // Check against both pending assignments AND existing activities on this day
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
        showToast(
          dupCount === ids.size
            ? `Already scheduled — ${dupCount === 1 ? 'target is' : `all ${dupCount} targets are`} already on this day`
            : `${toAdd.length} added, ${dupCount} already on this day`,
          'warning',
        )
      }

      setDragTargetId(null)
      setSelectedTargetIds(new Set())
    }
  }, [dragPending, dragActivityId, dragTargetId, selectedTargetIds, dayAssignments, activities, patchActivity, showToast])

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

  const completedCount = stats?.byStatus?.realizat ?? 0
  const completionRate = stats?.total ? Math.round((completedCount / stats.total) * 100) : 0
  const coveragePct = coverage ? Math.round(coverage.percentage) : 0

  return (
    <>
    <ToastContainer />
    <div className="flex h-full flex-col">
      <div className="flex flex-1 min-h-0 flex-col lg:flex-row">

        {/* ═══ LEFT PANEL: Map + Target List ═══ */}
        <section className="hidden lg:flex lg:w-[45%] xl:w-[40%] flex-col border-r border-slate-200 bg-white">
          {/* Map toolbar */}
          <div className="p-3 border-b border-slate-200 flex items-center justify-between shrink-0">
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  type="text"
                  value={targetSearch}
                  onChange={(e) => setTargetSearch(e.target.value)}
                  placeholder="Search targets..."
                  className="pl-8 pr-3 py-1.5 border border-slate-300 rounded-md text-sm w-48 focus:outline-none focus:ring-1 focus:ring-teal-500 focus:border-teal-500"
                />
              </div>
              <div className="flex items-center gap-1 bg-slate-100 p-0.5 rounded border border-slate-200">
                {['a', 'b', 'c'].map((p) => (
                  <button
                    key={p}
                    onClick={() => setPriorityFilter(priorityFilter === p ? '' : p)}
                    className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
                      priorityFilter === p
                        ? 'bg-white shadow-sm text-slate-700'
                        : 'text-slate-500 hover:text-slate-700'
                    }`}
                  >
                    {p.toUpperCase()}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Map */}
          <div className="flex-1 relative min-h-0">
            <MapContainer className="h-full">
              {filteredTargets.map((t) => (
                <TargetMarker
                  key={t.id}
                  lat={getLat(t.fields)!}
                  lng={getLng(t.fields)!}
                  name={t.name}
                  priority={getClassification(t.fields)}
                  selected={selectedTargetIds.has(t.id)}
                  highlighted={hoveredTargetId === t.id}
                  onClick={() => toggleTarget(t.id)}
                  onHover={(h) => setHoveredTargetId(h ? t.id : null)}
                />
              ))}
            </MapContainer>

            {/* Map legend */}
            <div className="absolute left-4 bottom-4 bg-white/90 backdrop-blur p-2 rounded-md shadow-sm border border-slate-200 z-20 text-xs">
              <div className="font-medium text-slate-700 mb-1">Priority</div>
              <div className="flex flex-col gap-1.5">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-red-500 border border-white shadow-sm" /> A (High)
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-amber-500 border border-white shadow-sm" /> B (Med)
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-slate-400 border border-white shadow-sm" /> C (Low)
                </div>
              </div>
            </div>
          </div>

          {/* Target list panel */}
          <div className="h-2/5 bg-white border-t border-slate-200 flex flex-col shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)]">
            {/* Panel header */}
            <div className="px-4 py-2 border-b border-slate-200 bg-slate-50 flex items-center justify-between shrink-0">
              <div className="flex items-center gap-2">
                {selectedTargetIds.size > 0 && (
                  <Badge variant="primary">{selectedTargetIds.size} Selected</Badge>
                )}
                <span className="text-sm font-medium text-slate-700">
                  Targets ({filteredTargets.length})
                </span>
              </div>
              <div className="flex items-center gap-2">
                {selectedTargetIds.size > 0 && (
                  <>
                    <button onClick={clearSelection} className="text-xs font-medium text-slate-500 hover:text-slate-700">
                      Clear
                    </button>
                    <button
                      onClick={() => setBulkModalOpen(true)}
                      className="px-3 py-1.5 bg-teal-600 hover:bg-teal-700 text-white text-xs font-medium rounded shadow-sm transition-colors flex items-center gap-1"
                    >
                      <CalendarPlus size={12} />
                      Bulk Schedule
                    </button>
                  </>
                )}
              </div>
            </div>

            {/* Drag hint */}
            {selectedTargetIds.size > 0 && (
              <div className="px-4 py-1.5 bg-slate-50 border-b border-slate-100">
                <span className="text-[10px] text-slate-400">
                  <Info size={12} className="inline text-slate-400 mr-1" />You can also drag and drop targets onto the calendar
                </span>
              </div>
            )}

            {/* Target list */}
            <div className="flex-1 overflow-y-auto">
              <ul>
                {filteredTargets.map((t) => {
                  const priority = getClassification(t.fields)
                  const dotClass = priorityDot[priority] ?? priorityDot.c
                  const lastVisit = visitMap.get(t.id)
                  const days = lastVisit ? daysAgo(lastVisit) : null
                  const isSelected = selectedTargetIds.has(t.id)
                  const isHovered = hoveredTargetId === t.id

                  return (
                    <li
                      key={t.id}
                      draggable
                      role="button"
                      tabIndex={0}
                      onDragStart={(e) => {
                        e.dataTransfer.setData('text/plain', t.id)
                        setDragTargetId(t.id)
                      }}
                      onDragEnd={() => setDragTargetId(null)}
                      onClick={() => toggleTarget(t.id)}
                      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleTarget(t.id) } }}
                      onMouseEnter={() => setHoveredTargetId(t.id)}
                      onMouseLeave={() => setHoveredTargetId(null)}
                      className={`px-3 py-2 border-b border-slate-50 flex items-center gap-2 text-xs cursor-pointer transition-colors ${
                        isSelected
                          ? 'bg-teal-50 border-l-2 border-l-teal-500'
                          : isHovered
                            ? 'bg-slate-50'
                            : 'hover:bg-slate-50'
                      }`}
                    >
                      {isSelected && (
                        <GripVertical size={12} className="text-slate-400 shrink-0 cursor-grab" />
                      )}
                      <div className={`w-2.5 h-2.5 rounded-full ${dotClass} shrink-0`} />
                      <div className="min-w-0 flex-1">
                        <div className="text-sm font-medium text-slate-800 truncate">{t.name}</div>
                        <div className="text-[10px] text-slate-500 truncate">
                          {(t.fields.address as string) ?? ''}
                          {days != null && (
                            <span className="ml-1">
                              · Last visit: {days}d ago
                            </span>
                          )}
                          {days == null && <span className="ml-1">· No visits</span>}
                        </div>
                      </div>
                      <span
                        className={`text-[9px] font-bold uppercase px-1.5 py-0.5 rounded shrink-0 ${
                          priority === 'a'
                            ? 'bg-red-100 text-red-700'
                            : priority === 'b'
                              ? 'bg-amber-100 text-amber-700'
                              : 'bg-slate-100 text-slate-600'
                        }`}
                      >
                        {priority.toUpperCase()}
                      </span>
                    </li>
                  )
                })}
                {filteredTargets.length === 0 && (
                  <li className="p-4 text-center text-sm text-slate-400">No targets found</li>
                )}
              </ul>
            </div>
          </div>
        </section>

        {/* ═══ RIGHT PANEL: Calendar ═══ */}
        <section className="flex-1 flex flex-col bg-slate-50">
          {/* Calendar header */}
          <div className="px-4 py-3 border-b border-slate-200 bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-2 shrink-0 md:px-6">
            <div className="flex flex-wrap items-center gap-2 md:gap-3">
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
          <label id="field-bulk-visit-type" className="mb-1.5 block text-sm font-medium text-slate-700">Visit type</label>
          <div role="group" aria-labelledby="field-bulk-visit-type" className="flex rounded-lg border border-slate-200 overflow-hidden">
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
                    p === 'a' ? 'bg-red-100 text-red-700'
                      : p === 'b' ? 'bg-amber-100 text-amber-700'
                        : 'bg-slate-100 text-slate-600'
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
