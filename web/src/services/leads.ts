import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
  type UseMutationResult,
} from '@tanstack/react-query'
import { apiGet, apiPost, apiPatch, apiDelete } from './api'
import type {
  Lead,
  CreateLeadInput,
  UpdateLeadInput,
  LeadListParams,
} from '@/types/lead'
import type { PaginatedResponse } from '@/types/api'

// ── Query keys ────────────────────────────────────────────────────────────────

export const leadKeys = {
  all: ['leads'] as const,
  lists: () => [...leadKeys.all, 'list'] as const,
  list: (params: LeadListParams) => [...leadKeys.lists(), params] as const,
  details: () => [...leadKeys.all, 'detail'] as const,
  detail: (id: string) => [...leadKeys.details(), id] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildLeadPath(params: LeadListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.status) qs.set('status', params.status)
  if (params.assigneeId) qs.set('assignee', params.assigneeId)
  const query = qs.toString()
  return query ? `/leads?${query}` : '/leads'
}

export function fetchLeads(
  params: LeadListParams = {},
): Promise<PaginatedResponse<Lead>> {
  return apiGet<PaginatedResponse<Lead>>(buildLeadPath(params))
}

export function fetchLead(id: string): Promise<Lead> {
  return apiGet<Lead>(`/leads/${id}`)
}

export function createLead(input: CreateLeadInput): Promise<Lead> {
  return apiPost<Lead>('/leads', input)
}

export function updateLead({ id, ...input }: UpdateLeadInput): Promise<Lead> {
  return apiPatch<Lead>(`/leads/${id}`, input)
}

export function deleteLead(id: string): Promise<void> {
  return apiDelete(`/leads/${id}`)
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useLeads(
  params: LeadListParams = {},
): UseQueryResult<PaginatedResponse<Lead>> {
  return useQuery({
    queryKey: leadKeys.list(params),
    queryFn: () => fetchLeads(params),
  })
}

export function useLead(id: string): UseQueryResult<Lead> {
  return useQuery({
    queryKey: leadKeys.detail(id),
    queryFn: () => fetchLead(id),
    enabled: Boolean(id),
  })
}

export function useCreateLead(): UseMutationResult<Lead, Error, CreateLeadInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createLead,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: leadKeys.lists() })
    },
  })
}

export function useUpdateLead(): UseMutationResult<Lead, Error, UpdateLeadInput> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateLead,
    onSuccess: (updated) => {
      queryClient.setQueryData(leadKeys.detail(updated.id), updated)
      void queryClient.invalidateQueries({ queryKey: leadKeys.lists() })
    },
  })
}

export function useDeleteLead(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteLead,
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: leadKeys.detail(id) })
      void queryClient.invalidateQueries({ queryKey: leadKeys.lists() })
    },
  })
}
