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

type activityRepository struct {
	pool *pgxpool.Pool
}

const activityColumns = `
	id::TEXT, activity_type, status, due_date, duration,
	COALESCE(routing, ''), fields,
	COALESCE(target_id::TEXT, ''), creator_id::TEXT,
	COALESCE(joint_visit_user_id::TEXT, ''), COALESCE(team_id::TEXT, ''),
	submitted_at, created_at, updated_at, deleted_at`

func scanActivity(row pgx.Row) (*domain.Activity, error) {
	var a domain.Activity
	var fieldsJSON []byte
	err := row.Scan(
		&a.ID, &a.ActivityType, &a.Status, &a.DueDate, &a.Duration,
		&a.Routing, &fieldsJSON,
		&a.TargetID, &a.CreatorID,
		&a.JointVisitUID, &a.TeamID,
		&a.SubmittedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
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
	return &a, nil
}

func (r *activityRepository) Get(ctx context.Context, id string) (*domain.Activity, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+activityColumns+` FROM activities WHERE id = $1::UUID AND deleted_at IS NULL`,
		id,
	)
	return scanActivity(row)
}

func (r *activityRepository) List(ctx context.Context, scope rbac.ActivityScope, filter store.ActivityFilter, page, limit int) (*store.ActivityPage, error) {
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

	// Always exclude soft-deleted.
	conditions = append(conditions, "deleted_at IS NULL")

	// RBAC scope.
	if !scope.AllActivities {
		var scopeParts []string
		if len(scope.CreatorIDs) > 0 {
			placeholders := make([]string, len(scope.CreatorIDs))
			for i, id := range scope.CreatorIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("(creator_id::TEXT = ANY(ARRAY[%s]) OR joint_visit_user_id::TEXT = ANY(ARRAY[%s]))",
				strings.Join(placeholders, ","), strings.Join(placeholders, ",")))
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
			return &store.ActivityPage{Activities: []*domain.Activity{}, Total: 0, Page: page, Limit: limit}, nil
		}
	}

	// Filters.
	if filter.ActivityType != nil {
		conditions = append(conditions, fmt.Sprintf("activity_type = $%d", argIdx))
		args = append(args, *filter.ActivityType)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}
	if filter.TargetID != nil {
		conditions = append(conditions, fmt.Sprintf("target_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TargetID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM activities ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting activities: %w", err)
	}

	listQuery := `SELECT ` + activityColumns + ` FROM activities ` + where +
		fmt.Sprintf(` ORDER BY due_date DESC, created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
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
		`INSERT INTO activities (activity_type, status, due_date, duration, routing, fields,
		     target_id, creator_id, joint_visit_user_id, team_id, submitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6,
		     $7::UUID, $8::UUID, $9::UUID, $10::UUID, $11)
		 RETURNING `+activityColumns,
		a.ActivityType, a.Status, a.DueDate, a.Duration,
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
		`UPDATE activities
		 SET activity_type = $1, status = $2, due_date = $3, duration = $4,
		     routing = $5, fields = $6,
		     target_id = $7::UUID, joint_visit_user_id = $8::UUID,
		     team_id = $9::UUID, submitted_at = $10,
		     updated_at = NOW()
		 WHERE id = $11::UUID AND deleted_at IS NULL
		 RETURNING `+activityColumns,
		a.ActivityType, a.Status, a.DueDate, a.Duration,
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
