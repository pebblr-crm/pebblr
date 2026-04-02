import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { buildQuery } from '@/lib/query'
import type {
  ActivityStatsResponse,
  CoverageResponse,
  FrequencyResponse,
  RecoveryBalanceResponse,
  DashboardFilter,
} from '@/types/dashboard'

export const dashboardKeys = {
  all: ['dashboard'] as const,
  activities: (f: DashboardFilter) => [...dashboardKeys.all, 'activities', f] as const,
  coverage: (f: DashboardFilter) => [...dashboardKeys.all, 'coverage', f] as const,
  frequency: (f: DashboardFilter) => [...dashboardKeys.all, 'frequency', f] as const,
  recovery: (f: DashboardFilter) => [...dashboardKeys.all, 'recovery', f] as const,
}

export function useActivityStats(filter: DashboardFilter = {}) {
  return useQuery({
    queryKey: dashboardKeys.activities(filter),
    queryFn: () => api.get<ActivityStatsResponse>(buildQuery('/dashboard/activities', { ...filter })),
  })
}

export function useCoverage(filter: DashboardFilter = {}) {
  return useQuery({
    queryKey: dashboardKeys.coverage(filter),
    queryFn: () => api.get<CoverageResponse>(buildQuery('/dashboard/coverage', { ...filter })),
  })
}

export function useFrequency(filter: DashboardFilter = {}) {
  return useQuery({
    queryKey: dashboardKeys.frequency(filter),
    queryFn: () => api.get<FrequencyResponse>(buildQuery('/dashboard/frequency', { ...filter })),
  })
}

export function useRecoveryBalance(filter: DashboardFilter = {}) {
  return useQuery({
    queryKey: dashboardKeys.recovery(filter),
    queryFn: () => api.get<RecoveryBalanceResponse>(buildQuery('/dashboard/recovery', { ...filter })),
  })
}
