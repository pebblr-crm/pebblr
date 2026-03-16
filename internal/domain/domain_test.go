package domain_test

import (
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
)

func TestLeadStatusValid(t *testing.T) {
	t.Parallel()
	valid := []domain.LeadStatus{
		domain.LeadStatusNew,
		domain.LeadStatusAssigned,
		domain.LeadStatusInProgress,
		domain.LeadStatusVisited,
		domain.LeadStatusClosedWon,
		domain.LeadStatusClosedLost,
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

	if domain.LeadStatus("unknown").Valid() {
		t.Error("expected unknown status to be invalid")
	}
}

func TestLeadStatusTerminal(t *testing.T) {
	t.Parallel()
	if !domain.LeadStatusClosedWon.Terminal() {
		t.Error("closed_won should be terminal")
	}
	if !domain.LeadStatusClosedLost.Terminal() {
		t.Error("closed_lost should be terminal")
	}
	if domain.LeadStatusNew.Terminal() {
		t.Error("new should not be terminal")
	}
}

func TestCustomerTypeValid(t *testing.T) {
	t.Parallel()
	valid := []domain.CustomerType{
		domain.CustomerTypeRetail,
		domain.CustomerTypeWholesale,
		domain.CustomerTypeHospitality,
		domain.CustomerTypeInstitutional,
		domain.CustomerTypeOther,
	}
	for _, ct := range valid {
		ct := ct
		t.Run(string(ct), func(t *testing.T) {
			t.Parallel()
			if !ct.Valid() {
				t.Errorf("expected %q to be valid", ct)
			}
		})
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
