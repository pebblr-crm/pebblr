import type { User } from './user'

export interface Team {
  id: string
  name: string
  managerId: string
}

export interface TeamDetail {
  team: Team
  members: User[]
}
