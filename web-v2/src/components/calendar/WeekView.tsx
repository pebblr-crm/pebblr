import { useMemo } from 'react'
import type { Activity } from '@/types/activity'
import { Badge } from '@/components/ui/Badge'

interface WeekViewProps {
  weekStart: Date
  activities: Activity[]
  onActivityClick?: (activity: Activity) => void
  onDayClick?: (date: string) => void
}

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri']

function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

function addDays(d: Date, n: number): Date {
  const result = new Date(d)
  result.setDate(result.getDate() + n)
  return result
}

const statusVariant: Record<string, 'default' | 'primary' | 'success' | 'danger' | 'warning'> = {
  planificat: 'primary',
  realizat: 'success',
  anulat: 'danger',
}

export function WeekView({ weekStart, activities, onActivityClick, onDayClick }: WeekViewProps) {
  const days = useMemo(() => {
    return DAY_NAMES.map((name, i) => {
      const date = addDays(weekStart, i)
      const dateStr = formatDate(date)
      const dayActivities = activities.filter((a) => a.dueDate === dateStr)
      return { name, date, dateStr, activities: dayActivities }
    })
  }, [weekStart, activities])

  return (
    <div className="grid grid-cols-5 gap-px rounded-xl border border-slate-200 bg-slate-200 overflow-hidden">
      {days.map((day) => (
        <div key={day.dateStr} className="bg-white min-h-[200px] flex flex-col">
          <button
            onClick={() => onDayClick?.(day.dateStr)}
            className="flex items-center justify-between border-b border-slate-100 px-3 py-2 text-left hover:bg-slate-50"
          >
            <span className="text-xs font-medium text-slate-500">{day.name}</span>
            <span className="text-xs text-slate-400">
              {day.date.getDate()}/{day.date.getMonth() + 1}
            </span>
          </button>
          <div className="flex-1 space-y-1 p-2">
            {day.activities.map((activity) => (
              <button
                key={activity.id}
                onClick={() => onActivityClick?.(activity)}
                className="w-full rounded-lg border border-slate-100 bg-slate-50 px-2 py-1.5 text-left text-xs hover:bg-slate-100 transition-colors"
              >
                <div className="font-medium text-slate-700 truncate">
                  {activity.targetName ?? activity.label ?? activity.activityType}
                </div>
                <Badge variant={statusVariant[activity.status] ?? 'default'} className="mt-1">
                  {activity.status}
                </Badge>
              </button>
            ))}
            {day.activities.length === 0 && (
              <div className="flex h-full items-center justify-center text-xs text-slate-300">
                No visits
              </div>
            )}
          </div>
          <div className="border-t border-slate-100 px-3 py-1 text-xs text-slate-400">
            {day.activities.length} visit{day.activities.length !== 1 ? 's' : ''}
          </div>
        </div>
      ))}
    </div>
  )
}
