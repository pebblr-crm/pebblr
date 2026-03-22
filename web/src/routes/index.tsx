import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Filter, Download } from 'lucide-react'
import { Route as rootRoute } from './__root'
import { TeamPerformanceCard } from '../components/dashboard/TeamPerformanceCard'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { useTeamMembers } from '../services/teams'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

function DashboardPage() {
  const { data: teamData, isLoading: teamLoading } = useTeamMembers({ limit: 10 })

  const teamMembers = teamData?.items ?? []

  if (teamLoading) {
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
      </section>

      <div>
        <TeamPerformanceCard members={teamMembers} />
      </div>
    </motion.div>
  )
}
