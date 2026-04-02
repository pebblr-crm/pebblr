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
  activities: (f: DashboardFilter) => ['dashboard', 'activities', f] as const,
  coverage: (f: DashboardFilter) => ['dashboard', 'coverage', f] as const,
  frequency: (f: DashboardFilter) => ['dashboard', 'frequency', f] as const,
  recovery: (f: DashboardFilter) => ['dashboard', 'recovery', f] as const,
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
