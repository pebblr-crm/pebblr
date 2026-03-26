import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { TenantConfig } from '@/types/config'

export const configKeys = {
  all: ['config'] as const,
}

export function useConfig() {
  return useQuery({
    queryKey: configKeys.all,
    queryFn: () => api.get<TenantConfig>('/config'),
    staleTime: Infinity,
  })
}
