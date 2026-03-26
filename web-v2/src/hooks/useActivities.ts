import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Activity, CreateActivityInput, ActivityListParams } from '@/types/activity'
import type { PaginatedResponse } from '@/types/api'

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (params: ActivityListParams) => [...activityKeys.lists(), params] as const,
  detail: (id: string) => [...activityKeys.all, 'detail', id] as const,
}

function buildQuery(params: ActivityListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.activityType) qs.set('activityType', params.activityType)
  if (params.status) qs.set('status', params.status)
  if (params.creatorId) qs.set('creatorId', params.creatorId)
  if (params.targetId) qs.set('targetId', params.targetId)
  if (params.teamId) qs.set('teamId', params.teamId)
  if (params.dateFrom) qs.set('dateFrom', params.dateFrom)
  if (params.dateTo) qs.set('dateTo', params.dateTo)
  const q = qs.toString()
  return q ? `/activities?${q}` : '/activities'
}

export function useActivities(params: ActivityListParams = {}) {
  return useQuery({
    queryKey: activityKeys.list(params),
    queryFn: () => api.get<PaginatedResponse<Activity>>(buildQuery(params)),
  })
}

export function useActivity(id: string) {
  return useQuery({
    queryKey: activityKeys.detail(id),
    queryFn: () => api.get<{ activity: Activity }>(`/activities/${id}`).then((r) => r.activity),
    enabled: !!id,
  })
}

export function useCreateActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateActivityInput) => api.post<Activity>('/activities', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: activityKeys.lists() }),
  })
}

export function usePatchActivityStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status, fields }: { id: string; status: string; fields?: Record<string, unknown> }) =>
      api.patch<void>(`/activities/${id}/status`, { status, fields }),
    onSuccess: () => qc.invalidateQueries({ queryKey: activityKeys.all }),
  })
}

export function useSubmitActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.post<void>(`/activities/${id}/submit`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: activityKeys.all }),
  })
}

export function useCloneWeek() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { sourceWeekStart: string; targetWeekStart: string }) =>
      api.post<void>('/activities/clone-week', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: activityKeys.lists() }),
  })
}
