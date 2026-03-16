package events_test

import (
	"testing"

	"github.com/pebblr/pebblr/internal/events"
)

func TestEventTypeValid(t *testing.T) {
	valid := []events.EventType{
		events.EventTypeCreated,
		events.EventTypeAssigned,
		events.EventTypeStatusChanged,
		events.EventTypeNoteAdded,
		events.EventTypeVisited,
		events.EventTypeClosed,
	}
	for _, et := range valid {
		if !et.Valid() {
			t.Errorf("expected %q to be valid", et)
		}
	}

	if events.EventType("unknown").Valid() {
		t.Error("unknown event type should be invalid")
	}
}
