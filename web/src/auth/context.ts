import { createContext, useContext } from 'react'
import type { AuthenticatedUser, Role } from '@/types/user'

export interface AuthContextValue {
  user: AuthenticatedUser | null
  role: Role | null
  isDemoMode: boolean
  demoLogin: (userId: string) => Promise<void>
  demoLogout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
