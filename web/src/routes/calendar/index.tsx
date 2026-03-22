import { useState } from 'react'
import { createRoute } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ChevronLeft, ChevronRight, PlusCircle, CalendarDays, TrendingUp } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { CalendarGrid } from '../../components/calendar/CalendarGrid'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useCalendarEvents } from '../../services/calendar'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/calendar',
  component: CalendarPage,
})

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

export function CalendarPage() {
  const now = new Date()
  const [year, setYear] = useState(now.getFullYear())
  const [month, setMonth] = useState(now.getMonth() + 1)

  const { data: events = [], isLoading } = useCalendarEvents({ year, month })

  function prevMonth() {
    if (month === 1) { setMonth(12); setYear((y) => y - 1) }
    else setMonth((m) => m - 1)
  }

  function nextMonth() {
    if (month === 12) { setMonth(1); setYear((y) => y + 1) }
    else setMonth((m) => m + 1)
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className="p-8 space-y-6"
    >
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-4xl font-headline font-extrabold text-primary tracking-tight">
            Sales Calendar
          </h1>
          <p className="text-on-surface-variant mt-1">Manage team visits and client appointments.</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 bg-surface-container-lowest px-4 py-2 rounded-xl shadow-sm">
            <button className="text-on-surface-variant hover:text-primary" onClick={prevMonth}>
              <ChevronLeft className="w-5 h-5" />
            </button>
            <span className="font-headline font-bold text-primary px-2">
              {MONTH_NAMES[month - 1]} {year}
            </span>
            <button className="text-on-surface-variant hover:text-primary" onClick={nextMonth}>
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>
          <button className="flex items-center gap-2 bg-gradient-to-br from-primary to-primary-container text-white py-2.5 px-6 rounded-xl font-bold text-sm shadow-md">
            <PlusCircle className="w-4 h-4" />
            Quick Add
          </button>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-6">
        <div className="col-span-12 lg:col-span-3 space-y-6">
          <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
            <h3 className="font-headline font-bold text-primary mb-4">Status Legend</h3>
            <div className="space-y-3">
              {[
                { label: 'Completed Visit', color: 'bg-tertiary-container' },
                { label: 'Scheduled', color: 'bg-amber-500' },
                { label: 'In Progress', color: 'bg-primary-container' },
                { label: 'Demo', color: 'bg-primary-fixed' },
              ].map((item) => (
                <div key={item.label} className="flex items-center gap-2">
                  <span className={`w-3 h-3 rounded-full ${item.color}`} />
                  <span className="text-[10px] font-medium text-on-surface-variant">{item.label}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-100">
            <h4 className="font-headline font-bold text-primary mb-4 text-sm">Daily Pulse</h4>
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <CalendarDays className="w-4 h-4 text-primary mt-0.5" />
                <div>
                  <p className="text-xs font-bold text-on-surface">
                    {events.filter((e) => {
                      const d = new Date(e.startTime)
                      return d.getDate() === now.getDate() && d.getMonth() === now.getMonth()
                    }).length} Events Today
                  </p>
                  <p className="text-[10px] text-on-surface-variant">
                    {MONTH_NAMES[month - 1]} {year}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <TrendingUp className="w-4 h-4 text-tertiary-fixed-dim mt-0.5" />
                <div>
                  <p className="text-xs font-bold text-on-surface">{events.length} Events This Month</p>
                  <p className="text-[10px] text-on-surface-variant">Across all team members</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="col-span-12 lg:col-span-9">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" label="Loading calendar..." />
            </div>
          ) : (
            <CalendarGrid events={events} year={year} month={month} />
          )}
        </div>
      </div>
    </motion.div>
  )
}
