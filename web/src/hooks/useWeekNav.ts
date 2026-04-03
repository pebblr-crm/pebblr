import { useState, useMemo, useCallback } from 'react'
import { getMonday, addDays, formatDate } from '@/lib/dates'

/** Shared week-navigation state used by planner and rep drill-down pages. */
export function useWeekNav() {
  const [weekStart, setWeekStart] = useState(() => getMonday(new Date()))

  const weekEnd = useMemo(() => addDays(weekStart, 4), [weekStart])
  const dateFrom = formatDate(weekStart)
  const dateTo = formatDate(weekEnd)

  const prevWeek = useCallback(() => setWeekStart((w) => addDays(w, -7)), [])
  const nextWeek = useCallback(() => setWeekStart((w) => addDays(w, 7)), [])
  const goToday = useCallback(() => setWeekStart(getMonday(new Date())), [])

  return { weekStart, setWeekStart, weekEnd, dateFrom, dateTo, prevWeek, nextWeek, goToday }
}
