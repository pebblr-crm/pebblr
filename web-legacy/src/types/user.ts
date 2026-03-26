/**
 * User and RBAC types.
 */

export type Role = 'admin' | 'manager' | 'rep'

export interface User {
  id: string
  /** Display name from Azure AD */
  displayName: string
  /** Email / UPN from Azure AD */
  email: string
  role: Role
}

/**
 * The currently authenticated user context, populated from the OIDC token.
 */
export interface AuthenticatedUser extends User {
  /** Raw Azure AD object ID */
  oid: string
  /** Access token (bearer) for API calls */
  accessToken: string
  /** Token expiry (unix ms) */
  expiresAt: number
}
