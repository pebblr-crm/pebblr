import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type { TeamMember, TeamListParams } from '@/types/team'
import type { PaginatedResponse } from '@/types/api'

export const teamKeys = {
  all: ['teams'] as const,
  lists: () => [...teamKeys.all, 'list'] as const,
  list: (params: TeamListParams) => [...teamKeys.lists(), params] as const,
  details: () => [...teamKeys.all, 'detail'] as const,
  detail: (id: string) => [...teamKeys.details(), id] as const,
}

function buildTeamPath(params: TeamListParams): string {
  const qs = new URLSearchParams()
  if (params.page !== undefined) qs.set('page', String(params.page))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  const query = qs.toString()
  return query ? `/teams?${query}` : '/teams'
}

export function fetchTeamMembers(
  params: TeamListParams = {},
): Promise<PaginatedResponse<TeamMember>> {
  return api.get<PaginatedResponse<TeamMember>>(buildTeamPath(params))
}

export function useTeamMembers(
  params: TeamListParams = {},
): UseQueryResult<PaginatedResponse<TeamMember>> {
  return useQuery({
    queryKey: teamKeys.list(params),
    queryFn: () => fetchTeamMembers(params),
  })
}
