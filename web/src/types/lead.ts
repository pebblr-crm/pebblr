

/**
 * Lead domain types — mirror the Go backend domain model.
 */

export type LeadStatus =
  | 'new'
  | 'contacted'
  | 'qualified'
  | 'proposal'
  | 'won'
  | 'lost'

export interface Lead {
  id: string
  companyName: string
  contactName: string
  contactEmail: string
  /** Primary contact phone */
  phone?: string
  status: LeadStatus
  assigneeId: string
  notes: string
  /** Physical address of the customer site */
  address?: string
  createdAt: string
  updatedAt: string
}

export interface CreateLeadInput {
  companyName: string
  contactName: string
  contactEmail: string
  phone?: string
  status?: LeadStatus
  assigneeId: string
  notes?: string
  address?: string
}

export interface UpdateLeadInput {
  companyName?: string
  contactName?: string
  contactEmail?: string
  phone?: string
  status?: LeadStatus
  assigneeId?: string
  notes?: string
  address?: string
}

export interface LeadsListResponse {
  leads: Lead[]
  total: number
  page: number
  limit: number
}

export interface LeadListParams {
  page?: number
  limit?: number
  status?: LeadStatus
  assigneeId?: string
}
