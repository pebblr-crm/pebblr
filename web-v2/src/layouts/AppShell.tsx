import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'

interface AppShellProps {
  children: ReactNode
  currentPath: string
}

export function AppShell({ children, currentPath }: AppShellProps) {
  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar currentPath={currentPath} />
      <main className="flex-1 overflow-auto">
        {children}
      </main>
    </div>
  )
}
