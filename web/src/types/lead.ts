

/**
 * Lead domain types — mirror the Go backend domain model.
 */

export type LeadStatus =
  | 'new'
  | 'assigned'
  | 'in_progress'
  | 'visited'
  | 'closed_won'
  | 'closed_lost'

export interface Lead {
  id: string
  title: string
  description: string
  status: LeadStatus
  assigneeId: string
  teamId: string
  customerId: string
  customerType: string
  createdAt: string
  updatedAt: string
}

export interface CreateLeadInput {
  title: string
  description?: string
  status?: LeadStatus
  assigneeId: string
  teamId: string
  customerId: string
  customerType: string
}

export interface UpdateLeadInput {
  id: string
  title?: string
  description?: string
  status?: LeadStatus
  assigneeId?: string
  teamId?: string
  customerId?: string
  customerType?: string
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
