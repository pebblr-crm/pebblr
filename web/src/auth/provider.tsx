import { useState, useEffect, useCallback, useRef, type ReactNode } from 'react'
import { setTokenProvider } from '@/api/client'
import { AuthContext } from './context'
import type { AuthenticatedUser, Role } from '@/types/user'

const DEMO_SESSION_KEY = 'pebblr_demo_user'

function restoreDemoSession(): AuthenticatedUser | null {
  try {
    const stored = sessionStorage.getItem(DEMO_SESSION_KEY)
    if (stored) {
      return JSON.parse(stored) as AuthenticatedUser
    }
  } catch {
    // Ignore parse errors
  }
  return null
}

function saveDemoSession(user: AuthenticatedUser): void {
  try {
    sessionStorage.setItem(DEMO_SESSION_KEY, JSON.stringify(user))
  } catch {
    // Ignore storage errors
  }
}

function clearDemoSession(): void {
  try {
    sessionStorage.removeItem(DEMO_SESSION_KEY)
  } catch {
    // Ignore storage errors
  }
}

function buildStaticUser(): AuthenticatedUser | null {
  const staticToken: string | undefined = import.meta.env.VITE_STATIC_TOKEN
  if (!staticToken) return null
  return {
    id: 'static-dev-user',
    name: 'Dev Admin',
    displayName: 'Dev Admin',
    email: 'admin@pebblr.dev',
    role: 'admin',
    oid: 'a0000000-0000-0000-0000-000000000001',
    accessToken: staticToken,
    expiresAt: Date.now() + 365 * 24 * 60 * 60 * 1000,
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true'

  // Ref keeps the token accessible to the synchronous token-provider callback
  // without introducing module-level mutable state.
  const tokenRef = useRef<string | null>(null)

  const [user, setUser] = useState<AuthenticatedUser | null>(() => {
    const initial = isDemoMode ? restoreDemoSession() : buildStaticUser()
    tokenRef.current = initial?.accessToken ?? null
    return initial
  })

  // Keep tokenRef in sync whenever user changes
  useEffect(() => {
    tokenRef.current = user?.accessToken ?? null
  }, [user])

  // Wire up the API client's token provider once on mount
  useEffect(() => {
    setTokenProvider(() => tokenRef.current)
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

    const newUser: AuthenticatedUser = {
      id: data.account.id,
      name: data.account.name,
      displayName: data.account.name,
      email: data.account.email,
      role: data.account.role as Role,
      oid: data.account.id,
      accessToken: data.token,
      expiresAt: Date.now() + 24 * 60 * 60 * 1000,
    }
    saveDemoSession(newUser)
    setUser(newUser)
  }, [])

  const demoLogout = useCallback(() => {
    clearDemoSession()
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
