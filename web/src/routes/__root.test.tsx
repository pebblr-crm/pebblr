import { render, screen } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'

// --- Mock dependencies ---
vi.mock('@/layouts/AppShell', () => ({
  AppShell: ({ children, currentPath }: { children: React.ReactNode; currentPath: string }) => (
    <div data-testid="app-shell" data-path={currentPath}>{children}</div>
  ),
}))

vi.mock('@/components/ui/ErrorBoundary', () => ({
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="error-boundary">{children}</div>
  ),
}))

let capturedComponent: React.ComponentType | null = null
let mockPathname = '/dashboard'

vi.mock('@tanstack/react-router', () => ({
  createRootRoute: (opts: { component?: React.ComponentType }) => {
    if (opts?.component) capturedComponent = opts.component
    return {}
  },
  Outlet: () => <div data-testid="outlet">outlet</div>,
  useRouterState: () => ({
    location: { pathname: mockPathname },
  }),
}))

// Force module evaluation to capture the component
await import('./__root')

describe('RootLayout', () => {
  it('renders AppShell with Outlet for non-sign-in paths', () => {
    mockPathname = '/dashboard'
    const Component = capturedComponent!
    render(<Component />)

    expect(screen.getByTestId('app-shell')).toBeInTheDocument()
    expect(screen.getByTestId('error-boundary')).toBeInTheDocument()
    expect(screen.getByTestId('outlet')).toBeInTheDocument()
  })

  it('renders only Outlet for /sign-in path', () => {
    mockPathname = '/sign-in'
    const Component = capturedComponent!
    render(<Component />)

    expect(screen.queryByTestId('app-shell')).not.toBeInTheDocument()
    expect(screen.getByTestId('outlet')).toBeInTheDocument()
  })
})
