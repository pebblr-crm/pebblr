import { Link, useRouterState } from '@tanstack/react-router'
import type { ReactNode } from 'react'

interface LayoutProps {
  children: ReactNode
}

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/leads', label: 'Leads' },
] as const

export function Layout({ children }: LayoutProps) {
  const { location } = useRouterState()

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <span className="sidebar-logo">Pebblr</span>
        </div>
        <nav className="sidebar-nav">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`sidebar-nav-item${location.pathname === item.to ? ' active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="main-content">{children}</main>
    </div>
  )
}
