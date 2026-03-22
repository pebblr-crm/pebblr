import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from '@tanstack/react-query'
import { api } from './api'
import type {
  Target,
  CreateTargetInput,
  UpdateTargetInput,
  TargetListParams,
} from '@/types/target'
import type { PaginatedResponse } from '@/types/api'

// ── Query keys ────────────────────────────────────────────────────────────────

export const targetKeys = {
  all: ['targets'] as const,
  lists: () => [...targetKeys.all, 'list'] as const,
  list: (params: TargetListParams) => [...targetKeys.lists(), params] as const,
  details: () => [...targetKeys.all, 'detail'] as const,
  detail: (id: string) => [...targetKeys.details(), id] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildTargetPath(params: TargetListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.type) qs.set('type', params.type)
  if (params.assignee) qs.set('assignee', params.assignee)
  if (params.q) qs.set('q', params.q)
  const query = qs.toString()
  return query ? `/targets?${query}` : '/targets'
}

interface TargetDetailResponse {
  target: Target
}

export function fetchTargets(
  params: TargetListParams = {},
): Promise<PaginatedResponse<Target>> {
  return api.get<PaginatedResponse<Target>>(buildTargetPath(params))
}

export function fetchTarget(id: string): Promise<Target> {
  return api.get<TargetDetailResponse>(`/targets/${id}`).then((r) => r.target)
}

export function createTarget(input: CreateTargetInput): Promise<Target> {
  return api.post<TargetDetailResponse>('/targets', input).then((r) => r.target)
}

export function updateTarget({ id, ...input }: UpdateTargetInput): Promise<Target> {
  return api.put<TargetDetailResponse>(`/targets/${id}`, input).then((r) => r.target)
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useTargets(
  params: TargetListParams = {},
): UseQueryResult<PaginatedResponse<Target>> {
  return useQuery({
    queryKey: targetKeys.list(params),
    queryFn: () => fetchTargets(params),
  })
}

export function useTarget(id: string): UseQueryResult<Target> {
  return useQuery({
    queryKey: targetKeys.detail(id),
    queryFn: () => fetchTarget(id),
    enabled: Boolean(id),
  })
}

export function useCreateTarget(): UseMutationResult<Target, Error, CreateTargetInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createTarget,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: targetKeys.lists() })
    },
  })
}

export function useUpdateTarget(): UseMutationResult<Target, Error, UpdateTargetInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateTarget,
    onSuccess: (updated) => {
      queryClient.setQueryData(targetKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: targetKeys.lists() })
    },
  })
}
