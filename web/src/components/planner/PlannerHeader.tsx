import { Button } from '@/components/ui/Button'
import { ChevronLeft, ChevronRight, Copy, CalendarDays, CalendarPlus } from 'lucide-react'

export interface PlannerHeaderProps {
  readonly weekStart: Date
  readonly weekEnd: Date
  readonly totalAssigned: number
  readonly onPrevWeek: () => void
  readonly onNextWeek: () => void
  readonly onGoToday: () => void
  readonly onCloneWeek: () => void
  readonly onCreateActivities: () => void
  readonly cloneWeekPending: boolean
  readonly batchCreatePending: boolean
}

export function PlannerHeader({
  weekStart,
  weekEnd,
  totalAssigned,
  onPrevWeek,
  onNextWeek,
  onGoToday,
  onCloneWeek,
  onCreateActivities,
  cloneWeekPending,
  batchCreatePending,
}: PlannerHeaderProps) {
  return (
    <div className="px-4 py-3 border-b border-slate-200 bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-2 shrink-0 md:px-6">
      <div className="flex flex-wrap items-center gap-2 md:gap-3">
        <div className="flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50">
          <button onClick={onPrevWeek} className="p-1.5 hover:bg-slate-100 rounded-l-lg" aria-label="Previous week">
            <ChevronLeft size={16} />
          </button>
          <span className="px-2 text-xs font-medium text-slate-700 md:px-3 md:text-sm">
            {weekStart.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} — {weekEnd.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}
          </span>
          <button onClick={onNextWeek} className="p-1.5 hover:bg-slate-100 rounded-r-lg" aria-label="Next week">
            <ChevronRight size={16} />
          </button>
        </div>
        <Button variant="ghost" size="sm" onClick={onGoToday}>
          <CalendarDays size={14} />
          Today
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="secondary" size="sm" onClick={onCloneWeek} disabled={cloneWeekPending}>
          <Copy size={14} />
          Clone Week
        </Button>
        {totalAssigned > 0 && (
          <Button
            variant="primary"
            size="sm"
            onClick={onCreateActivities}
            disabled={batchCreatePending}
          >
            <CalendarPlus size={14} />
            Create {totalAssigned} Visit{totalAssigned === 1 ? '' : 's'}
          </Button>
        )}
      </div>
    </div>
  )
}
