package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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
	id::TEXT, COALESCE(external_id, ''), target_type, name, fields,
	COALESCE(assignee_id::TEXT, ''), COALESCE(team_id::TEXT, ''),
	imported_at, created_at, updated_at`

func scanTarget(row pgx.Row) (*domain.Target, error) {
	var t domain.Target
	var fieldsJSON []byte
	err := row.Scan(
		&t.ID, &t.ExternalID, &t.TargetType, &t.Name, &fieldsJSON,
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
		`INSERT INTO targets (external_id, target_type, name, fields, assignee_id, team_id, imported_at)
		 VALUES ($1, $2, $3, $4, $5::UUID, $6::UUID, $7)
		 RETURNING `+targetColumns,
		nullIfEmpty(t.ExternalID), t.TargetType, t.Name, fieldsJSON,
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

func (r *targetRepository) Upsert(ctx context.Context, targets []*domain.Target) (*store.ImportResult, error) {
	if len(targets) == 0 {
		return &store.ImportResult{}, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning import transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a no-op

	result := &store.ImportResult{}

	for _, t := range targets {
		fieldsJSON, err := json.Marshal(t.Fields)
		if err != nil {
			return nil, fmt.Errorf("marshalling target fields: %w", err)
		}

		var isNew bool
		row := tx.QueryRow(ctx,
			`INSERT INTO targets (external_id, target_type, name, fields, assignee_id, team_id, imported_at)
			 VALUES ($1, $2, $3, $4, $5::UUID, $6::UUID, NOW())
			 ON CONFLICT (target_type, external_id) WHERE external_id IS NOT NULL
			 DO UPDATE SET name = EXCLUDED.name, fields = EXCLUDED.fields,
			     assignee_id = EXCLUDED.assignee_id, team_id = EXCLUDED.team_id,
			     imported_at = NOW(), updated_at = NOW()
			 RETURNING `+targetColumns+`,
			     (xmax = 0) AS is_new`,
			nullIfEmpty(t.ExternalID), t.TargetType, t.Name, fieldsJSON,
			nullIfEmpty(t.AssigneeID), nullIfEmpty(t.TeamID),
		)

		imported, err := scanTargetWithFlag(row, &isNew)
		if err != nil {
			return nil, fmt.Errorf("upserting target %q: %w", t.ExternalID, err)
		}

		result.Imported = append(result.Imported, imported)
		if isNew {
			result.Created++
		} else {
			result.Updated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing import transaction: %w", err)
	}

	return result, nil
}

func (r *targetRepository) VisitStatus(ctx context.Context, scope rbac.TargetScope, fieldTypes []string) ([]store.TargetVisitStatus, error) {
	if len(fieldTypes) == 0 {
		return nil, nil
	}

	args := []any{}
	argIdx := 1
	var conditions []string

	// Scope targets.
	if !scope.AllTargets {
		if len(scope.AssigneeIDs) > 0 {
			phs := make([]string, len(scope.AssigneeIDs))
			for i, id := range scope.AssigneeIDs {
				phs[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("t.assignee_id::TEXT = ANY(ARRAY[%s])", strings.Join(phs, ",")))
		}
		if len(scope.TeamIDs) > 0 {
			phs := make([]string, len(scope.TeamIDs))
			for i, id := range scope.TeamIDs {
				phs[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("t.team_id::TEXT = ANY(ARRAY[%s])", strings.Join(phs, ",")))
		}
		if len(conditions) == 0 {
			return nil, nil
		}
	}

	// Field activity types filter.
	args = append(args, fieldTypes)
	typesArg := fmt.Sprintf("$%d", argIdx)

	where := ""
	if len(conditions) > 0 {
		where = " AND (" + strings.Join(conditions, " OR ") + ")"
	}

	query := `SELECT t.id::TEXT, MAX(a.due_date)
		FROM targets t
		JOIN activities a ON a.target_id = t.id AND a.deleted_at IS NULL AND a.activity_type = ANY(` + typesArg + `)
		WHERE TRUE` + where + `
		GROUP BY t.id`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying visit status: %w", err)
	}
	defer rows.Close()

	var result []store.TargetVisitStatus
	for rows.Next() {
		var vs store.TargetVisitStatus
		if err := rows.Scan(&vs.TargetID, &vs.LastVisitDate); err != nil {
			return nil, fmt.Errorf("scanning visit status: %w", err)
		}
		result = append(result, vs)
	}
	return result, rows.Err()
}

func (r *targetRepository) FrequencyStatus(ctx context.Context, scope rbac.TargetScope, fieldTypes []string, dateFrom, dateTo time.Time) ([]store.TargetFrequencyStatus, error) {
	if len(fieldTypes) == 0 {
		return nil, nil
	}

	args := []any{}
	argIdx := 1
	var conditions []string

	// Scope targets.
	if !scope.AllTargets {
		if len(scope.AssigneeIDs) > 0 {
			phs := make([]string, len(scope.AssigneeIDs))
			for i, id := range scope.AssigneeIDs {
				phs[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("t.assignee_id::TEXT = ANY(ARRAY[%s])", strings.Join(phs, ",")))
		}
		if len(scope.TeamIDs) > 0 {
			phs := make([]string, len(scope.TeamIDs))
			for i, id := range scope.TeamIDs {
				phs[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("t.team_id::TEXT = ANY(ARRAY[%s])", strings.Join(phs, ",")))
		}
		if len(conditions) == 0 {
			return nil, nil
		}
	}

	// Only targets with a classification.
	conditions = append(conditions, "t.fields->>'potential' IS NOT NULL", "t.fields->>'potential' != ''")

	// Field activity types and date range for the LEFT JOIN.
	args = append(args, fieldTypes)
	typesArg := fmt.Sprintf("$%d", argIdx)
	argIdx++

	args = append(args, dateFrom)
	dateFromArg := fmt.Sprintf("$%d", argIdx)
	argIdx++

	args = append(args, dateTo)
	dateToArg := fmt.Sprintf("$%d", argIdx)

	where := ""
	if len(conditions) > 0 {
		where = " AND (" + strings.Join(conditions, " AND ") + ")"
	}

	query := `SELECT t.id::TEXT, t.fields->>'potential', COUNT(a.id)
		FROM targets t
		LEFT JOIN activities a ON a.target_id = t.id
			AND a.deleted_at IS NULL
			AND a.activity_type = ANY(` + typesArg + `)
			AND a.due_date >= ` + dateFromArg + `
			AND a.due_date <= ` + dateToArg + `
		WHERE TRUE` + where + `
		GROUP BY t.id, t.fields->>'potential'`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying frequency status: %w", err)
	}
	defer rows.Close()

	var result []store.TargetFrequencyStatus
	for rows.Next() {
		var fs store.TargetFrequencyStatus
		if err := rows.Scan(&fs.TargetID, &fs.Classification, &fs.VisitCount); err != nil {
			return nil, fmt.Errorf("scanning frequency status: %w", err)
		}
		result = append(result, fs)
	}
	return result, rows.Err()
}

func scanTargetWithFlag(row pgx.Row, flag *bool) (*domain.Target, error) {
	var t domain.Target
	var fieldsJSON []byte
	err := row.Scan(
		&t.ID, &t.ExternalID, &t.TargetType, &t.Name, &fieldsJSON,
		&t.AssigneeID, &t.TeamID,
		&t.ImportedAt, &t.CreatedAt, &t.UpdatedAt,
		flag,
	)
	if err != nil {
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
