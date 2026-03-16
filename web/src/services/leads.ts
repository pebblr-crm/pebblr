import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './api'
import type {
  Lead,
  CreateLeadInput,
  UpdateLeadInput,
  LeadsListResponse,
} from '../types/lead'

const LEADS_KEY = 'leads'

interface LeadsParams {
  page?: number
  limit?: number
  status?: string
  assignee?: string
}

export function useLeads(params?: LeadsParams) {
  const searchParams = new URLSearchParams()
  if (params?.page) searchParams.set('page', String(params.page))
  if (params?.limit) searchParams.set('limit', String(params.limit))
  if (params?.status) searchParams.set('status', params.status)
  if (params?.assignee) searchParams.set('assignee', params.assignee)
  const query = searchParams.toString()
  const url = `/api/v1/leads${query ? `?${query}` : ''}`

  return useQuery<LeadsListResponse>({
    queryKey: [LEADS_KEY, params],
    queryFn: () => api.get<LeadsListResponse>(url),
  })
}

export function useLead(id: string) {
  return useQuery<Lead>({
    queryKey: [LEADS_KEY, id],
    queryFn: () => api.get<Lead>(`/api/v1/leads/${id}`),
    enabled: !!id,
  })
}

export function useCreateLead() {
  const queryClient = useQueryClient()
  return useMutation<Lead, Error, CreateLeadInput>({
    mutationFn: (input) => api.post<Lead>('/api/v1/leads', input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [LEADS_KEY] })
    },
  })
}

export function useUpdateLead(id: string) {
  const queryClient = useQueryClient()
  return useMutation<Lead, Error, UpdateLeadInput>({
    mutationFn: (input) => api.patch<Lead>(`/api/v1/leads/${id}`, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [LEADS_KEY] })
    },
  })
}

export function useDeleteLead() {
  const queryClient = useQueryClient()
  return useMutation<void, Error, string>({
    mutationFn: (id) => api.delete<void>(`/api/v1/leads/${id}`),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: [LEADS_KEY] })
    },
  })
}
