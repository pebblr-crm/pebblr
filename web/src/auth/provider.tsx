import { useState, useEffect, useCallback, type ReactNode } from 'react'
import { setTokenProvider } from '@/api/client'
import { AuthContext } from './context'
import type { AuthenticatedUser, Role } from '@/types/user'

const VALID_ROLES: readonly string[] = ['admin', 'manager', 'rep'] as const

function isRole(value: string): value is Role {
  return VALID_ROLES.includes(value)
}

function parseRole(value: string): Role {
  if (isRole(value)) return value
  console.warn(`Unknown role "${value}", defaulting to "rep"`)
  return 'rep'
}

const DEMO_SESSION_KEY = 'pebblr_demo_user'

let _currentUser: AuthenticatedUser | null = null

function restoreDemoSession(): void {
  try {
    const stored = sessionStorage.getItem(DEMO_SESSION_KEY)
    if (stored) {
      _currentUser = JSON.parse(stored) as AuthenticatedUser
    }
  } catch {
    // Ignore parse errors
  }
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

function initStaticAuth(): void {
  const staticToken: string | undefined = import.meta.env.VITE_STATIC_TOKEN
  if (staticToken) {
    _currentUser = {
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
  setTokenProvider(() => _currentUser?.accessToken ?? null)
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true'
  const [user, setUser] = useState<AuthenticatedUser | null>(() => {
    if (!isDemoMode) {
      initStaticAuth()
    } else {
      restoreDemoSession()
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
      name: data.account.name,
      displayName: data.account.name,
      email: data.account.email,
      role: parseRole(data.account.role),
      oid: data.account.id,
      accessToken: data.token,
      expiresAt: Date.now() + 24 * 60 * 60 * 1000,
    }
    saveDemoSession(_currentUser)
    setUser(_currentUser)
  }, [])

  const demoLogout = useCallback(() => {
    _currentUser = null
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
