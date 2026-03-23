import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Route as rootRoute } from './__root'
import { StatCard } from '../components/dashboard/StatCard'
import { PeriodSelector } from '../components/dashboard/PeriodSelector'
import { UserStatsTable } from '../components/dashboard/UserStatsTable'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { useDashboardStats, useCoverageStats, useUserStats } from '../services/dashboard'
import { useConfig } from '../services/config'
import type { TenantConfig } from '@/types/config'
import type { DashboardStatsResponse } from '@/types/dashboard'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function currentPeriod(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function resolveStatusLabel(config: TenantConfig | undefined, key: string): string {
  if (!config) return key
  const st = config.activities.statuses.find((s) => s.key === key)
  return st?.label ?? key
}

function StatusBreakdown({
  stats,
  config,
}: {
  stats: DashboardStatsResponse
  config?: TenantConfig
}) {
  if (stats.stats.byStatus.length === 0) return null

  return (
    <div className="bg-surface-container-low p-6 rounded-xl">
      <h3 className="text-lg font-bold text-primary font-headline mb-4">By Status</h3>
      <div className="space-y-3">
        {stats.stats.byStatus.map((sc) => {
          const pct = stats.stats.total > 0 ? Math.round((sc.count / stats.stats.total) * 100) : 0
          return (
            <div key={sc.status}>
              <div className="flex justify-between items-center mb-1">
                <span className="text-sm font-medium text-on-surface">
                  {resolveStatusLabel(config, sc.status)}
                </span>
                <span className="text-sm font-bold text-on-surface-variant">
                  {sc.count} ({pct}%)
                </span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary rounded-full transition-all"
                  style={{ width: `${pct}%` }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function CategoryBreakdown({ stats }: { stats: DashboardStatsResponse }) {
  if (stats.byCategory.length === 0) return null

  const CATEGORY_LABELS: Record<string, string> = {
    field: 'Field Activities',
    non_field: 'Non-field Activities',
  }
  const CATEGORY_COLORS: Record<string, string> = {
    field: 'bg-amber-400',
    non_field: 'bg-blue-400',
  }

  return (
    <div className="bg-surface-container-low p-6 rounded-xl">
      <h3 className="text-lg font-bold text-primary font-headline mb-4">Field vs Non-field</h3>
      <div className="space-y-3">
        {stats.byCategory.map((cc) => {
          const pct = stats.stats.total > 0 ? Math.round((cc.count / stats.stats.total) * 100) : 0
          return (
            <div key={cc.category}>
              <div className="flex justify-between items-center mb-1">
                <span className="text-sm font-medium text-on-surface">
                  {CATEGORY_LABELS[cc.category] ?? cc.category}
                </span>
                <span className="text-sm font-bold text-on-surface-variant">
                  {cc.count} ({pct}%)
                </span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all ${CATEGORY_COLORS[cc.category] ?? 'bg-primary'}`}
                  style={{ width: `${pct}%` }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function DashboardPage() {
  const [period, setPeriod] = useState(currentPeriod)
  const params = useMemo(() => ({ period }), [period])

  const { data: config } = useConfig()
  const { data: stats, isLoading: statsLoading } = useDashboardStats(params)
  const { data: coverage, isLoading: coverageLoading } = useCoverageStats(params)
  const { data: userStatsData, isLoading: userStatsLoading } = useUserStats(params)

  const isLoading = statsLoading || coverageLoading || userStatsLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" label="Loading dashboard..." />
      </div>
    )
  }

  const submitRate =
    stats && stats.stats.total > 0
      ? Math.round((stats.stats.submittedCount / stats.stats.total) * 100)
      : 0

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 space-y-8"
    >
      <section>
        <div className="flex justify-between items-end mb-6">
          <div>
            <h1 className="text-3xl font-extrabold text-primary tracking-tight font-headline">
              Dashboard
            </h1>
            <p className="text-on-surface-variant">
              Activity metrics for the selected period
            </p>
          </div>
          <PeriodSelector period={period} onPeriodChange={setPeriod} />
        </div>
      </section>

      {/* KPI stat cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Total Activities"
          value={String(stats?.stats.total ?? 0)}
          variant="primary"
        />
        <StatCard
          label="Submitted"
          value={String(stats?.stats.submittedCount ?? 0)}
          progress={submitRate}
        />
        <StatCard
          label="Target Coverage"
          value={coverage ? `${Math.round(coverage.coveragePercent)}%` : '—'}
          change={coverage ? `${coverage.visitedTargets}/${coverage.totalTargets}` : undefined}
          progress={coverage ? Math.round(coverage.coveragePercent) : undefined}
        />
        <StatCard
          label="Targets Visited"
          value={String(coverage?.visitedTargets ?? 0)}
        />
      </div>

      {/* Status + category breakdowns */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {stats && <StatusBreakdown stats={stats} config={config} />}
        {stats && <CategoryBreakdown stats={stats} />}
      </div>

      {/* Per-user stats table */}
      {userStatsData && (
        <UserStatsTable users={userStatsData.users} config={config} />
      )}
    </motion.div>
  )
}
