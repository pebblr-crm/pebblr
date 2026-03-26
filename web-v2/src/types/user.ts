export type Role = 'admin' | 'manager' | 'rep'

export interface User {
  id: string
  displayName: string
  email: string
  role: Role
  teamIds?: string[]
  avatar?: string
  onlineStatus?: 'online' | 'away' | 'offline'
}

export interface AuthenticatedUser extends User {
  oid: string
  accessToken: string
  expiresAt: number
}

export interface CurrentUser {
  id: string
  email: string
  name: string
  role: Role
  teamIds: string[]
}
