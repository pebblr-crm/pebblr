package service_test

import "github.com/pebblr/pebblr/internal/domain"

func adminUser() *domain.User {
	return &domain.User{
		ID:      "admin-1",
		Name:    "Admin User",
		Role:    domain.RoleAdmin,
		TeamIDs: []string{"team-1"},
	}
}

func managerUser() *domain.User {
	return &domain.User{
		ID:      "mgr-1",
		Name:    "Manager User",
		Role:    domain.RoleManager,
		TeamIDs: []string{"team-1"},
	}
}

func repUser() *domain.User {
	return &domain.User{
		ID:      "rep-1",
		Name:    "Rep User",
		Role:    domain.RoleRep,
		TeamIDs: []string{"team-1"},
	}
}
