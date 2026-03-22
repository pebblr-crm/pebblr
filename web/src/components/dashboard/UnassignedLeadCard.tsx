import { Filter, Send } from 'lucide-react'
import type { Lead } from '@/types/lead'
import type { TeamMember } from '@/types/team'

interface UnassignedLeadItemProps {
  lead: Lead
  teamMembers: TeamMember[]
}

function UnassignedLeadItem({ lead, teamMembers }: UnassignedLeadItemProps) {
  return (
    <div className="p-4 rounded-xl bg-surface-container-low border-l-4 border-primary group">
      <div className="flex justify-between items-start mb-2">
        <span className="text-[10px] font-bold bg-white px-2 py-1 rounded-md text-primary shadow-sm uppercase">
          {lead.status}
        </span>
        <span className="text-[10px] font-medium text-on-surface-variant">
          {new Date(lead.createdAt).toLocaleDateString()}
        </span>
      </div>
      <p className="font-bold text-on-surface leading-tight">{lead.title}</p>
      <p className="text-xs text-on-surface-variant mb-4">{lead.customerType}</p>
      <div className="flex items-center space-x-2">
        <select className="flex-1 bg-white border-none rounded-lg text-xs font-medium py-2 focus:ring-1 focus:ring-primary shadow-sm outline-none">
          <option>Assign Agent...</option>
          {teamMembers.map((m) => (
            <option key={m.id}>{m.name}</option>
          ))}
        </select>
        <button className="p-2 bg-primary text-white rounded-lg hover:bg-primary-container transition-colors">
          <Send className="w-4 h-4" />
        </button>
      </div>
    </div>
  )
}

interface UnassignedLeadsCardProps {
  leads: Lead[]
  teamMembers: TeamMember[]
}

export function UnassignedLeadsCard({ leads, teamMembers }: UnassignedLeadsCardProps) {
  return (
    <div className="bg-white p-6 rounded-xl shadow-sm border border-slate-50 flex flex-col h-full">
      <div className="flex justify-between items-center mb-6">
        <div className="flex items-center space-x-2">
          <h3 className="text-lg font-bold text-primary font-headline">Unassigned Leads</h3>
          <span className="bg-primary-fixed text-primary px-2 py-0.5 rounded-full text-[10px] font-bold">
            {leads.length}
          </span>
        </div>
        <button className="text-primary hover:text-primary-container">
          <Filter className="w-4 h-4" />
        </button>
      </div>
      <div className="space-y-4 flex-1">
        {leads.map((lead) => (
          <UnassignedLeadItem key={lead.id} lead={lead} teamMembers={teamMembers} />
        ))}
        {leads.length === 0 && (
          <p className="text-sm text-on-surface-variant text-center py-8">No unassigned leads.</p>
        )}
      </div>
      <button className="mt-6 w-full py-3 text-[10px] font-bold text-primary hover:bg-slate-50 rounded-xl transition-colors uppercase tracking-widest border border-slate-100">
        View Entire Queue
      </button>
    </div>
  )
}
