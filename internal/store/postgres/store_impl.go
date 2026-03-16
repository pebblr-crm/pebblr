package postgres

import (
	"github.com/pebblr/pebblr/internal/store"
)

// Verify DB implements store.Store at compile time.
var _ store.Store = (*DB)(nil)

func (db *DB) Leads() store.LeadRepository {
	return &leadRepository{pool: db.pool}
}

func (db *DB) Users() store.UserRepository {
	return &userRepository{pool: db.pool}
}

func (db *DB) Teams() store.TeamRepository {
	return &teamRepository{pool: db.pool}
}

func (db *DB) Events() store.EventRepository {
	return &eventRepository{pool: db.pool}
}
