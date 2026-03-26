import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { User } from '@/types/user'

export const userKeys = {
  all: ['users'] as const,
  detail: (id: string) => [...userKeys.all, id] as const,
}

export function useUsers() {
  return useQuery({
    queryKey: userKeys.all,
    queryFn: () => api.get<{ items: User[]; total: number }>('/users'),
  })
}

export function useUser(id: string) {
  return useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => api.get<User>(`/users/${id}`),
    enabled: !!id,
  })
}
