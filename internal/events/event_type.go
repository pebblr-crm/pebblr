package events

// EventType classifies what happened in a lead lifecycle event.
type EventType string

const (
	// EventTypeCreated fires when a lead is first created.
	EventTypeCreated EventType = "created"
	// EventTypeAssigned fires when a lead is assigned or reassigned to a rep.
	EventTypeAssigned EventType = "assigned"
	// EventTypeStatusChanged fires when a lead's status transitions.
	EventTypeStatusChanged EventType = "status_changed"
	// EventTypeNoteAdded fires when a note is added to a lead.
	EventTypeNoteAdded EventType = "note_added"
	// EventTypeVisited fires when a rep marks a lead as visited.
	EventTypeVisited EventType = "visited"
	// EventTypeClosed fires when a lead reaches a terminal status.
	EventTypeClosed EventType = "closed"
)

// Valid returns true if the event type is a recognized value.
func (t EventType) Valid() bool {
	switch t {
	case EventTypeCreated, EventTypeAssigned, EventTypeStatusChanged,
		EventTypeNoteAdded, EventTypeVisited, EventTypeClosed:
		return true
	}
	return false
}
