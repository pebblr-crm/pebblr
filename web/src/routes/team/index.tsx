import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { UserPlus } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { MemberCard } from '../../components/MemberCard'
import { useTeamMembers } from '../../services/teams'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/team',
  component: TeamPage,
})

export function TeamPage() {
  const { t } = useTranslation()
  const { data, isLoading, isError } = useTeamMembers({ limit: 50 })
  const members = data?.items ?? []

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 space-y-6 sm:space-y-8"
    >
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-end gap-4">
        <div>
          <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight text-primary leading-tight font-headline">
            {t('team.title')}
          </h1>
          <p className="text-on-surface-variant mt-1 font-medium">
            {members.length > 0 ? t('team.subtitleWithCount', { count: members.length }) : t('team.subtitleEmpty')}
          </p>
        </div>
        <button className="px-5 py-2.5 bg-primary text-white rounded-xl text-sm font-semibold flex items-center gap-2 shadow-sm hover:opacity-90 transition-opacity">
          <UserPlus className="w-4 h-4" />
          {t('team.addMember')}
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label={t('team.loading')} />
        </div>
      ) : isError ? (
        <div
          data-testid="team-error"
          className="text-center py-12 text-error"
        >
          {t('team.error')}
        </div>
      ) : members.length === 0 ? (
        <div
          data-testid="team-empty"
          className="text-center py-12 text-on-surface-variant"
        >
          {t('team.empty')}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {members.map((member) => (
            <MemberCard key={member.id} member={member} />
          ))}
        </div>
      )}
    </motion.div>
  )
}
