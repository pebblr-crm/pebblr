package domain_test

import (
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
)

func TestActivityIsSubmitted(t *testing.T) {
	t.Parallel()
	a := &domain.Activity{}
	if a.IsSubmitted() {
		t.Error("activity without SubmittedAt should not be submitted")
	}

	now := time.Now()
	a.SubmittedAt = &now
	if !a.IsSubmitted() {
		t.Error("activity with SubmittedAt should be submitted")
	}
}

func TestOnlineStatusValid(t *testing.T) {
	t.Parallel()
	valid := []domain.OnlineStatus{
		domain.OnlineStatusOnline,
		domain.OnlineStatusAway,
		domain.OnlineStatusOffline,
	}
	for _, s := range valid {
		s := s
		t.Run(string(s), func(t *testing.T) {
			t.Parallel()
			if !s.Valid() {
				t.Errorf("expected %q to be valid", s)
			}
		})
	}
	if domain.OnlineStatus("busy").Valid() {
		t.Error("expected unknown online status to be invalid")
	}
}

func TestRoleValid(t *testing.T) {
	t.Parallel()
	for _, r := range []domain.Role{domain.RoleRep, domain.RoleManager, domain.RoleAdmin} {
		r := r
		t.Run(string(r), func(t *testing.T) {
			t.Parallel()
			if !r.Valid() {
				t.Errorf("expected role %q to be valid", r)
			}
		})
	}
	if domain.Role("superuser").Valid() {
		t.Error("unknown role should be invalid")
	}
}

func TestValidateBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		bound   map[string]any
		wantErr bool
	}{
		{name: "nil boundary is valid", bound: nil, wantErr: false},
		{name: "empty boundary is valid", bound: map[string]any{}, wantErr: false},
		{name: "valid polygon", bound: map[string]any{
			"type":        "Polygon",
			"coordinates": []any{[]any{[]any{0.0, 0.0}}},
		}, wantErr: false},
		{name: "valid point", bound: map[string]any{
			"type":        "Point",
			"coordinates": []any{0.0, 0.0},
		}, wantErr: false},
		{name: "valid geometry collection", bound: map[string]any{
			"type":       "GeometryCollection",
			"geometries": []any{},
		}, wantErr: false},
		{name: "missing type", bound: map[string]any{
			"coordinates": []any{0.0, 0.0},
		}, wantErr: true},
		{name: "unknown type", bound: map[string]any{
			"type":        "Hexagon",
			"coordinates": []any{},
		}, wantErr: true},
		{name: "missing coordinates", bound: map[string]any{
			"type": "Polygon",
		}, wantErr: true},
		{name: "geometry collection without geometries", bound: map[string]any{
			"type": "GeometryCollection",
		}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			territory := &domain.Territory{Boundary: tt.bound}
			err := territory.ValidateBoundary()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBoundary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

