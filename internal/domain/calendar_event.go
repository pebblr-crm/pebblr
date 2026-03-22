package domain

import "time"

// CalendarEventType categorises the kind of calendar event.
type CalendarEventType string

const (
	// CalendarEventTypeSync is a team synchronisation meeting.
	CalendarEventTypeSync CalendarEventType = "sync"
	// CalendarEventTypeVisit is a customer site visit.
	CalendarEventTypeVisit CalendarEventType = "visit"
	// CalendarEventTypeReview is a performance or account review.
	CalendarEventTypeReview CalendarEventType = "review"
	// CalendarEventTypeCallback is a scheduled callback with a lead or customer.
	CalendarEventTypeCallback CalendarEventType = "callback"
	// CalendarEventTypeLunch is a lunch meeting.
	CalendarEventTypeLunch CalendarEventType = "lunch"
	// CalendarEventTypeDemo is a product demonstration.
	CalendarEventTypeDemo CalendarEventType = "demo"
)

// Valid returns true if the CalendarEventType is a recognised value.
func (t CalendarEventType) Valid() bool {
	switch t {
	case CalendarEventTypeSync, CalendarEventTypeVisit, CalendarEventTypeReview,
		CalendarEventTypeCallback, CalendarEventTypeLunch, CalendarEventTypeDemo:
		return true
	}
	return false
}

// CalendarEvent represents a scheduled event associated with a sales activity.
type CalendarEvent struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	EventType CalendarEventType `json:"eventType"`
	StartTime time.Time         `json:"startTime"`
	// EndTime is optional; nil means no defined end time.
	EndTime   *time.Time `json:"endTime,omitempty"`
	Client    string     `json:"client"`
	CreatorID string     `json:"creatorId"`
	// TeamID is optional; associates the event with a team.
	TeamID    string    `json:"teamId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
