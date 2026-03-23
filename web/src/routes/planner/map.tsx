import { createRoute } from '@tanstack/react-router'
import { useState, useMemo, useCallback } from 'react'
import { APIProvider, Map as GoogleMap, AdvancedMarker } from '@vis.gl/react-google-maps'
import { MapPin, Check, GripVertical } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useTargets, useTargetVisitStatus } from '../../services/targets'
import { useBatchCreateActivities } from '../../services/activities'
import { useConfig } from '../../services/config'
import { useToast } from '../../hooks/useToast'
import type { Target } from '@/types/target'

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

function isWithinCadence(lastVisit: string | undefined, cadenceDays: number): boolean {
  if (!lastVisit) return false
  const last = new Date(lastVisit)
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - cadenceDays)
  return last >= cutoff
}

// ── Main component ───────────────────────────────────────────────────────────

function MapPlannerPage() {
  const { data: config, isLoading: configLoading } = useConfig()
  const { data: targetsResult, isLoading: targetsLoading } = useTargets({ limit: 200 })
  const { data: visitStatus } = useTargetVisitStatus()
  const batchCreate = useBatchCreateActivities()
  const { showToast } = useToast()

  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [weekOffset, setWeekOffset] = useState(0)
  // dayAssignments: dateString → set of target IDs
  const [dayAssignments, setDayAssignments] = useState<Record<string, string[]>>({})
  const [dragTargetId, setDragTargetId] = useState<string | null>(null)

  const cadenceDays = config?.rules?.visit_cadence_days ?? 21
  const targets = targetsResult?.items ?? []

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

  const toggleTarget = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const handleDrop = useCallback((_dateStr: string, _targetId: string) => {
    // Drop all selected targets onto the day, not just the dragged one.
    setDayAssignments((prev) => {
      const arr = prev[_dateStr] ?? []
      const existing = new Set(arr)
      const toAdd = Array.from(selectedIds).filter((id) => !existing.has(id))
      if (toAdd.length === 0) return prev
      return { ...prev, [_dateStr]: [...arr, ...toAdd] }
    })
    setSelectedIds(new Set())
  }, [selectedIds])

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
        showToast(`Created ${createdCount} activities.`)
      }
      setDayAssignments({})
      setSelectedIds(new Set())
    } catch {
      showToast('Failed to create activities.')
    }
  }, [dayAssignments, batchCreate, showToast])

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
                const isCadenced = isWithinCadence(lastVisit, cadenceDays)

                return (
                  <AdvancedMarker
                    key={target.id}
                    position={{ lat, lng }}
                    onClick={() => {
                      if (!isCadenced && !isAssigned) toggleTarget(target.id)
                    }}
                    title={`${target.name}${isCadenced ? ' (recently visited)' : ''}`}
                  >
                    <div
                      className={`w-7 h-7 rounded-full flex items-center justify-center border-2 transition-colors cursor-pointer ${
                        isAssigned
                          ? 'bg-green-500 border-green-700 text-white'
                          : isSelected
                            ? 'bg-primary border-primary text-white'
                            : isCadenced
                              ? 'bg-slate-200 border-slate-300 text-slate-400 cursor-not-allowed'
                              : target.targetType === 'pharmacy'
                                ? 'bg-amber-100 border-amber-400 text-amber-700'
                                : 'bg-blue-100 border-blue-400 text-blue-700'
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
            <div className="p-3 border-b border-slate-100">
              <h2 className="text-sm font-bold text-on-surface">
                Selected ({selectedIds.size})
              </h2>
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
                    onToggle={() => toggleTarget(id)}
                    onDragStart={() => setDragTargetId(id)}
                    onDragEnd={() => setDragTargetId(null)}
                  />
                )
              })}
              {selectedIds.size > 0 && geoTargets.length > selectedIds.size && (
                <div className="border-t border-slate-100 px-3 py-1">
                  <span className="text-[10px] uppercase tracking-widest text-slate-400 font-bold">
                    Available
                  </span>
                </div>
              )}
              {/* Unselected, non-cadenced targets */}
              {geoTargets
                .filter((t) => !selectedIds.has(t.id) && !assignedIds.has(t.id))
                .map((target) => {
                  const lastVisit = visitStatusMap.get(target.id)
                  const isCadenced = isWithinCadence(lastVisit, cadenceDays)
                  return (
                    <TargetListItem
                      key={target.id}
                      target={target}
                      lastVisit={lastVisit}
                      isSelected={false}
                      isCadenced={isCadenced}
                      onToggle={() => {
                        if (!isCadenced) toggleTarget(target.id)
                      }}
                      onDragStart={() => setDragTargetId(target.id)}
                      onDragEnd={() => setDragTargetId(null)}
                    />
                  )
                })}
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
                  targetMap={targetMap}
                  dragTargetId={dragTargetId}
                  onDrop={(id) => handleDrop(dateStr, id)}
                  onRemove={(id) => removeFromDay(dateStr, id)}
                />
              )
            })}
          </div>
        </div>
      </div>
    </APIProvider>
  )
}

// ── Target list item ─────────────────────────────────────────────────────────

interface TargetListItemProps {
  target: Target
  lastVisit?: string
  isSelected: boolean
  isCadenced?: boolean
  onToggle: () => void
  onDragStart: () => void
  onDragEnd: () => void
}

function TargetListItem({
  target,
  lastVisit,
  isSelected,
  isCadenced,
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
      className={`flex items-center gap-2 px-3 py-2 border-b border-slate-50 text-xs cursor-pointer transition-colors ${
        isCadenced
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
              — last visit: {new Date(lastVisit).toLocaleDateString()}
            </span>
          )}
        </p>
      </div>
      <span
        className={`text-[9px] font-bold uppercase px-1.5 py-0.5 rounded ${
          target.targetType === 'pharmacy'
            ? 'bg-amber-100 text-amber-700'
            : 'bg-blue-100 text-blue-700'
        }`}
      >
        {(target.fields?.potential as string) ?? target.targetType}
      </span>
    </div>
  )
}

// ── Day drop zone ────────────────────────────────────────────────────────────

interface DayDropZoneProps {
  label: string
  targetIds: string[]
  targetMap: Map<string, Target>
  dragTargetId: string | null
  onDrop: (targetId: string) => void
  onRemove: (targetId: string) => void
}

function DayDropZone({
  label,
  targetIds,
  targetMap,
  dragTargetId,
  onDrop,
  onRemove,
}: DayDropZoneProps) {
  const [dragOver, setDragOver] = useState(false)

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
        const id = dragTargetId ?? e.dataTransfer.getData('text/plain')
        if (id) onDrop(id)
      }}
      className={`rounded-lg border-2 border-dashed p-2 min-h-[100px] transition-colors ${
        dragOver
          ? 'border-primary bg-primary/5'
          : 'border-slate-200 bg-white'
      }`}
    >
      <p className="text-[10px] font-bold text-slate-500 mb-1">{label}</p>
      <p className="text-[9px] text-slate-400 mb-1">{targetIds.length} targets</p>
      <div className="space-y-1">
        {targetIds.map((id) => {
          const t = targetMap.get(id)
          if (!t) return null
          return (
            <div
              key={id}
              className="flex items-center gap-1 bg-slate-50 rounded px-1.5 py-0.5 text-[10px]"
            >
              <span className="truncate flex-1">{t.name}</span>
              <button
                type="button"
                onClick={() => onRemove(id)}
                className="text-slate-400 hover:text-error shrink-0"
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
