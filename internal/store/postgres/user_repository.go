package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func (r *userRepository) GetByID(_ context.Context, _ string) (*domain.User, error) {
	// TODO: implement
	return nil, nil
}

func (r *userRepository) GetByExternalID(_ context.Context, _ string) (*domain.User, error) {
	// TODO: implement
	return nil, nil
}

func (r *userRepository) List(_ context.Context) ([]*domain.User, error) {
	// TODO: implement
	return nil, nil
}

func (r *userRepository) Upsert(_ context.Context, user *domain.User) (*domain.User, error) {
	// TODO: implement
	return user, nil
}
