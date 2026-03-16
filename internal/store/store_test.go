package store_test

import (
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/store"
)

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	if !errors.Is(store.ErrNotFound, store.ErrNotFound) {
		t.Error("ErrNotFound should match itself")
	}
	if !errors.Is(store.ErrConflict, store.ErrConflict) {
		t.Error("ErrConflict should match itself")
	}
	if errors.Is(store.ErrNotFound, store.ErrConflict) {
		t.Error("ErrNotFound should not match ErrConflict")
	}
}
