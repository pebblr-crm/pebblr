import { createRoute } from '@tanstack/react-router'
import { useState, useMemo, useCallback, useEffect } from 'react'
import { APIProvider, Map as GoogleMap, AdvancedMarker } from '@vis.gl/react-google-maps'
import { MapPin, Check, GripVertical, FolderOpen, Save, Trash2 } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useTargets, useTargetVisitStatus, useTargetFrequencyStatus } from '../../services/targets'
import { useActivities, useBatchCreateActivities, usePatchActivity } from '../../services/activities'
import { useConfig } from '../../services/config'
import { useToast } from '../../hooks/useToast'
import { getStatusDotColor, getStatusLabel, getTypeCategory, getTypeLabel } from '@/utils/config'
import type { TenantConfig } from '@/types/config'
import type { Activity } from '@/types/activity'
import type { Target } from '@/types/target'
import { usePlannerState } from '@/contexts/planner'
import { useCollections, useCreateCollection, useDeleteCollection } from '@/services/collections'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner/map',
  component: MapPlannerPage,
})

const BUCHAREST_CENTER = { lat: 44.4268, lng: 26.1025 }
const API_KEY = import.meta.env.VITE_GOOGLE_MAPS_API_KEY ?? ''

// ── Helpers ──────────────────────────────────────────────────────────────────

function getWeekDates(baseDate: Date): Array<{ date: Date; label: string }> {
  const monday = new Date(baseDate)
  const day = monday.getDay()
  const diff = day === 0 ? -6 : 1 - day
  monday.setDate(monday.getDate() + diff)

  return Array.from({ length: 5 }, (_, i) => {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    return {
      date: d,
      label: d.toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short' }),
    }
  })
}

function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

