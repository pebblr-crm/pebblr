import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type {
  DashboardStatsResponse,
  CoverageStats,
  UserStatsResponse,
  DashboardParams,
} from '@/types/dashboard'

// ── Query keys ────────────────────────────────────────────────────────────────

export const dashboardKeys = {
  all: ['dashboard'] as const,
  stats: (params: DashboardParams) => [...dashboardKeys.all, 'stats', params] as const,
  coverage: (params: DashboardParams) => [...dashboardKeys.all, 'coverage', params] as const,
  userStats: (params: DashboardParams) => [...dashboardKeys.all, 'user-stats', params] as const,
}

// ── API functions ─────────────────────────────────────────────────────────────

function buildDashboardPath(endpoint: string, params: DashboardParams): string {
  const qs = new URLSearchParams()
  qs.set('period', params.period)
  if (params.teamId) qs.set('teamId', params.teamId)
  if (params.creatorId) qs.set('creatorId', params.creatorId)
  return `/dashboard/${endpoint}?${qs.toString()}`
}

export function fetchDashboardStats(params: DashboardParams): Promise<DashboardStatsResponse> {
  return api.get<DashboardStatsResponse>(buildDashboardPath('stats', params))
}

export function fetchCoverageStats(params: DashboardParams): Promise<CoverageStats> {
  return api.get<CoverageStats>(buildDashboardPath('coverage', params))
}

export function fetchUserStats(params: DashboardParams): Promise<UserStatsResponse> {
  return api.get<UserStatsResponse>(buildDashboardPath('user-stats', params))
}

// ── TanStack Query hooks ──────────────────────────────────────────────────────

export function useDashboardStats(
  params: DashboardParams,
): UseQueryResult<DashboardStatsResponse> {
  return useQuery({
    queryKey: dashboardKeys.stats(params),
    queryFn: () => fetchDashboardStats(params),
    enabled: Boolean(params.period),
  })
}

export function useCoverageStats(
  params: DashboardParams,
): UseQueryResult<CoverageStats> {
  return useQuery({
    queryKey: dashboardKeys.coverage(params),
    queryFn: () => fetchCoverageStats(params),
    enabled: Boolean(params.period),
  })
}

export function useUserStats(
  params: DashboardParams,
): UseQueryResult<UserStatsResponse> {
  return useQuery({
    queryKey: dashboardKeys.userStats(params),
    queryFn: () => fetchUserStats(params),
    enabled: Boolean(params.period),
  })
}
