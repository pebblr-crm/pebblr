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
  /** Display name for the lead / prospect */
  name: string
  /** Company or organisation */
  company: string
  /** Primary contact email */
  email: string
  /** Primary contact phone */
  phone?: string
  status: LeadStatus
  /** ID of the rep assigned to this lead */
  assigneeId: string
  /** Free-form notes */
  notes?: string
  /** Physical address of the customer site */
  address?: string
  createdAt: string
  updatedAt: string
}

export interface CreateLeadInput {
  name: string
  company: string
  email: string
  phone?: string
  status?: LeadStatus
  assigneeId: string
  notes?: string
  address?: string
}

export interface UpdateLeadInput extends Partial<CreateLeadInput> {
  id: string
}

export interface LeadListParams {
  page?: number
  limit?: number
  status?: LeadStatus
  assigneeId?: string
}
