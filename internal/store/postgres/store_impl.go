package postgres

import (
	"github.com/pebblr/pebblr/internal/store"
)

// Verify DB implements store.Store at compile time.
var _ store.Store = (*DB)(nil)

// Users returns the PostgreSQL-backed user repository.
func (db *DB) Users() store.UserRepository {
	return &userRepository{pool: db.pool}
}

// Teams returns the PostgreSQL-backed team repository.
func (db *DB) Teams() store.TeamRepository {
	return &teamRepository{pool: db.pool}
}

// Targets returns the PostgreSQL-backed target repository.
func (db *DB) Targets() store.TargetRepository {
	return &targetRepository{pool: db.pool}
}

// Activities returns the PostgreSQL-backed activity repository.
func (db *DB) Activities() store.ActivityRepository {
	return &activityRepository{pool: db.pool}
}

// Audit returns the PostgreSQL-backed audit repository.
func (db *DB) Audit() store.AuditRepository {
	return &auditRepository{pool: db.pool}
}

// Dashboard returns the PostgreSQL-backed dashboard repository.
func (db *DB) Dashboard() store.DashboardRepository {
	return &dashboardRepository{pool: db.pool}
}

// Collections returns the PostgreSQL-backed collection repository.
func (db *DB) Collections() store.CollectionRepository {
	return &collectionRepository{pool: db.pool}
}
