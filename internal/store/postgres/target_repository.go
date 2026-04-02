package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// errFmtMarshalTargetFields is the error format for target field marshalling failures.
const errFmtMarshalTargetFields = "marshalling target fields: %w"

type targetRepository struct {
	pool dbPool
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

// targetQueryBuilder accumulates SQL conditions and positional arguments
// for building target list queries.
type targetQueryBuilder struct {
	conditions []string
	args       []any
	argIdx     int
}

func newTargetQueryBuilder() *targetQueryBuilder {
	return &targetQueryBuilder{argIdx: 1}
}

func (b *targetQueryBuilder) addCondition(sql string, val any) {
	b.conditions = append(b.conditions, fmt.Sprintf(sql, b.argIdx))
	b.args = append(b.args, val)
	b.argIdx++
}

func (b *targetQueryBuilder) whereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conditions, " AND ")
}

// applyScope applies RBAC scope conditions to the query builder.
// Returns false if the scope excludes all targets (empty result).
func (b *targetQueryBuilder) applyScope(scope rbac.TargetScope) bool {
	scopeSQL, outArgs, nextIdx := targetScopeConditionsAliased("", scope, b.args, b.argIdx)
	b.args = outArgs
	b.argIdx = nextIdx
	if scopeSQL != "" {
		b.conditions = append(b.conditions, scopeSQL)
		return true
	}
	return scope.AllTargets
}

// applyTargetFilter adds filter conditions to the query builder.
func (b *targetQueryBuilder) applyFilter(filter store.TargetFilter) {
	if filter.TargetType != nil {
		b.addCondition("target_type = $%d", *filter.TargetType)
	}
	if filter.AssigneeID != nil {
		b.addCondition("assignee_id::TEXT = $%d", *filter.AssigneeID)
	}
	if filter.TeamID != nil {
		b.addCondition("team_id::TEXT = $%d", *filter.TeamID)
	}
	if filter.Query != nil {
		b.addCondition("name ILIKE $%d", "%"+*filter.Query+"%")
	}
}

func (r *targetRepository) List(ctx context.Context, scope rbac.TargetScope, filter store.TargetFilter, page, limit int) (*store.TargetPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	qb := newTargetQueryBuilder()

	if !qb.applyScope(scope) {
		return &store.TargetPage{Targets: []*domain.Target{}, Total: 0, Page: page, Limit: limit}, nil
	}

	qb.applyFilter(filter)

	where := qb.whereClause()

	countQuery := `SELECT COUNT(*) FROM targets` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, qb.args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting targets: %w", err)
	}

	listQuery := `SELECT ` + targetColumns + ` FROM targets` + where +
		fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, qb.argIdx, qb.argIdx+1)
	qb.args = append(qb.args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, qb.args...)
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
		return nil, fmt.Errorf(errFmtMarshalTargetFields, err)
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
		return nil, fmt.Errorf(errFmtMarshalTargetFields, err)
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
			return nil, fmt.Errorf(errFmtMarshalTargetFields, err)
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

// targetScopeResult holds the output of building a target RBAC scope clause.
type targetScopeResult struct {
	conditions []string
	args       []any
	argIdx     int
	empty      bool // true if the scope excludes all targets
}

// buildTargetScopeConditions generates SQL conditions and args for a target RBAC scope.
// The prefix (e.g. "t.") is prepended to column names. It delegates to the shared
// targetScopeConditionsAliased function to avoid duplicating scope-building logic.
func buildTargetScopeConditions(scope rbac.TargetScope, prefix string, startArgIdx int) targetScopeResult {
	// Convert prefix "t." to alias "t" (strip trailing dot).
	alias := strings.TrimSuffix(prefix, ".")

	scopeSQL, outArgs, nextIdx := targetScopeConditionsAliased(alias, scope, nil, startArgIdx)

	result := targetScopeResult{
		args:   outArgs,
		argIdx: nextIdx,
	}
	if scopeSQL != "" {
		result.conditions = []string{scopeSQL}
	} else if !scope.AllTargets {
		result.empty = true
	}
	return result
}

func (r *targetRepository) VisitStatus(ctx context.Context, scope rbac.TargetScope, fieldTypes []string) ([]store.TargetVisitStatus, error) {
	if len(fieldTypes) == 0 {
		return nil, nil
	}

	sr := buildTargetScopeConditions(scope, "t.", 1)
	if sr.empty {
		return nil, nil
	}

	args := sr.args
	argIdx := sr.argIdx

	args = append(args, fieldTypes)
	typesArg := fmt.Sprintf("$%d", argIdx)

	where := ""
	if len(sr.conditions) > 0 {
		where = " AND (" + strings.Join(sr.conditions, " OR ") + ")"
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating visit status: %w", err)
	}
	return result, nil
}

func (r *targetRepository) FrequencyStatus(ctx context.Context, scope rbac.TargetScope, fieldTypes []string, dateFrom, dateTo time.Time) ([]store.TargetFrequencyStatus, error) {
	if len(fieldTypes) == 0 {
		return nil, nil
	}

	sr := buildTargetScopeConditions(scope, "t.", 1)
	if sr.empty {
		return nil, nil
	}

	conditions := sr.conditions
	args := sr.args
	argIdx := sr.argIdx

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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating frequency status: %w", err)
	}
	return result, nil
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
