import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'

export interface CurrentUser {
  id: string
  email: string
  name: string
  role: string
  teamIds: string[]
}

export function useCurrentUser(): UseQueryResult<CurrentUser> {
  return useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<CurrentUser>('/me'),
    staleTime: 5 * 60 * 1000, // user identity rarely changes
  })
}
