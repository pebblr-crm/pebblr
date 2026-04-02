import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { buildQuery } from '@/lib/query'
import type { AuditEntry, AuditListParams, AuditStatus } from '@/types/audit'
import type { PaginatedResponse } from '@/types/api'

export const auditKeys = {
  all: ['audit'] as const,
  lists: () => [...auditKeys.all, 'list'] as const,
  list: (params: AuditListParams) => [...auditKeys.lists(), params] as const,
}

export function useAuditLog(params: AuditListParams = {}) {
  return useQuery({
    queryKey: auditKeys.list(params),
    queryFn: () => api.get<PaginatedResponse<AuditEntry>>(buildQuery('/audit', {
      page: params.page,
      limit: params.limit,
      entityType: params.entityType,
      actorId: params.actorId,
      status: params.status,
    })),
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
