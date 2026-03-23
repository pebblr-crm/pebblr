import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from '@tanstack/react-query'
import { api } from './api'
import type {
  Activity,
  CreateActivityInput,
  UpdateActivityInput,
  ActivityListParams,
  StatusPatchInput,
} from '@/types/activity'
import type { PaginatedResponse } from '@/types/api'

// ── Query keys ────────────────────────────────────────────────────────────────

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (params: ActivityListParams) => [...activityKeys.lists(), params] as const,
  details: () => [...activityKeys.all, 'detail'] as const,
  detail: (id: string) => [...activityKeys.details(), id] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildActivityPath(params: ActivityListParams): string {
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
  const query = qs.toString()
  return query ? `/activities?${query}` : '/activities'
}

interface ActivityDetailResponse {
  activity: Activity
}

export function fetchActivities(
  params: ActivityListParams = {},
): Promise<PaginatedResponse<Activity>> {
  return api.get<PaginatedResponse<Activity>>(buildActivityPath(params))
}

export function fetchActivity(id: string): Promise<Activity> {
  return api.get<ActivityDetailResponse>(`/activities/${id}`).then((r) => r.activity)
}

export function createActivity(input: CreateActivityInput): Promise<Activity> {
  return api.post<ActivityDetailResponse>('/activities', input).then((r) => r.activity)
}

export function updateActivity({ id, ...input }: UpdateActivityInput): Promise<Activity> {
  return api.put<ActivityDetailResponse>(`/activities/${id}`, input).then((r) => r.activity)
}

export function deleteActivity(id: string): Promise<void> {
  return api.delete<void>(`/activities/${id}`)
}

export function submitActivity(id: string): Promise<Activity> {
  return api.post<ActivityDetailResponse>(`/activities/${id}/submit`, {}).then((r) => r.activity)
}

export function patchActivityStatus({ id, status }: StatusPatchInput): Promise<Activity> {
  return api.patch<ActivityDetailResponse>(`/activities/${id}/status`, { status }).then((r) => r.activity)
}

export function patchActivity({ id, ...input }: Partial<UpdateActivityInput> & { id: string }): Promise<Activity> {
  return api.patch<ActivityDetailResponse>(`/activities/${id}`, input).then((r) => r.activity)
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useActivities(
  params: ActivityListParams = {},
): UseQueryResult<PaginatedResponse<Activity>> {
  return useQuery({
    queryKey: activityKeys.list(params),
    queryFn: () => fetchActivities(params),
  })
}

export function useActivity(id: string): UseQueryResult<Activity> {
  return useQuery({
    queryKey: activityKeys.detail(id),
    queryFn: () => fetchActivity(id),
    enabled: Boolean(id),
  })
}

export function useCreateActivity(): UseMutationResult<Activity, Error, CreateActivityInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createActivity,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function useUpdateActivity(): UseMutationResult<Activity, Error, UpdateActivityInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateActivity,
    onSuccess: (updated) => {
      queryClient.setQueryData(activityKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function useDeleteActivity(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteActivity,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function useSubmitActivity(): UseMutationResult<Activity, Error, string> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: submitActivity,
    onSuccess: (updated) => {
      queryClient.setQueryData(activityKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function usePatchActivityStatus(): UseMutationResult<Activity, Error, StatusPatchInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: patchActivityStatus,
    onSuccess: (updated) => {
      queryClient.setQueryData(activityKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

// ── Batch create ─────────────────────────────────────────────────────────────

export interface BatchCreateItem {
  targetId: string
  dueDate: string
}

export interface BatchCreateResult {
  created: Activity[]
  errors: Array<{ targetId: string; error: string }>
}

export function batchCreateActivities(items: BatchCreateItem[]): Promise<BatchCreateResult> {
  return api.post<BatchCreateResult>('/activities/batch', { items })
}

export function useBatchCreateActivities(): UseMutationResult<BatchCreateResult, Error, BatchCreateItem[]> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: batchCreateActivities,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}

export function usePatchActivity(): UseMutationResult<Activity, Error, Partial<UpdateActivityInput> & { id: string }> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: patchActivity,
    onSuccess: (updated) => {
      queryClient.setQueryData(activityKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: activityKeys.lists() })
    },
  })
}
