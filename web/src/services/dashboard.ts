import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type { DashboardStats } from '@/types/dashboard'

export const dashboardKeys = {
  all: ['dashboard'] as const,
  stats: () => [...dashboardKeys.all, 'stats'] as const,
}

export function fetchDashboardStats(): Promise<DashboardStats> {
  return api.get<DashboardStats>('/dashboard/stats')
}

export function useDashboardStats(): UseQueryResult<DashboardStats> {
  return useQuery({
    queryKey: dashboardKeys.stats(),
    queryFn: fetchDashboardStats,
  })
}
