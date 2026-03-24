import { useState, useRef, useEffect } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { LayoutDashboard, Users, CalendarDays, Settings, HelpCircle, Target, ClipboardList, MapPin, Moon, Sun, X, ArrowLeftRight } from 'lucide-react'
import { useTheme } from '@/contexts/theme'
import { isDemoMode, demoLogout, getCurrentUser } from '@/services/auth'

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/targets', label: 'Targets', icon: Target },
  { to: '/activities/new', label: 'New Activity', icon: ClipboardList },
  { to: '/planner', label: 'Planner', icon: CalendarDays },
  { to: '/planner/map', label: 'Map Planner', icon: MapPin },
  { to: '/team', label: 'Team', icon: Users },
] as const

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const { location } = useRouterState()
  const { theme, toggle } = useTheme()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const settingsRef = useRef<HTMLDivElement>(null)

  // Close popover on outside click
  useEffect(() => {
    if (!settingsOpen) return
    function handleClick(e: MouseEvent) {
      if (settingsRef.current && !settingsRef.current.contains(e.target as Node)) {
        setSettingsOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [settingsOpen])

  // Close mobile sidebar on navigation
  const prevPathname = useRef(location.pathname)
  useEffect(() => {
    if (location.pathname !== prevPathname.current) {
      onClose()
      prevPathname.current = location.pathname
    }
  }, [location.pathname, onClose])

  return (
    <>
      {/* Backdrop — mobile only */}
      {open && (
        <div
          className="fixed inset-0 bg-black/30 z-40 lg:hidden"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      <aside
        className={`
          fixed inset-y-0 left-0 w-64 bg-white border-r border-slate-100 flex flex-col shrink-0 overflow-y-auto z-50
          transform transition-transform duration-200 ease-in-out
          lg:static lg:translate-x-0
          ${open ? 'translate-x-0' : '-translate-x-full'}
        `}
      >
        <div className="p-6">
          <div className="flex items-center justify-between mb-8">
            <div className="text-xl font-bold tracking-tight text-primary font-headline">Pebblr</div>
            <button
              onClick={onClose}
              className="p-1 text-slate-400 hover:text-on-surface rounded-lg lg:hidden"
              aria-label="Close menu"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <nav className="space-y-1">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.to || location.pathname.startsWith(item.to + '/')
              return (
                <Link
                  key={item.to}
                  to={item.to}
                  className={`w-full flex items-center px-4 py-3 transition-all rounded-xl group relative no-underline ${
                    isActive
                      ? 'bg-blue-50 text-primary font-bold'
                      : 'text-slate-500 hover:bg-slate-50'
                  }`}
                >
                  {isActive && (
                    <div className="absolute left-0 w-1 h-8 bg-primary rounded-r-full" />
                  )}
                  <Icon
                    className={`mr-3 w-5 h-5 ${
                      isActive ? 'text-primary' : 'text-slate-400 group-hover:text-primary'
                    }`}
                  />
                  <span className="font-headline text-sm font-medium">{item.label}</span>
                </Link>
              )
            })}
          </nav>
        </div>

        <div className="mt-auto p-6 border-t border-slate-50">
          <nav className="space-y-1">
            <div ref={settingsRef} className="relative">
              <button
                onClick={() => setSettingsOpen((o) => !o)}
                className="w-full flex items-center px-4 py-3 text-slate-500 hover:bg-slate-50 transition-colors rounded-xl group"
              >
                <Settings className="mr-3 w-5 h-5 text-slate-400 group-hover:text-primary" />
                <span className="font-headline text-sm font-medium">Settings</span>
              </button>
              {settingsOpen && (
                <div className="absolute bottom-full left-0 mb-2 w-56 bg-surface-container-lowest rounded-xl shadow-lg border border-slate-100 p-3 space-y-2 z-50">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">Theme</span>
                    <button
                      onClick={toggle}
                      className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium bg-surface-container-high text-on-surface hover:opacity-80 transition-opacity"
                    >
                      {theme === 'dark' ? (
                        <><Sun className="w-3.5 h-3.5" /> Light</>
                      ) : (
                        <><Moon className="w-3.5 h-3.5" /> Dark</>
                      )}
                    </button>
                  </div>
                </div>
              )}
            </div>
            {isDemoMode() && (
              <button
                onClick={demoLogout}
                className="w-full flex items-center px-4 py-3 text-slate-500 hover:bg-slate-50 transition-colors rounded-xl group"
              >
                <ArrowLeftRight className="mr-3 w-5 h-5 text-slate-400 group-hover:text-primary" />
                <span className="font-headline text-sm font-medium flex-1 text-left">Switch Account</span>
                {getCurrentUser() && (
                  <span className="text-xs text-on-surface-variant truncate max-w-24">
                    {getCurrentUser()?.displayName}
                  </span>
                )}
              </button>
            )}
            <button className="w-full flex items-center px-4 py-3 text-slate-500 hover:bg-slate-50 transition-colors rounded-xl group">
              <HelpCircle className="mr-3 w-5 h-5 text-slate-400 group-hover:text-primary" />
              <span className="font-headline text-sm font-medium">Help</span>
            </button>
          </nav>
        </div>
      </aside>
    </>
  )
}
