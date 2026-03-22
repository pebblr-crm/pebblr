package rbac_test

import (
	"context"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

func TestRepCanOnlySeeOwnTargets(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownTarget := &domain.Target{ID: "target-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherTarget := &domain.Target{ID: "target-2", AssigneeID: "rep-2", TeamID: "team-1"}

	if !enforcer.CanViewTarget(ctx, rep, ownTarget) {
		t.Error("rep should be able to view own target")
	}
	if enforcer.CanViewTarget(ctx, rep, otherTarget) {
		t.Error("rep should not be able to view another rep's target")
	}
}

func TestManagerCanSeeTeamTargets(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	teamTarget := &domain.Target{ID: "target-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherTeamTarget := &domain.Target{ID: "target-2", AssigneeID: "rep-2", TeamID: "team-2"}

	if !enforcer.CanViewTarget(ctx, manager, teamTarget) {
		t.Error("manager should see team target")
	}
	if enforcer.CanViewTarget(ctx, manager, otherTeamTarget) {
		t.Error("manager should not see target from another team")
	}
}

func TestAdminCanSeeAllTargets(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	anyTarget := &domain.Target{ID: "target-1", AssigneeID: "someone", TeamID: "any-team"}

	if !enforcer.CanViewTarget(ctx, admin, anyTarget) {
		t.Error("admin should see all targets")
	}
}

func TestScopeTargetQueryForRep(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeTargetQuery(ctx, rep)

	if scope.AllTargets {
		t.Error("rep scope should not be all targets")
	}
	if len(scope.AssigneeIDs) != 1 || scope.AssigneeIDs[0] != rep.ID {
		t.Errorf("rep scope should restrict to own ID, got %v", scope.AssigneeIDs)
	}
}

func TestContextUserRoundtrip(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "user-1", Role: domain.RoleRep}
	ctx := rbac.WithUser(context.Background(), user)

	got, err := rbac.UserFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("expected user ID %q, got %q", user.ID, got.ID)
	}
}
