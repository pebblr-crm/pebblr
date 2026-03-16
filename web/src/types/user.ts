export type Role = 'rep' | 'manager' | 'admin'

export interface User {
  id: string
  email: string
  displayName: string
  role: Role
}
