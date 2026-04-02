import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Territory, CreateTerritoryInput } from '@/types/territory'

export const territoryKeys = {
  all: ['territories'] as const,
  lists: () => [...territoryKeys.all, 'list'] as const,
  details: () => [...territoryKeys.all, 'detail'] as const,
  detail: (id: string) => [...territoryKeys.details(), id] as const,
}

export function useTerritories() {
  return useQuery({
    queryKey: territoryKeys.lists(),
    queryFn: () => api.get<{ items: Territory[]; total: number }>('/territories'),
  })
}

export function useTerritory(id: string) {
  return useQuery({
    queryKey: territoryKeys.detail(id),
    queryFn: () => api.get<Territory>(`/territories/${id}`),
    enabled: !!id,
  })
}

export function useCreateTerritory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateTerritoryInput) => api.post<Territory>('/territories', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: territoryKeys.lists() }),
  })
}

export function useDeleteTerritory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/territories/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: territoryKeys.lists() }),
  })
}
