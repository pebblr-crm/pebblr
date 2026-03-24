import { useState, useEffect } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight, ArrowLeft, PlusCircle } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useActivities } from '../../services/activities'
import { useConfig } from '../../services/config'
import { formatDate, addDays } from '@/utils/date'
import {
  getActivityTitle,
  getTypeCategory,
  getStatusLabel,
  getStatusBadgeColor,
  CATEGORY_COLORS,
  getMonthName,
} from '@/utils/config'
import { usePlannerState } from '@/contexts/planner'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner/daily',
  component: PlannerDailyPage,
})

const DAY_NAME_KEYS = [
  'weekdaysFull.sunday', 'weekdaysFull.monday', 'weekdaysFull.tuesday',
  'weekdaysFull.wednesday', 'weekdaysFull.thursday', 'weekdaysFull.friday',
  'weekdaysFull.saturday',
] as const

export function PlannerDailyPage() {
  const { t } = useTranslation()
  const { setFrom } = usePlannerState()
  const [currentDate, setCurrentDate] = useState(() => new Date())
  const dateStr = formatDate(currentDate)

  useEffect(() => { setFrom('daily') }, [setFrom])

  const { data: config } = useConfig()
  const { data, isLoading } = useActivities({ dateFrom: dateStr, dateTo: dateStr, limit: 50 })
  const activities = data?.items ?? []

  const dayName = t(DAY_NAME_KEYS[currentDate.getDay()])
  const monthName = getMonthName(currentDate.getMonth())
  const isToday = formatDate(new Date()) === dateStr

  function prevDay() { setCurrentDate((d) => addDays(d, -1)) }
  function nextDay() { setCurrentDate((d) => addDays(d, 1)) }
  function goToToday() { setCurrentDate(new Date()) }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-4xl mx-auto w-full space-y-6"
    >
      {/* Back to planner */}
      <Link
        to="/planner"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {t('daily.backToPlanner')}
      </Link>

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-headline font-extrabold text-primary tracking-tight">
            {dayName}
          </h1>
          <p className="text-on-surface-variant mt-1">
            {currentDate.getDate()} {monthName} {currentDate.getFullYear()}
            {isToday && <span className="ml-2 text-xs font-bold text-primary">({t('common.today')})</span>}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {!isToday && (
            <button
              onClick={goToToday}
              className="px-3 py-1.5 text-xs font-bold text-primary border border-primary rounded-lg hover:bg-primary-fixed transition-colors"
            >
              {t('common.today')}
            </button>
          )}
          <div className="flex items-center gap-2 bg-surface-container-lowest px-4 py-2 rounded-xl shadow-sm">
            <button className="text-on-surface-variant hover:text-primary" onClick={prevDay} aria-label={t('pagination.previousDay')}>
              <ChevronLeft className="w-5 h-5" />
            </button>
            <span className="font-headline font-bold text-primary px-2 text-sm" data-testid="daily-date">
              {dateStr}
            </span>
            <button className="text-on-surface-variant hover:text-primary" onClick={nextDay} aria-label={t('pagination.nextDay')}>
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>
          <Link
            to="/activities/new"
            search={{ date: dateStr }}
            className="flex items-center gap-2 bg-gradient-to-br from-primary to-primary-container text-white py-2.5 px-6 rounded-xl font-bold text-sm shadow-md no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            {t('nav.newActivity')}
          </Link>
        </div>
      </div>

      {/* Activity list */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label={t('activityList.loadingActivities')} />
        </div>
      ) : activities.length === 0 ? (
        <div data-testid="empty-state" className="bg-surface-container-lowest p-12 rounded-xl shadow-sm border border-slate-100 text-center">
          <p className="text-on-surface-variant text-sm">{t('daily.noActivities')}</p>
          <Link
            to="/activities/new"
            search={{ date: dateStr }}
            className="inline-flex items-center gap-2 mt-4 text-sm font-medium text-primary hover:underline no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            {t('daily.createActivity')}
          </Link>
        </div>
      ) : (
        <div className="space-y-1.5" data-testid="daily-activities">
          {activities.map((activity) => {
            const title = getActivityTitle(config, activity)
            const category = getTypeCategory(config?.activities, activity.activityType)
            const style = CATEGORY_COLORS[category] ?? CATEGORY_COLORS.field
            const statusLabel = getStatusLabel(config?.activities, activity.status)
            const statusStyle = getStatusBadgeColor(config?.activities, activity.status)

            return (
              <Link
                key={activity.id}
                to="/activities/$activityId"
                params={{ activityId: activity.id }}
                className={`block px-3 py-2 rounded-lg border-l-2 ${style} hover:shadow-sm transition-shadow no-underline`}
                data-testid="daily-activity-row"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 min-w-0">
                    <h3 className="text-xs font-bold text-on-surface truncate">{title}</h3>
                  </div>
                  <div className="flex items-center gap-1.5 shrink-0 ml-2">
                    <span className={`px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-tight ${statusStyle}`}>
                      {statusLabel}
                    </span>
                    {activity.submittedAt && (
                      <span className="px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-tight bg-slate-200 text-slate-600">
                        {t('daily.submitted')}
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
          {t('daily.activitiesScheduled', { count: activities.length })}
        </div>
      )}
    </motion.div>
  )
}
