import { useState, useMemo, useRef } from 'react'
import { X, GraduationCap, Briefcase, Palmtree, Users, Car } from 'lucide-react'
import { formatDate, addDays } from '@/lib/dates'
import type { Activity } from '@/types/activity'
import type { Target } from '@/types/target'

interface WeekViewProps {
  readonly weekStart: Date
  readonly activities: readonly Activity[]
  /** Pending assignments: dateStr → target IDs not yet created */
  readonly dayAssignments?: Readonly<Record<string, string[]>>
  readonly targetMap?: Map<string, Target>
  /** Whether something is currently being dragged */
  readonly isDragging?: boolean
  /** ID of the activity currently being dragged (to dim its source card) */
  readonly draggingActivityId?: string | null
  /** Pending assignment currently being dragged: { sourceDate, targetId } */
  readonly draggingPending?: { sourceDate: string; targetId: string } | null
  /** Max visits per day from tenant config (default 10) */
  readonly maxPerDay?: number
  readonly onActivityClick?: (activity: Activity) => void
  readonly onDayClick?: (date: string) => void
  readonly onDrop?: (dateStr: string) => void
  readonly onRemoveAssignment?: (dateStr: string, targetId: string) => void
  readonly onActivityDragStart?: (activityId: string) => void
  readonly onActivityDragEnd?: () => void
  readonly onPendingDragStart?: (sourceDate: string, targetId: string) => void
  readonly onPendingDragEnd?: () => void
}

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri']

const priorityBorder: Record<string, string> = {
  a: 'border-l-red-500',
  b: 'border-l-amber-500',
  c: 'border-l-slate-400',
}

function getTargetPriority(activity: Activity): string {
  const p = activity.targetSummary?.fields?.potential as string | undefined
  return (p ?? 'c').toLowerCase()
}

/** Non-visit activity types that block part or all of a day */
const BLOCKER_TYPES = new Set(['training', 'team_meeting', 'administrative', 'vacation', 'business_travel'])

function isBlocker(activity: Activity): boolean {
  return BLOCKER_TYPES.has(activity.activityType)
}

function isVisit(activity: Activity): boolean {
  return activity.activityType === 'visit'
}

const blockerIcons: Record<string, typeof GraduationCap> = {
  training: GraduationCap,
  team_meeting: Users,
  administrative: Briefcase,
  vacation: Palmtree,
  business_travel: Car,
}

const blockerLabels: Record<string, string> = {
  training: 'Training',
  team_meeting: 'Team Meeting',
  administrative: 'Administrative',
  vacation: 'Vacation',
  business_travel: 'Business Travel',
}

const HATCH_PATTERN = 'repeating-linear-gradient(45deg, transparent, transparent 10px, rgba(0,0,0,0.04) 10px, rgba(0,0,0,0.04) 20px)'

/** Compute blocker-based capacity constraints for a day */
function computeCapacity(blockers: readonly Activity[], maxPerDay: number) {
  const blockerCapacity = blockers.reduce((sum, b) => {
    return sum + (b.duration === 'full_day' ? 1 : 0.5)
  }, 0)
  const cappedCapacity = Math.min(blockerCapacity, 1)
  const isFullyBlocked = cappedCapacity >= 1
  const isHalfBlocked = cappedCapacity >= 0.5 && !isFullyBlocked
  let maxVisits: number
  if (isFullyBlocked) maxVisits = 0
  else if (isHalfBlocked) maxVisits = Math.floor(maxPerDay / 2)
  else maxVisits = maxPerDay
  return { isFullyBlocked, isHalfBlocked, maxVisits }
}

/** Build the header label text for a day column */
function buildHeaderLabel(isFullyBlocked: boolean, isHalfBlocked: boolean, visitCount: number): string {
  if (isFullyBlocked) return 'Blocked'
  if (isHalfBlocked) return `${visitCount} visit${visitCount === 1 ? '' : 's'} + blocker`
  if (visitCount > 0) return `${visitCount} visit${visitCount === 1 ? '' : 's'} planned`
  return '0 visits planned'
}

/** Resolve the drop zone border/background classes */
function dropZoneClasses(overCapacity: boolean, dragOver: boolean, isDragging: boolean): string {
  if (overCapacity) return 'border-red-300 border-solid bg-red-50/30'
  if (dragOver) return 'border-teal-500 bg-teal-50/50 border-solid'
  if (isDragging) return 'border-dashed border-teal-200 bg-teal-50/20'
  return 'border-slate-200 border-solid bg-slate-50'
}

