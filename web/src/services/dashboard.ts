import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type {
  ActivityStatsResponse,
  CoverageResponse,
  FrequencyResponse,
  DashboardFilter,
} from '@/types/dashboard'

// ── Query keys ────────────────────────────────────────────────────────────────

export const dashboardKeys = {
  all: ['dashboard'] as const,
  activities: (filter: DashboardFilter) => [...dashboardKeys.all, 'activities', filter] as const,
  coverage: (filter: DashboardFilter) => [...dashboardKeys.all, 'coverage', filter] as const,
  frequency: (filter: DashboardFilter) => [...dashboardKeys.all, 'frequency', filter] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildDashboardPath(endpoint: string, filter: DashboardFilter): string {
  const qs = new URLSearchParams()
  if (filter.period) qs.set('period', filter.period)
  if (filter.dateFrom) qs.set('dateFrom', filter.dateFrom)
  if (filter.dateTo) qs.set('dateTo', filter.dateTo)
  if (filter.userId) qs.set('userId', filter.userId)
  if (filter.teamId) qs.set('teamId', filter.teamId)
  const query = qs.toString()
  return query ? `/dashboard/${endpoint}?${query}` : `/dashboard/${endpoint}`
}

export function fetchActivityStats(filter: DashboardFilter): Promise<ActivityStatsResponse> {
  return api.get<ActivityStatsResponse>(buildDashboardPath('activities', filter))
}

export function fetchCoverage(filter: DashboardFilter): Promise<CoverageResponse> {
  return api.get<CoverageResponse>(buildDashboardPath('coverage', filter))
}

export function fetchFrequency(filter: DashboardFilter): Promise<FrequencyResponse> {
  return api.get<FrequencyResponse>(buildDashboardPath('frequency', filter))
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useActivityStats(
  filter: DashboardFilter,
): UseQueryResult<ActivityStatsResponse> {
  return useQuery({
    queryKey: dashboardKeys.activities(filter),
    queryFn: () => fetchActivityStats(filter),
  })
}

export function useCoverage(
  filter: DashboardFilter,
): UseQueryResult<CoverageResponse> {
  return useQuery({
    queryKey: dashboardKeys.coverage(filter),
    queryFn: () => fetchCoverage(filter),
  })
}

export function useFrequency(
  filter: DashboardFilter,
): UseQueryResult<FrequencyResponse> {
  return useQuery({
    queryKey: dashboardKeys.frequency(filter),
    queryFn: () => fetchFrequency(filter),
  })
}
