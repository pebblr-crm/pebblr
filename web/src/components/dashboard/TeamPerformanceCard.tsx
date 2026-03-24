import { useTranslation } from 'react-i18next'
import { MemberRow } from '../MemberCard'
import type { TeamMember } from '@/types/team'

interface TeamPerformanceCardProps {
  members: TeamMember[]
}

export function TeamPerformanceCard({ members }: TeamPerformanceCardProps) {
  const { t } = useTranslation()

  return (
    <div className="bg-surface-container-low p-6 rounded-xl">
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-xl font-bold text-primary font-headline">{t('dashboardCards.teamPerformance')}</h3>
        <div className="flex bg-white p-1 rounded-lg border border-slate-100">
          <button className="px-3 py-1 text-[10px] font-bold bg-slate-100 text-primary rounded">{t('dashboardCards.weekly')}</button>
          <button className="px-3 py-1 text-[10px] font-medium text-slate-500 hover:text-primary">{t('dashboardCards.monthly')}</button>
        </div>
      </div>
      <div className="space-y-4">
        {members.map((member) => (
          <MemberRow key={member.id} member={member} />
        ))}
      </div>
    </div>
  )
}
