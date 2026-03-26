import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { AuditEntry, AuditListParams, AuditStatus } from '@/types/audit'
import type { PaginatedResponse } from '@/types/api'

export const auditKeys = {
  all: ['audit'] as const,
  list: (params: AuditListParams) => [...auditKeys.all, params] as const,
}

function buildQuery(params: AuditListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.entityType) qs.set('entityType', params.entityType)
  if (params.actorId) qs.set('actorId', params.actorId)
  if (params.status) qs.set('status', params.status)
  const q = qs.toString()
  return q ? `/audit?${q}` : '/audit'
}

export function useAuditLog(params: AuditListParams = {}) {
  return useQuery({
    queryKey: auditKeys.list(params),
    queryFn: () => api.get<PaginatedResponse<AuditEntry>>(buildQuery(params)),
  })
}

export function useUpdateAuditStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: AuditStatus }) =>
      api.patch<void>(`/audit/${id}/status`, { status }),
    onSuccess: () => qc.invalidateQueries({ queryKey: auditKeys.all }),
  })
}
