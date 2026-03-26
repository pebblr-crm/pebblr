export type AuditStatus = 'pending' | 'accepted' | 'false_positive'

export interface AuditEntry {
  id: string
  entityType: string
  entityId: string
  eventType: string
  actorId: string
  oldValue?: Record<string, unknown>
  newValue?: Record<string, unknown>
  status: AuditStatus
  reviewedBy?: string
  reviewedAt?: string
  createdAt: string
}

export interface AuditListParams {
  page?: number
  limit?: number
  entityType?: string
  actorId?: string
  status?: AuditStatus
}
