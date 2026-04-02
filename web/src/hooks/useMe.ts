import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { CurrentUser } from '@/types/user'

export const meKeys = {
  all: ['me'] as const,
}

export function useCurrentUser() {
  return useQuery({
    queryKey: meKeys.all,
    queryFn: () => api.get<CurrentUser>('/me'),
    staleTime: 5 * 60 * 1000,
  })
}
