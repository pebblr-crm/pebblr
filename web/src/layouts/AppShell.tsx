import { useState, type ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { Menu } from 'lucide-react'

interface AppShellProps {
  children: ReactNode
  currentPath: string
}

export function AppShell({ children, currentPath }: AppShellProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="flex h-screen bg-slate-50">
      {/* Mobile header */}
      <div className="fixed inset-x-0 top-0 z-30 flex h-14 items-center gap-2 border-b border-slate-200 bg-white px-4 md:hidden">
        <button
          onClick={() => setSidebarOpen(true)}
          className="rounded-lg p-2 text-slate-600 hover:bg-slate-100"
          aria-label="Open menu"
        >
          <Menu size={20} />
        </button>
        <div className="h-7 w-7 rounded-lg bg-teal-600 text-white flex items-center justify-center text-xs font-bold">
          P
        </div>
        <span className="text-base font-semibold text-slate-900">Pebblr</span>
        <span className="rounded bg-teal-50 px-1.5 py-0.5 text-xs font-medium text-teal-700">v2</span>
      </div>

      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={`fixed inset-y-0 left-0 z-50 transform transition-transform duration-200 ease-in-out md:static md:translate-x-0 ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <Sidebar currentPath={currentPath} onNavigate={() => setSidebarOpen(false)} />
      </div>

      <main className="flex-1 overflow-auto pt-14 md:pt-0">
        {children}
      </main>
    </div>
  )
}
