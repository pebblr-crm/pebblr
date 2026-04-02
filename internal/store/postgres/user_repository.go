package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type userRepository struct {
	pool dbPool
}

const (
	userColumns = `id, external_id, email, name, role, avatar, online_status`
	errScanUser = "scanning user: %w"
)

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.ExternalID, &u.Email, &u.Name, &u.Role, &u.Avatar, &u.OnlineStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf(errScanUser, err)
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1::UUID`,
		id,
	)
	return scanUser(row)
}

func (r *userRepository) GetByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE external_id = $1`,
		externalID,
	)
	return scanUser(row)
}

func (r *userRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.external_id, u.email, u.name, u.role, u.avatar, u.online_status,
		        COALESCE(array_agg(tm.team_id::TEXT) FILTER (WHERE tm.team_id IS NOT NULL), '{}')
		 FROM users u
		 LEFT JOIN team_members tm ON u.id = tm.user_id
		 GROUP BY u.id
		 ORDER BY u.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.ExternalID, &u.Email, &u.Name, &u.Role, &u.Avatar, &u.OnlineStatus, &u.TeamIDs)
		if err != nil {
			return nil, fmt.Errorf(errScanUser, err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating users: %w", err)
	}
	return users, nil
}

func (r *userRepository) ListPaginated(ctx context.Context, page, limit int) (*store.UserPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.external_id, u.email, u.name, u.role, u.avatar, u.online_status,
		        COALESCE(array_agg(tm.team_id::TEXT) FILTER (WHERE tm.team_id IS NOT NULL), '{}')
		 FROM users u
		 LEFT JOIN team_members tm ON u.id = tm.user_id
		 GROUP BY u.id
		 ORDER BY u.name
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.ExternalID, &u.Email, &u.Name, &u.Role, &u.Avatar, &u.OnlineStatus, &u.TeamIDs)
		if err != nil {
			return nil, fmt.Errorf(errScanUser, err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating users: %w", err)
	}

	return &store.UserPage{Users: users, Total: total, Page: page, Limit: limit}, nil
}

func (r *userRepository) Upsert(ctx context.Context, user *domain.User) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (external_id, email, name, role, avatar, online_status)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (external_id) DO UPDATE
		   SET email = EXCLUDED.email,
		       name  = EXCLUDED.name,
		       role  = EXCLUDED.role,
		       avatar = EXCLUDED.avatar,
		       online_status = EXCLUDED.online_status,
		       updated_at = NOW()
		 RETURNING `+userColumns,
		user.ExternalID, user.Email, user.Name, string(user.Role),
		user.Avatar, string(user.OnlineStatus),
	)
	return scanUser(row)
}
