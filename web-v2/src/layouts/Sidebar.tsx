import { useTranslation } from 'react-i18next'
import { useAuth } from '@/auth/context'
import {
  Map,
  Target,
  CalendarCheck,
  LayoutDashboard,
  Globe,
  Settings,
  FileText,
  LogOut,
} from 'lucide-react'

interface NavItem {
  label: string
  href: string
  icon: React.ReactNode
  roles: readonly string[]
}

export function Sidebar({ currentPath }: { currentPath: string }) {
  const { t } = useTranslation()
  const { role, isDemoMode, demoLogout } = useAuth()

  const navItems: NavItem[] = [
    { label: t('nav.planner'), href: '/planner', icon: <Map size={20} />, roles: ['rep'] },
    { label: t('nav.targets'), href: '/targets', icon: <Target size={20} />, roles: ['rep'] },
    { label: t('nav.activities'), href: '/activities', icon: <CalendarCheck size={20} />, roles: ['rep'] },
    { label: t('nav.dashboard'), href: '/dashboard', icon: <LayoutDashboard size={20} />, roles: ['manager'] },
    { label: t('nav.coverage'), href: '/coverage', icon: <Globe size={20} />, roles: ['manager'] },
    { label: t('nav.console'), href: '/console', icon: <Settings size={20} />, roles: ['admin'] },
    { label: t('nav.audit'), href: '/audit', icon: <FileText size={20} />, roles: ['admin'] },
  ]

  const visibleItems = navItems.filter(
    (item) => role === 'admin' || item.roles.includes(role ?? ''),
  )

  return (
    <aside className="flex h-full w-60 flex-col border-r border-slate-200 bg-white">
      <div className="flex h-14 items-center gap-2 border-b border-slate-200 px-4">
        <div className="h-8 w-8 rounded-lg bg-teal-600 text-white flex items-center justify-center text-sm font-bold">
          P
        </div>
        <span className="text-lg font-semibold text-slate-900">Pebblr</span>
        <span className="ml-1 rounded bg-teal-50 px-1.5 py-0.5 text-xs font-medium text-teal-700">v2</span>
      </div>

      <nav className="flex-1 space-y-1 p-3">
        {visibleItems.map((item) => {
          const active = currentPath.startsWith(item.href)
          return (
            <a
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                active
                  ? 'bg-teal-50 text-teal-700'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              }`}
            >
              {item.icon}
              {item.label}
            </a>
          )
        })}
      </nav>

      <div className="border-t border-slate-200 p-3 space-y-1">
        <a
          href="/?ui=v1"
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-500 hover:bg-slate-50"
        >
          {t('common.switchToV1')}
        </a>
        {isDemoMode && (
          <button
            onClick={demoLogout}
            className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-500 hover:bg-slate-50"
          >
            <LogOut size={20} />
            {t('nav.signOut')}
          </button>
        )}
      </div>
    </aside>
  )
}
