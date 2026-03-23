import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Route as rootRoute } from './__root'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { StatCard } from '../components/dashboard/StatCard'
import { PeriodSelector } from '../components/dashboard/PeriodSelector'
import { ActivityStatsCard } from '../components/dashboard/ActivityStatsCard'
import { CoverageCard } from '../components/dashboard/CoverageCard'
import { FrequencyTable } from '../components/dashboard/FrequencyTable'
import { TeamPerformanceCard } from '../components/dashboard/TeamPerformanceCard'
import { useActivityStats, useCoverage, useFrequency } from '../services/dashboard'
import { useConfig } from '../services/config'
import { useTeamMembers } from '../services/teams'
import type { DashboardFilter } from '../types/dashboard'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function currentPeriod(): string {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

export function DashboardPage() {
  const [period, setPeriod] = useState(currentPeriod)

  const filter: DashboardFilter = { period }

  const { data: activityStats, isLoading: statsLoading } = useActivityStats(filter)
  const { data: coverage, isLoading: coverageLoading } = useCoverage(filter)
  const { data: frequency, isLoading: frequencyLoading } = useFrequency(filter)
  const { data: config } = useConfig()
  const { data: teamData, isLoading: teamLoading } = useTeamMembers({ limit: 10 })

  const teamMembers = teamData?.items ?? []
  const isLoading = statsLoading || coverageLoading || frequencyLoading || teamLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" label="Loading dashboard..." />
      </div>
    )
  }

  const planned = activityStats?.byStatus['planificat'] ?? 0
  const realized = activityStats?.byStatus['realizat'] ?? 0
  const realizationPct = planned > 0 ? Math.round((realized / planned) * 100) : 0

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
              Command Center
            </h1>
            <p className="text-on-surface-variant">
              Activity-based KPIs for your team
            </p>
          </div>
          <PeriodSelector period={period} onPeriodChange={setPeriod} />
        </div>
      </section>

      {/* Top-level KPI cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard
          label="Planned"
          value={String(planned)}
          variant="default"
        />
        <StatCard
          label="Realized"
          value={String(realized)}
          variant="primary"
          progress={realizationPct}
        />
        <StatCard
          label="Realization Rate"
          value={`${realizationPct}%`}
          progress={realizationPct}
        />
      </div>

      {/* Detailed cards */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {activityStats && (
          <ActivityStatsCard data={activityStats} activitiesConfig={config?.activities} />
        )}
        {coverage && <CoverageCard data={coverage} />}
        {frequency && <FrequencyTable items={frequency.items} />}
      </div>

      {/* Team performance */}
      {teamMembers.length > 0 && (
        <TeamPerformanceCard members={teamMembers} />
      )}
    </motion.div>
  )
}
