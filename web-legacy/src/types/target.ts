/**
 * Target domain types — mirror the Go backend domain model.
 * Targets are entities that reps visit (doctors, pharmacies, etc.).
 */

export interface Target {
  id: string
  targetType: string
  name: string
  fields: Record<string, unknown>
  assigneeId: string
  teamId: string
  importedAt?: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateTargetInput {
  targetType: string
  name: string
  fields?: Record<string, unknown>
  assigneeId?: string
  teamId?: string
}

export interface UpdateTargetInput {
  id: string
  targetType: string
  name: string
  fields?: Record<string, unknown>
  assigneeId?: string
  teamId?: string
}

export interface TargetListParams {
  page?: number
  limit?: number
  type?: string
  assignee?: string
  q?: string
}

export interface AssignTargetInput {
  id: string
  assigneeId: string
  teamId?: string
}

export interface TargetFrequencyItem {
  targetId: string
  classification: string
  visitCount: number
  required: number
  compliance: number
}