/** Haversine distance in km between two lat/lng points. */
function haversineKm(lat1: number, lng1: number, lat2: number, lng2: number): number {
  const R = 6371
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLng = ((lng2 - lng1) * Math.PI) / 180
  const a =
    Math.sin(dLat / 2) ** 2 +
    Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLng / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function isWithinCadence(lastVisit: string | undefined, cadenceDays: number, referenceDate: Date): boolean {
  if (!lastVisit) return false
  const last = new Date(lastVisit)
  const cutoff = new Date(referenceDate)
  cutoff.setDate(cutoff.getDate() - cadenceDays)
  return last >= cutoff
}

// ── Main component ───────────────────────────────────────────────────────────

function MapPlannerPage() {
  const { data: config, isLoading: configLoading } = useConfig()
  const { data: targetsResult, isLoading: targetsLoading } = useTargets({ limit: 200 })
  const { data: visitStatus } = useTargetVisitStatus()
  const batchCreate = useBatchCreateActivities()
  const { data: collections } = useCollections()
  const createCollection = useCreateCollection()
  const deleteCollectionMut = useDeleteCollection()
  const { showToast } = useToast()
  const { state: plannerState, setWeek, setFrom } = usePlannerState()

  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [weekOffset, setWeekOffset] = useState(() => {
    if (!plannerState.week) return 0
    const target = new Date(plannerState.week + 'T00:00:00')
    const now = new Date()
    const diffDays = Math.round((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    return Math.round(diffDays / 7)
  })

  // Compute week range for activity query.
  const weekRange = useMemo(() => {
    const base = new Date()
    base.setDate(base.getDate() + weekOffset * 7)
    const days = getWeekDates(base)
    return {
      dateFrom: formatDate(days[0].date),
      dateTo: formatDate(days[days.length - 1].date),
    }
  }, [weekOffset])

  // Sync week to context for back navigation
  useEffect(() => {
    setWeek(weekRange.dateFrom)
    setFrom('map')
  }, [weekRange.dateFrom, setWeek, setFrom])

  const { data: weekActivities } = useActivities({
    dateFrom: weekRange.dateFrom,
    dateTo: weekRange.dateTo,
    limit: 200,
  })
  // dayAssignments: dateString → set of target IDs
  const [dayAssignments, setDayAssignments] = useState<Record<string, string[]>>({})
  const [dragTargetId, setDragTargetId] = useState<string | null>(null)
  const [dragActivityId, setDragActivityId] = useState<string | null>(null)
  const [hoveredTargetId, setHoveredTargetId] = useState<string | null>(null)
  const [showCadenced, setShowCadenced] = useState(false)
  const [savingCollection, setSavingCollection] = useState(false)
  const [collectionName, setCollectionName] = useState('')
  const [dragCollectionId, setDragCollectionId] = useState<string | null>(null)

  const patchActivity = usePatchActivity()

  const cadenceDays = config?.rules?.visit_cadence_days ?? 21
  const targets = useMemo(() => targetsResult?.items ?? [], [targetsResult?.items])

  // Build visit status lookup
  const visitStatusMap = useMemo(() => {
    const map = new Map<string, string>()
    if (visitStatus) {
      for (const vs of visitStatus) {
        map.set(vs.targetId, vs.lastVisitDate)
      }
    }
    return map
  }, [visitStatus])

  // Frequency compliance for current month — drives marker color intensity
  const currentPeriod = useMemo(() => {
    const base = new Date()
    base.setDate(base.getDate() + weekOffset * 7)
    return `${base.getFullYear()}-${String(base.getMonth() + 1).padStart(2, '0')}`
  }, [weekOffset])
  const { data: frequencyData } = useTargetFrequencyStatus(currentPeriod)
  const frequencyMap = useMemo(() => {
    const map = new Map<string, number>()
    if (frequencyData) {
      for (const item of frequencyData) {
        map.set(item.targetId, item.compliance)
      }
    }
    return map
  }, [frequencyData])

  // Targets with geo coordinates
  const geoTargets = useMemo(
    () => targets.filter((t) => t.fields?.lat != null && t.fields?.lng != null),
    [targets],
  )

  // Already assigned target IDs (across all days)
  const assignedIds = useMemo(() => {
    const ids = new Set<string>()
    for (const arr of Object.values(dayAssignments)) {
      for (const id of arr) ids.add(id)
    }
    return ids
  }, [dayAssignments])

  const weekDates = useMemo(() => {
    const base = new Date()
    base.setDate(base.getDate() + weekOffset * 7)
    return getWeekDates(base)
  }, [weekOffset])

  // Centroid of selected targets — used to sort available targets by proximity.
  const selectionCentroid = useMemo(() => {
    if (selectedIds.size === 0) return null
    let sumLat = 0
    let sumLng = 0
    let count = 0
    for (const id of selectedIds) {
      const t = targets.find((x) => x.id === id)
      if (t?.fields?.lat != null && t?.fields?.lng != null) {
        sumLat += t.fields.lat as number
        sumLng += t.fields.lng as number
        count++
      }
    }
    if (count === 0) return null
    return { lat: sumLat / count, lng: sumLng / count }
  }, [selectedIds, targets])

  // Available targets with computed distances, sorted by proximity when a selection exists.
  const availableWithDistance = useMemo(() => {
    const available = geoTargets.filter((t) => !selectedIds.has(t.id) && !assignedIds.has(t.id))

    const items = available.map((t) => {
      const dist = selectionCentroid
        ? haversineKm(selectionCentroid.lat, selectionCentroid.lng, t.fields.lat as number, t.fields.lng as number)
        : undefined
      const isCadenced = isWithinCadence(visitStatusMap.get(t.id), cadenceDays, weekDates[0].date)
      return { target: t, distance: dist, isCadenced }
    })

    if (selectionCentroid) {
      items.sort((a, b) => (a.distance ?? Infinity) - (b.distance ?? Infinity))
    }

    return items
  }, [geoTargets, selectedIds, assignedIds, selectionCentroid, visitStatusMap, cadenceDays, weekDates])

  // Existing activities grouped by date string.
  const activitiesByDate = useMemo(() => {
    const map = new Map<string, Activity[]>()
    for (const a of weekActivities?.items ?? []) {
      const dateStr = a.dueDate.slice(0, 10)
      const arr = map.get(dateStr) ?? []
      arr.push(a)
      map.set(dateStr, arr)
    }
    return map
  }, [weekActivities])

  const toggleTarget = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  // Handle collection drop onto a day slot
  const handleCollectionDrop = useCallback((dateStr: string) => {
    if (!dragCollectionId) return
    const col = collections?.find((c) => c.id === dragCollectionId)
    if (!col) return
    setDayAssignments((prev) => {
      const arr = prev[dateStr] ?? []
      const existing = new Set(arr)
      const toAdd = col.targetIds.filter((id) => !existing.has(id))
      if (toAdd.length === 0) return prev
      return { ...prev, [dateStr]: [...arr, ...toAdd] }
    })
    setDragCollectionId(null)
  }, [dragCollectionId, collections])

  const handleDrop = useCallback((dateStr: string) => {
    // If an existing activity is being dragged, reschedule it.
    if (dragActivityId) {
      patchActivity.mutate(
        { id: dragActivityId, dueDate: dateStr },
        {
          onSuccess: () => showToast('Activity rescheduled.', 'success'),
          onError: () => showToast('Failed to reschedule activity.'),
        },
      )
      setDragActivityId(null)
      return
    }
    // If a collection is being dragged, assign all its targets.
    if (dragCollectionId) {
      handleCollectionDrop(dateStr)
      return
    }
    // Otherwise drop all selected targets onto the day.
    setDayAssignments((prev) => {
      const arr = prev[dateStr] ?? []
      const existing = new Set(arr)
      const toAdd = Array.from(selectedIds).filter((id) => !existing.has(id))
      if (toAdd.length === 0) return prev
      return { ...prev, [dateStr]: [...arr, ...toAdd] }
    })
    setSelectedIds(new Set())
  }, [selectedIds, dragActivityId, dragCollectionId, patchActivity, showToast, handleCollectionDrop])

  const removeFromDay = useCallback((dateStr: string, targetId: string) => {
    setDayAssignments((prev) => {
      const arr = (prev[dateStr] ?? []).filter((id) => id !== targetId)
      const next = { ...prev }
      if (arr.length === 0) delete next[dateStr]
      else next[dateStr] = arr
      return next
    })
  }, [])

  const handleCreateActivities = useCallback(async () => {
    const items: Array<{ targetId: string; dueDate: string }> = []
    for (const [date, ids] of Object.entries(dayAssignments)) {
      for (const id of ids) {
        items.push({ targetId: id, dueDate: date })
      }
    }
    if (items.length === 0) return

    try {
      const result = await batchCreate.mutateAsync(items)
      const createdCount = result.created?.length ?? 0
      const errorCount = result.errors?.length ?? 0
      if (errorCount > 0) {
        showToast(`Created ${createdCount} activities, ${errorCount} failed.`)
      } else {
        showToast(`Created ${createdCount} activities.`, 'success')
      }
      setDayAssignments({})
      setSelectedIds(new Set())
    } catch {
      showToast('Failed to create activities.')
    }
  }, [dayAssignments, batchCreate, showToast])

  const handleSaveCollection = useCallback(() => {
    if (!collectionName.trim() || selectedIds.size === 0) return
    createCollection.mutate(
      { name: collectionName.trim(), targetIds: Array.from(selectedIds) },
      {
        onSuccess: () => {
          showToast(`Collection "${collectionName.trim()}" saved.`, 'success')
          setSavingCollection(false)
          setCollectionName('')
        },
        onError: () => showToast('Failed to save collection.'),
      },
    )
  }, [collectionName, selectedIds, createCollection, showToast])

  const handleLoadCollection = useCallback((targetIds: string[]) => {
    setSelectedIds(new Set(targetIds))
  }, [])

  const handleDeleteCollection = useCallback((id: string, name: string) => {
    deleteCollectionMut.mutate(id, {
      onSuccess: () => showToast(`Collection "${name}" deleted.`, 'success'),
      onError: () => showToast('Failed to delete collection.'),
    })
  }, [deleteCollectionMut, showToast])

  if (configLoading || targetsLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <LoadingSpinner size="lg" label="Loading map planner..." />
      </div>
    )
  }

  if (!API_KEY) {
    return (
      <div className="p-8 text-center text-on-surface-variant">
        Google Maps API key is not configured. Set VITE_GOOGLE_MAPS_API_KEY in your environment.
      </div>
    )
  }

  const targetMap = new Map(targets.map((t) => [t.id, t]))

  const totalAssigned = Object.values(dayAssignments).reduce((sum, arr) => sum + arr.length, 0)

  return (
    <APIProvider apiKey={API_KEY}>
      <div className="flex flex-col h-[calc(100vh-64px)]">
        {/* Top section: map + target list */}
        <div className="flex flex-1 min-h-0">
          {/* Map (3/4) */}
          <div className="w-3/4 relative">
            <GoogleMap
              defaultCenter={BUCHAREST_CENTER}
              defaultZoom={12}
              mapId="pebblr-map-planner"
              gestureHandling="greedy"
              disableDefaultUI={false}
              className="w-full h-full"
            >
              {geoTargets.map((target) => {
                const lat = target.fields.lat as number
                const lng = target.fields.lng as number
                const isSelected = selectedIds.has(target.id)
                const isAssigned = assignedIds.has(target.id)
                const lastVisit = visitStatusMap.get(target.id)
                const isCadenced = isWithinCadence(lastVisit, cadenceDays, weekDates[0].date)
                const isHovered = hoveredTargetId === target.id
                const compliance = frequencyMap.get(target.id)

                const cadenceLocked = isCadenced && !showCadenced

                if (cadenceLocked && !isSelected && !isAssigned) return null

                // Frequency-based marker color: red = behind, amber = partial, green = on track
                const frequencyMarkerColor =
                  compliance == null
                    ? target.targetType === 'pharmacy'
                      ? 'bg-amber-100 border-amber-400 text-amber-700'
                      : 'bg-blue-100 border-blue-400 text-blue-700'
                    : compliance >= 80
                      ? 'bg-emerald-200 border-emerald-500 text-emerald-800'
                      : compliance >= 50
                        ? 'bg-amber-100 border-amber-400 text-amber-700'
                        : 'bg-red-200 border-red-500 text-red-800'

                return (
                  <AdvancedMarker
                    key={target.id}
                    position={{ lat, lng }}
                    onClick={() => {
                      if (!cadenceLocked && !isAssigned) toggleTarget(target.id)
                    }}
                    title={`${target.name}${compliance != null ? ` (${Math.round(compliance)}% compliance)` : ''}${isCadenced ? ' (recently visited)' : ''}`}
                  >
                    <div
                      onMouseEnter={() => setHoveredTargetId(target.id)}
                      onMouseLeave={() => setHoveredTargetId(null)}
                      className={`rounded-full flex items-center justify-center border-2 transition-all cursor-pointer ${
                        isHovered ? 'w-9 h-9 shadow-lg ring-2 ring-primary/40' : 'w-7 h-7'
                      } ${
                        isAssigned
                          ? 'bg-green-500 border-green-700 text-white'
                          : isSelected
                            ? 'bg-primary border-primary text-white'
                            : cadenceLocked
                              ? 'bg-slate-200 border-slate-300 text-slate-400 cursor-not-allowed'
                              : isCadenced
                                ? 'bg-slate-100 border-slate-300 text-slate-500 border-dashed'
                                : frequencyMarkerColor
                      }`}
                    >
                      {isAssigned ? (
                        <Check className="w-3.5 h-3.5" />
                      ) : isSelected ? (
                        <Check className="w-3.5 h-3.5" />
                      ) : (
                        <MapPin className="w-3.5 h-3.5" />
                      )}
                    </div>
                  </AdvancedMarker>
                )
              })}
            </GoogleMap>
          </div>

          {/* Target list (1/4) */}
          <div className="w-1/4 border-l border-slate-200 flex flex-col bg-white">
            {/* Collections section */}
            {collections && collections.length > 0 && (
              <div className="border-b border-slate-100">
                <div className="px-3 py-1.5">
                  <span className="text-[10px] uppercase tracking-widest text-slate-400 font-bold">
                    Collections
                  </span>
                </div>
                {collections.map((col) => (
                  <div
                    key={col.id}
                    draggable
                    onDragStart={() => setDragCollectionId(col.id)}
                    onDragEnd={() => setDragCollectionId(null)}
                    className="flex items-center gap-2 px-3 py-1.5 hover:bg-slate-50 cursor-grab text-xs border-b border-slate-50"
                  >
                    <FolderOpen className="w-3 h-3 text-slate-400 shrink-0" />
                    <button
                      type="button"
                      onClick={() => handleLoadCollection(col.targetIds)}
                      className="flex-1 text-left font-medium text-on-surface truncate hover:text-primary"
                    >
                      {col.name}
                    </button>
                    <span className="text-[9px] text-slate-400 shrink-0">{col.targetIds.length}</span>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteCollection(col.id, col.name)
                      }}
                      className="text-slate-300 hover:text-error shrink-0"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            <div className="p-3 border-b border-slate-100">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-bold text-on-surface">
                  Selected ({selectedIds.size})
                </h2>
                {selectedIds.size > 0 && !savingCollection && (
                  <button
                    type="button"
                    onClick={() => setSavingCollection(true)}
                    className="text-[10px] text-primary font-bold flex items-center gap-1 hover:opacity-70"
                  >
                    <Save className="w-3 h-3" />
                    Save
                  </button>
                )}
              </div>
              {savingCollection && (
                <div className="mt-2 flex gap-1">
                  <input
                    type="text"
                    value={collectionName}
                    onChange={(e) => setCollectionName(e.target.value)}
                    placeholder="Collection name..."
                    className="flex-1 px-2 py-1 text-xs border border-slate-200 rounded"
                    onKeyDown={(e) => { if (e.key === 'Enter') handleSaveCollection() }}
                    autoFocus
                  />
                  <button
                    type="button"
                    onClick={handleSaveCollection}
                    disabled={!collectionName.trim() || createCollection.isPending}
                    className="px-2 py-1 text-xs font-bold text-white bg-primary rounded disabled:opacity-50"
                  >
                    {createCollection.isPending ? '...' : 'Save'}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setSavingCollection(false); setCollectionName('') }}
                    className="px-2 py-1 text-xs text-slate-400 hover:text-slate-600"
                  >
                    ×
                  </button>
                </div>
              )}
            </div>
            <div className="flex-1 overflow-y-auto">
              {/* Selected targets first */}
              {Array.from(selectedIds).map((id) => {
                const target = targetMap.get(id)
                if (!target) return null
                return (
                  <TargetListItem
                    key={id}
                    target={target}
                    lastVisit={visitStatusMap.get(id)}
                    isSelected
                    isHovered={hoveredTargetId === id}
                    compliance={frequencyMap.get(id)}
                    onHover={(h) => setHoveredTargetId(h ? id : null)}
                    onToggle={() => toggleTarget(id)}
                    onDragStart={() => setDragTargetId(id)}
                    onDragEnd={() => setDragTargetId(null)}
                  />
                )
              })}
              {/* Available targets sorted by proximity to selection */}
              {(() => {
                const open = availableWithDistance.filter((x) => !x.isCadenced)
                const cadenced = availableWithDistance.filter((x) => x.isCadenced)

                return (
                  <>
                    {(selectedIds.size > 0 || open.length > 0) && open.length > 0 && (
                      <div className="border-t border-slate-100 px-3 py-1">
                        <span className="text-[10px] uppercase tracking-widest text-slate-400 font-bold">
                          {selectionCentroid ? 'Nearby' : 'Available'}
                        </span>
                      </div>
                    )}
                    {open.map(({ target, distance }) => (
                      <TargetListItem
                        key={target.id}
                        target={target}
                        lastVisit={visitStatusMap.get(target.id)}
                        isSelected={false}
                        isHovered={hoveredTargetId === target.id}
                        distanceKm={distance}
                        compliance={frequencyMap.get(target.id)}
                        onHover={(h) => setHoveredTargetId(h ? target.id : null)}
                        onToggle={() => toggleTarget(target.id)}
                        onDragStart={() => setDragTargetId(target.id)}
                        onDragEnd={() => setDragTargetId(null)}
                      />
                    ))}

                    {cadenced.length > 0 && (
                      <CadencedSection
                        items={cadenced}
                        showCadenced={showCadenced}
                        setShowCadenced={setShowCadenced}
                        visitStatusMap={visitStatusMap}
                        frequencyMap={frequencyMap}
                        hoveredTargetId={hoveredTargetId}
                        setHoveredTargetId={setHoveredTargetId}
                        toggleTarget={toggleTarget}
                        setDragTargetId={setDragTargetId}
                      />
                    )}
                  </>
                )
              })()}
            </div>
          </div>
        </div>

        {/* Bottom: Week calendar drop zones */}
        <div className="border-t border-slate-200 bg-slate-50 p-3">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setWeekOffset((w) => w - 1)}
                className="px-2 py-1 text-xs rounded border border-slate-200 hover:bg-slate-100"
              >
                Prev
              </button>
              <button
                onClick={() => setWeekOffset(0)}
                className="px-2 py-1 text-xs rounded border border-slate-200 hover:bg-slate-100"
              >
                This week
              </button>
              <button
                onClick={() => setWeekOffset((w) => w + 1)}
                className="px-2 py-1 text-xs rounded border border-slate-200 hover:bg-slate-100"
              >
                Next
              </button>
            </div>
            {totalAssigned > 0 && (
              <button
                onClick={() => void handleCreateActivities()}
                disabled={batchCreate.isPending}
                className="px-4 py-2 text-sm font-bold text-white bg-primary rounded-xl hover:bg-primary/90 disabled:opacity-50"
              >
                {batchCreate.isPending
                  ? 'Creating...'
                  : `Create ${totalAssigned} activities`}
              </button>
            )}
          </div>
          <div className="grid grid-cols-5 gap-2">
            {weekDates.map(({ date, label }) => {
              const dateStr = formatDate(date)
              const dayTargets = dayAssignments[dateStr] ?? []
              return (
                <DayDropZone
                  key={dateStr}
                  label={label}
                  targetIds={dayTargets}
                  existingActivities={activitiesByDate.get(dateStr) ?? []}
                  targetMap={targetMap}
                  config={config!}
                  dragTargetId={dragTargetId}
                  dragActivityId={dragActivityId}
                  dragCollectionId={dragCollectionId}
                  onDrop={() => handleDrop(dateStr)}
                  onRemove={(id) => removeFromDay(dateStr, id)}
                  onActivityDragStart={(id) => setDragActivityId(id)}
                  onActivityDragEnd={() => setDragActivityId(null)}
                />
              )
            })}
          </div>
        </div>
      </div>
    </APIProvider>
  )
}

// ── Cadenced targets rollout ──────────────────────────────────────────────────

interface CadencedSectionProps {
  items: Array<{ target: Target; distance?: number }>
  showCadenced: boolean
  setShowCadenced: (v: boolean) => void
  visitStatusMap: Map<string, string>
  frequencyMap: Map<string, number>
  hoveredTargetId: string | null
  setHoveredTargetId: (id: string | null) => void
  toggleTarget: (id: string) => void
  setDragTargetId: (id: string | null) => void
}

function CadencedSection({
  items,
  showCadenced,
  setShowCadenced,
  visitStatusMap,
  frequencyMap,
  hoveredTargetId,
  setHoveredTargetId,
  toggleTarget,
  setDragTargetId,
}: CadencedSectionProps) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="border-t border-slate-100">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center justify-between px-3 py-1.5 hover:bg-slate-50 transition-colors"
      >
        <span className="text-[10px] uppercase tracking-widest text-slate-400 font-bold">
          Recently visited ({items.length})
        </span>
        <span className="text-[10px] text-slate-400">{expanded ? '▲' : '▼'}</span>
      </button>
      {expanded && (
        <>
          <label className="flex items-center gap-1.5 px-3 py-1 text-[10px] text-slate-500 cursor-pointer">
            <input
              type="checkbox"
              checked={showCadenced}
              onChange={(e) => setShowCadenced(e.target.checked)}
              className="rounded border-slate-300"
            />
            Allow scheduling
          </label>
          {items.map(({ target, distance }) => (
            <TargetListItem
              key={target.id}
              target={target}
              lastVisit={visitStatusMap.get(target.id)}
              isSelected={false}
              isCadenced={!showCadenced}
              isHovered={hoveredTargetId === target.id}
              distanceKm={distance}
              compliance={frequencyMap.get(target.id)}
              onHover={(h) => setHoveredTargetId(h ? target.id : null)}
              onToggle={() => {
                if (showCadenced) toggleTarget(target.id)
              }}
              onDragStart={() => setDragTargetId(target.id)}
              onDragEnd={() => setDragTargetId(null)}
            />
          ))}
        </>
      )}
    </div>
  )
}

