import type { CalendarEvent } from '@/types/calendar'

interface EventCardProps {
  event: CalendarEvent
}

const eventStyles: Record<CalendarEvent['type'], string> = {
  visit: 'bg-amber-100 border-amber-500',
  sync: 'bg-tertiary-container/10 border-tertiary-container',
  review: 'bg-primary-container/10 border-primary-container',
  callback: 'bg-secondary-container/20 border-secondary',
  lunch: 'bg-orange-100 border-orange-400',
  demo: 'bg-primary-fixed/40 border-primary',
}

export function EventCard({ event }: EventCardProps) {
  const style = eventStyles[event.type] ?? 'bg-slate-100 border-slate-400'
  return (
    <div className={`p-1.5 rounded-lg border-l-4 ${style}`}>
      <p className="text-[8px] font-bold uppercase tracking-tighter opacity-70">{event.time}</p>
      <p className="text-[10px] font-semibold truncate">{event.title}</p>
    </div>
  )
}
