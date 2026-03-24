import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import { ActivityCard } from './ActivityCard'
import { formatDateStr, extractDate } from '@/utils/date'

interface MonthGridProps {
  readonly activities: Activity[]
  readonly year: number
  readonly month: number
  readonly config?: TenantConfig
}

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

function startDayOffset(year: number, month: number): number {
  // day of week for 1st of month: 0=Sun..6=Sat → convert to Mon=0..Sun=6
  const dow = new Date(year, month - 1, 1).getDay()
  return dow === 0 ? 6 : dow - 1
}

export function MonthGrid({ activities, year, month, config }: Readonly<MonthGridProps>) {
  const daysInMonth = new Date(year, month, 0).getDate()
  const offset = startDayOffset(year, month)
  const today = new Date()
  const isCurrentMonth = today.getMonth() + 1 === month && today.getFullYear() === year

  return (
    <div className="bg-surface-container-lowest rounded-xl shadow-sm border border-slate-100 overflow-x-auto" data-testid="month-grid">
      <div className="grid grid-cols-7 bg-surface-container-low border-b border-slate-100 min-w-[560px]">
        {DAYS.map((day) => (
          <div key={day} className="py-4 text-center text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">
            {day}
          </div>
        ))}
      </div>
      <div className="grid grid-cols-7 min-w-[560px]">
        {/* empty cells before first day */}
        {Array.from({ length: offset }).map((_, padIdx) => (
          <div key={`pad-${padIdx}`} className="min-h-[120px] p-2 border-r border-b border-slate-50 bg-slate-50/30" />
        ))}
        {Array.from({ length: daysInMonth }).map((_, i) => {
          const day = i + 1
          const dateStr = formatDateStr(year, month, day)
          const dayActivities = activities.filter((a) => extractDate(a.dueDate) === dateStr)
          const isToday = isCurrentMonth && today.getDate() === day

          return (
            <div
              key={day}
              className={`min-h-[120px] p-2 border-r border-b border-slate-50 relative hover:bg-surface-container-low/30 transition-colors ${isToday ? 'bg-primary/5' : ''}`}
            >
              <span
                className={`text-sm font-medium ${
                  isToday
                    ? 'flex items-center justify-center w-7 h-7 bg-primary text-white rounded-full'
                    : 'text-on-surface-variant'
                }`}
              >
                {day}
              </span>
              <div className="mt-1 space-y-1">
                {dayActivities.map((a) => (
                  <ActivityCard key={a.id} activity={a} config={config} />
                ))}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
