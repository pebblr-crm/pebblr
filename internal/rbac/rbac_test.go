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

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownTarget := &domain.Target{ID: "target-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherTarget := &domain.Target{ID: "target-2", AssigneeID: "rep-2", TeamID: "team-1"}

	if !enforcer.CanViewTarget(rep, ownTarget) {
		t.Error("rep should be able to view own target")
	}
	if enforcer.CanViewTarget(rep, otherTarget) {
		t.Error("rep should not be able to view another rep's target")
	}
}

func TestManagerCanSeeTeamTargets(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	teamTarget := &domain.Target{ID: "target-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherTeamTarget := &domain.Target{ID: "target-2", AssigneeID: "rep-2", TeamID: "team-2"}

	if !enforcer.CanViewTarget(manager, teamTarget) {
		t.Error("manager should see team target")
	}
	if enforcer.CanViewTarget(manager, otherTeamTarget) {
		t.Error("manager should not see target from another team")
	}
}

func TestAdminCanSeeAllTargets(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	anyTarget := &domain.Target{ID: "target-1", AssigneeID: "someone", TeamID: "any-team"}

	if !enforcer.CanViewTarget(admin, anyTarget) {
		t.Error("admin should see all targets")
	}
}

func TestScopeTargetQueryForRep(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeTargetQuery(rep)

	if scope.AllTargets {
		t.Error("rep scope should not be all targets")
	}
	if scope.DenyAll {
		t.Error("rep scope should not be deny-all")
	}
	if len(scope.AssigneeIDs) != 1 || scope.AssigneeIDs[0] != rep.ID {
		t.Errorf("rep scope should restrict to own ID, got %v", scope.AssigneeIDs)
	}
}

func TestScopeTargetQueryDenyAllForUnknownRole(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	unknown := &domain.User{ID: "x", Role: domain.Role("bogus")}
	scope := enforcer.ScopeTargetQuery(unknown)

	if !scope.DenyAll {
		t.Error("unknown role should produce deny-all scope")
	}
}

// ── Activity RBAC Tests ─────────────────────────────────────────────────

func TestRepCanOnlySeeOwnActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	jointActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", JointVisitUID: "rep-1", TeamID: "team-1"}
	otherActivity := &domain.Activity{ID: "act-3", CreatorID: "rep-2", TeamID: "team-1"}

	if !enforcer.CanViewActivity(rep, ownActivity) {
		t.Error("rep should view own activity")
	}
	if !enforcer.CanViewActivity(rep, jointActivity) {
		t.Error("rep should view joint visit activity")
	}
	if enforcer.CanViewActivity(rep, otherActivity) {
		t.Error("rep should not view another rep's activity")
	}
}

func TestRepCanOnlyUpdateOwnActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	jointActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", JointVisitUID: "rep-1", TeamID: "team-1"}

	if !enforcer.CanUpdateActivity(rep, ownActivity) {
		t.Error("rep should update own activity")
	}
	if enforcer.CanUpdateActivity(rep, jointActivity) {
		t.Error("rep should not update activity they didn't create (even as joint visitor)")
	}
}

func TestManagerCanSeeTeamActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	teamActivity := &domain.Activity{ID: "act-1", CreatorID: "rep-1", TeamID: "team-1"}
	otherTeamActivity := &domain.Activity{ID: "act-2", CreatorID: "rep-2", TeamID: "team-2"}

	if !enforcer.CanViewActivity(manager, teamActivity) {
		t.Error("manager should see team activity")
	}
	if enforcer.CanViewActivity(manager, otherTeamActivity) {
		t.Error("manager should not see activity from another team")
	}
}

func TestAdminCanSeeAllActivities(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	anyActivity := &domain.Activity{ID: "act-1", CreatorID: "someone", TeamID: "any-team"}

	if !enforcer.CanViewActivity(admin, anyActivity) {
		t.Error("admin should see all activities")
	}
	if !enforcer.CanUpdateActivity(admin, anyActivity) {
		t.Error("admin should update all activities")
	}
	if !enforcer.CanDeleteActivity(admin, anyActivity) {
		t.Error("admin should delete all activities")
	}
}

func TestScopeActivityQueryForRep(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeActivityQuery(rep)

	if scope.AllActivities {
		t.Error("rep scope should not be all activities")
	}
	if scope.DenyAll {
		t.Error("rep scope should not be deny-all")
	}
	if len(scope.CreatorIDs) != 1 || scope.CreatorIDs[0] != rep.ID {
		t.Errorf("rep scope should restrict to own ID, got %v", scope.CreatorIDs)
	}
}

func TestScopeActivityQueryForAdmin(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	scope := enforcer.ScopeActivityQuery(admin)

	if !scope.AllActivities {
		t.Error("admin scope should be all activities")
	}
}

func TestScopeActivityQueryDenyAllForUnknownRole(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	unknown := &domain.User{ID: "x", Role: domain.Role("bogus")}
	scope := enforcer.ScopeActivityQuery(unknown)

	if !scope.DenyAll {
		t.Error("unknown role should produce deny-all scope")
	}
}

func TestNilActorDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	target := &domain.Target{ID: "t-1", AssigneeID: "rep-1", TeamID: "team-1"}
	activity := &domain.Activity{ID: "a-1", CreatorID: "rep-1", TeamID: "team-1"}

	if enforcer.CanViewTarget(nil, target) {
		t.Error("nil actor must not view targets")
	}
	if enforcer.CanUpdateTarget(nil, target) {
		t.Error("nil actor must not update targets")
	}
	if enforcer.CanViewActivity(nil, activity) {
		t.Error("nil actor must not view activities")
	}
	if enforcer.CanUpdateActivity(nil, activity) {
		t.Error("nil actor must not update activities")
	}
	if enforcer.CanDeleteActivity(nil, activity) {
		t.Error("nil actor must not delete activities")
	}
}

func TestInvalidRoleDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	bogus := &domain.User{ID: "x", Role: domain.Role("superuser")}
	target := &domain.Target{ID: "t-1", AssigneeID: "x", TeamID: "team-1"}
	activity := &domain.Activity{ID: "a-1", CreatorID: "x", TeamID: "team-1"}

	if enforcer.CanViewTarget(bogus, target) {
		t.Error("invalid role must not view targets")
	}
	if enforcer.CanViewActivity(bogus, activity) {
		t.Error("invalid role must not view activities")
	}

	scope := enforcer.ScopeTargetQuery(bogus)
	if scope.AllTargets || len(scope.AssigneeIDs) > 0 {
		t.Error("invalid role scope must match nothing")
	}
}

func TestNilResourceDeniesAccess(t *testing.T) {
	t.Parallel()
	enforcer := rbac.NewEnforcer()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}

	if enforcer.CanViewTarget(admin, nil) {
		t.Error("nil target must deny access even for admin")
	}
	if enforcer.CanViewActivity(admin, nil) {
		t.Error("nil activity must deny access even for admin")
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
