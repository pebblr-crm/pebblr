/**
 * Activity types — mirror the Go backend domain model.
 */

/** Read-only snapshot of a target embedded in an activity response. */
export interface TargetSummary {
  id: string
  targetType: string
  name: string
  fields: Record<string, unknown>
}

export interface Activity {
  id: string
  activityType: string
  label?: string
  status: string
  dueDate: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  targetId?: string
  targetName?: string
  targetSummary?: TargetSummary
  creatorId: string
  jointVisitUserId?: string
  teamId?: string
  submittedAt?: string
  createdAt: string
  updatedAt: string
}

export interface CreateActivityInput {
  activityType: string
  label?: string
  status: string
  dueDate: string
  duration: string
  routing?: string
  fields: Record<string, unknown>
  targetId?: string
  jointVisitUserId?: string
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
