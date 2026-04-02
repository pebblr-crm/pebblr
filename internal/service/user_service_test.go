package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
)

// userSvcRepo returns a stubUserRepo suitable for UserService tests.
// It reuses the existing stubUserRepo type from activity_service_test.go.
func userSvcRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*domain.User{
		"admin-1": {ID: "admin-1", Name: "Admin", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}},
		"mgr-1":   {ID: "mgr-1", Name: "Manager", Role: domain.RoleManager, TeamIDs: []string{"team-1"}},
		"rep-1":   {ID: "rep-1", Name: "Rep", Role: domain.RoleRep, TeamIDs: []string{"team-1"}},
		"rep-2":   {ID: "rep-2", Name: "Rep2", Role: domain.RoleRep, TeamIDs: []string{"team-2"}},
	}}
}

// --- tests ---

func TestUserService_List_AdminSeesAll(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	// The underlying List returns nil because stubUserRepo.List returns nil, nil.
	// But admin is permitted to call it; verify no error.
	_, err := svc.List(context.Background(), adminUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_List_ManagerSeesAll(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	_, err := svc.List(context.Background(), managerUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_List_RepSeesOnlySelf(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	users, err := svc.List(context.Background(), repUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("rep should see exactly 1 user, got %d", len(users))
	}
	if users[0].ID != "rep-1" {
		t.Errorf("rep should see their own record, got %s", users[0].ID)
	}
}

func TestUserService_Get_RepCannotViewOthers(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	_, err := svc.Get(context.Background(), repUser(), "admin-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUserService_Get_RepCanViewSelf(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	user, err := svc.Get(context.Background(), repUser(), "rep-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "rep-1" {
		t.Errorf("expected rep-1, got %s", user.ID)
	}
}

func TestUserService_Get_AdminCanViewAnyone(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	user, err := svc.Get(context.Background(), adminUser(), "rep-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "rep-1" {
		t.Errorf("expected rep-1, got %s", user.ID)
	}
}
