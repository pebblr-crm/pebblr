/**
 * Dashboard types — mirror the Go backend dashboard response models.
 */

export interface StatusCount {
  status: string
  count: number
}

export interface TypeCount {
  activityType: string
  count: number
}

export interface CategoryCount {
  category: string
  count: number
}

export interface ActivityStats {
  total: number
  submittedCount: number
  byStatus: StatusCount[]
  byType: TypeCount[]
}

export interface DashboardStatsResponse {
  period: string
  dateFrom: string
  dateTo: string
  stats: ActivityStats
  byCategory: CategoryCount[]
}

export interface CoverageStats {
  totalTargets: number
  visitedTargets: number
  coveragePercent: number
}

export interface UserActivityStats {
  userId: string
  userName: string
  total: number
  byStatus: StatusCount[]
}

export interface UserStatsResponse {
  users: UserActivityStats[]
}

export interface DashboardParams {
  period: string
  teamId?: string
  creatorId?: string
}
