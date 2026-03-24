package postgres

import (
	"context"
	"errors"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// errNotImplemented is returned by team repository methods that have not yet
// been wired to the database. This ensures callers fail fast with a clear
// error rather than silently receiving nil results.
var errNotImplemented = errors.New("team repository: not implemented")

type teamRepository struct {
	pool dbPool
}

func (r *teamRepository) Get(_ context.Context, _ string) (*domain.Team, error) {
	return nil, store.ErrNotFound
}

func (r *teamRepository) List(_ context.Context) ([]*domain.Team, error) {
	return []*domain.Team{}, nil
}

func (r *teamRepository) Create(_ context.Context, _ *domain.Team) (*domain.Team, error) {
	return nil, errNotImplemented
}

func (r *teamRepository) Update(_ context.Context, _ *domain.Team) (*domain.Team, error) {
	return nil, errNotImplemented
}

func (r *teamRepository) Delete(_ context.Context, _ string) error {
	return errNotImplemented
}

func (r *teamRepository) ListMembers(_ context.Context, _ string) ([]*domain.User, error) {
	return []*domain.User{}, nil
}
