package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

type leadRepository struct {
	pool *pgxpool.Pool
}

func (r *leadRepository) Get(_ context.Context, _ string) (*domain.Lead, error) {
	// TODO: implement
	return nil, store.ErrNotFound
}

func (r *leadRepository) List(_ context.Context, _ rbac.LeadScope, _ store.LeadFilter, _, _ int) (*store.LeadPage, error) {
	// TODO: implement
	return &store.LeadPage{}, nil
}

func (r *leadRepository) Create(_ context.Context, lead *domain.Lead) (*domain.Lead, error) {
	// TODO: implement
	return lead, nil
}

func (r *leadRepository) Update(_ context.Context, lead *domain.Lead) (*domain.Lead, error) {
	// TODO: implement
	return lead, nil
}

func (r *leadRepository) Delete(_ context.Context, _ string) error {
	// TODO: implement
	return nil
}
