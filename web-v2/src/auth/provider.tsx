import { useState, useEffect, useCallback, type ReactNode } from 'react'
import { setTokenProvider } from '@/api/client'
import { AuthContext } from './context'
import type { AuthenticatedUser, Role } from '@/types/user'

let _currentUser: AuthenticatedUser | null = null

function initStaticAuth(): void {
  const staticToken: string | undefined = import.meta.env.VITE_STATIC_TOKEN
  if (staticToken) {
    _currentUser = {
      id: 'static-dev-user',
      displayName: 'Dev Admin',
      email: 'admin@pebblr.dev',
      role: 'admin',
      oid: 'a0000000-0000-0000-0000-000000000001',
      accessToken: staticToken,
      expiresAt: Date.now() + 365 * 24 * 60 * 60 * 1000,
    }
  }
  setTokenProvider(() => _currentUser?.accessToken ?? null)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true'
  const [user, setUser] = useState<AuthenticatedUser | null>(() => {
    if (!isDemoMode) {
      initStaticAuth()
    } else {
      setTokenProvider(() => _currentUser?.accessToken ?? null)
    }
    return _currentUser
  })

  useEffect(() => {
    setTokenProvider(() => _currentUser?.accessToken ?? null)
  }, [])

  const demoLogin = useCallback(async (userId: string) => {
    const response = await fetch('/demo/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: userId }),
    })
    if (!response.ok) throw new Error(`Demo login failed: ${response.status}`)

    const data = (await response.json()) as {
      token: string
      account: { id: string; name: string; email: string; role: string }
    }

    _currentUser = {
      id: data.account.id,
      displayName: data.account.name,
      email: data.account.email,
      role: data.account.role as Role,
      oid: data.account.id,
      accessToken: data.token,
      expiresAt: Date.now() + 24 * 60 * 60 * 1000,
    }
    setUser(_currentUser)
  }, [])

  const demoLogout = useCallback(() => {
    _currentUser = null
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider
      value={{
        user,
        role: user?.role ?? null,
        isDemoMode,
        demoLogin,
        demoLogout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
