package postgres

import (
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestDB_InitRepos(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("creating pgxmock pool: %v", err)
	}

	db := &DB{pool: mock}
	db.initRepos()

	// Verify all repository accessors return non-nil.
	if db.Users() == nil {
		t.Error("Users() returned nil")
	}
	if db.Teams() == nil {
		t.Error("Teams() returned nil")
	}
	if db.Targets() == nil {
		t.Error("Targets() returned nil")
	}
	if db.Activities() == nil {
		t.Error("Activities() returned nil")
	}
	if db.Audit() == nil {
		t.Error("Audit() returned nil")
	}
	if db.Dashboard() == nil {
		t.Error("Dashboard() returned nil")
	}
	if db.Collections() == nil {
		t.Error("Collections() returned nil")
	}
	if db.Territories() == nil {
		t.Error("Territories() returned nil")
	}
}
