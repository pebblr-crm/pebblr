package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
)

type teamRepository struct {
	pool *pgxpool.Pool
}

func (r *teamRepository) Get(_ context.Context, _ string) (*domain.Team, error) {
	// TODO: implement
	return nil, nil
}

func (r *teamRepository) List(_ context.Context) ([]*domain.Team, error) {
	// TODO: implement
	return nil, nil
}

func (r *teamRepository) Create(_ context.Context, team *domain.Team) (*domain.Team, error) {
	// TODO: implement
	return team, nil
}

func (r *teamRepository) Update(_ context.Context, team *domain.Team) (*domain.Team, error) {
	// TODO: implement
	return team, nil
}

func (r *teamRepository) Delete(_ context.Context, _ string) error {
	// TODO: implement
	return nil
}

func (r *teamRepository) ListMembers(_ context.Context, _ string) ([]*domain.User, error) {
	// TODO: implement
	return nil, nil
}
