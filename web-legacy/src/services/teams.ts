import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type { TeamMember, TeamListParams, MemberStatus } from '@/types/team'
import type { PaginatedResponse } from '@/types/api'

export const teamKeys = {
  all: ['teams'] as const,
  lists: () => [...teamKeys.all, 'list'] as const,
  list: (params: TeamListParams) => [...teamKeys.lists(), params] as const,
  details: () => [...teamKeys.all, 'detail'] as const,
  detail: (id: string) => [...teamKeys.details(), id] as const,
}

/** Raw user shape returned by GET /api/v1/users */
interface ApiUser {
  id: string
  email: string
  name: string
  role: string
  teamIds: string[]
  avatar: string
  onlineStatus: MemberStatus
}

interface ApiUserList {
  items: ApiUser[]
  total: number
}

function mapUserToMember(u: ApiUser): TeamMember {
  return {
    id: u.id,
    name: u.name,
    role: u.role,
    avatar: u.avatar,
    status: u.onlineStatus ?? 'offline',
    metrics: { assigned: 0, completed: 0, efficiency: 0 },
  }
}

export async function fetchTeamMembers(
  _params: TeamListParams = {},
): Promise<PaginatedResponse<TeamMember>> {
  const data = await api.get<ApiUserList>('/users')
  const members = data.items.map(mapUserToMember)
  return { items: members, total: data.total, page: 1, limit: members.length }
}

export function useTeamMembers(
  params: TeamListParams = {},
): UseQueryResult<PaginatedResponse<TeamMember>> {
  return useQuery({
    queryKey: teamKeys.list(params),
    queryFn: () => fetchTeamMembers(params),
  })
}
