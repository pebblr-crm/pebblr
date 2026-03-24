import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
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
import { translateConfigLabel } from '@/utils/config'
import type { DashboardFilter } from '../types/dashboard'
import { formatPeriod } from '../utils/date'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function currentPeriod(): string {
  return formatPeriod(new Date())
}

export function DashboardPage() {
  const { t } = useTranslation()
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
        <LoadingSpinner size="lg" label={t('dashboard.loading')} />
      </div>
    )
  }

  const statuses = config?.activities.statuses ?? []
  const transitions = config?.activities.status_transitions ?? {}
  const initialStatus = statuses.find((s) => s.initial)
  const plannedKey = initialStatus?.key ?? statuses[0]?.key
  // The first allowed transition from the initial status is the "completed" outcome.
  const completedKey = plannedKey ? transitions[plannedKey]?.[0] : undefined
  const planned = plannedKey ? (activityStats?.byStatus[plannedKey] ?? 0) : 0
  const completed = completedKey ? (activityStats?.byStatus[completedKey] ?? 0) : 0
  const realizationPct = planned > 0 ? Math.round((completed / planned) * 100) : 0

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 space-y-6 sm:space-y-8"
    >
      <section>
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-end gap-4 mb-6">
          <div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-primary tracking-tight font-headline">
              {t('dashboard.title')}
            </h1>
            <p className="text-on-surface-variant text-sm sm:text-base">
              {t('dashboard.subtitle')}
            </p>
          </div>
          <PeriodSelector period={period} onPeriodChange={setPeriod} />
        </div>
      </section>

      {/* Top-level KPI cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard
          label={initialStatus ? translateConfigLabel(`status.${initialStatus.key}`, initialStatus.label) : (statuses[0] ? translateConfigLabel(`status.${statuses[0].key}`, statuses[0].label) : t('dashboard.planned'))}
          value={String(planned)}
          variant="default"
        />
        <StatCard
          label={completedKey ? translateConfigLabel(`status.${completedKey}`, statuses.find((s) => s.key === completedKey)?.label ?? completedKey) : t('dashboard.completed')}
          value={String(completed)}
          variant="primary"
          progress={realizationPct}
        />
        <StatCard
          label={t('dashboard.completionRate')}
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
