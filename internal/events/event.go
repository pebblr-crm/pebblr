package events

import "time"

// LeadEvent records a single lifecycle event on a lead.
type LeadEvent struct {
	ID        string
	LeadID    string
	EventType EventType
	// ActorID is the User.ID of the person who triggered the event.
	ActorID   string
	// Payload holds event-specific data encoded as JSON.
	Payload   []byte
	Timestamp time.Time
}
