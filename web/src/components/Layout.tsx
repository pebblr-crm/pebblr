import { useState, useCallback, type ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  const [menuOpen, setMenuOpen] = useState(false)
  const toggleMenu = useCallback(() => setMenuOpen((o) => !o), [])
  const closeMenu = useCallback(() => setMenuOpen(false), [])

  return (
    <div className="flex min-h-screen overflow-hidden bg-surface">
      <Sidebar open={menuOpen} onClose={closeMenu} />
      <div className="flex-1 flex flex-col h-screen overflow-hidden min-w-0">
        <TopBar onMenuToggle={toggleMenu} />
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  )
}
