import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import { ActivityCard } from './ActivityCard'
import { formatDate, addDays, extractDate } from '@/utils/date'

interface WeekGridProps {
  activities: Activity[]
  weekStart: Date
  config?: TenantConfig
}

const DAY_LABEL_KEYS = [
  'weekdays.mon', 'weekdays.tue', 'weekdays.wed', 'weekdays.thu',
  'weekdays.fri', 'weekdays.sat', 'weekdays.sun',
] as const

export function WeekGrid({ activities, weekStart, config }: WeekGridProps) {
  const { t } = useTranslation()
  const today = formatDate(new Date())
  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))

  return (
    <div className="bg-surface-container-lowest rounded-xl shadow-sm border border-slate-100 overflow-x-auto" data-testid="week-grid">
      <div className="grid grid-cols-7 divide-x divide-slate-100 h-[calc(100vh-220px)] min-h-[400px] min-w-[560px]">
        {days.map((day, i) => {
          const dateStr = formatDate(day)
          const isToday = dateStr === today
          const dayActivities = activities.filter((a) => extractDate(a.dueDate) === dateStr)

          return (
            <div key={i} className={`flex flex-col ${isToday ? 'bg-primary/5' : ''}`}>
              {/* Day header */}
              <div className="px-2 py-1.5 border-b border-slate-100 text-center shrink-0">
                <div className="text-[9px] font-bold text-on-surface-variant uppercase tracking-widest">
                  {t(DAY_LABEL_KEYS[i])}
                </div>
                <div
                  className={`text-sm font-bold ${
                    isToday
                      ? 'w-6 h-6 bg-primary text-white rounded-full flex items-center justify-center mx-auto text-xs'
                      : 'text-on-surface'
                  }`}
                >
                  {day.getDate()}
                </div>
                {dayActivities.length > 0 && (
                  <span className="text-[8px] text-on-surface-variant">{dayActivities.length}</span>
                )}
              </div>

              {/* Activities — scrollable */}
              <div className="flex-1 overflow-y-auto p-1 space-y-0.5">
                {dayActivities.map((a) => (
                  <ActivityCard key={a.id} activity={a} config={config} />
                ))}
                {dayActivities.length === 0 && (
                  <p className="text-[9px] text-on-surface-variant text-center mt-2 opacity-50">—</p>
                )}
              </div>

              {/* Quick add */}
              <div className="px-1 pb-1 shrink-0">
                <Link
                  to="/activities/new"
                  search={{ date: dateStr }}
                  className="block w-full text-center text-[9px] text-on-surface-variant hover:text-primary hover:bg-primary/5 rounded py-0.5 transition-colors no-underline"
                >
                  +
                </Link>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