export function WeekView({
  weekStart, activities, dayAssignments = {}, targetMap,
  isDragging = false, draggingActivityId, draggingPending,
  maxPerDay = 10,
  onActivityClick, onDayClick, onDrop, onRemoveAssignment,
  onActivityDragStart, onActivityDragEnd,
  onPendingDragStart, onPendingDragEnd,
}: WeekViewProps) {
  const days = useMemo(() => {
    return DAY_NAMES.map((name, i) => {
      const date = addDays(weekStart, i)
      const dateStr = formatDate(date)
      const dayActivities = activities.filter((a) => a.dueDate.slice(0, 10) === dateStr)
      const pendingIds = dayAssignments[dateStr] ?? []
      return { name, date, dateStr, activities: dayActivities, pendingIds }
    })
  }, [weekStart, activities, dayAssignments])

  const today = formatDate(new Date())

  return (
    <div className="grid grid-cols-5 gap-3 min-h-full">
      {days.map((day) => {
        const isToday = day.dateStr === today
        const visits = day.activities.filter(isVisit)
        const blockers = day.activities.filter(isBlocker)
        const totalCount = visits.length + day.pendingIds.length
        return (
          <DayColumn
            key={day.dateStr}
            dayName={day.name}
            date={day.date}
            dateStr={day.dateStr}
            isToday={isToday}
            visits={visits}
            blockers={blockers}
            pendingIds={day.pendingIds}
            targetMap={targetMap}
            visitCount={totalCount}
            isDragging={isDragging}
            draggingActivityId={draggingActivityId}
            draggingPending={draggingPending}
            onDayClick={onDayClick}
            onActivityClick={onActivityClick}
            onDrop={onDrop}
            onRemoveAssignment={onRemoveAssignment}
            onActivityDragStart={onActivityDragStart}
            onActivityDragEnd={onActivityDragEnd}
            onPendingDragStart={onPendingDragStart}
            onPendingDragEnd={onPendingDragEnd}
            maxPerDay={maxPerDay}
          />
        )
      })}
    </div>
  )
}

/* ── Single day column with drop zone ── */

interface DayColumnProps {
  readonly dayName: string
  readonly date: Date
  readonly dateStr: string
  readonly isToday: boolean
  readonly visits: readonly Activity[]
  readonly blockers: readonly Activity[]
  readonly pendingIds: readonly string[]
  readonly targetMap?: Map<string, Target>
  readonly visitCount: number
  readonly isDragging: boolean
  readonly draggingActivityId?: string | null
  readonly draggingPending?: { sourceDate: string; targetId: string } | null
  readonly maxPerDay?: number
  readonly onDayClick?: (date: string) => void
  readonly onActivityClick?: (activity: Activity) => void
  readonly onDrop?: (dateStr: string) => void
  readonly onRemoveAssignment?: (dateStr: string, targetId: string) => void
  readonly onActivityDragStart?: (activityId: string) => void
  readonly onActivityDragEnd?: () => void
  readonly onPendingDragStart?: (sourceDate: string, targetId: string) => void
  readonly onPendingDragEnd?: () => void
}