// ── Target list item ─────────────────────────────────────────────────────────

interface TargetListItemProps {
  target: Target
  lastVisit?: string
  isSelected: boolean
  isCadenced?: boolean
  isHovered?: boolean
  distanceKm?: number
  compliance?: number
  onHover?: (hovered: boolean) => void
  onToggle: () => void
  onDragStart: () => void
  onDragEnd: () => void
}

function TargetListItem({
  target,
  lastVisit,
  isSelected,
  isCadenced,
  isHovered,
  distanceKm,
  compliance,
  onHover,
  onToggle,
  onDragStart,
  onDragEnd,
}: TargetListItemProps) {
  return (
    <div
      draggable={isSelected}
      onDragStart={(e) => {
        e.dataTransfer.setData('text/plain', target.id)
        onDragStart()
      }}
      onDragEnd={onDragEnd}
      onClick={onToggle}
      onMouseEnter={() => onHover?.(true)}
      onMouseLeave={() => onHover?.(false)}
      className={`flex items-center gap-2 px-3 py-2 border-b border-slate-50 text-xs cursor-pointer transition-colors ${
        isHovered
          ? 'bg-primary/10 ring-1 ring-inset ring-primary/30'
          : isCadenced
            ? 'opacity-40 cursor-not-allowed'
            : isSelected
              ? 'bg-primary-fixed'
              : 'hover:bg-slate-50'
      }`}
    >
      {isSelected && <GripVertical className="w-3 h-3 text-slate-400 shrink-0 cursor-grab" />}
      <div className="min-w-0 flex-1">
        <p className="font-medium text-on-surface truncate">{target.name}</p>
        <p className="text-[10px] text-slate-400 truncate">
          {(target.fields?.address as string) ?? ''}
          {lastVisit && (
            <span className="ml-1">
              — last: {new Date(lastVisit).toLocaleDateString()}
            </span>
          )}
        </p>
      </div>
      <div className="flex flex-col items-end shrink-0 gap-0.5">
        <span
          className={`text-[9px] font-bold uppercase px-1.5 py-0.5 rounded ${
            target.targetType === 'pharmacy'
              ? 'bg-amber-100 text-amber-700'
              : 'bg-blue-100 text-blue-700'
          }`}
        >
          {(target.fields?.potential as string) ?? target.targetType}
        </span>
        {compliance != null && (
          <span
            className={`text-[9px] font-bold px-1 py-0.5 rounded ${
              compliance >= 80
                ? 'text-emerald-600'
                : compliance >= 50
                  ? 'text-amber-600'
                  : 'text-red-600'
            }`}
          >
            {Math.round(compliance)}%
          </span>
        )}
        {distanceKm != null && (
          <span className="text-[9px] text-slate-400">
            {distanceKm < 1 ? `${Math.round(distanceKm * 1000)}m` : `${distanceKm.toFixed(1)}km`}
          </span>
        )}
      </div>
    </div>
  )
}

