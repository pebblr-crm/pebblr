package events

import "time"

// LeadEvent records a single lifecycle event on a lead.
type LeadEvent struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"leadId"`
	EventType EventType `json:"eventType"`
	// ActorID is the User.ID of the person who triggered the event.
	ActorID string `json:"actorId"`
	// Payload holds event-specific data encoded as JSON.
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
