import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Target, TargetListParams, CreateTargetInput, TargetFrequencyItem } from '@/types/target'
import type { PaginatedResponse } from '@/types/api'

export const targetKeys = {
  all: ['targets'] as const,
  lists: () => [...targetKeys.all, 'list'] as const,
  list: (params: TargetListParams) => [...targetKeys.lists(), params] as const,
  detail: (id: string) => [...targetKeys.all, 'detail', id] as const,
  visitStatus: () => [...targetKeys.all, 'visit-status'] as const,
  frequencyStatus: () => [...targetKeys.all, 'frequency-status'] as const,
}

function buildQuery(params: TargetListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.type) qs.set('type', params.type)
  if (params.assignee) qs.set('assignee', params.assignee)
  if (params.q) qs.set('q', params.q)
  const q = qs.toString()
  return q ? `/targets?${q}` : '/targets'
}

export function useTargets(params: TargetListParams = {}) {
  return useQuery({
    queryKey: targetKeys.list(params),
    queryFn: () => api.get<PaginatedResponse<Target>>(buildQuery(params)),
  })
}

export function useTarget(id: string) {
  return useQuery({
    queryKey: targetKeys.detail(id),
    queryFn: () => api.get<{ target: Target }>(`/targets/${id}`).then((r) => r.target),
    enabled: !!id,
  })
}

export function useCreateTarget() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTargetInput) =>
      api.post<{ target: Target }>('/targets', input).then((r) => r.target),
    onSuccess: () => qc.invalidateQueries({ queryKey: targetKeys.lists() }),
  })
}

export function useTargetVisitStatus() {
  return useQuery({
    queryKey: targetKeys.visitStatus(),
    queryFn: () => api.get<{ items: { targetId: string; lastVisitDate: string }[] }>('/targets/visit-status'),
  })
}

export function useTargetFrequencyStatus() {
  return useQuery({
    queryKey: targetKeys.frequencyStatus(),
    queryFn: () => api.get<{ items: TargetFrequencyItem[] }>('/targets/frequency-status'),
  })
}
