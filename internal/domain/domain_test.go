package domain_test

import (
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
)

func TestCalendarEventTypeValid(t *testing.T) {
	t.Parallel()
	valid := []domain.CalendarEventType{
		domain.CalendarEventTypeSync,
		domain.CalendarEventTypeVisit,
		domain.CalendarEventTypeReview,
		domain.CalendarEventTypeCallback,
		domain.CalendarEventTypeLunch,
		domain.CalendarEventTypeDemo,
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
	if domain.CalendarEventType("unknown").Valid() {
		t.Error("expected unknown event type to be invalid")
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
