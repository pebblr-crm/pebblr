package rbac_test

import (
	"context"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

func TestRepCanOnlySeeOwnLeads(t *testing.T) {
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	ownLead := &domain.Lead{ID: "lead-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherLead := &domain.Lead{ID: "lead-2", AssigneeID: "rep-2", TeamID: "team-1"}

	if !enforcer.CanViewLead(ctx, rep, ownLead) {
		t.Error("rep should be able to view own lead")
	}
	if enforcer.CanViewLead(ctx, rep, otherLead) {
		t.Error("rep should not be able to view another rep's lead")
	}
}

func TestManagerCanSeeTeamLeads(t *testing.T) {
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	manager := &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
	teamLead := &domain.Lead{ID: "lead-1", AssigneeID: "rep-1", TeamID: "team-1"}
	otherTeamLead := &domain.Lead{ID: "lead-2", AssigneeID: "rep-2", TeamID: "team-2"}

	if !enforcer.CanViewLead(ctx, manager, teamLead) {
		t.Error("manager should see team lead")
	}
	if enforcer.CanViewLead(ctx, manager, otherTeamLead) {
		t.Error("manager should not see lead from another team")
	}
}

func TestAdminCanSeeAllLeads(t *testing.T) {
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	admin := &domain.User{ID: "admin-1", Role: domain.RoleAdmin}
	anyLead := &domain.Lead{ID: "lead-1", AssigneeID: "someone", TeamID: "any-team"}

	if !enforcer.CanViewLead(ctx, admin, anyLead) {
		t.Error("admin should see all leads")
	}
}

func TestScopeLeadQueryForRep(t *testing.T) {
	enforcer := rbac.NewEnforcer()
	ctx := context.Background()

	rep := &domain.User{ID: "rep-1", Role: domain.RoleRep}
	scope := enforcer.ScopeLeadQuery(ctx, rep)

	if scope.AllLeads {
		t.Error("rep scope should not be all leads")
	}
	if len(scope.AssigneeIDs) != 1 || scope.AssigneeIDs[0] != rep.ID {
		t.Errorf("rep scope should restrict to own ID, got %v", scope.AssigneeIDs)
	}
}

func TestContextUserRoundtrip(t *testing.T) {
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
