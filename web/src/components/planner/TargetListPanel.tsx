import { useCallback } from 'react'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { Badge } from '@/components/ui/Badge'
import { priorityDot } from '@/lib/styles'
import { daysAgo } from '@/lib/helpers'
import { getLat, getLng, getClassification } from '@/lib/target-fields'
import { Search, Info, GripVertical, CalendarPlus } from 'lucide-react'
import type { Target } from '@/types/target'

export interface TargetListPanelProps {
  readonly filteredTargets: Target[]
  readonly selectedTargetIds: Set<string>
  readonly hoveredTargetId: string | null
  readonly targetSearch: string
  readonly priorityFilter: string
  readonly visitMap: Map<string, string>
  readonly onTargetSearchChange: (value: string) => void
  readonly onPriorityFilterChange: (value: string) => void
  readonly onToggleTarget: (id: string) => void
  readonly onClearSelection: () => void
  readonly onHoverTarget: (id: string | null) => void
  readonly onDragTargetStart: (id: string) => void
  readonly onDragTargetEnd: () => void
  readonly onBulkSchedule: () => void
}

export function TargetListPanel({
  filteredTargets,
  selectedTargetIds,
  hoveredTargetId,
  targetSearch,
  priorityFilter,
  visitMap,
  onTargetSearchChange,
  onPriorityFilterChange,
  onToggleTarget,
  onClearSelection,
  onHoverTarget,
  onDragTargetStart,
  onDragTargetEnd,
  onBulkSchedule,
}: TargetListPanelProps) {
  const handleDragStart = useCallback(
    (e: React.DragEvent, id: string) => {
      e.dataTransfer.setData('text/plain', id)
      onDragTargetStart(id)
    },
    [onDragTargetStart],
  )

  return (
    <section className="hidden lg:flex lg:w-[45%] xl:w-[40%] flex-col border-r border-slate-200 bg-white">
      {/* Map toolbar */}
      <div className="p-3 border-b border-slate-200 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              value={targetSearch}
              onChange={(e) => onTargetSearchChange(e.target.value)}
              placeholder="Search targets..."
              className="pl-8 pr-3 py-1.5 border border-slate-300 rounded-md text-sm w-48 focus:outline-none focus:ring-1 focus:ring-teal-500 focus:border-teal-500"
            />
          </div>
          <div className="flex items-center gap-1 bg-slate-100 p-0.5 rounded border border-slate-200">
            {['a', 'b', 'c'].map((p) => (
              <button
                key={p}
                onClick={() => onPriorityFilterChange(priorityFilter === p ? '' : p)}
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
          {filteredTargets.map((t) => {
            const lat = getLat(t.fields)
            const lng = getLng(t.fields)
            if (lat == null || lng == null) return null
            return (
              <TargetMarker
                key={t.id}
                lat={lat}
                lng={lng}
                name={t.name}
                priority={getClassification(t.fields)}
                selected={selectedTargetIds.has(t.id)}
                highlighted={hoveredTargetId === t.id}
                onClick={() => onToggleTarget(t.id)}
                onHover={(h) => onHoverTarget(h ? t.id : null)}
              />
            )
          })}
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
                <button onClick={onClearSelection} className="text-xs font-medium text-slate-500 hover:text-slate-700">
                  Clear
                </button>
                <button
                  onClick={onBulkSchedule}
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

              let stateClass: string
              if (isSelected) stateClass = 'bg-teal-50 border-l-2 border-l-teal-500'
              else if (isHovered) stateClass = 'bg-slate-50'
              else stateClass = 'hover:bg-slate-50'

              return (
                <li
                  key={t.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, t.id)}
                  onDragEnd={onDragTargetEnd}
                  onMouseEnter={() => onHoverTarget(t.id)}
                  onMouseLeave={() => onHoverTarget(null)}
                  className={`px-3 py-2 border-b border-slate-50 flex items-center gap-2 text-xs cursor-pointer transition-colors ${stateClass}`}
                >
                <button
                  type="button"
                  tabIndex={0}
                  onClick={() => onToggleTarget(t.id)}
                  className="px-3 py-2 flex items-center gap-2 w-full text-left cursor-pointer"
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
                      { a: 'bg-red-100 text-red-700', b: 'bg-amber-100 text-amber-700', c: 'bg-slate-100 text-slate-600' }[priority] ?? 'bg-slate-100 text-slate-600'
                    }`}
                  >
                    {priority.toUpperCase()}
                  </span>
                </button>
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
  )
}
