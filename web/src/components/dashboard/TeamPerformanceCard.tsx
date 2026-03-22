import { MoreVertical } from 'lucide-react'
import type { TeamMember } from '@/types/team'

interface TeamMemberRowProps {
  member: TeamMember
}

function TeamMemberRow({ member }: TeamMemberRowProps) {
  return (
    <div className="bg-white p-4 rounded-xl flex items-center justify-between group hover:shadow-md transition-shadow">
      <div className="flex items-center space-x-4">
        <div className="relative">
          <img
            src={member.avatar}
            alt={member.name}
            className="w-12 h-12 rounded-full object-cover"
            referrerPolicy="no-referrer"
          />
          <span
            className={`absolute bottom-0 right-0 w-3 h-3 border-2 border-white rounded-full ${
              member.status === 'online' ? 'bg-tertiary-fixed' : 'bg-amber-400'
            }`}
          />
        </div>
        <div>
          <p className="font-bold text-on-surface">{member.name}</p>
          <p className="text-xs text-on-surface-variant">{member.role}</p>
        </div>
      </div>
      <div className="flex space-x-12">
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">Assigned</p>
          <p className="font-extrabold text-primary">{member.metrics.assigned}</p>
        </div>
        <div className="text-center">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1">Completed</p>
          <p className="font-extrabold text-tertiary-container">{member.metrics.completed}</p>
        </div>
        <div className="w-32 hidden md:block">
          <p className="text-[10px] font-bold text-on-surface-variant uppercase mb-1 text-right">Efficiency</p>
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

interface TeamPerformanceCardProps {
  members: TeamMember[]
}

export function TeamPerformanceCard({ members }: TeamPerformanceCardProps) {
  return (
    <div className="bg-surface-container-low p-6 rounded-xl">
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-xl font-bold text-primary font-headline">Team Performance</h3>
        <div className="flex bg-white p-1 rounded-lg border border-slate-100">
          <button className="px-3 py-1 text-[10px] font-bold bg-slate-100 text-primary rounded">WEEKLY</button>
          <button className="px-3 py-1 text-[10px] font-medium text-slate-500 hover:text-primary">MONTHLY</button>
        </div>
      </div>
      <div className="space-y-4">
        {members.map((member) => (
          <TeamMemberRow key={member.id} member={member} />
        ))}
      </div>
    </div>
  )
}
