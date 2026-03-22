import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { api } from './api'
import type { CalendarEvent, CalendarEventListParams } from '@/types/calendar'

interface CalendarEventListResponse {
  items: CalendarEvent[]
  total: number
  page: number
  limit: number
}

export const calendarKeys = {
  all: ['events'] as const,
  lists: () => [...calendarKeys.all, 'list'] as const,
  list: (params: CalendarEventListParams) => [...calendarKeys.lists(), params] as const,
}

function buildCalendarPath(params: CalendarEventListParams): string {
  const qs = new URLSearchParams()
  if (params.month !== undefined) qs.set('month', String(params.month))
  if (params.year !== undefined) qs.set('year', String(params.year))
  if (params.assigneeId) qs.set('assignee', params.assigneeId)
  const query = qs.toString()
  return query ? `/events?${query}` : '/events'
}

export async function fetchCalendarEvents(
  params: CalendarEventListParams = {},
): Promise<CalendarEvent[]> {
  const res = await api.get<CalendarEventListResponse>(buildCalendarPath(params))
  return res.items
}

export function useCalendarEvents(
  params: CalendarEventListParams = {},
): UseQueryResult<CalendarEvent[]> {
  return useQuery({
    queryKey: calendarKeys.list(params),
    queryFn: () => fetchCalendarEvents(params),
  })
}
