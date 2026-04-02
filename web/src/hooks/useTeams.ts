import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Team, TeamDetail } from '@/types/team'

export const teamKeys = {
  all: ['teams'] as const,
  lists: () => [...teamKeys.all, 'list'] as const,
  details: () => [...teamKeys.all, 'detail'] as const,
  detail: (id: string) => [...teamKeys.details(), id] as const,
}

export function useTeams() {
  return useQuery({
    queryKey: teamKeys.lists(),
    queryFn: () => api.get<{ items: Team[]; total: number }>('/teams'),
  })
}

export function useTeam(id: string) {
  return useQuery({
    queryKey: teamKeys.detail(id),
    queryFn: () => api.get<TeamDetail>(`/teams/${id}`),
    enabled: !!id,
  })
}
