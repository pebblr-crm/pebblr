export type LeadStatus = 'new' | 'contacted' | 'qualified' | 'won' | 'lost'

export interface Lead {
  id: string
  companyName: string
  contactName: string
  contactEmail: string
  status: LeadStatus
  assigneeId: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface CreateLeadInput {
  companyName: string
  contactName: string
  contactEmail: string
  status: LeadStatus
  assigneeId: string
  notes?: string
}

export interface UpdateLeadInput {
  companyName?: string
  contactName?: string
  contactEmail?: string
  status?: LeadStatus
  assigneeId?: string
  notes?: string
}

export interface LeadsListResponse {
  leads: Lead[]
  total: number
  page: number
  limit: number
}
