import { useState } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ChevronLeft, ChevronRight, ArrowLeft, PlusCircle } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useActivities } from '../../services/activities'
import { useConfig } from '../../services/config'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner/daily',
  component: PlannerDailyPage,
})

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

const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

const categoryColors: Record<string, string> = {
  field: 'border-amber-500 bg-amber-50',
  non_field: 'border-blue-400 bg-blue-50',
}

const statusBadge: Record<string, string> = {
  planificat: 'bg-amber-100 text-amber-700',
  realizat: 'bg-emerald-100 text-emerald-700',
  anulat: 'bg-red-100 text-red-700',
}

export function PlannerDailyPage() {
  const [currentDate, setCurrentDate] = useState(() => new Date())
  const dateStr = formatDate(currentDate)

  const { data: config } = useConfig()
  const { data, isLoading } = useActivities({ dateFrom: dateStr, dateTo: dateStr, limit: 50 })
  const activities = data?.items ?? []

  const dayName = DAY_NAMES[currentDate.getDay()]
  const monthName = MONTH_NAMES[currentDate.getMonth()]
  const isToday = formatDate(new Date()) === dateStr

  function prevDay() { setCurrentDate((d) => addDays(d, -1)) }
  function nextDay() { setCurrentDate((d) => addDays(d, 1)) }
  function goToToday() { setCurrentDate(new Date()) }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 max-w-4xl mx-auto w-full space-y-6"
    >
      {/* Back to planner */}
      <Link
        to="/planner"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to planner
      </Link>

      {/* Header */}
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-headline font-extrabold text-primary tracking-tight">
            {dayName}
          </h1>
          <p className="text-on-surface-variant mt-1">
            {currentDate.getDate()} {monthName} {currentDate.getFullYear()}
            {isToday && <span className="ml-2 text-xs font-bold text-primary">(Today)</span>}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {!isToday && (
            <button
              onClick={goToToday}
              className="px-3 py-1.5 text-xs font-bold text-primary border border-primary rounded-lg hover:bg-primary-fixed transition-colors"
            >
              Today
            </button>
          )}
          <div className="flex items-center gap-2 bg-surface-container-lowest px-4 py-2 rounded-xl shadow-sm">
            <button className="text-on-surface-variant hover:text-primary" onClick={prevDay} aria-label="Previous day">
              <ChevronLeft className="w-5 h-5" />
            </button>
            <span className="font-headline font-bold text-primary px-2 text-sm" data-testid="daily-date">
              {dateStr}
            </span>
            <button className="text-on-surface-variant hover:text-primary" onClick={nextDay} aria-label="Next day">
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>
          <Link
            to="/activities/new"
            search={{ date: dateStr }}
            className="flex items-center gap-2 bg-gradient-to-br from-primary to-primary-container text-white py-2.5 px-6 rounded-xl font-bold text-sm shadow-md no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            New Activity
          </Link>
        </div>
      </div>

      {/* Activity list */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label="Loading activities..." />
        </div>
      ) : activities.length === 0 ? (
        <div data-testid="empty-state" className="bg-surface-container-lowest p-12 rounded-xl shadow-sm border border-slate-100 text-center">
          <p className="text-on-surface-variant text-sm">No activities scheduled for this day.</p>
          <Link
            to="/activities/new"
            search={{ date: dateStr }}
            className="inline-flex items-center gap-2 mt-4 text-sm font-medium text-primary hover:underline no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            Create an activity
          </Link>
        </div>
      ) : (
        <div className="space-y-3" data-testid="daily-activities">
          {activities.map((activity) => {
            const typeConfig = config?.activities.types.find((t) => t.key === activity.activityType)
            const typeLabel = typeConfig?.label ?? activity.activityType
            const category = typeConfig?.category ?? 'field'
            const style = categoryColors[category] ?? categoryColors.field
            const statusLabel = config?.activities.statuses.find((s) => s.key === activity.status)?.label ?? activity.status
            const statusStyle = statusBadge[activity.status] ?? 'bg-slate-100 text-slate-600'
            const durationLabel = config?.activities.durations.find((d) => d.key === activity.duration)?.label ?? activity.duration

            return (
              <Link
                key={activity.id}
                to="/activities/$activityId"
                params={{ activityId: activity.id }}
                className={`block p-5 rounded-xl border-l-4 ${style} shadow-sm hover:shadow-md transition-shadow no-underline`}
                data-testid="daily-activity-row"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="text-sm font-bold text-on-surface">{typeLabel}</h3>
                    <p className="text-xs text-on-surface-variant mt-1">{durationLabel}</p>
                    {activity.routing && (
                      <p className="text-xs text-on-surface-variant mt-0.5">Route: {activity.routing}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-tight ${statusStyle}`}>
                      {statusLabel}
                    </span>
                    {activity.submittedAt && (
                      <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-tight bg-slate-200 text-slate-600">
                        Submitted
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            )
          })}
        </div>
      )}

      {/* Summary */}
      {activities.length > 0 && (
        <div className="text-xs text-on-surface-variant text-center" data-testid="daily-summary">
          {activities.length} {activities.length === 1 ? 'activity' : 'activities'} scheduled
        </div>
      )}
    </motion.div>
  )
}
