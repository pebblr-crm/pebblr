import { render, screen, fireEvent } from '@testing-library/react'
import { vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Mock dependencies
const mockDemoLogout = vi.fn()
const mockClear = vi.fn()

vi.mock('@/auth/context', () => ({
  useAuth: () => ({
    role: 'admin',
    isDemoMode: true,
    demoLogout: mockDemoLogout,
  }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const map: Record<string, string> = {
        'nav.planner': 'Planner',
        'nav.targets': 'Targets',
        'nav.activities': 'Activities',
        'nav.dashboard': 'Dashboard',
        'nav.coverage': 'Coverage',
        'nav.console': 'Console',
        'nav.audit': 'Audit',
        'nav.signOut': 'Sign Out',
      }
      return map[key] ?? key
    },
  }),
}))

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, onClick, ...props }: Record<string, unknown>) => (
    <a href={to as string} onClick={onClick as () => void} {...props}>{children as React.ReactNode}</a>
  ),
}))

import { Sidebar } from './Sidebar'

function renderSidebar(props: { currentPath: string; onNavigate?: () => void }) {
  const qc = new QueryClient()
  // Spy on qc.clear
  qc.clear = mockClear
  return render(
    <QueryClientProvider client={qc}>
      <Sidebar {...props} />
    </QueryClientProvider>,
  )
}

describe('Sidebar', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders the Pebblr brand', () => {
    renderSidebar({ currentPath: '/dashboard' })
    expect(screen.getByText('Pebblr')).toBeInTheDocument()
  })

  it('renders visible nav items for admin role', () => {
    renderSidebar({ currentPath: '/dashboard' })
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Coverage')).toBeInTheDocument()
    expect(screen.getByText('Console')).toBeInTheDocument()
    expect(screen.getByText('Audit')).toBeInTheDocument()
  })

  it('shows Sign Out button in demo mode', () => {
    renderSidebar({ currentPath: '/dashboard' })
    expect(screen.getByText('Sign Out')).toBeInTheDocument()
  })

  it('calls demoLogout and clears query cache on logout', () => {
    renderSidebar({ currentPath: '/dashboard' })
    fireEvent.click(screen.getByText('Sign Out'))
    expect(mockDemoLogout).toHaveBeenCalledOnce()
    expect(mockClear).toHaveBeenCalledOnce()
  })

  it('calls onNavigate when close button is clicked', () => {
    const onNavigate = vi.fn()
    renderSidebar({ currentPath: '/dashboard', onNavigate })
    fireEvent.click(screen.getByLabelText('Close menu'))
    expect(onNavigate).toHaveBeenCalledOnce()
  })
})
