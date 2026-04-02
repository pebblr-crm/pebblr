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

type activityRepository struct {
	pool dbPool
}

const activityColumns = `
	a.id::TEXT, a.activity_type, COALESCE(a.label, ''), a.status, a.due_date, a.duration,
	COALESCE(a.routing, ''), a.fields,
	COALESCE(a.target_id::TEXT, ''), COALESCE(t.name, ''), a.creator_id::TEXT,
	COALESCE(a.joint_visit_user_id::TEXT, ''), COALESCE(a.team_id::TEXT, ''),
	a.submitted_at, a.created_at, a.updated_at, a.deleted_at,
	COALESCE(t.target_type, ''), COALESCE(t.fields, '{}'::JSONB)`

const activityFrom = ` FROM activities a LEFT JOIN targets t ON a.target_id = t.id`

func scanActivity(row pgx.Row) (*domain.Activity, error) {
	var a domain.Activity
	var fieldsJSON []byte
	var targetType string
	var targetFieldsJSON []byte
	err := row.Scan(
		&a.ID, &a.ActivityType, &a.Label, &a.Status, &a.DueDate, &a.Duration,
		&a.Routing, &fieldsJSON,
		&a.TargetID, &a.TargetName, &a.CreatorID,
		&a.JointVisitUID, &a.TeamID,
		&a.SubmittedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		&targetType, &targetFieldsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("scanning activity: %w", err)
	}
	a.Fields = make(map[string]any)
	if len(fieldsJSON) > 0 {
		if err := json.Unmarshal(fieldsJSON, &a.Fields); err != nil {
			return nil, fmt.Errorf("unmarshalling activity fields: %w", err)
		}
	}
	// Build embedded TargetSummary when the activity has a linked target.
	if a.TargetID != "" {
		ts := &domain.TargetSummary{
			ID:         a.TargetID,
			TargetType: targetType,
			Name:       a.TargetName,
			Fields:     make(map[string]any),
		}
		if len(targetFieldsJSON) > 0 {
			if err := json.Unmarshal(targetFieldsJSON, &ts.Fields); err != nil {
				return nil, fmt.Errorf("unmarshalling target fields: %w", err)
			}
		}
		a.TargetSummary = ts
	}
	return &a, nil
}

func (r *activityRepository) Get(ctx context.Context, id string) (*domain.Activity, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+activityColumns+activityFrom+` WHERE a.id = $1::UUID AND a.deleted_at IS NULL`,
		id,
	)
	return scanActivity(row)
}

// activityQueryBuilder accumulates SQL conditions and positional arguments
// for building activity list queries.
type activityQueryBuilder struct {
	conditions []string
	args       []any
	argIdx     int
}

func newActivityQueryBuilder() *activityQueryBuilder {
	return &activityQueryBuilder{argIdx: 1}
}

func (b *activityQueryBuilder) addCondition(sql string, val any) {
	b.conditions = append(b.conditions, fmt.Sprintf(sql, b.argIdx))
	b.args = append(b.args, val)
	b.argIdx++
}

func (b *activityQueryBuilder) whereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conditions, " AND ")
}

