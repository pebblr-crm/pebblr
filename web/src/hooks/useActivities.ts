import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { buildQuery } from '@/lib/query'
import type { Activity, CreateActivityInput, ActivityListParams } from '@/types/activity'
import type { PaginatedResponse } from '@/types/api'

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (params: ActivityListParams) => [...activityKeys.lists(), params] as const,
  detail: (id: string) => [...activityKeys.all, 'detail', id] as const,
}

export function useActivities(params: ActivityListParams = {}) {
  return useQuery({
    queryKey: activityKeys.list(params),
    queryFn: () => api.get<PaginatedResponse<Activity>>(buildQuery('/activities', {
      page: params.page,
      limit: params.limit,
      activityType: params.activityType,
      status: params.status,
      creatorId: params.creatorId,
      targetId: params.targetId,
      teamId: params.teamId,
      dateFrom: params.dateFrom,
      dateTo: params.dateTo,
    })),
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

export interface BatchCreateItem {
  targetId: string
  dueDate: string
  fields?: Record<string, unknown>
}

interface BatchCreateResult {
  created: Activity[]
  errors: Array<{ targetId: string; error: string }>
}

export function useBatchCreateActivities() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (items: BatchCreateItem[]) =>
      api.post<BatchCreateResult>('/activities/batch', { items }),
    onSuccess: () => qc.invalidateQueries({ queryKey: activityKeys.lists() }),
  })
}

export function usePatchActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: { id: string; dueDate?: string; status?: string; fields?: Record<string, unknown> }) =>
      api.patch<void>(`/activities/${id}`, body),
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
