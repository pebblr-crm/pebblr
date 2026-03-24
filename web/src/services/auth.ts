

/**
 * Authentication service.
 *
 * Supports three modes:
 * - Static token (VITE_STATIC_TOKEN) for local dev / E2E
 * - Demo mode (VITE_DEMO_MODE) for self-service demo environments
 * - Azure AD MSAL (future) for production
 */

import { setTokenProvider } from './api'
import type { AuthenticatedUser } from '@/types/user'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface AuthConfig {
  /** Azure AD tenant ID */
  tenantId: string
  /** Application (client) ID registered in Entra ID */
  clientId: string
  /** OAuth2 redirect URI */
  redirectUri: string
  /** API scope to request (e.g. "api://<client-id>/Leads.Read") */
  apiScope: string
}

export interface AuthAccount {
  oid: string
  email: string
  displayName: string
}

// ---------------------------------------------------------------------------
// Module state (no global singletons — configure once at app startup)
// ---------------------------------------------------------------------------

let _currentUser: AuthenticatedUser | null = null
let _isDemoMode = false
let _onAuthChange: (() => void) | null = null

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Returns true if the app is running in demo mode.
 */
export function isDemoMode(): boolean {
  return _isDemoMode
}

/**
 * Register a callback invoked when the auth state changes (login/logout).
 */
export function onAuthChange(cb: () => void): void {
  _onAuthChange = cb
}

/**
 * Initialise the auth module.  Call once in main.tsx before rendering.
 *
 * In dev/E2E builds, VITE_STATIC_TOKEN is set at build time and used as
 * the Bearer token for all API requests. In demo mode, auth is deferred
 * until the user picks an account. In production, this will be replaced
 * by real MSAL/OIDC initialisation.
 */
export async function initAuth(_config: AuthConfig): Promise<void> {
  _isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true'

  if (_isDemoMode) {
    // In demo mode, no token until the user picks an account.
    setTokenProvider(() => _currentUser?.accessToken ?? null)
    return
  }

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

/**
 * Sign in as a demo user by their user ID.
 * Calls the demo token endpoint and stores the resulting token.
 */
export async function demoLogin(userId: string): Promise<void> {
  const response = await fetch('/demo/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id: userId }),
  })

  if (!response.ok) {
    throw new Error(`Demo login failed: ${response.status}`)
  }

  const data = (await response.json()) as {
    token: string
    account: { id: string; name: string; email: string; role: string }
  }

  _currentUser = {
    id: data.account.id,
    displayName: data.account.name,
    email: data.account.email,
    role: data.account.role as 'admin' | 'manager' | 'rep',
    oid: data.account.id,
    accessToken: data.token,
    expiresAt: Date.now() + 24 * 60 * 60 * 1000,
  }

  _onAuthChange?.()
}

/**
 * Sign out the current demo user, returning to the account picker.
 */
export function demoLogout(): void {
  _currentUser = null
  _onAuthChange?.()
}

/**
 * Returns the currently authenticated user, or null if unauthenticated.
 */
export function getCurrentUser(): AuthenticatedUser | null {
  return _currentUser
}

/**
 * Retrieve the current MSAL account info.
 *
 * TODO: Implement with msalInstance.getActiveAccount().
 */
export async function getAccount(): Promise<AuthAccount | null> {
  return null
}

/**
 * Silently acquire a fresh access token.
 * Returns null if the user is not authenticated.
 *
 * TODO: implement with msalInstance.acquireTokenSilent().
 */
export async function getAccessToken(): Promise<string> {
  if (_currentUser?.accessToken) {
    return _currentUser.accessToken
  }
  throw new Error('Auth not yet implemented')
}

/**
 * Trigger interactive login.
 *
 * TODO: implement with msalInstance.loginRedirect() or loginPopup().
 */
export async function login(): Promise<void> {
  throw new Error('login() not yet implemented — wire up MSAL')
}

/**
 * Sign out the current user.
 *
 * TODO: implement with msalInstance.logoutRedirect().
 */
export async function logout(): Promise<void> {
  _currentUser = null
  throw new Error('logout() not yet implemented — wire up MSAL')
}

/**
 * Silently acquire a fresh access token (returns null if unauthenticated).
 *
 * TODO: implement with msalInstance.acquireTokenSilent().
 */
export async function acquireToken(): Promise<string | null> {
  if (!_currentUser) return null
  // TODO: check token expiry, refresh if needed
  return _currentUser.accessToken
}
