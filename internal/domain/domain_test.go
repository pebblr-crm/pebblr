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

func TestUserValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid user", func(t *testing.T) {
		t.Parallel()
		u := &domain.User{ID: "u-1", ExternalID: "ext-1", Role: domain.RoleRep}
		if err := u.Validate(); err != nil {
			t.Errorf("expected valid, got %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		t.Parallel()
		u := &domain.User{ID: "", ExternalID: "ext-1", Role: domain.RoleRep}
		if err := u.Validate(); err == nil {
			t.Error("expected error for empty ID")
		}
	})
	t.Run("empty external ID", func(t *testing.T) {
		t.Parallel()
		u := &domain.User{ID: "u-1", ExternalID: "", Role: domain.RoleRep}
		if err := u.Validate(); err == nil {
			t.Error("expected error for empty external ID")
		}
	})
	t.Run("invalid role", func(t *testing.T) {
		t.Parallel()
		u := &domain.User{ID: "u-1", ExternalID: "ext-1", Role: domain.Role("superuser")}
		if err := u.Validate(); err == nil {
			t.Error("expected error for invalid role")
		}
	})
}

func TestActivityPatchApplyToNilFields(t *testing.T) {
	t.Parallel()
	// Verify that applying a patch with FieldsPresent=true to an activity
	// with nil Fields map initializes the map rather than panicking.
	dst := &domain.Activity{Status: "planned"}
	patch := &domain.ActivityPatch{
		FieldsPresent: true,
		Fields:        map[string]any{"key": "value"},
	}
	patch.ApplyTo(dst)
	if dst.Fields["key"] != "value" {
		t.Errorf("expected field key=value, got %v", dst.Fields["key"])
	}
}

func TestActivityPatchApplyToAllFields(t *testing.T) {
	t.Parallel()

	newDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	newStatus := "completed"
	newDuration := "half_day"
	newRouting := "week-2"
	newTargetID := "target-99"
	newJointVisit := "user-42"

	dst := &domain.Activity{
		Status:        "planned",
		DueDate:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Duration:      "full_day",
		Routing:       "week-1",
		TargetID:      "target-1",
		JointVisitUID: "user-1",
		Fields:        map[string]any{"existing": "keep", "removeme": "gone"},
	}

	patch := &domain.ActivityPatch{
		Status:        &newStatus,
		DueDate:       &newDate,
		Duration:      &newDuration,
		Routing:       &newRouting,
		TargetID:      &newTargetID,
		JointVisitUID: &newJointVisit,
		FieldsPresent: true,
		Fields:        map[string]any{"new": "val", "removeme": nil},
	}

	patch.ApplyTo(dst)

	if dst.Status != newStatus {
		t.Errorf("Status: got %q, want %q", dst.Status, newStatus)
	}
	if !dst.DueDate.Equal(newDate) {
		t.Errorf("DueDate: got %v, want %v", dst.DueDate, newDate)
	}
	if dst.Duration != newDuration {
		t.Errorf("Duration: got %q, want %q", dst.Duration, newDuration)
	}
	if dst.Routing != newRouting {
		t.Errorf("Routing: got %q, want %q", dst.Routing, newRouting)
	}
	if dst.TargetID != newTargetID {
		t.Errorf("TargetID: got %q, want %q", dst.TargetID, newTargetID)
	}
	if dst.JointVisitUID != newJointVisit {
		t.Errorf("JointVisitUID: got %q, want %q", dst.JointVisitUID, newJointVisit)
	}
	// Existing field untouched by patch should remain.
	if dst.Fields["existing"] != "keep" {
		t.Errorf("existing field should be untouched, got %v", dst.Fields["existing"])
	}
	// New field should be added.
	if dst.Fields["new"] != "val" {
		t.Errorf("new field: got %v, want val", dst.Fields["new"])
	}
	// Nil-value field should be removed.
	if _, ok := dst.Fields["removeme"]; ok {
		t.Error("removeme field should have been deleted")
	}
}

func TestActivityPatchApplyToNoOp(t *testing.T) {
	t.Parallel()
	// A completely empty patch should leave the activity unchanged.
	dst := &domain.Activity{
		Status:   "planned",
		Duration: "full_day",
		TargetID: "target-1",
	}
	patch := &domain.ActivityPatch{}
	patch.ApplyTo(dst)

	if dst.Status != "planned" {
		t.Errorf("Status should be unchanged, got %q", dst.Status)
	}
	if dst.Duration != "full_day" {
		t.Errorf("Duration should be unchanged, got %q", dst.Duration)
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

