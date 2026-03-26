package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type territoryRepository struct {
	pool dbPool
}

func (r *territoryRepository) Get(ctx context.Context, id string) (*domain.Territory, error) {
	var t domain.Territory
	var boundaryJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id::TEXT, name, team_id::TEXT, region, boundary, created_at, updated_at
		 FROM territories WHERE id = $1::UUID`, id,
	).Scan(&t.ID, &t.Name, &t.TeamID, &t.Region, &boundaryJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("getting territory: %w", err)
	}
	if len(boundaryJSON) > 0 {
		t.Boundary = make(map[string]any)
		if err := json.Unmarshal(boundaryJSON, &t.Boundary); err != nil {
			return nil, fmt.Errorf("unmarshalling territory boundary: %w", err)
		}
	}
	return &t, nil
}

func (r *territoryRepository) List(ctx context.Context, filter store.TerritoryFilter) ([]*domain.Territory, error) {
	query := `SELECT id::TEXT, name, team_id::TEXT, region, boundary, created_at, updated_at
	          FROM territories WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d::UUID", argIdx)
		args = append(args, *filter.TeamID)
		argIdx++
	}
	if filter.Region != nil {
		query += fmt.Sprintf(" AND region = $%d", argIdx)
		args = append(args, *filter.Region)
		argIdx++
	}
	_ = argIdx

	query += " ORDER BY name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing territories: %w", err)
	}
	defer rows.Close()

	var territories []*domain.Territory
	for rows.Next() {
		var t domain.Territory
		var boundaryJSON []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.TeamID, &t.Region, &boundaryJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning territory: %w", err)
		}
		if len(boundaryJSON) > 0 {
			t.Boundary = make(map[string]any)
			if err := json.Unmarshal(boundaryJSON, &t.Boundary); err != nil {
				return nil, fmt.Errorf("unmarshalling territory boundary: %w", err)
			}
		}
		territories = append(territories, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating territories: %w", err)
	}
	return territories, nil
}

func (r *territoryRepository) Create(ctx context.Context, t *domain.Territory) (*domain.Territory, error) {
	boundaryJSON, err := marshalJSONField(t.Boundary)
	if err != nil {
		return nil, fmt.Errorf("marshalling territory boundary: %w", err)
	}

	var created domain.Territory
	var retBoundary []byte
	err = r.pool.QueryRow(ctx,
		`INSERT INTO territories (name, team_id, region, boundary)
		 VALUES ($1, $2::UUID, $3, $4)
		 RETURNING id::TEXT, name, team_id::TEXT, region, boundary, created_at, updated_at`,
		t.Name, t.TeamID, t.Region, boundaryJSON,
	).Scan(&created.ID, &created.Name, &created.TeamID, &created.Region, &retBoundary, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating territory: %w", err)
	}
	if len(retBoundary) > 0 {
		created.Boundary = make(map[string]any)
		if err := json.Unmarshal(retBoundary, &created.Boundary); err != nil {
			return nil, fmt.Errorf("unmarshalling created territory boundary: %w", err)
		}
	}
	return &created, nil
}

func (r *territoryRepository) Update(ctx context.Context, t *domain.Territory) (*domain.Territory, error) {
	boundaryJSON, err := marshalJSONField(t.Boundary)
	if err != nil {
		return nil, fmt.Errorf("marshalling territory boundary: %w", err)
	}

	var updated domain.Territory
	var retBoundary []byte
	err = r.pool.QueryRow(ctx,
		`UPDATE territories SET name = $1, team_id = $2::UUID, region = $3, boundary = $4, updated_at = NOW()
		 WHERE id = $5::UUID
		 RETURNING id::TEXT, name, team_id::TEXT, region, boundary, created_at, updated_at`,
		t.Name, t.TeamID, t.Region, boundaryJSON, t.ID,
	).Scan(&updated.ID, &updated.Name, &updated.TeamID, &updated.Region, &retBoundary, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("updating territory: %w", err)
	}
	if len(retBoundary) > 0 {
		updated.Boundary = make(map[string]any)
		if err := json.Unmarshal(retBoundary, &updated.Boundary); err != nil {
			return nil, fmt.Errorf("unmarshalling updated territory boundary: %w", err)
		}
	}
	return &updated, nil
}

func (r *territoryRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM territories WHERE id = $1::UUID`, id)
	if err != nil {
		return fmt.Errorf("deleting territory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// marshalJSONField marshals a map to JSON, returning nil for nil/empty maps.
func marshalJSONField(m map[string]any) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
