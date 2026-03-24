import { MoreVertical, User } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { TeamMember, MemberStatus } from '@/types/team'

const STATUS_COLORS: Record<MemberStatus, string> = {
  online: 'bg-tertiary-fixed',
  away: 'bg-amber-400',
  offline: 'bg-slate-300',
}

interface MemberAvatarProps {
  avatar: string
  name: string
  status: MemberStatus
  /** px size class, e.g. "w-12 h-12" or "w-14 h-14" */
  size?: string
}

export function MemberAvatar({ avatar, name, status, size = 'w-12 h-12' }: MemberAvatarProps) {
  return (
    <div className="relative">
      {avatar ? (
        <img
          src={avatar}
          alt={name}
          className={`${size} rounded-full object-cover`}
          referrerPolicy="no-referrer"
        />
      ) : (
        <div className={`${size} rounded-full bg-surface-container-high flex items-center justify-center`}>
          <User className="w-1/2 h-1/2 text-on-surface-variant" />
        </div>
      )}
      <span
        className={`absolute bottom-0 right-0 w-3 h-3 border-2 border-white rounded-full ${STATUS_COLORS[status]}`}
      />
    </div>
  )
}

/** Compact row variant — used on the dashboard Team Performance panel. */
export function MemberRow({ member }: { member: TeamMember }) {
  const { t } = useTranslation()
  return (
    <div className="bg-white p-4 rounded-xl flex items-center justify-between group hover:shadow-md transition-shadow">
      <div className="flex items-center space-x-4">
        <MemberAvatar avatar={member.avatar} name={member.name} status={member.status} />
        <div>
          <p className="font-bold text-on-surface">{member.name}</p>
          <p className="text-xs text-on-surface-variant">{member.role}</p>
        </div>
      </div>
      <div className="flex space-x-12">
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">{t('team.assigned')}</p>
          <p className="font-extrabold text-primary">{member.metrics.assigned}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">{t('team.completed')}</p>
          <p className="font-extrabold text-tertiary-container">{member.metrics.completed}</p>
        </div>
        <div className="w-32 hidden md:block">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1 text-right">{t('team.efficiency')}</p>
          <div className="h-1.5 w-full bg-slate-100 rounded-full mt-2">
            <div
              className="h-full bg-primary rounded-full"
              style={{ width: `${member.metrics.efficiency}%` }}
            />
          </div>
        </div>
      </div>
      <button className="p-2 text-slate-400 hover:text-primary transition-colors">
        <MoreVertical className="w-5 h-5" />
      </button>
    </div>
  )
}

/** Full card variant — used on the Team Management page. */
export function MemberCard({ member }: { member: TeamMember }) {
  const { t } = useTranslation()
  const completionRate = member.metrics.assigned > 0
    ? Math.round((member.metrics.completed / member.metrics.assigned) * 100)
    : 0

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-3">
          <MemberAvatar avatar={member.avatar} name={member.name} status={member.status} size="w-14 h-14" />
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
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">{t('team.assigned')}</p>
          <p className="text-lg font-extrabold text-primary font-headline">{member.metrics.assigned}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">{t('team.completed')}</p>
          <p className="text-lg font-extrabold text-tertiary-container font-headline">{member.metrics.completed}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">{t('team.efficiency')}</p>
          <p className="text-lg font-extrabold text-on-surface font-headline">{member.metrics.efficiency}%</p>
        </div>
      </div>

      <div className="mt-4">
        <div className="flex justify-between items-center mb-1">
          <span className="text-[10px] font-medium text-on-surface-variant">{t('team.completionRate')}</span>
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