// ── Day drop zone ────────────────────────────────────────────────────────────

interface DayDropZoneProps {
  label: string
  targetIds: string[]
  existingActivities: Activity[]
  targetMap: Map<string, Target>
  config: TenantConfig
  dragTargetId: string | null
  dragActivityId: string | null
  dragCollectionId: string | null
  onDrop: () => void
  onRemove: (targetId: string) => void
  onActivityDragStart: (activityId: string) => void
  onActivityDragEnd: () => void
}

function DayDropZone({
  label,
  targetIds,
  existingActivities,
  targetMap,
  config,
  dragTargetId,
  dragActivityId,
  dragCollectionId,
  onDrop,
  onRemove,
  onActivityDragStart,
  onActivityDragEnd,
}: DayDropZoneProps) {
  const [dragOver, setDragOver] = useState(false)
  const totalCount = existingActivities.length + targetIds.length
  const ac = config.activities

  // Closed statuses cannot be rescheduled.
  const closedStatuses = new Set(
    ac.statuses.filter((s) => s.submittable).map((s) => s.key),
  )

  return (
    <div
      onDragOver={(e) => {
        e.preventDefault()
        setDragOver(true)
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={(e) => {
        e.preventDefault()
        setDragOver(false)
        if (dragTargetId || dragActivityId || dragCollectionId) onDrop()
      }}
      className={`rounded-lg border-2 border-dashed p-1.5 min-h-[80px] transition-colors ${
        dragOver
          ? 'border-primary bg-primary/5'
          : 'border-slate-200 bg-white'
      }`}
    >
      <div className="flex items-center justify-between mb-0.5">
        <p className="text-[10px] font-bold text-slate-500">{label}</p>
        <span className={`text-[9px] font-bold px-1 py-0.5 rounded ${totalCount > 0 ? 'bg-slate-100 text-slate-600' : 'text-slate-400'}`}>{totalCount}</span>
      </div>

      <div className="space-y-0.5 max-h-[120px] overflow-y-auto">
        {/* Existing scheduled activities */}
        {existingActivities.map((a) => {
          const dotColor = getStatusDotColor(ac, a.status)
          const statusLabel = getStatusLabel(ac, a.status)
          const category = getTypeCategory(ac, a.activityType)
          const bgColor = category === 'field' ? 'bg-amber-50' : 'bg-blue-50'
          const textColor = category === 'field' ? 'text-amber-900' : 'text-blue-900'
          const isClosed = closedStatuses.has(a.status) || Boolean(a.submittedAt)
          const canDrag = !isClosed

          return (
            <div
              key={a.id}
              draggable={canDrag}
              onDragStart={() => { if (canDrag) onActivityDragStart(a.id) }}
              onDragEnd={onActivityDragEnd}
              className={`flex items-center gap-1 ${bgColor} rounded px-1.5 py-0.5 text-[10px] ${textColor} ${
                canDrag ? 'cursor-grab' : ''
              }`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${dotColor} shrink-0`} />
              <span className="truncate flex-1">
                {a.targetName
                  ? `${getTypeLabel(ac, a.activityType)} — ${a.targetName}`
                  : getTypeLabel(ac, a.activityType)}
              </span>
              <span className="text-[8px] opacity-60 shrink-0">{statusLabel}</span>
            </div>
          )
        })}

        {/* Newly assigned targets (pending creation) */}
        {targetIds.map((id) => {
          const t = targetMap.get(id)
          if (!t) return null
          return (
            <div
              key={id}
              className="flex items-center gap-1 bg-primary/10 rounded px-1.5 py-0.5 text-[10px] text-primary"
            >
              <span className="w-1.5 h-1.5 rounded-full bg-primary shrink-0" />
              <span className="truncate flex-1">{t.name}</span>
              <button
                type="button"
                onClick={() => onRemove(id)}
                className="text-primary/50 hover:text-error shrink-0"
              >
                ×
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
