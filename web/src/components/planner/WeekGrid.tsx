import { Link } from '@tanstack/react-router'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import { ActivityCard } from './ActivityCard'

interface WeekGridProps {
  activities: Activity[]
  weekStart: Date
  config?: TenantConfig
}

const DAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function addDays(d: Date, n: number): Date {
  const result = new Date(d)
  result.setDate(result.getDate() + n)
  return result
}

export function WeekGrid({ activities, weekStart, config }: WeekGridProps) {
  const today = formatDate(new Date())
  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))

  return (
    <div className="bg-surface-container-lowest rounded-xl shadow-sm border border-slate-100 overflow-hidden" data-testid="week-grid">
      <div className="grid grid-cols-7 divide-x divide-slate-100">
        {days.map((day, i) => {
          const dateStr = formatDate(day)
          const isToday = dateStr === today
          const dayActivities = activities.filter((a) => a.dueDate === dateStr)

          return (
            <div key={i} className={`min-h-[400px] ${isToday ? 'bg-primary/5' : ''}`}>
              {/* Day header */}
              <div className="p-3 border-b border-slate-100 text-center">
                <div className="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">
                  {DAY_LABELS[i]}
                </div>
                <div
                  className={`text-lg font-bold mt-1 ${
                    isToday
                      ? 'w-9 h-9 bg-primary text-white rounded-full flex items-center justify-center mx-auto'
                      : 'text-on-surface'
                  }`}
                >
                  {day.getDate()}
                </div>
              </div>

              {/* Activities */}
              <div className="p-2 space-y-1">
                {dayActivities.map((a) => (
                  <ActivityCard key={a.id} activity={a} config={config} />
                ))}
                {dayActivities.length === 0 && (
                  <p className="text-[10px] text-on-surface-variant text-center mt-4 opacity-50">No activities</p>
                )}
              </div>

              {/* Quick add */}
              <div className="px-2 pb-2">
                <Link
                  to="/activities/new"
                  search={{ date: dateStr }}
                  className="block w-full text-center text-[10px] text-on-surface-variant hover:text-primary hover:bg-primary/5 rounded py-1 transition-colors no-underline"
                >
                  + Add
                </Link>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
