import { Search, Bell, History } from 'lucide-react'

export function TopBar() {
  return (
    <header className="h-16 w-full bg-white/80 backdrop-blur-md sticky top-0 z-40 flex justify-between items-center px-8 border-b border-slate-100">
      <div className="flex items-center flex-1">
        <div className="relative w-96">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 w-4 h-4" />
          <input
            type="text"
            placeholder="Search leads or team members..."
            className="w-full pl-10 pr-4 py-2 bg-surface-container-highest border-none rounded-xl focus:ring-2 focus:ring-primary/20 focus:bg-white transition-all text-sm outline-none"
          />
        </div>
      </div>

      <div className="flex items-center space-x-4">
        <button className="p-2 text-slate-500 hover:bg-slate-50 rounded-lg relative">
          <Bell className="w-5 h-5" />
          <span className="absolute top-2 right-2 w-2 h-2 bg-error rounded-full border-2 border-white" />
        </button>
        <button className="p-2 text-slate-500 hover:bg-slate-50 rounded-lg">
          <History className="w-5 h-5" />
        </button>
        <div className="h-8 w-[1px] bg-slate-100 mx-2" />
        <div className="w-8 h-8 rounded-full bg-primary-fixed flex items-center justify-center text-primary font-bold text-xs">
          U
        </div>
      </div>
    </header>
  )
}
