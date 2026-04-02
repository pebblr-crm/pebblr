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

func TestNilActorDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	target := &domain.Target{ID: "t-1", AssigneeID: "rep-1", TeamID: "team-1"}
	activity := &domain.Activity{ID: "a-1", CreatorID: "rep-1", TeamID: "team-1"}

	if enforcer.CanViewTarget(ctx, nil, target) {
		t.Error("nil actor must not view targets")
	}
	if enforcer.CanUpdateTarget(ctx, nil, target) {
		t.Error("nil actor must not update targets")
	}
	if enforcer.CanViewActivity(ctx, nil, activity) {
		t.Error("nil actor must not view activities")
	}
	if enforcer.CanUpdateActivity(ctx, nil, activity) {
		t.Error("nil actor must not update activities")
	}
	if enforcer.CanDeleteActivity(ctx, nil, activity) {
		t.Error("nil actor must not delete activities")
	}

	scope := enforcer.ScopeTargetQuery(ctx, nil)
	if scope.AllTargets || len(scope.AssigneeIDs) > 0 || len(scope.TeamIDs) > 0 {
		t.Error("nil actor scope must match nothing")
	}

	actScope := enforcer.ScopeActivityQuery(ctx, nil)
	if actScope.AllActivities || len(actScope.CreatorIDs) > 0 || len(actScope.TeamIDs) > 0 {
		t.Error("nil actor activity scope must match nothing")
	}
}

func TestInvalidRoleDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	bogus := &domain.User{ID: "x", Role: domain.Role("superuser")}
	target := &domain.Target{ID: "t-1", AssigneeID: "x", TeamID: "team-1"}
	activity := &domain.Activity{ID: "a-1", CreatorID: "x", TeamID: "team-1"}

	if enforcer.CanViewTarget(ctx, bogus, target) {
		t.Error("invalid role must not view targets")
	}
	if enforcer.CanViewActivity(ctx, bogus, activity) {
		t.Error("invalid role must not view activities")
	}

	scope := enforcer.ScopeTargetQuery(ctx, bogus)
	if scope.AllTargets || len(scope.AssigneeIDs) > 0 {
		t.Error("invalid role scope must match nothing")
	}
}

func TestNilResourceDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}

	if enforcer.CanViewTarget(ctx, admin, nil) {
		t.Error("nil target must deny access even for admin")
	}
	if enforcer.CanViewActivity(ctx, admin, nil) {
		t.Error("nil activity must deny access even for admin")
	}
}

func TestRepActivityScopeIncludesJointVisit(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeActivityQuery(ctx, rep)

	if scope.JointVisitUID != rep.ID {
		t.Errorf("rep scope should set JointVisitUID=%q, got %q", rep.ID, scope.JointVisitUID)
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
