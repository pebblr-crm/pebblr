// Azure AD MSAL integration placeholder
// TODO: Integrate with @azure/msal-browser

export interface AuthAccount {
  oid: string
  email: string
  displayName: string
}

export async function getAccessToken(): Promise<string> {
  // TODO: Implement MSAL token acquisition
  throw new Error('Auth not yet implemented')
}

export async function getAccount(): Promise<AuthAccount | null> {
  // TODO: Implement MSAL account retrieval
  return null
}

export async function login(): Promise<void> {
  // TODO: Implement MSAL login redirect/popup
  throw new Error('Auth not yet implemented')
}

export async function logout(): Promise<void> {
  // TODO: Implement MSAL logout
  throw new Error('Auth not yet implemented')
}
