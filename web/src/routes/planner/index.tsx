import { useState, useMemo } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ChevronLeft, ChevronRight, PlusCircle, CalendarDays, TrendingUp } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useActivities } from '../../services/activities'
import { useConfig } from '../../services/config'
import { MonthGrid } from '../../components/planner/MonthGrid'
import { WeekGrid } from '../../components/planner/WeekGrid'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/planner',
  component: PlannerPage,
})

type ViewMode = 'week' | 'month'

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

/** Returns the Monday of the week containing `date`. */
function getMonday(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = day === 0 ? -6 : 1 - day
  d.setDate(d.getDate() + diff)
  d.setHours(0, 0, 0, 0)
  return d
}

function addDays(d: Date, n: number): Date {
  const result = new Date(d)
  result.setDate(result.getDate() + n)
  return result
}

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function PlannerPage() {
  const now = new Date()
  const [viewMode, setViewMode] = useState<ViewMode>('week')
  const [year, setYear] = useState(now.getFullYear())
  const [month, setMonth] = useState(now.getMonth() + 1)
  const [weekStart, setWeekStart] = useState(() => getMonday(now))

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
    ? `${MONTH_NAMES[month - 1]} ${year}`
    : `${formatDate(weekStart)} — ${formatDate(addDays(weekStart, 6))}`

  // Stats
  const todayStr = formatDate(now)
  const todayCount = activities.filter((a) => a.dueDate.split('T')[0] === todayStr).length

  // Status legend from config
  const statusLegend = config?.activities.statuses.map((s) => {
    const dot = s.key === 'realizat' ? 'bg-emerald-500'
      : s.key === 'anulat' ? 'bg-red-400'
        : 'bg-amber-500'
    return { label: s.label, dot }
  }) ?? []

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="p-8 space-y-6"
    >
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-4xl font-headline font-extrabold text-primary tracking-tight">
            Planner
          </h1>
          <p className="text-on-surface-variant mt-1">Plan and track field activities.</p>
        </div>
        <div className="flex items-center gap-3">
          {/* View toggle */}
          <div className="flex bg-surface-container-low rounded-lg p-0.5" data-testid="view-toggle">
            <button
              onClick={() => setViewMode('week')}
              className={`px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${
                viewMode === 'week' ? 'bg-primary text-white' : 'text-on-surface-variant hover:text-primary'
              }`}
            >
              Week
            </button>
            <button
              onClick={() => setViewMode('month')}
              className={`px-3 py-1.5 text-xs font-bold rounded-md transition-colors ${
                viewMode === 'month' ? 'bg-primary text-white' : 'text-on-surface-variant hover:text-primary'
              }`}
            >
              Month
            </button>
          </div>

          {/* Today button */}
          <button
            onClick={goToToday}
            className="px-3 py-1.5 text-xs font-bold text-primary border border-primary rounded-lg hover:bg-primary-fixed transition-colors"
          >
            Today
          </button>

          {/* Period navigation */}
          <div className="flex items-center gap-2 bg-surface-container-lowest px-4 py-2 rounded-xl shadow-sm">
            <button className="text-on-surface-variant hover:text-primary" onClick={prevPeriod} aria-label="Previous period">
              <ChevronLeft className="w-5 h-5" />
            </button>
            <span className="font-headline font-bold text-primary px-2 min-w-[180px] text-center text-sm" data-testid="period-label">
              {periodLabel}
            </span>
            <button className="text-on-surface-variant hover:text-primary" onClick={nextPeriod} aria-label="Next period">
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>

          {/* New activity */}
          <Link
            to="/activities/new"
            className="flex items-center gap-2 bg-gradient-to-br from-primary to-primary-container text-white py-2.5 px-6 rounded-xl font-bold text-sm shadow-md no-underline"
          >
            <PlusCircle className="w-4 h-4" />
            New Activity
          </Link>
        </div>
      </div>

      {/* Body */}
      <div className="grid grid-cols-12 gap-6">
        {/* Sidebar */}
        <div className="col-span-12 lg:col-span-3 space-y-6">
          {/* Status legend */}
          <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
            <h3 className="font-headline font-bold text-primary mb-4 text-sm">Status Legend</h3>
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
            <h3 className="font-headline font-bold text-primary mb-4 text-sm">Categories</h3>
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <span className="w-3 h-3 rounded-sm bg-amber-500" />
                <span className="text-[10px] font-medium text-on-surface-variant">Field activities</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-3 h-3 rounded-sm bg-blue-400" />
                <span className="text-[10px] font-medium text-on-surface-variant">Non-field activities</span>
              </div>
            </div>
          </div>

          {/* Daily pulse */}
          <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
            <h4 className="font-headline font-bold text-primary mb-4 text-sm">Daily Pulse</h4>
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <CalendarDays className="w-4 h-4 text-primary mt-0.5" />
                <div>
                  <p className="text-xs font-bold text-on-surface">{todayCount} Today</p>
                  <p className="text-[10px] text-on-surface-variant">{formatDate(now)}</p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <TrendingUp className="w-4 h-4 text-tertiary-fixed-dim mt-0.5" />
                <div>
                  <p className="text-xs font-bold text-on-surface">{activities.length} In View</p>
                  <p className="text-[10px] text-on-surface-variant">{periodLabel}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Calendar grid */}
        <div className="col-span-12 lg:col-span-9">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" label="Loading planner..." />
            </div>
          ) : viewMode === 'month' ? (
            <MonthGrid activities={activities} year={year} month={month} config={config} />
          ) : (
            <WeekGrid activities={activities} weekStart={weekStart} config={config} />
          )}
        </div>
      </div>
    </motion.div>
  )
}
