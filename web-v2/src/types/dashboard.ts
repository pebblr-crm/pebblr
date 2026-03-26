export interface ActivityStatsResponse {
  byStatus: Record<string, number>
  byCategory: Record<string, number>
  total: number
}

export interface CoverageResponse {
  totalTargets: number
  visitedTargets: number
  percentage: number
}

export interface FrequencyItem {
  classification: string
  targetCount: number
  totalVisits: number
  required: number
  compliance: number
}

export interface FrequencyResponse {
  items: FrequencyItem[]
}

export interface RecoveryClaimInterval {
  weekendDate: string
  claimFrom: string
  claimBy: string
  claimed: boolean
}

export interface RecoveryBalanceResponse {
  earned: number
  taken: number
  balance: number
  intervals: RecoveryClaimInterval[]
}

export interface DashboardFilter {
  period?: string
  dateFrom?: string
  dateTo?: string
  userId?: string
  teamId?: string
}
