package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type collectionRepository struct {
	pool dbPool
}

func (r *collectionRepository) Create(ctx context.Context, c *domain.Collection) (*domain.Collection, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	row := tx.QueryRow(ctx,
		`INSERT INTO target_collections (name, creator_id, team_id)
		 VALUES ($1, $2::UUID, $3::UUID)
		 RETURNING id::TEXT, name, creator_id::TEXT, team_id::TEXT, created_at, updated_at`,
		c.Name, c.CreatorID, c.TeamID,
	)

	var out domain.Collection
	if err := row.Scan(&out.ID, &out.Name, &out.CreatorID, &out.TeamID, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, fmt.Errorf("inserting collection: %w", err)
	}

	if err := r.replaceItems(ctx, tx, out.ID, c.TargetIDs); err != nil {
		return nil, err
	}
	out.TargetIDs = c.TargetIDs

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return &out, nil
}

func (r *collectionRepository) List(ctx context.Context, filter store.CollectionFilter) ([]*domain.Collection, error) {
	args := []any{}
	argIdx := 1
	var conditions []string

	if filter.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("c.creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("c.team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := `SELECT c.id::TEXT, c.name, c.creator_id::TEXT, c.team_id::TEXT,
	                 c.created_at, c.updated_at,
	                 COALESCE(array_agg(ci.target_id::TEXT) FILTER (WHERE ci.target_id IS NOT NULL), '{}')
	          FROM target_collections c
	          LEFT JOIN target_collection_items ci ON ci.collection_id = c.id
	          ` + where + `
	          GROUP BY c.id
	          ORDER BY c.name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing collections: %w", err)
	}
	defer rows.Close()

	var result []*domain.Collection
	for rows.Next() {
		var c domain.Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatorID, &c.TeamID, &c.CreatedAt, &c.UpdatedAt, &c.TargetIDs); err != nil {
			return nil, fmt.Errorf("scanning collection: %w", err)
		}
		if c.TargetIDs == nil {
			c.TargetIDs = []string{}
		}
		result = append(result, &c)
	}
	return result, rows.Err()
}

func (r *collectionRepository) Get(ctx context.Context, id string) (*domain.Collection, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT c.id::TEXT, c.name, c.creator_id::TEXT, c.team_id::TEXT, c.created_at, c.updated_at,
		        COALESCE(array_agg(ci.target_id::TEXT) FILTER (WHERE ci.target_id IS NOT NULL), '{}')
		 FROM target_collections c
		 LEFT JOIN target_collection_items ci ON ci.collection_id = c.id
		 WHERE c.id = $1::UUID
		 GROUP BY c.id`,
		id,
	)

	var c domain.Collection
	if err := row.Scan(&c.ID, &c.Name, &c.CreatorID, &c.TeamID, &c.CreatedAt, &c.UpdatedAt, &c.TargetIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("getting collection: %w", err)
	}
	if c.TargetIDs == nil {
		c.TargetIDs = []string{}
	}
	return &c, nil
}

func (r *collectionRepository) Update(ctx context.Context, c *domain.Collection) (*domain.Collection, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	row := tx.QueryRow(ctx,
		`UPDATE target_collections SET name = $1, updated_at = NOW()
		 WHERE id = $2::UUID
		 RETURNING id::TEXT, name, creator_id::TEXT, team_id::TEXT, created_at, updated_at`,
		c.Name, c.ID,
	)

	var out domain.Collection
	if err := row.Scan(&out.ID, &out.Name, &out.CreatorID, &out.TeamID, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("updating collection: %w", err)
	}

	if c.TargetIDs != nil {
		if err := r.replaceItems(ctx, tx, out.ID, c.TargetIDs); err != nil {
			return nil, err
		}
		out.TargetIDs = c.TargetIDs
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return &out, nil
}

func (r *collectionRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM target_collections WHERE id = $1::UUID`, id)
	if err != nil {
		return fmt.Errorf("deleting collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// replaceItems deletes all existing items and inserts the new set within the given transaction.
func (r *collectionRepository) replaceItems(ctx context.Context, tx pgx.Tx, collectionID string, targetIDs []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM target_collection_items WHERE collection_id = $1::UUID`, collectionID); err != nil {
		return fmt.Errorf("clearing collection items: %w", err)
	}

	if len(targetIDs) == 0 {
		return nil
	}

	// Batch insert.
	args := []any{collectionID}
	vals := make([]string, len(targetIDs))
	for i, tid := range targetIDs {
		args = append(args, tid)
		vals[i] = fmt.Sprintf("($1::UUID, $%d::UUID)", i+2)
	}

	query := `INSERT INTO target_collection_items (collection_id, target_id) VALUES ` + strings.Join(vals, ", ")
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("inserting collection items: %w", err)
	}
	return nil
}
