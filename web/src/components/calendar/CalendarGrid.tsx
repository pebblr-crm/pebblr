import type { CalendarEvent } from '@/types/calendar'
import { EventCard } from './EventCard'

interface CalendarGridProps {
  events: CalendarEvent[]
  year: number
  month: number
}

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

export function CalendarGrid({ events, year, month }: CalendarGridProps) {
  // month is 1-based
  const daysInMonth = new Date(year, month, 0).getDate()

  return (
    <div className="bg-surface-container-lowest rounded-xl shadow-sm border border-slate-100 overflow-hidden">
      <div className="grid grid-cols-7 bg-surface-container-low border-b border-slate-100">
        {DAYS.map((day) => (
          <div key={day} className="py-4 text-center text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">
            {day}
          </div>
        ))}
      </div>
      <div className="grid grid-cols-7">
        {Array.from({ length: daysInMonth }).map((_, i) => {
          const day = i + 1
          const dayEvents = events.filter((e) => {
            const d = new Date(e.date)
            return d.getDate() === day && d.getMonth() + 1 === month
          })
          const isToday =
            new Date().getDate() === day &&
            new Date().getMonth() + 1 === month &&
            new Date().getFullYear() === year

          return (
            <div
              key={i}
              className={`min-h-[140px] p-2 border-r border-b border-slate-50 relative group hover:bg-surface-container-low/30 transition-colors ${isToday ? 'bg-primary/5' : ''}`}
            >
              <span
                className={`text-sm font-medium ${
                  isToday
                    ? 'flex items-center justify-center w-7 h-7 bg-primary text-white rounded-full'
                    : 'text-on-surface-variant'
                }`}
              >
                {day}
              </span>
              <div className="mt-2 space-y-1">
                {dayEvents.map((event) => (
                  <EventCard key={event.id} event={event} />
                ))}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
