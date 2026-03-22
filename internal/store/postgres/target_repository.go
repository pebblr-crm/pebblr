package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

type targetRepository struct {
	pool *pgxpool.Pool
}

const targetColumns = `
	id::TEXT, target_type, name, fields,
	COALESCE(assignee_id::TEXT, ''), COALESCE(team_id::TEXT, ''),
	imported_at, created_at, updated_at`

func scanTarget(row pgx.Row) (*domain.Target, error) {
	var t domain.Target
	var fieldsJSON []byte
	err := row.Scan(
		&t.ID, &t.TargetType, &t.Name, &fieldsJSON,
		&t.AssigneeID, &t.TeamID,
		&t.ImportedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("scanning target: %w", err)
	}
	t.Fields = make(map[string]any)
	if len(fieldsJSON) > 0 {
		if err := json.Unmarshal(fieldsJSON, &t.Fields); err != nil {
			return nil, fmt.Errorf("unmarshalling target fields: %w", err)
		}
	}
	return &t, nil
}

func (r *targetRepository) Get(ctx context.Context, id string) (*domain.Target, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+targetColumns+` FROM targets WHERE id = $1::UUID`,
		id,
	)
	return scanTarget(row)
}

func (r *targetRepository) List(ctx context.Context, scope rbac.TargetScope, filter store.TargetFilter, page, limit int) (*store.TargetPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{}
	argIdx := 1
	var conditions []string

	// RBAC scope
	if !scope.AllTargets {
		var scopeParts []string
		if len(scope.AssigneeIDs) > 0 {
			placeholders := make([]string, len(scope.AssigneeIDs))
			for i, id := range scope.AssigneeIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("assignee_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		}
		if len(scope.TeamIDs) > 0 {
			placeholders := make([]string, len(scope.TeamIDs))
			for i, id := range scope.TeamIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("team_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		}
		if len(scopeParts) > 0 {
			conditions = append(conditions, "("+strings.Join(scopeParts, " OR ")+")")
		} else {
			return &store.TargetPage{Targets: []*domain.Target{}, Total: 0, Page: page, Limit: limit}, nil
		}
	}

	// Filters
	if filter.TargetType != nil {
		conditions = append(conditions, fmt.Sprintf("target_type = $%d", argIdx))
		args = append(args, *filter.TargetType)
		argIdx++
	}
	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id::TEXT = $%d", argIdx))
		args = append(args, *filter.AssigneeID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
		argIdx++
	}
	if filter.Query != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Query+"%")
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM targets ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting targets: %w", err)
	}

	listQuery := `SELECT ` + targetColumns + ` FROM targets ` + where +
		fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying targets: %w", err)
	}
	defer rows.Close()

	var targets []*domain.Target
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating targets: %w", err)
	}

	return &store.TargetPage{Targets: targets, Total: total, Page: page, Limit: limit}, nil
}

func (r *targetRepository) Create(ctx context.Context, t *domain.Target) (*domain.Target, error) {
	fieldsJSON, err := json.Marshal(t.Fields)
	if err != nil {
		return nil, fmt.Errorf("marshalling target fields: %w", err)
	}

	row := r.pool.QueryRow(ctx,
		`INSERT INTO targets (target_type, name, fields, assignee_id, team_id, imported_at)
		 VALUES ($1, $2, $3, $4::UUID, $5::UUID, $6)
		 RETURNING `+targetColumns,
		t.TargetType, t.Name, fieldsJSON,
		nullIfEmpty(t.AssigneeID), nullIfEmpty(t.TeamID), t.ImportedAt,
	)
	return scanTarget(row)
}

func (r *targetRepository) Update(ctx context.Context, t *domain.Target) (*domain.Target, error) {
	fieldsJSON, err := json.Marshal(t.Fields)
	if err != nil {
		return nil, fmt.Errorf("marshalling target fields: %w", err)
	}

	row := r.pool.QueryRow(ctx,
		`UPDATE targets
		 SET target_type = $1, name = $2, fields = $3,
		     assignee_id = $4::UUID, team_id = $5::UUID,
		     updated_at = NOW()
		 WHERE id = $6::UUID
		 RETURNING `+targetColumns,
		t.TargetType, t.Name, fieldsJSON,
		nullIfEmpty(t.AssigneeID), nullIfEmpty(t.TeamID),
		t.ID,
	)
	return scanTarget(row)
}
