package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type teamRepository struct {
	pool dbPool
}

func (r *teamRepository) Get(ctx context.Context, id string) (*domain.Team, error) {
	var t domain.Team
	err := r.pool.QueryRow(ctx,
		`SELECT id::TEXT, name, manager_id::TEXT FROM teams WHERE id = $1::UUID`,
		id,
	).Scan(&t.ID, &t.Name, &t.ManagerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("getting team: %w", err)
	}
	return &t, nil
}

func (r *teamRepository) List(ctx context.Context) ([]*domain.Team, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::TEXT, name, manager_id::TEXT FROM teams ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing teams: %w", err)
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ManagerID); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating teams: %w", err)
	}
	return teams, nil
}

func (r *teamRepository) Create(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	var t domain.Team
	err := r.pool.QueryRow(ctx,
		`INSERT INTO teams (name, manager_id)
		 VALUES ($1, $2::UUID)
		 RETURNING id::TEXT, name, manager_id::TEXT`,
		team.Name, team.ManagerID,
	).Scan(&t.ID, &t.Name, &t.ManagerID)
	if err != nil {
		return nil, fmt.Errorf("creating team: %w", err)
	}
	return &t, nil
}

func (r *teamRepository) Update(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	var t domain.Team
	err := r.pool.QueryRow(ctx,
		`UPDATE teams SET name = $1, manager_id = $2::UUID, updated_at = NOW()
		 WHERE id = $3::UUID
		 RETURNING id::TEXT, name, manager_id::TEXT`,
		team.Name, team.ManagerID, team.ID,
	).Scan(&t.ID, &t.Name, &t.ManagerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("updating team: %w", err)
	}
	return &t, nil
}

func (r *teamRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM teams WHERE id = $1::UUID`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting team: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *teamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.external_id, u.email, u.name, u.role, u.avatar, u.online_status
		 FROM users u
		 JOIN team_members tm ON u.id = tm.user_id
		 WHERE tm.team_id = $1::UUID
		 ORDER BY u.name`,
		teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating team members: %w", err)
	}
	return users, nil
}
