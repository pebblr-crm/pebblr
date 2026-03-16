package events_test

import (
	"testing"

	"github.com/pebblr/pebblr/internal/events"
)

func TestEventTypeValid(t *testing.T) {
	t.Parallel()
	valid := []events.EventType{
		events.EventTypeCreated,
		events.EventTypeAssigned,
		events.EventTypeStatusChanged,
		events.EventTypeNoteAdded,
		events.EventTypeVisited,
		events.EventTypeClosed,
	}
	for _, et := range valid {
		et := et
		t.Run(string(et), func(t *testing.T) {
			t.Parallel()
			if !et.Valid() {
				t.Errorf("expected %q to be valid", et)
			}
		})
	}

	if events.EventType("unknown").Valid() {
		t.Error("unknown event type should be invalid")
	}
}
