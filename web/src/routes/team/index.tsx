import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { MoreVertical, UserPlus } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useTeamMembers } from '../../services/teams'
import type { TeamMember } from '../../types/team'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/team',
  component: TeamPage,
})

function MemberStatusDot({ status }: { status: TeamMember['status'] }) {
  const colorMap = {
    online: 'bg-tertiary-fixed',
    away: 'bg-amber-400',
    offline: 'bg-slate-300',
  }
  return <span className={`w-2.5 h-2.5 rounded-full ${colorMap[status]}`} />
}

function TeamMemberCard({ member }: { member: TeamMember }) {
  const completionRate = member.metrics.assigned > 0
    ? Math.round((member.metrics.completed / member.metrics.assigned) * 100)
    : 0

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-3">
          <div className="relative">
            <img
              src={member.avatar}
              alt={member.name}
              className="w-14 h-14 rounded-full object-cover"
              referrerPolicy="no-referrer"
            />
            <span className="absolute bottom-0 right-0 w-3.5 h-3.5 border-2 border-white rounded-full">
              <MemberStatusDot status={member.status} />
            </span>
          </div>
          <div>
            <p className="font-bold text-on-surface">{member.name}</p>
            <p className="text-xs text-on-surface-variant">{member.role}</p>
          </div>
        </div>
        <button className="p-2 text-slate-400 hover:text-primary transition-colors">
          <MoreVertical className="w-4 h-4" />
        </button>
      </div>

      <div className="grid grid-cols-3 gap-4 mt-4">
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">Assigned</p>
          <p className="text-lg font-extrabold text-primary font-headline">{member.metrics.assigned}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">Completed</p>
          <p className="text-lg font-extrabold text-tertiary-container font-headline">{member.metrics.completed}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">Efficiency</p>
          <p className="text-lg font-extrabold text-on-surface font-headline">{member.metrics.efficiency}%</p>
        </div>
      </div>

      <div className="mt-4">
        <div className="flex justify-between items-center mb-1">
          <span className="text-[10px] font-medium text-on-surface-variant">Completion rate</span>
          <span className="text-[10px] font-bold text-primary">{completionRate}%</span>
        </div>
        <div className="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-primary rounded-full"
            style={{ width: `${completionRate}%` }}
          />
        </div>
      </div>
    </div>
  )
}

export function TeamPage() {
  const { data, isLoading, isError } = useTeamMembers({ limit: 50 })
  const members = data?.items ?? []

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 space-y-8"
    >
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-extrabold tracking-tight text-primary leading-tight font-headline">
            Team Management
          </h1>
          <p className="text-on-surface-variant mt-1 font-medium">
            {members.length > 0 ? `Managing ${members.length} team members` : 'Overview of your sales team'}
          </p>
        </div>
        <button className="px-5 py-2.5 bg-primary text-white rounded-xl text-sm font-semibold flex items-center gap-2 shadow-sm hover:opacity-90 transition-opacity">
          <UserPlus className="w-4 h-4" />
          Add Member
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label="Loading team members..." />
        </div>
      ) : isError ? (
        <div
          data-testid="team-error"
          className="text-center py-12 text-error"
        >
          Failed to load team members. Please try again.
        </div>
      ) : members.length === 0 ? (
        <div
          data-testid="team-empty"
          className="text-center py-12 text-on-surface-variant"
        >
          No team members found.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {members.map((member) => (
            <TeamMemberCard key={member.id} member={member} />
          ))}
        </div>
      )}
    </motion.div>
  )
}
