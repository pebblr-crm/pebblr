import { render, screen, act } from '@testing-library/react'
import { vi } from 'vitest'
import { AuthProvider } from './provider'
import { useAuth } from './context'

// Mock the setTokenProvider from api/client
const mockSetTokenProvider = vi.fn()
vi.mock('@/api/client', () => ({
  setTokenProvider: (...args: unknown[]) => mockSetTokenProvider(...args),
}))

// Helper component to read auth context
function AuthConsumer() {
  const { user, role, isDemoMode, demoLogout } = useAuth()
  return (
    <div>
      <span data-testid="user">{user?.name ?? 'none'}</span>
      <span data-testid="role">{role ?? 'none'}</span>
      <span data-testid="demo">{String(isDemoMode)}</span>
      <button onClick={demoLogout}>Logout</button>
    </div>
  )
}

function setEnv(token: string, demo: string) {
  vi.stubEnv('VITE_STATIC_TOKEN', token)
  vi.stubEnv('VITE_DEMO_MODE', demo)
}

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('provides null user when no static token and not demo mode', () => {
    setEnv('', '')

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('user')).toHaveTextContent('none')
    expect(screen.getByTestId('role')).toHaveTextContent('none')
  })

  it('builds static user when VITE_STATIC_TOKEN is set', () => {
    setEnv('my-static-token', '')

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('user')).toHaveTextContent('Dev Admin')
    expect(screen.getByTestId('role')).toHaveTextContent('admin')
  })

  it('restores demo session from sessionStorage', () => {
    setEnv('', 'true')

    const demoUser = {
      id: 'demo-1',
      name: 'Demo Rep',
      displayName: 'Demo Rep',
      email: 'demo@test.com',
      role: 'rep',
      oid: 'demo-1',
      accessToken: 'demo-token',
      expiresAt: Date.now() + 60000,
    }
    sessionStorage.setItem('pebblr_demo_user', JSON.stringify(demoUser))

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('user')).toHaveTextContent('Demo Rep')
    expect(screen.getByTestId('role')).toHaveTextContent('rep')
    expect(screen.getByTestId('demo')).toHaveTextContent('true')
  })

  it('sets token provider on mount', () => {
    setEnv('', '')

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(mockSetTokenProvider).toHaveBeenCalledWith(expect.any(Function))
  })

  it('clears demo session on logout', () => {
    setEnv('', 'true')

    const demoUser = {
      id: 'demo-1',
      name: 'Demo Rep',
      displayName: 'Demo Rep',
      email: 'demo@test.com',
      role: 'rep',
      oid: 'demo-1',
      accessToken: 'demo-token',
      expiresAt: Date.now() + 60000,
    }
    sessionStorage.setItem('pebblr_demo_user', JSON.stringify(demoUser))

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('user')).toHaveTextContent('Demo Rep')

    act(() => {
      screen.getByText('Logout').click()
    })

    expect(screen.getByTestId('user')).toHaveTextContent('none')
    expect(sessionStorage.getItem('pebblr_demo_user')).toBeNull()
  })

  it('handles corrupted session storage gracefully', () => {
    setEnv('', 'true')

    sessionStorage.setItem('pebblr_demo_user', 'NOT VALID JSON')

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('user')).toHaveTextContent('none')
  })
})
