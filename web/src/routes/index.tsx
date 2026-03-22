import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Filter, Download } from 'lucide-react'
import { Route as rootRoute } from './__root'
import { StatCard } from '../components/dashboard/StatCard'
import { TeamPerformanceCard } from '../components/dashboard/TeamPerformanceCard'
import { UnassignedLeadsCard } from '../components/dashboard/UnassignedLeadCard'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { useDashboardStats } from '../services/dashboard'
import { useTeamMembers } from '../services/teams'
import { useLeads } from '../services/leads'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: teamData, isLoading: teamLoading } = useTeamMembers({ limit: 10 })
  const { data: leadsData, isLoading: leadsLoading } = useLeads({ status: 'new', limit: 5 })

  const isLoading = statsLoading || teamLoading || leadsLoading
  const teamMembers = teamData?.items ?? []
  const unassignedLeads = leadsData?.items ?? []

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" label="Loading dashboard..." />
      </div>
    )
  }

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
              Managing {teamMembers.length} agents across regional territories
            </p>
          </div>
          <div className="flex space-x-3">
            <button className="px-4 py-2 bg-surface-container-high text-on-surface font-semibold rounded-xl text-sm flex items-center hover:bg-surface-container-highest transition-colors">
              <Filter className="w-4 h-4 mr-2" />
              Filter View
            </button>
            <button className="px-4 py-2 bg-surface-container-high text-on-surface font-semibold rounded-xl text-sm flex items-center hover:bg-surface-container-highest transition-colors">
              <Download className="w-4 h-4 mr-2" />
              Export Report
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <StatCard
            label="Total Active Leads"
            value={stats ? String(stats.totalActiveLeads) : '—'}
            change="+12%"
            progress={75}
          />
          <StatCard
            label="Conversion Rate"
            value={stats ? `${stats.conversionRate}%` : '—'}
            change="+3.2%"
            progress={50}
          />
          <StatCard
            label="Avg. Response Time"
            value={stats ? `${stats.avgResponseTimeMinutes}m` : '—'}
            change="-4m"
            progress={66}
          />
          <StatCard
            label="Unassigned Queue"
            value={stats ? String(stats.unassignedCount) : '—'}
            change="High Priority"
            variant="primary"
          />
        </div>
      </section>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <TeamPerformanceCard members={teamMembers} />
        </div>
        <div>
          <UnassignedLeadsCard leads={unassignedLeads} teamMembers={teamMembers} />
        </div>
      </div>
    </motion.div>
  )
}
