import { useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useActivityStats, useCoverage, useFrequency, useRecoveryBalance } from '@/hooks/useDashboard'
import { useTeams } from '@/hooks/useTeams'
import { StatCard } from '@/components/data/StatCard'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import type { FrequencyItem } from '@/types/dashboard'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: DashboardPage,
})

const freqColumnHelper = createColumnHelper<FrequencyItem>()

function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useActivityStats({})
  const { data: coverage } = useCoverage({})
  const { data: frequency } = useFrequency({})
  const { data: recovery } = useRecoveryBalance({})
  const { data: teamsData } = useTeams()

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
          const v = Math.round(info.getValue() * 100)
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

  return (
    <div className="p-4 space-y-4 md:p-6 md:space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Team Dashboard</h1>
          <p className="text-sm text-slate-500">
            {teamsData?.items?.length ?? 0} teams
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
  )
}
