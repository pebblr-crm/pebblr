package events

import "time"

// LeadEvent records a single lifecycle event on a lead.
type LeadEvent struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	EventType EventType `json:"event_type"`
	// ActorID is the User.ID of the person who triggered the event.
	ActorID string `json:"actor_id"`
	// Payload holds event-specific data encoded as JSON.
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
