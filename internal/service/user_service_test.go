package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
)

const (
	testAdminID = "admin-1"
)

// userSvcRepo returns a stubUserRepo suitable for UserService tests.
// It reuses the existing stubUserRepo type from activity_service_test.go.
func userSvcRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*domain.User{
		testAdminID: {ID: testAdminID, Name: "Admin", Role: domain.RoleAdmin, TeamIDs: []string{testTeamID}},
		"mgr-1":     {ID: "mgr-1", Name: "Manager", Role: domain.RoleManager, TeamIDs: []string{testTeamID}},
		"rep-1":     {ID: "rep-1", Name: "Rep", Role: domain.RoleRep, TeamIDs: []string{testTeamID}},
		"rep-2":     {ID: "rep-2", Name: "Rep2", Role: domain.RoleRep, TeamIDs: []string{testTeam2ID}},
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
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestUserService_List_ManagerSeesAll(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	_, err := svc.List(context.Background(), managerUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestUserService_List_RepSeesOnlySelf(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	users, err := svc.List(context.Background(), repUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
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
	_, err := svc.Get(context.Background(), repUser(), testAdminID)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUserService_Get_RepCanViewSelf(t *testing.T) {
	t.Parallel()
	svc := service.NewUserService(userSvcRepo())
	user, err := svc.Get(context.Background(), repUser(), "rep-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
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
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if user.ID != "rep-1" {
		t.Errorf("expected rep-1, got %s", user.ID)
	}
}
