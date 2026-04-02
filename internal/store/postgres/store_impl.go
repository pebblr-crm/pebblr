package postgres

import (
	"github.com/pebblr/pebblr/internal/store"
)

// Verify DB implements store.Store at compile time.
var _ store.Store = (*DB)(nil)

// repos holds cached repository instances to avoid per-call allocations.
type repos struct {
	users       userRepository
	teams       teamRepository
	targets     targetRepository
	activities  activityRepository
	audit       auditRepository
	dashboard   dashboardRepository
	collections collectionRepository
	territories territoryRepository
}

// initRepos initialises the cached repository instances on the DB.
func (db *DB) initRepos() {
	db.r = repos{
		users:       userRepository{pool: db.pool},
		teams:       teamRepository{pool: db.pool},
		targets:     targetRepository{pool: db.pool},
		activities:  activityRepository{pool: db.pool},
		audit:       auditRepository{pool: db.pool},
		dashboard:   dashboardRepository{pool: db.pool},
		collections: collectionRepository{pool: db.pool},
		territories: territoryRepository{pool: db.pool},
	}
}

// Users returns the PostgreSQL-backed user repository.
func (db *DB) Users() store.UserRepository { return &db.r.users }

// Teams returns the PostgreSQL-backed team repository.
func (db *DB) Teams() store.TeamRepository { return &db.r.teams }

// Targets returns the PostgreSQL-backed target repository.
func (db *DB) Targets() store.TargetRepository { return &db.r.targets }

// Activities returns the PostgreSQL-backed activity repository.
func (db *DB) Activities() store.ActivityRepository { return &db.r.activities }

// Audit returns the PostgreSQL-backed audit repository.
func (db *DB) Audit() store.AuditRepository { return &db.r.audit }

// Dashboard returns the PostgreSQL-backed dashboard repository.
func (db *DB) Dashboard() store.DashboardRepository { return &db.r.dashboard }

// Collections returns the PostgreSQL-backed collection repository.
func (db *DB) Collections() store.CollectionRepository { return &db.r.collections }

// Territories returns the PostgreSQL-backed territory repository.
func (db *DB) Territories() store.TerritoryRepository { return &db.r.territories }