// buildActivityScopeConditions applies RBAC scope conditions to the query builder.
// Returns false if the scope excludes all activities (empty result).
func (b *activityQueryBuilder) applyScope(scope rbac.ActivityScope) bool {
	if scope.DenyAll {
		return false
	}
	if scope.AllActivities {
		return true
	}

	var scopeParts []string
	if len(scope.CreatorIDs) > 0 {
		placeholders := make([]string, len(scope.CreatorIDs))
		for i, id := range scope.CreatorIDs {
			placeholders[i] = fmt.Sprintf("$%d", b.argIdx)
			b.args = append(b.args, id)
			b.argIdx++
		}
		joined := strings.Join(placeholders, ",")
		scopeParts = append(scopeParts, fmt.Sprintf("(a.creator_id::TEXT = ANY(ARRAY[%s]) OR a.joint_visit_user_id::TEXT = ANY(ARRAY[%s]))",
			joined, joined))
	}
	if len(scope.TeamIDs) > 0 {
		placeholders := make([]string, len(scope.TeamIDs))
		for i, id := range scope.TeamIDs {
			placeholders[i] = fmt.Sprintf("$%d", b.argIdx)
			b.args = append(b.args, id)
			b.argIdx++
		}
		scopeParts = append(scopeParts, fmt.Sprintf("a.team_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
	}
	if len(scopeParts) == 0 {
		return false
	}
	b.conditions = append(b.conditions, "("+strings.Join(scopeParts, " OR ")+")")
	return true
}

// applyActivityFilter adds filter conditions to the query builder.
func (b *activityQueryBuilder) applyFilter(filter store.ActivityFilter) {
	if filter.ActivityType != nil {
		b.addCondition("a.activity_type = $%d", *filter.ActivityType)
	}
	if filter.Status != nil {
		b.addCondition("a.status = $%d", *filter.Status)
	}
	if filter.CreatorID != nil {
		b.addCondition("a.creator_id::TEXT = $%d", *filter.CreatorID)
	}
	if filter.TargetID != nil {
		b.addCondition("a.target_id::TEXT = $%d", *filter.TargetID)
	}
	if filter.TeamID != nil {
		b.addCondition("a.team_id::TEXT = $%d", *filter.TeamID)
	}
	if filter.DateFrom != nil {
		b.addCondition("a.due_date >= $%d", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		b.addCondition("a.due_date <= $%d", *filter.DateTo)
	}
}

func (r *activityRepository) List(ctx context.Context, scope rbac.ActivityScope, filter store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	qb := newActivityQueryBuilder()
	qb.conditions = append(qb.conditions, "a.deleted_at IS NULL")

	if !qb.applyScope(scope) {
		return &store.ActivityPage{Activities: []*domain.Activity{}, Total: 0, Page: page, Limit: limit}, nil
	}

	qb.applyFilter(filter)

	where := qb.whereClause()

	countQuery := `SELECT COUNT(*)` + activityFrom + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, qb.args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting activities: %w", err)
	}

	listQuery := `SELECT ` + activityColumns + activityFrom + where +
		fmt.Sprintf(` ORDER BY a.due_date DESC, a.created_at DESC LIMIT $%d OFFSET $%d`, qb.argIdx, qb.argIdx+1)
	qb.args = append(qb.args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("querying activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.Activity
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating activities: %w", err)
	}

	return &store.ActivityPage{Activities: activities, Total: total, Page: page, Limit: limit}, nil
}

func (r *activityRepository) Create(ctx context.Context, a *domain.Activity) (*domain.Activity, error) {
	fieldsJSON, err := json.Marshal(a.Fields)
	if err != nil {
		return nil, fmt.Errorf("marshalling activity fields: %w", err)
	}

	row := r.pool.QueryRow(ctx,
		`WITH ins AS (
		   INSERT INTO activities (activity_type, label, status, due_date, duration, routing, fields,
		       target_id, creator_id, joint_visit_user_id, team_id, submitted_at)
		   VALUES ($1, $2, $3, $4, $5, $6, $7,
		       $8::UUID, $9::UUID, $10::UUID, $11::UUID, $12)
		   RETURNING *
		 )
		 SELECT `+activityColumns+` FROM ins a LEFT JOIN targets t ON a.target_id = t.id`,
		a.ActivityType, a.Label, a.Status, a.DueDate, a.Duration,
		nullIfEmpty(a.Routing), fieldsJSON,
		nullIfEmpty(a.TargetID), a.CreatorID,
		nullIfEmpty(a.JointVisitUID), nullIfEmpty(a.TeamID),
		a.SubmittedAt,
	)
	return scanActivity(row)
}

func (r *activityRepository) Update(ctx context.Context, a *domain.Activity) (*domain.Activity, error) {
	fieldsJSON, err := json.Marshal(a.Fields)
	if err != nil {
		return nil, fmt.Errorf("marshalling activity fields: %w", err)
	}

	row := r.pool.QueryRow(ctx,
		`WITH upd AS (
		   UPDATE activities
		   SET activity_type = $1, label = $2, status = $3, due_date = $4, duration = $5,
		       routing = $6, fields = $7,
		       target_id = $8::UUID, joint_visit_user_id = $9::UUID,
		       team_id = $10::UUID, submitted_at = $11,
		       updated_at = NOW()
		   WHERE id = $12::UUID AND deleted_at IS NULL
		   RETURNING *
		 )
		 SELECT `+activityColumns+` FROM upd a LEFT JOIN targets t ON a.target_id = t.id`,
		a.ActivityType, a.Label, a.Status, a.DueDate, a.Duration,
		nullIfEmpty(a.Routing), fieldsJSON,
		nullIfEmpty(a.TargetID), nullIfEmpty(a.JointVisitUID),
		nullIfEmpty(a.TeamID), a.SubmittedAt,
		a.ID,
	)
	return scanActivity(row)
}

func (r *activityRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE activities SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1::UUID AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting activity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (r *activityRepository) CountByDate(ctx context.Context, creatorID string, date time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM activities
		 WHERE creator_id = $1::UUID AND due_date = $2 AND deleted_at IS NULL`,
		creatorID, date,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting activities by date: %w", err)
	}
	return count, nil
}

func (r *activityRepository) HasActivityWithTypes(ctx context.Context, creatorID string, date time.Time, types []string) (bool, error) {
	if len(types) == 0 {
		return false, nil
	}
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM activities
			WHERE creator_id = $1::UUID AND due_date = $2 AND activity_type = ANY($3) AND deleted_at IS NULL
		)`,
		creatorID, date, types,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking activities by type: %w", err)
	}
	return exists, nil
}

func (r *activityRepository) ExistsForTargetOnDate(ctx context.Context, creatorID, targetID string, date time.Time) (bool, error) {
	if targetID == "" {
		return false, nil
	}
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM activities
			WHERE creator_id = $1::UUID AND target_id = $2::UUID AND due_date = $3 AND deleted_at IS NULL
		)`,
		creatorID, targetID, date,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking duplicate activity: %w", err)
	}
	return exists, nil
}
