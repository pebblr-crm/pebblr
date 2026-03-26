import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Territory, CreateTerritoryInput } from '@/types/territory'

export const territoryKeys = {
  all: ['territories'] as const,
  detail: (id: string) => [...territoryKeys.all, id] as const,
}

export function useTerritories() {
  return useQuery({
    queryKey: territoryKeys.all,
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
    onSuccess: () => qc.invalidateQueries({ queryKey: territoryKeys.all }),
  })
}

export function useDeleteTerritory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/territories/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: territoryKeys.all }),
  })
}
