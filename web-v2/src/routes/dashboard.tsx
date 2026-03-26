import { useState, useMemo, useCallback } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useActivityStats, useCoverage, useFrequency, useRecoveryBalance } from '@/hooks/useDashboard'
import { useActivities } from '@/hooks/useActivities'
import { useUsers } from '@/hooks/useUsers'
import { useTeams } from '@/hooks/useTeams'
import { useAuth } from '@/auth/context'
import { WeekView } from '@/components/calendar/WeekView'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { StatCard } from '@/components/data/StatCard'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import { getMonday, addDays, formatDate } from '@/lib/dates'
import { ChevronLeft, ChevronRight, CalendarDays, Users, CalendarRange } from 'lucide-react'
import type { FrequencyItem } from '@/types/dashboard'
import type { Activity } from '@/types/activity'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: DashboardPage,
})

const freqColumnHelper = createColumnHelper<FrequencyItem>()

/* ── Helpers ── */

function getMonthStart(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), 1)
}

function getMonthEnd(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth() + 1, 0)
}

/* ── Main Page ── */

function DashboardPage() {
  useAuth() // ensure authenticated
  const [selectedRep, setSelectedRep] = useState('')
  const [viewMode, setViewMode] = useState<'week' | 'month'>('week')
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()))
  const [monthDate, setMonthDate] = useState(() => new Date())
  const [detailActivityId, setDetailActivityId] = useState<string | null>(null)

  const { data: stats, isLoading: statsLoading } = useActivityStats({})
  const { data: coverage } = useCoverage({})
  const { data: frequency } = useFrequency({})
  const { data: recovery } = useRecoveryBalance({})
  const { data: teamsData } = useTeams()
  const { data: usersData } = useUsers()

  const repUsers = useMemo(
    () => (usersData?.items ?? []).filter((u) => u.role === 'rep'),
    [usersData],
  )

  // Calendar date range
  const weekEnd = useMemo(() => addDays(weekStart, 4), [weekStart])
  const monthStart = useMemo(() => getMonthStart(monthDate), [monthDate])
  const monthEnd = useMemo(() => getMonthEnd(monthDate), [monthDate])

  const calendarDateFrom = viewMode === 'week' ? formatDate(weekStart) : formatDate(monthStart)
  const calendarDateTo = viewMode === 'week' ? formatDate(weekEnd) : formatDate(monthEnd)

  // Fetch activities for selected rep
  const { data: repActivities, isLoading: repLoading } = useActivities({
    creatorId: selectedRep || undefined,
    dateFrom: calendarDateFrom,
    dateTo: calendarDateTo,
    limit: 200,
  })

  const activities = useMemo(() => repActivities?.items ?? [], [repActivities])

  // Navigation
  const prevWeek = useCallback(() => setWeekStart((w) => addDays(w, -7)), [])
  const nextWeek = useCallback(() => setWeekStart((w) => addDays(w, 7)), [])
  const goToday = useCallback(() => {
    setWeekStart(getMonday(new Date()))
    setMonthDate(new Date())
  }, [])
  const prevMonth = useCallback(() => setMonthDate((d) => new Date(d.getFullYear(), d.getMonth() - 1, 1)), [])
  const nextMonth = useCallback(() => setMonthDate((d) => new Date(d.getFullYear(), d.getMonth() + 1, 1)), [])

  // Jump from month week click to that week
  const handleMonthWeekClick = useCallback((mondayStr: string) => {
    const [y, m, d] = mondayStr.split('-').map(Number)
    const monday = new Date(y, m - 1, d)
    monday.setHours(0, 0, 0, 0)
    setWeekStart(monday)
    setViewMode('week')
  }, [])

  // Frequency columns
  const freqColumns = useMemo(
    () => [
      freqColumnHelper.accessor('classification', {
        header: 'Classification',
        cell: (info) => {
          const v = info.getValue()
          const variant = v === 'A' ? 'danger' : v === 'B' ? 'warning' : 'default'
          return <Badge variant={variant}>{v}</Badge>
        },
      }),
      freqColumnHelper.accessor('targetCount', { header: 'Targets' }),
      freqColumnHelper.accessor('totalVisits', { header: 'Visits' }),
      freqColumnHelper.accessor('required', { header: 'Required' }),
      freqColumnHelper.accessor('compliance', {
        header: 'Compliance',
        cell: (info) => {
          const v = Math.round(info.getValue())
          const color = v >= 80 ? 'text-emerald-600' : v >= 50 ? 'text-amber-600' : 'text-red-600'
          return <span className={`font-semibold ${color}`}>{v}%</span>
        },
      }),
    ],
    [],
  )

  if (statsLoading) return <Spinner />

  const completedCount = stats?.byStatus?.realizat ?? 0
  const completionRate = stats?.total ? Math.round((completedCount / stats.total) * 100) : 0
  const selectedRepName = repUsers.find((u) => u.id === selectedRep)?.name ?? 'All Reps'

  return (
    <div className="flex-1 overflow-auto">
      <div className="p-4 space-y-4 md:p-6 md:space-y-6 max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-slate-900">Team Dashboard</h1>
            <p className="text-sm text-slate-500">
              {teamsData?.items?.length ?? 0} teams · {repUsers.length} reps
            </p>
          </div>
        </div>

        {/* KPI cards */}
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
          <StatCard
            label="Cycle Compliance"
            value={`${completionRate}%`}
            subtitle={`${completedCount} of ${stats?.total ?? 0} completed`}
            trend={completionRate >= 70 ? 'up' : 'down'}
          />
          <StatCard
            label="Coverage"
            value={coverage ? `${Math.round(coverage.percentage)}%` : '-'}
            subtitle={coverage ? `${coverage.visitedTargets} of ${coverage.totalTargets} visited` : undefined}
            trend={coverage && coverage.percentage >= 70 ? 'up' : 'down'}
          />
          <StatCard
            label="Week Progress"
            value={`${completedCount} / ${stats?.total ?? 0}`}
            subtitle="visits this period"
          />
          <StatCard
            label="Recovery Balance"
            value={recovery ? `${recovery.balance} days` : '-'}
            subtitle={recovery ? `${recovery.earned} earned` : undefined}
            trend="neutral"
          />
        </div>

        {/* Rep Calendar Section */}
        <Card className="!p-0 overflow-hidden">
          {/* Calendar toolbar */}
          <div className="px-4 py-3 border-b border-slate-200 bg-slate-50 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <Users size={16} className="text-slate-500" />
              <select
                value={selectedRep}
                onChange={(e) => setSelectedRep(e.target.value)}
                className="text-sm border border-slate-300 rounded-md py-1.5 px-3 bg-white focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none"
              >
                <option value="">All Reps</option>
                {repUsers.map((u) => (
                  <option key={u.id} value={u.id}>{u.name || u.displayName}</option>
                ))}
              </select>

              {/* View mode toggle */}
              <div className="flex bg-slate-100 rounded-md p-0.5 border border-slate-200">
                <button
                  onClick={() => setViewMode('week')}
                  className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                    viewMode === 'week' ? 'bg-white shadow-sm text-slate-700' : 'text-slate-500 hover:text-slate-700'
                  }`}
                >
                  <CalendarDays size={12} className="inline mr-1" />
                  Week
                </button>
                <button
                  onClick={() => setViewMode('month')}
                  className={`px-2.5 py-1 rounded text-xs font-medium transition-colors ${
                    viewMode === 'month' ? 'bg-white shadow-sm text-slate-700' : 'text-slate-500 hover:text-slate-700'
                  }`}
                >
                  <CalendarRange size={12} className="inline mr-1" />
                  Month
                </button>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <div className="flex items-center gap-1 rounded-lg border border-slate-200 bg-white">
                <button onClick={viewMode === 'week' ? prevWeek : prevMonth} className="p-1.5 hover:bg-slate-100 rounded-l-lg">
                  <ChevronLeft size={16} />
                </button>
                <span className="px-3 text-sm font-medium text-slate-700">
                  {viewMode === 'week'
                    ? `${weekStart.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })} — ${weekEnd.toLocaleDateString('en-GB', { month: 'short', day: 'numeric' })}`
                    : monthDate.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' })
                  }
                </span>
                <button onClick={viewMode === 'week' ? nextWeek : nextMonth} className="p-1.5 hover:bg-slate-100 rounded-r-lg">
                  <ChevronRight size={16} />
                </button>
              </div>
              <Button variant="ghost" size="sm" onClick={goToday}>Today</Button>
            </div>
          </div>

          {/* Calendar body */}
          <div className="p-4">
            {repLoading ? (
              <div className="py-12 flex justify-center"><Spinner /></div>
            ) : selectedRep === '' && viewMode === 'week' ? (
              <div className="py-12 text-center text-sm text-slate-400">Select a rep to view their calendar.</div>
            ) : viewMode === 'week' ? (
              <WeekView
                weekStart={weekStart}
                activities={activities}
                onActivityClick={(a) => setDetailActivityId(a.id)}
              />
            ) : (
              <MonthGrid
                monthDate={monthDate}
                activities={activities}
                onWeekClick={handleMonthWeekClick}
                selectedRep={selectedRep}
                repName={selectedRepName}
              />
            )}
          </div>
        </Card>

        {/* Activity breakdown */}
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-6">
          <Card>
            <h3 className="mb-4 text-sm font-semibold text-slate-900">Activity by Status</h3>
            {stats?.byStatus && (
              <div className="space-y-3">
                {Object.entries(stats.byStatus).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className={`h-3 w-3 rounded-full ${status === 'realizat' ? 'bg-emerald-500' : status === 'anulat' ? 'bg-red-500' : 'bg-blue-500'}`} />
                      <span className="text-sm text-slate-700 capitalize">{status}</span>
                    </div>
                    <span className="text-sm font-medium text-slate-900">{count}</span>
                  </div>
                ))}
              </div>
            )}
          </Card>

          <Card>
            <h3 className="mb-4 text-sm font-semibold text-slate-900">Activity by Category</h3>
            {stats?.byCategory && (
              <div className="space-y-3">
                {Object.entries(stats.byCategory).map(([category, count]) => (
                  <div key={category} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className={`h-3 w-3 rounded-full ${category === 'field' ? 'bg-amber-500' : 'bg-blue-500'}`} />
                      <span className="text-sm text-slate-700 capitalize">{category}</span>
                    </div>
                    <span className="text-sm font-medium text-slate-900">{count}</span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>

        {/* Frequency compliance table */}
        {frequency?.items && frequency.items.length > 0 && (
          <div>
            <h3 className="mb-3 text-sm font-semibold text-slate-900">Frequency Compliance by Classification</h3>
            <DataTable data={frequency.items} columns={freqColumns} />
          </div>
        )}
      </div>

      {/* Activity detail modal */}
      <ActivityDetailModal activityId={detailActivityId} onClose={() => setDetailActivityId(null)} />
    </div>
  )
}

/* ── Month Grid ── */

interface MonthGridProps {
  monthDate: Date
  activities: Activity[]
  onWeekClick: (mondayStr: string) => void
  selectedRep: string
  repName: string
}

function MonthGrid({ monthDate, activities, onWeekClick, selectedRep, repName }: MonthGridProps) {
  const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const today = formatDate(new Date())

  // Build the calendar grid
  const weeks = useMemo(() => {
    const first = new Date(monthDate.getFullYear(), monthDate.getMonth(), 1)
    const last = new Date(monthDate.getFullYear(), monthDate.getMonth() + 1, 0)
    // Start from Monday of the first week
    const startDay = first.getDay()
    const gridStart = new Date(first)
    gridStart.setDate(first.getDate() - (startDay === 0 ? 6 : startDay - 1))

    const result: Date[][] = []
    const current = new Date(gridStart)
    while (current <= last || result.length < 5) {
      const week: Date[] = []
      for (let i = 0; i < 7; i++) {
        week.push(new Date(current))
        current.setDate(current.getDate() + 1)
      }
      result.push(week)
      if (current > last && result.length >= 4) break
    }
    return result
  }, [monthDate])

  // Count activities per day (convert UTC dueDate to local date string)
  const dayCounts = useMemo(() => {
    const counts = new Map<string, { visits: number; nonField: number; total: number }>()
    for (const a of activities) {
      const dateStr = formatDate(new Date(a.dueDate))
      const entry = counts.get(dateStr) ?? { visits: 0, nonField: 0, total: 0 }
      if (a.activityType === 'visit') entry.visits++
      else entry.nonField++
      entry.total++
      counts.set(dateStr, entry)
    }
    return counts
  }, [activities])

  const currentMonth = monthDate.getMonth()

  if (!selectedRep) {
    return <div className="py-12 text-center text-sm text-slate-400">Select a rep to view their month.</div>
  }

  return (
    <div>
      <p className="text-sm text-slate-600 mb-4">
        Viewing <span className="font-medium text-slate-800">{repName}</span> — click a week to drill in.
      </p>
      <div className="border border-slate-200 rounded-lg overflow-hidden">
        {/* Day names header */}
        <div className="grid grid-cols-7 bg-slate-50 border-b border-slate-200">
          {DAY_NAMES.map((name) => (
            <div key={name} className="px-2 py-2 text-center text-xs font-medium text-slate-500 uppercase">
              {name}
            </div>
          ))}
        </div>

        {/* Weeks */}
        {weeks.map((week, wi) => {
          const weekMonday = formatDate(week[0])
          return (
            <div
              key={wi}
              className="grid grid-cols-7 border-b border-slate-100 last:border-0 cursor-pointer hover:bg-slate-50/50 transition-colors"
              onClick={() => onWeekClick(weekMonday)}
            >
              {week.map((day) => {
                const dateStr = formatDate(day)
                const isCurrentMonth = day.getMonth() === currentMonth
                const isToday = dateStr === today
                const counts = dayCounts.get(dateStr)

                return (
                  <div
                    key={dateStr}
                    className={`min-h-[80px] p-2 border-r border-slate-100 last:border-0 ${
                      !isCurrentMonth ? 'bg-slate-50/50' : ''
                    }`}
                  >
                    <div className={`text-xs font-medium mb-1 ${
                      isToday ? 'text-white bg-teal-600 w-6 h-6 rounded-full flex items-center justify-center'
                        : isCurrentMonth ? 'text-slate-700' : 'text-slate-300'
                    }`}>
                      {day.getDate()}
                    </div>
                    {counts && isCurrentMonth && (
                      <div className="space-y-0.5">
                        {counts.visits > 0 && (
                          <div className="flex items-center gap-1">
                            <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
                            <span className="text-[10px] text-slate-600">{counts.visits} visit{counts.visits !== 1 ? 's' : ''}</span>
                          </div>
                        )}
                        {counts.nonField > 0 && (
                          <div className="flex items-center gap-1">
                            <span className="w-1.5 h-1.5 rounded-full bg-blue-400" />
                            <span className="text-[10px] text-slate-600">{counts.nonField} other</span>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )
        })}
      </div>
    </div>
  )
}
