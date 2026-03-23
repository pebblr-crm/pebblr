/**
 * Activity types — mirror the Go backend domain model.
 */

export interface Activity {
  id: string
  activityType: string
  status: string
  dueDate: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  targetId?: string
  creatorId: string
  teamId?: string
  submittedAt?: string
  createdAt: string
  updatedAt: string
}

export interface CreateActivityInput {
  activityType: string
  status: string
  dueDate: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  targetId?: string
}

export interface UpdateActivityInput extends CreateActivityInput {
  id: string
}

export interface ActivityListParams {
  page?: number
  limit?: number
  activityType?: string
  status?: string
  creatorId?: string
  targetId?: string
  teamId?: string
  dateFrom?: string
  dateTo?: string
}

export interface StatusPatchInput {
  id: string
  status: string
}

export interface ValidationFieldError {
  field: string
  message: string
}

export interface ActivityValidationError {
  error: { code: string; message: string }
  fields: ValidationFieldError[]
}
