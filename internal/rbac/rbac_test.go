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

// ── Activity RBAC Tests ─────────────────────────────────────────────────

func TestRepCanOnlySeeOwnActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	jointActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", JointVisitUID: "rep-1", TeamID: "team-1"}
	otherActivity := &domain.Activity{ID: "act-3", CreatorID: "rep-2", TeamID: "team-1"}

	if !enforcer.CanViewActivity(ctx, rep, ownActivity) {
		t.Error("rep should view own activity")
	}
	if !enforcer.CanViewActivity(ctx, rep, jointActivity) {
		t.Error("rep should view joint visit activity")
	}
	if enforcer.CanViewActivity(ctx, rep, otherActivity) {
		t.Error("rep should not view another rep's activity")
	}
}

func TestRepCanOnlyUpdateOwnActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	jointActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", JointVisitUID: "rep-1", TeamID: "team-1"}

	if !enforcer.CanUpdateActivity(ctx, rep, ownActivity) {
		t.Error("rep should update own activity")
	}
	if enforcer.CanUpdateActivity(ctx, rep, jointActivity) {
		t.Error("rep should not update activity they didn't create (even as joint visitor)")
	}
}

func TestManagerCanSeeTeamActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	teamActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	otherTeamActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", TeamID: "team-2"}

	if !enforcer.CanViewActivity(ctx, manager, teamActivity) {
		t.Error("manager should see team activity")
	}
	if enforcer.CanViewActivity(ctx, manager, otherTeamActivity) {
		t.Error("manager should not see activity from another team")
	}
}

func TestAdminCanSeeAllActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	anyActivity := &domain.Activity{ID: "act-1", CreatorID: "someone", TeamID: "any-team"}

	if !enforcer.CanViewActivity(ctx, admin, anyActivity) {
		t.Error("admin should see all activities")
	}
	if !enforcer.CanUpdateActivity(ctx, admin, anyActivity) {
		t.Error("admin should update all activities")
	}
	if !enforcer.CanDeleteActivity(ctx, admin, anyActivity) {
		t.Error("admin should delete all activities")
	}
}

func TestScopeActivityQueryForRep(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeActivityQuery(ctx, rep)

	if scope.AllActivities {
		t.Error("rep scope should not be all activities")
	}
	if len(scope.CreatorIDs) != 1 || scope.CreatorIDs[0] != rep.ID {
		t.Errorf("rep scope should restrict to own ID, got %v", scope.CreatorIDs)
	}
}

func TestScopeActivityQueryForAdmin(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	scope := enforcer.ScopeActivityQuery(ctx, admin)

	if !scope.AllActivities {
		t.Error("admin scope should be all activities")
	}
}

func TestScopeTargetQueryForManager(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1", "team-2"}}
	scope := enforcer.ScopeTargetQuery(ctx, manager)

	if scope.AllTargets {
		t.Error("manager scope should not be all targets")
	}
	if len(scope.TeamIDs) != 2 {
		t.Errorf("manager scope should restrict to team IDs, got %v", scope.TeamIDs)
	}
}

func TestScopeTargetQueryForAdmin(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	scope := enforcer.ScopeTargetQuery(ctx, admin)

	if !scope.AllTargets {
		t.Error("admin scope should be all targets")
	}
}

func TestScopeActivityQueryForManager(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	scope := enforcer.ScopeActivityQuery(ctx, manager)

	if scope.AllActivities {
		t.Error("manager scope should not be all activities")
	}
	if len(scope.TeamIDs) != 1 || scope.TeamIDs[0] != "team-1" {
		t.Errorf("manager scope should restrict to team IDs, got %v", scope.TeamIDs)
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

func TestUserFromContextMissing(t *testing.T) {
	t.Parallel()
	_, err := rbac.UserFromContext(context.Background())
	if err == nil {
		t.Fatal("expected error when no user is in context")
	}
}
