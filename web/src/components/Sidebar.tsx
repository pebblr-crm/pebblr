import { Link, useRouterState } from '@tanstack/react-router'
import { LayoutDashboard, Users, CalendarDays, Settings, HelpCircle, Target, ClipboardList, MapPin } from 'lucide-react'

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/targets', label: 'Targets', icon: Target },
  { to: '/activities/new', label: 'New Activity', icon: ClipboardList },
  { to: '/planner', label: 'Planner', icon: CalendarDays },
  { to: '/planner/map', label: 'Map Planner', icon: MapPin },
  { to: '/team', label: 'Team', icon: Users },
] as const

export function Sidebar() {
  const { location } = useRouterState()

  return (
    <aside className="w-64 h-screen border-r border-slate-100 bg-white flex flex-col shrink-0 overflow-y-auto z-50">
      <div className="p-6">
        <div className="text-xl font-bold tracking-tight text-primary mb-8 font-headline">Pebblr</div>

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
          <button className="w-full flex items-center px-4 py-3 text-slate-500 hover:bg-slate-50 transition-colors rounded-xl group">
            <Settings className="mr-3 w-5 h-5 text-slate-400 group-hover:text-primary" />
            <span className="font-headline text-sm font-medium">Settings</span>
          </button>
          <button className="w-full flex items-center px-4 py-3 text-slate-500 hover:bg-slate-50 transition-colors rounded-xl group">
            <HelpCircle className="mr-3 w-5 h-5 text-slate-400 group-hover:text-primary" />
            <span className="font-headline text-sm font-medium">Help</span>
          </button>
        </nav>
      </div>
    </aside>
  )
}
