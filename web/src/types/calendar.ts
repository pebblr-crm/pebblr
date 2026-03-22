export type CalendarEventType = 'sync' | 'visit' | 'review' | 'callback' | 'lunch' | 'demo'

export interface CalendarEvent {
  id: string
  title: string
  type: CalendarEventType
  time: string
  date: string
  client: string
}

export interface CalendarEventListParams {
  month?: number
  year?: number
  assigneeId?: string
}
