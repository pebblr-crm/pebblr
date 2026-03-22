
/**
 * Lead domain types — mirror the Go backend domain model.
 * Field names use camelCase (frontend convention), mapping to the
 * snake_case json tags on the Go domain.Lead struct.
 */

export type LeadStatus =
  | 'new'
  | 'assigned'
  | 'in_progress'
  | 'visited'
  | 'closed_won'
  | 'closed_lost'

export type CustomerType =
  | 'retail'
  | 'wholesale'
  | 'hospitality'
  | 'institutional'
  | 'other'

export interface Lead {
  id: string
  title: string
  description: string
  status: LeadStatus
  assigneeId: string
  teamId: string
  customerId: string
  customerType: CustomerType
  company: string
  industry: string
  location: string
  valueCents: number
  initials: string
  createdAt: string
  updatedAt: string
  deletedAt?: string | null
}

export interface CreateLeadInput {
  title: string
  description?: string
  status?: LeadStatus
  assigneeId: string
  teamId: string
  customerId: string
  customerType: CustomerType
}

export interface UpdateLeadInput {
  id: string
  title?: string
  description?: string
  status?: LeadStatus
  assigneeId?: string
  teamId?: string
  customerId?: string
  customerType?: CustomerType
}

export interface LeadListParams {
  page?: number
  limit?: number
  status?: LeadStatus
  assigneeId?: string
}