function DayColumn({
  dayName, date, dateStr, isToday,
  visits, blockers, pendingIds, targetMap, visitCount, isDragging, draggingActivityId, draggingPending,
  maxPerDay = 10,
  onDayClick, onActivityClick, onDrop, onRemoveAssignment,
  onActivityDragStart, onActivityDragEnd,
  onPendingDragStart, onPendingDragEnd,
}: DayColumnProps) {
  const [dragOver, setDragOver] = useState(false)
  const dragCounter = useRef(0)

  const { isFullyBlocked, isHalfBlocked, maxVisits } = computeCapacity(blockers, maxPerDay)
  const overCapacity = visitCount > maxVisits
  const headerLabel = buildHeaderLabel(isFullyBlocked, isHalfBlocked, visitCount)

  return (
    <div className="flex flex-col gap-2 min-h-[500px]">
      {/* Day header */}
      <button
        onClick={() => onDayClick?.(dateStr)}
        className="text-center pb-2 border-b border-slate-200"
      >
        <div className={`text-xs font-medium uppercase tracking-wider ${isToday ? 'text-teal-600' : 'text-slate-500'}`}>
          {dayName}
        </div>
        <div className={`text-lg font-semibold ${isToday ? 'text-teal-700' : 'text-slate-800'}`}>
          {date.getDate()}
        </div>
        <div className={`text-[10px] mt-1 ${overCapacity ? 'text-red-500 font-medium' : 'text-slate-400'}`}>
          {headerLabel}
        </div>
      </button>

      {/* Day body / drop zone */}
      <section
        aria-label={`${dayName} drop zone`}
        onDragEnter={(e) => { e.preventDefault(); dragCounter.current++; setDragOver(true) }}
        onDragOver={(e) => e.preventDefault()}
        onDragLeave={() => { dragCounter.current--; if (dragCounter.current === 0) setDragOver(false) }}
        onDrop={(e) => { e.preventDefault(); dragCounter.current = 0; setDragOver(false); onDrop?.(dateStr) }}
        className={`flex-1 rounded-md border-2 p-2 gap-2 relative transition-colors flex flex-col overflow-hidden ${dropZoneClasses(overCapacity, dragOver, isDragging)}`}
      >

        {/* Visit cards */}
        {visits.map((activity) => {
          const priority = getTargetPriority(activity)
          const borderClass = priorityBorder[priority] ?? priorityBorder.c
          return (
            <button
              key={activity.id}
              type="button"
              draggable
              onDragStart={() => onActivityDragStart?.(activity.id)}
              onDragEnd={() => onActivityDragEnd?.()}
              onClick={() => onActivityClick?.(activity)}
              className={`bg-white p-2 rounded border-l-4 ${borderClass} shadow-sm text-sm cursor-grab hover:shadow-md transition-all text-left w-full ${
                draggingActivityId === activity.id ? 'opacity-40 scale-95' : ''
              }`}
            >
              <div className="font-medium text-slate-800 truncate">
                {activity.targetName ?? activity.label ?? activity.activityType}
              </div>
            </button>
          )
        })}

        {/* Pending assignments */}
        {pendingIds.map((id) => {
          const t = targetMap?.get(id)
          if (!t) return null
          return (
            <li
              key={id}
              draggable
              onDragStart={() => onPendingDragStart?.(dateStr, id)}
              onDragEnd={() => onPendingDragEnd?.()}
              className={`flex items-center gap-1 bg-teal-50 border border-teal-200 rounded px-2 py-1.5 text-sm cursor-grab transition-all list-none ${
                draggingPending?.sourceDate === dateStr && draggingPending?.targetId === id
                  ? 'opacity-40 scale-95' : ''
              }`}
            >
              <span className="w-1.5 h-1.5 rounded-full bg-teal-600 shrink-0" />
              <span className="truncate flex-1 text-teal-800 font-medium">{t.name}</span>
              <button
                onClick={() => onRemoveAssignment?.(dateStr, id)}
                onKeyDown={(e) => { if (e.key === 'Delete' || e.key === 'Backspace') onRemoveAssignment?.(dateStr, id) }}
                className="text-teal-400 hover:text-red-500 shrink-0"
              >
                <X size={12} />
              </button>
            </li>
          )
        })}

        {/* Blocker blocks — inline, same spacing as visits */}
        {blockers.map((blocker) => {
          const Icon = blockerIcons[blocker.activityType] ?? Briefcase
          const label = blockerLabels[blocker.activityType] ?? blocker.activityType
          const details = (blocker.fields?.details as string) ?? ''
          const isFullDay = blocker.duration === 'full_day'
          return (
            <button
              key={blocker.id}
              type="button"
              onClick={() => onActivityClick?.(blocker)}
              className="rounded border border-slate-200 overflow-hidden cursor-pointer hover:border-slate-300 transition-colors relative flex-1 min-h-[70px] w-full"
              style={{
                backgroundImage: HATCH_PATTERN,
                backgroundColor: 'rgb(241 245 249)',
                maxHeight: isFullDay ? undefined : '50%',
              }}
            >
              <div className="h-full flex flex-col items-center justify-center text-center p-3">
                <div className="w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center text-slate-500 mb-1.5 shrink-0">
                  <Icon size={16} />
                </div>
                <span className="font-medium text-slate-700 text-xs">{label}</span>
                {details && <span className="text-[10px] text-slate-500 mt-0.5 truncate max-w-full">{details}</span>}
                <span className="text-[10px] text-slate-400 mt-0.5">
                  {isFullDay ? 'All Day' : 'Half Day'}
                </span>
              </div>
            </button>
          )
        })}

        {/* Empty state / drop hint */}
        {visits.length === 0 && pendingIds.length === 0 && blockers.length === 0 && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <span className={`font-medium text-sm ${isDragging ? 'text-teal-400' : 'text-slate-300'}`}>
              {isDragging ? 'Drop here' : 'No visits'}
            </span>
          </div>
        )}
      </section>
    </div>
  )
}
