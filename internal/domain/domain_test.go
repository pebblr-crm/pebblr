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

func TestActivityPrepareForResponse(t *testing.T) {
	t.Parallel()

	t.Run("nil fields map is initialized", func(t *testing.T) {
		t.Parallel()
		a := &domain.Activity{}
		a.PrepareForResponse()
		if a.Fields == nil {
			t.Fatal("Fields should be initialized after PrepareForResponse")
		}
	})

	t.Run("joint visit user ID is hoisted into fields", func(t *testing.T) {
		t.Parallel()
		a := &domain.Activity{JointVisitUID: "user-99"}
		a.PrepareForResponse()
		if a.Fields["joint_visit_user_id"] != "user-99" {
			t.Errorf("expected joint_visit_user_id in fields, got %v", a.Fields)
		}
	})

	t.Run("empty joint visit user ID is not hoisted", func(t *testing.T) {
		t.Parallel()
		a := &domain.Activity{}
		a.PrepareForResponse()
		if _, ok := a.Fields["joint_visit_user_id"]; ok {
			t.Error("joint_visit_user_id should not appear when JointVisitUID is empty")
		}
	})
}

func TestActivityPatchApplyTo(t *testing.T) {
	t.Parallel()

	t.Run("nil pointer fields are not applied", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{Status: "planned", Duration: "full_day"}
		p := &domain.ActivityPatch{}
		p.ApplyTo(dst)
		if dst.Status != "planned" || dst.Duration != "full_day" {
			t.Errorf("nil patch fields should leave dst unchanged, got status=%q duration=%q", dst.Status, dst.Duration)
		}
	})

	t.Run("non-nil pointer fields are applied", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{Status: "planned", Duration: "full_day", Routing: "week_1"}
		status := "completed"
		duration := "half_day"
		routing := "week_2"
		p := &domain.ActivityPatch{Status: &status, Duration: &duration, Routing: &routing}
		p.ApplyTo(dst)
		if dst.Status != "completed" {
			t.Errorf("expected status %q, got %q", "completed", dst.Status)
		}
		if dst.Duration != "half_day" {
			t.Errorf("expected duration %q, got %q", "half_day", dst.Duration)
		}
		if dst.Routing != "week_2" {
			t.Errorf("expected routing %q, got %q", "week_2", dst.Routing)
		}
	})

	t.Run("fields merge semantics: add and delete keys", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{
			Fields: map[string]any{"keep": "yes", "remove": "old"},
		}
		p := &domain.ActivityPatch{
			FieldsPresent: true,
			Fields:        map[string]any{"remove": nil, "add": "new"},
		}
		p.ApplyTo(dst)
		if dst.Fields["keep"] != "yes" {
			t.Error("absent keys should be left untouched")
		}
		if _, ok := dst.Fields["remove"]; ok {
			t.Error("nil value should delete the key")
		}
		if dst.Fields["add"] != "new" {
			t.Error("non-nil value should be set")
		}
	})

	t.Run("fields merge on nil dst.Fields initializes the map", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{}
		p := &domain.ActivityPatch{
			FieldsPresent: true,
			Fields:        map[string]any{"key": "val"},
		}
		p.ApplyTo(dst)
		if dst.Fields["key"] != "val" {
			t.Errorf("expected key=val in fields, got %v", dst.Fields)
		}
	})

	t.Run("TargetID and JointVisitUID are applied", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{}
		tid := "target-1"
		jvid := "user-2"
		p := &domain.ActivityPatch{TargetID: &tid, JointVisitUID: &jvid}
		p.ApplyTo(dst)
		if dst.TargetID != "target-1" {
			t.Errorf("expected targetId %q, got %q", "target-1", dst.TargetID)
		}
		if dst.JointVisitUID != "user-2" {
			t.Errorf("expected jointVisitUID %q, got %q", "user-2", dst.JointVisitUID)
		}
	})

	t.Run("DueDate is applied", func(t *testing.T) {
		t.Parallel()
		dst := &domain.Activity{}
		due := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		p := &domain.ActivityPatch{DueDate: &due}
		p.ApplyTo(dst)
		if !dst.DueDate.Equal(due) {
			t.Errorf("expected dueDate %v, got %v", due, dst.DueDate)
		}
	})
}

func TestRolePermissions(t *testing.T) {
	t.Parallel()
	repPerms := domain.RoleRep.Permissions()
	if len(repPerms) == 0 {
		t.Error("rep should have permissions")
	}

	adminPerms := domain.RoleAdmin.Permissions()
	managerPerms := domain.RoleManager.Permissions()

	if len(adminPerms) <= len(managerPerms) {
		t.Error("admin should have more permissions than manager")
	}
	if len(managerPerms) <= len(repPerms) {
		t.Error("manager should have more permissions than rep")
	}
}
