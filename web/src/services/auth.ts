

/**
 * Azure AD (Entra ID) MSAL authentication placeholder.
 *
 * In production this will use @azure/msal-browser to acquire tokens via
 * OIDC/OAuth2. The token is then passed to setTokenProvider in api.ts so
 * every API request includes a valid Bearer token.
 *
 * Local dev uses Kind + federated credentials (no static secrets).
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

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Initialise the auth module.  Call once in main.tsx before rendering.
 *
 * In dev/E2E builds, VITE_STATIC_TOKEN is set at build time and used as
 * the Bearer token for all API requests. In production, this will be
 * replaced by real MSAL/OIDC initialisation.
 */
export async function initAuth(_config: AuthConfig): Promise<void> {
  const staticToken = import.meta.env.VITE_STATIC_TOKEN as string | undefined
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
