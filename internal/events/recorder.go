package events

import "context"

// EventRecorder writes lead lifecycle events to persistent storage.
type EventRecorder interface {
	// Record persists a new event. Returns an error if the write fails.
	Record(ctx context.Context, event *LeadEvent) error
}
