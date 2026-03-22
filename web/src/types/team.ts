export type MemberStatus = 'online' | 'away' | 'offline'

export interface TeamMemberMetrics {
  assigned: number
  completed: number
  efficiency: number
}

export interface TeamMember {
  id: string
  name: string
  role: string
  avatar: string
  status: MemberStatus
  metrics: TeamMemberMetrics
}

export interface Team {
  id: string
  name: string
  members: TeamMember[]
}

export interface TeamListParams {
  page?: number
  limit?: number
}
