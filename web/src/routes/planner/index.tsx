import { useState, useMemo, useEffect } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight, PlusCircle, CalendarDays, TrendingUp, Sun } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useActivities } from '../../services/activities'
import { useRecoveryBalance } from '../../services/dashboard'
import { useConfig } from '../../services/config'
import { MonthGrid } from '../../components/planner/MonthGrid'
import { WeekGrid } from '../../components/planner/WeekGrid'
import { ActivityList } from '../../components/planner/ActivityList'
import { formatDate, addDays, getMonday, extractDate } from '@/utils/date'
import { getMonthName, getStatusDotColor, getDateLocale } from '@/utils/config'
import { usePlannerState } from '@/contexts/planner'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

type ViewMode = 'week' | 'month' | 'list'


export function PlannerPage() {
  const { t } = useTranslation()
  const { state: plannerState, setWeek, setFrom } = usePlannerState()
  const now = new Date()
  const [viewMode, setViewMode] = useState<ViewMode>('week')
  const [year, setYear] = useState(() => {
    if (plannerState.week) return new Date(plannerState.week + 'T00:00:00').getFullYear()
    return now.getFullYear()
  })
  const [month, setMonth] = useState(() => {
    if (plannerState.week) return new Date(plannerState.week + 'T00:00:00').getMonth() + 1
    return now.getMonth() + 1
  })
  const [weekStart, setWeekStart] = useState(() => {
    if (plannerState.week) return getMonday(new Date(plannerState.week + 'T00:00:00'))
    return getMonday(now)
  })

  // Sync week to context so child routes can navigate back
  useEffect(() => {
    setWeek(formatDate(weekStart))
    setFrom('planner')
  }, [weekStart, setWeek, setFrom])

  const { data: config } = useConfig()

  // Compute date range for the API query
  const { dateFrom, dateTo } = useMemo(() => {
    if (viewMode === 'month') {
      const from = `${year}-${String(month).padStart(2, '0')}-01`
      const lastDay = new Date(year, month, 0).getDate()
      const to = `${year}-${String(month).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`
      return { dateFrom: from, dateTo: to }
    }
    const weekEnd = addDays(weekStart, 6)
    return { dateFrom: formatDate(weekStart), dateTo: formatDate(weekEnd) }
  }, [viewMode, year, month, weekStart])

  const { data, isLoading } = useActivities({ dateFrom, dateTo, limit: 200 })
  const activities = data?.items ?? []

  const recoveryEnabled = config?.recovery?.weekend_activity_flag === true
  const { data: recoveryBalance } = useRecoveryBalance({ dateFrom, dateTo }, recoveryEnabled)

  function prevPeriod() {
    if (viewMode === 'month') {
      if (month === 1) { setMonth(12); setYear((y) => y - 1) }
      else setMonth((m) => m - 1)
    } else {
      setWeekStart((ws) => addDays(ws, -7))
    }
  }

  function nextPeriod() {
    if (viewMode === 'month') {
      if (month === 12) { setMonth(1); setYear((y) => y + 1) }
      else setMonth((m) => m + 1)
    } else {
      setWeekStart((ws) => addDays(ws, 7))
    }
  }

  function goToToday() {
    const today = new Date()
    setYear(today.getFullYear())
    setMonth(today.getMonth() + 1)
    setWeekStart(getMonday(today))
  }

  // Derive period label
  const periodLabel = viewMode === 'month'
    ? `${getMonthName(month - 1)} ${year}`
    : `${formatDate(weekStart)} — ${formatDate(addDays(weekStart, 6))}`

  // Stats
  const todayStr = formatDate(now)
  const todayCount = activities.filter((a) => extractDate(a.dueDate) === todayStr).length

  // Status legend from config
  const statusLegend = config?.activities.statuses.map((s) => ({
    label: s.label,
    dot: getStatusDotColor(config?.activities, s.key),
  })) ?? []

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="p-4 sm:p-8 space-y-6"
    >
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-4xl font-headline font-extrabold text-primary tracking-tight">
            {t('planner.title')}
          </h1>
          <p className="text-on-surface-variant mt-1">{t('planner.subtitle')}</p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {/* View toggle */}
          <div className="flex bg-surface-container-low rounded-lg p-0.5" data-testid="view-toggle">
            <button
              onClick={() => setViewMode('week')}
              className={`px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${
                viewMode === 'week' ? 'bg-primary text-white' : 'text-on-surface-variant hover:text-primary'
              }`}
            >
              {t('planner.week')}
            </button>
            <button
              onClick={() => setViewMode('month')}
              className={`px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${
                viewMode === 'month' ? 'bg-primary text-white' : 'text-on-surface-variant hover:text-primary'
              }`}
            >
              {t('planner.month')}
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${
                viewMode === 'list' ? 'bg-primary text-white' : 'text-on-surface-variant hover:text-primary'
              }`}
              data-testid="list-view-toggle"
            >
              {t('planner.myActivities')}
            </button>
          </div>

          {/* Today button + Period navigation (hidden in list mode) */}
          {viewMode !== 'list' && (
            <>
              <button
                onClick={goToToday}
                className="px-3 py-1.5 text-xs font-bold text-primary border border-primary rounded-lg hover:bg-primary-fixed transition-colors"
              >
                {t('common.today')}
              </button>
              <div className="flex items-center gap-2 bg-surface-container-lowest px-4 py-2 rounded-xl shadow-sm">
                <button className="text-on-surface-variant hover:text-primary" onClick={prevPeriod} aria-label={t('pagination.previousPeriod')}>
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <span className="font-headline font-bold text-primary px-2 min-w-[180px] text-center text-sm" data-testid="period-label">
                  {periodLabel}
                </span>
                <button className="text-on-surface-variant hover:text-primary" onClick={nextPeriod} aria-label={t('pagination.nextPeriod')}>
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
            </>
          )}

          {/* New activity */}
          <Link
            to="/activities/new"
            className="flex items-center gap-2 bg-gradient-to-br from-primary to-primary-container text-white py-2.5 px-6 rounded-xl font-bold text-sm shadow-md no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            {t('nav.newActivity')}
          </Link>
        </div>
      </div>

      {/* Body */}
      {viewMode === 'list' ? (
        <ActivityList />
      ) : (
        <div className="grid grid-cols-12 gap-6">
          {/* Sidebar */}
          <div className="col-span-12 lg:col-span-3 space-y-6">
            {/* Status legend */}
            <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
              <h3 className="font-headline font-bold text-primary mb-4 text-sm">{t('planner.statusLegend')}</h3>
              <div className="space-y-3">
                {statusLegend.map((item) => (
                  <div key={item.label} className="flex items-center gap-2">
                    <span className={`w-3 h-3 rounded-full ${item.dot}`} />
                    <span className="text-[10px] font-medium text-on-surface-variant">{item.label}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Category legend */}
            <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
              <h3 className="font-headline font-bold text-primary mb-4 text-sm">{t('planner.categories')}</h3>
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-sm bg-amber-500" />
                  <span className="text-[10px] font-medium text-on-surface-variant">{t('planner.fieldActivities')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-sm bg-blue-400" />
                  <span className="text-[10px] font-medium text-on-surface-variant">{t('planner.nonFieldActivities')}</span>
                </div>
              </div>
            </div>

            {/* Daily pulse */}
            <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
              <h4 className="font-headline font-bold text-primary mb-4 text-sm">{t('planner.dailyPulse')}</h4>
              <div className="space-y-4">
                <div className="flex items-start gap-3">
                  <CalendarDays className="w-4 h-4 text-primary mt-0.5" />
                  <div>
                    <p className="text-xs font-bold text-on-surface">{t('planner.todayCount', { count: todayCount })}</p>
                    <p className="text-[10px] text-on-surface-variant">{formatDate(now)}</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <TrendingUp className="w-4 h-4 text-tertiary-fixed-dim mt-0.5" />
                  <div>
                    <p className="text-xs font-bold text-on-surface">{t('planner.inView', { count: activities.length })}</p>
                    <p className="text-[10px] text-on-surface-variant">{periodLabel}</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Recovery balance — only shown when recovery is configured */}
            {recoveryEnabled && recoveryBalance && recoveryBalance.earned > 0 && (
              <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
                <h4 className="font-headline font-bold text-primary mb-4 text-sm">{t('planner.recoveryDays')}</h4>
                <div className="flex items-start gap-3 mb-3">
                  <Sun className="w-4 h-4 text-amber-500 mt-0.5" />
                  <div>
                    <p className="text-xs font-bold text-on-surface">
                      {t('planner.recoveryAvailable', { count: recoveryBalance.balance })}
                    </p>
                    <p className="text-[10px] text-on-surface-variant">
                      {t('planner.recoveryEarnedTaken', { earned: recoveryBalance.earned, taken: recoveryBalance.taken })}
                    </p>
                  </div>
                </div>
                {recoveryBalance.intervals.filter((iv) => !iv.claimed).length > 0 && (
                  <div className="space-y-1.5">
                    <p className="text-[9px] uppercase tracking-widest text-slate-400 font-bold">{t('planner.claimBy')}</p>
                    {recoveryBalance.intervals
                      .filter((iv) => !iv.claimed)
                      .map((iv) => (
                        <div key={iv.weekendDate} className="flex items-center justify-between text-[10px]">
                          <span className="text-on-surface-variant">
                            {new Date(iv.weekendDate).toLocaleDateString(getDateLocale(), { weekday: 'short', day: 'numeric', month: 'short' })}
                          </span>
                          <span className="font-bold text-amber-600">
                            {t('planner.byDate', { date: new Date(iv.claimBy).toLocaleDateString(getDateLocale(), { day: 'numeric', month: 'short' }) })}
                          </span>
                        </div>
                      ))}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Calendar grid */}
          <div className="col-span-12 lg:col-span-9">
            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <LoadingSpinner size="lg" label={t('planner.loading')} />
              </div>
            ) : viewMode === 'month' ? (
              <MonthGrid activities={activities} year={year} month={month} config={config} />
            ) : (
              <WeekGrid activities={activities} weekStart={weekStart} config={config} />
            )}
          </div>
        </div>
      )}
    </motion.div>
  )
}
