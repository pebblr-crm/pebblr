package postgres

import (
	"github.com/pebblr/pebblr/internal/store"
)

// Verify DB implements store.Store at compile time.
var _ store.Store = (*DB)(nil)

// Leads returns the PostgreSQL-backed lead repository.
func (db *DB) Leads() store.LeadRepository {
	return &leadRepository{pool: db.pool}
}

// Users returns the PostgreSQL-backed user repository.
func (db *DB) Users() store.UserRepository {
	return &userRepository{pool: db.pool}
}

// Teams returns the PostgreSQL-backed team repository.
func (db *DB) Teams() store.TeamRepository {
	return &teamRepository{pool: db.pool}
}

// Events returns the PostgreSQL-backed event repository.
func (db *DB) Events() store.EventRepository {
	return &eventRepository{pool: db.pool}
}

// Targets returns the PostgreSQL-backed target repository.
func (db *DB) Targets() store.TargetRepository {
	return &targetRepository{pool: db.pool}
}

// CalendarEvents returns the PostgreSQL-backed calendar event repository.
func (db *DB) CalendarEvents() store.CalendarEventRepository {
	return &calendarEventRepository{pool: db.pool}
}
