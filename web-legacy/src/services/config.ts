import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type { TenantConfig } from '@/types/config'

export const configKeys = {
  all: ['config'] as const,
}

function fetchConfig(): Promise<TenantConfig> {
  return api.get<TenantConfig>('/config')
}

export function useConfig(): UseQueryResult<TenantConfig> {
  return useQuery({
    queryKey: configKeys.all,
    queryFn: fetchConfig,
    staleTime: Infinity,
  })
}
