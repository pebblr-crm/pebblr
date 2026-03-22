export type CalendarEventType = 'call' | 'meeting' | 'sync' | 'visit' | 'review' | 'callback' | 'lunch' | 'demo' | 'other'

export interface CalendarEvent {
  id: string
  title: string
  eventType: CalendarEventType
  startTime: string
  endTime?: string
  client: string
  creatorId: string
  teamId?: string
  createdAt: string
  updatedAt: string
}

export interface CalendarEventListParams {
  month?: number
  year?: number
  assigneeId?: string
}
