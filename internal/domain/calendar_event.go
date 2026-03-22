package domain

import "time"

// CalendarEventType classifies the kind of calendar event.
type CalendarEventType string

const (
	// CalendarEventTypeCall represents a scheduled phone or video call.
	CalendarEventTypeCall CalendarEventType = "call"
	// CalendarEventTypeMeeting represents an in-person or virtual meeting.
	CalendarEventTypeMeeting CalendarEventType = "meeting"
	// CalendarEventTypeSync is a team synchronisation meeting.
	CalendarEventTypeSync CalendarEventType = "sync"
	// CalendarEventTypeVisit represents a field sales visit.
	CalendarEventTypeVisit CalendarEventType = "visit"
	// CalendarEventTypeReview is a performance or account review.
	CalendarEventTypeReview CalendarEventType = "review"
	// CalendarEventTypeCallback is a scheduled callback with a lead or customer.
	CalendarEventTypeCallback CalendarEventType = "callback"
	// CalendarEventTypeLunch is a lunch meeting.
	CalendarEventTypeLunch CalendarEventType = "lunch"
	// CalendarEventTypeDemo is a product demonstration.
	CalendarEventTypeDemo CalendarEventType = "demo"
	// CalendarEventTypeOther represents any other calendar event type.
	CalendarEventTypeOther CalendarEventType = "other"
)

// Valid returns true if the event type is a recognized CalendarEventType.
func (t CalendarEventType) Valid() bool {
	switch t {
	case CalendarEventTypeCall, CalendarEventTypeMeeting, CalendarEventTypeSync,
		CalendarEventTypeVisit, CalendarEventTypeReview, CalendarEventTypeCallback,
		CalendarEventTypeLunch, CalendarEventTypeDemo, CalendarEventTypeOther:
		return true
	}
	return false
}

// CalendarEvent represents a scheduled event in the sales calendar.
type CalendarEvent struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	EventType CalendarEventType `json:"eventType"`
	StartTime time.Time         `json:"startTime"`
	EndTime   *time.Time        `json:"endTime,omitempty"`
	// Client is the customer/client name associated with this event.
	Client string `json:"client,omitempty"`
	// CreatorID is the User.ID of the event creator.
	CreatorID string `json:"creatorId"`
	// TeamID optionally associates the event with a team.
	TeamID    string    `json:"teamId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
