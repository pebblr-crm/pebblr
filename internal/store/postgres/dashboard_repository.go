package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

const condDeletedNull = "deleted_at IS NULL"

type dashboardRepository struct {
	pool dbPool
}

// ActivityStats returns activity counts grouped by status and category for the given scope and filter.
func (r *dashboardRepository) ActivityStats(ctx context.Context, scope rbac.ActivityScope, filter store.DashboardFilter) (*store.ActivityStats, error) {
	qb := newQueryBuilder()
	qb.addRaw(condDeletedNull)

	if !qb.applyActivityScope("", scope) {
		return &store.ActivityStats{ByStatus: map[string]int{}, ByCategory: map[string]int{}}, nil
	}

	qb.applyDashboardFilter("", filter)
	where := qb.where()

	// By status
	byStatus := map[string]int{}
	statusQuery := `SELECT status, COUNT(*) FROM activities ` + where + ` GROUP BY status`
	statusRows, err := r.pool.Query(ctx, statusQuery, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by status: %w", err)
	}
	defer statusRows.Close()

	total := 0
	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scanning status row: %w", err)
		}
		byStatus[status] = count
		total += count
	}
	if err := statusRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating status rows: %w", err)
	}

	// By category (activity_type -> category mapping done at service layer, here just group by activity_type)
	byCategory := map[string]int{}
	typeQuery := `SELECT activity_type, COUNT(*) FROM activities ` + where + ` GROUP BY activity_type`
	typeRows, err := r.pool.Query(ctx, typeQuery, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var activityType string
		var count int
		if err := typeRows.Scan(&activityType, &count); err != nil {
			return nil, fmt.Errorf("scanning type row: %w", err)
		}
		byCategory[activityType] = count
	}
	if err := typeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating type rows: %w", err)
	}

	return &store.ActivityStats{ByStatus: byStatus, ByCategory: byCategory, Total: total}, nil
}

// CoverageStats returns the total vs visited target counts.
func (r *dashboardRepository) CoverageStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter store.DashboardFilter) (*store.CoverageStats, error) {
	// Count total targets in scope.
	tqb := newQueryBuilder()

	if !tqb.applyTargetScope("", targetScope) {
		return &store.CoverageStats{}, nil
	}

	if filter.UserID != nil {
		tqb.add("assignee_id::TEXT = $%d", *filter.UserID)
	}
	if filter.TeamID != nil {
		tqb.add("team_id::TEXT = $%d", *filter.TeamID)
	}

	tWhere := tqb.where()

	var totalTargets int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM targets `+tWhere, tqb.args...).Scan(&totalTargets); err != nil {
		return nil, fmt.Errorf("counting targets: %w", err)
	}

	// Count distinct targets visited (field activities with a target in the period).
	aqb := newQueryBuilder()
	aqb.addRaw("a.deleted_at IS NULL")
	aqb.addRaw("a.target_id IS NOT NULL")

	if !aqb.applyActivityScope("a.", scope) {
		return &store.CoverageStats{TotalTargets: totalTargets}, nil
	}

	aqb.add("a.due_date >= $%d", filter.DateFrom)
	aqb.add("a.due_date <= $%d", filter.DateTo)

	if filter.UserID != nil {
		aqb.add("a.creator_id::TEXT = $%d", *filter.UserID)
	}
	if filter.TeamID != nil {
		aqb.add("a.team_id::TEXT = $%d", *filter.TeamID)
	}

	aWhere := aqb.where()

	var visitedTargets int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT a.target_id) FROM activities a `+aWhere,
		aqb.args...,
	).Scan(&visitedTargets); err != nil {
		return nil, fmt.Errorf("counting visited targets: %w", err)
	}

	return &store.CoverageStats{TotalTargets: totalTargets, VisitedTargets: visitedTargets}, nil
}

// FrequencyStats returns per-classification visit counts.
func (r *dashboardRepository) FrequencyStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter store.DashboardFilter) ([]store.FrequencyRow, error) {
	// Use a single queryBuilder for the whole query so arg indices are
	// consistent across both the target WHERE and the activity JOIN ON.
	b := newQueryBuilder()

	// Target scope
	if !b.applyTargetScope("t.", targetScope) {
		return []store.FrequencyRow{}, nil
	}

	// Only include targets that have a classification.
	b.addRaw("t.fields->>'potential' IS NOT NULL")
	b.addRaw("t.fields->>'potential' != ''")

	if filter.UserID != nil {
		b.add("t.assignee_id::TEXT = $%d", *filter.UserID)
	}
	if filter.TeamID != nil {
		b.add("t.team_id::TEXT = $%d", *filter.TeamID)
	}

	tWhere := b.where()

	// Activity conditions for the LEFT JOIN (share the same args array).
	aConds := []string{"a.deleted_at IS NULL", "a.target_id = t.id"}

	// Activity scope - need to build inline since it shares args with the target part.
	if !scope.AllActivities {
		var actParts []string
		if len(scope.CreatorIDs) > 0 {
			phs := b.placeholders(scope.CreatorIDs)
			actParts = append(actParts,
				fmt.Sprintf("(a.creator_id::TEXT = ANY(ARRAY[%s]) OR a.joint_visit_user_id::TEXT = ANY(ARRAY[%s]))",
					phs, phs))
		}
		if len(scope.TeamIDs) > 0 {
			phs := b.placeholders(scope.TeamIDs)
			actParts = append(actParts,
				fmt.Sprintf("a.team_id::TEXT = ANY(ARRAY[%s])", phs))
		}
		if len(actParts) > 0 {
			aConds = append(aConds, "("+strings.Join(actParts, " OR ")+")")
		}
	}

	aConds = append(aConds, fmt.Sprintf("a.due_date >= $%d", b.argIdx))
	b.args = append(b.args, filter.DateFrom)
	b.argIdx++
	aConds = append(aConds, fmt.Sprintf("a.due_date <= $%d", b.argIdx))
	b.args = append(b.args, filter.DateTo)
	b.argIdx++

	joinOn := strings.Join(aConds, " AND ")

	query := fmt.Sprintf(`
		SELECT t.fields->>'potential' AS classification,
		       COUNT(DISTINCT t.id) AS target_count,
		       COUNT(a.id) AS total_visits
		FROM targets t
		LEFT JOIN activities a ON %s
		%s
		GROUP BY classification
		ORDER BY classification`,
		joinOn, tWhere)

	rows, err := r.pool.Query(ctx, query, b.args...)
	if err != nil {
		return nil, fmt.Errorf("querying frequency stats: %w", err)
	}
	defer rows.Close()

	var result []store.FrequencyRow
	for rows.Next() {
		var row store.FrequencyRow
		if err := rows.Scan(&row.Classification, &row.TargetCount, &row.TotalVisits); err != nil {
			return nil, fmt.Errorf("scanning frequency row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating frequency rows: %w", err)
	}

	return result, nil
}

// queryDueDates executes a query that selects due_date from activities and returns the dates.
func (r *dashboardRepository) queryDueDates(ctx context.Context, selectClause string, qb *queryBuilder, errContext string) ([]time.Time, error) {
	where := qb.where()
	query := selectClause + ` FROM activities ` + where + ` ORDER BY due_date`
	rows, err := r.pool.Query(ctx, query, qb.args...)
	if err != nil {
		return nil, fmt.Errorf("querying %s: %w", errContext, err)
	}
	defer rows.Close()

	var result []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scanning %s: %w", errContext, err)
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating %s: %w", errContext, err)
	}
	return result, nil
}

// WeekendFieldActivities returns dates of field-category activities on weekends.
func (r *dashboardRepository) WeekendFieldActivities(ctx context.Context, scope rbac.ActivityScope, fieldTypes []string, filter store.DashboardFilter) ([]store.WeekendActivity, error) {
	qb := newQueryBuilder()
	qb.addRaw(condDeletedNull)

	// Weekend: extract DOW (0=Sun, 6=Sat)
	qb.addRaw("EXTRACT(DOW FROM due_date) IN (0, 6)")

	// Field activity types
	qb.add("activity_type = ANY($%d)", fieldTypes)

	if !qb.applyActivityScope("", scope) {
		return nil, nil
	}

	qb.applyDashboardFilter("", filter)

	dates, err := r.queryDueDates(ctx, `SELECT DISTINCT due_date`, qb, "weekend field activities")
	if err != nil {
		return nil, err
	}

	result := make([]store.WeekendActivity, len(dates))
	for i, d := range dates {
		result[i] = store.WeekendActivity{DueDate: d}
	}
	return result, nil
}

// RecoveryActivities returns dates of recovery-type activities taken.
func (r *dashboardRepository) RecoveryActivities(ctx context.Context, scope rbac.ActivityScope, recoveryType string, filter store.DashboardFilter) ([]time.Time, error) {
	qb := newQueryBuilder()
	qb.addRaw(condDeletedNull)

	qb.add("activity_type = $%d", recoveryType)

	if !qb.applyActivityScope("", scope) {
		return nil, nil
	}

	qb.applyDashboardFilter("", filter)

	return r.queryDueDates(ctx, `SELECT due_date`, qb, "recovery activities")
}


// buildPlaceholders appends ids to args starting at argIdx and returns
// comma-separated "$N" placeholders along with the updated args and next index.
func buildPlaceholders(ids []string, args []any, argIdx int) (ph string, outArgs []any, nextIdx int) {
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", argIdx)
		args = append(args, id)
		argIdx++
	}
	return strings.Join(placeholders, ","), args, argIdx
}

// activityScopeConditionsAliased builds the RBAC scope WHERE clause for activities with optional table alias.
func activityScopeConditionsAliased(alias string, scope rbac.ActivityScope, args []any, argIdx int) (sql string, outArgs []any, nextIdx int) {
	if scope.DenyAll {
		return "FALSE", args, argIdx
	}
	if scope.AllActivities {
		return "", args, argIdx
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	var parts []string
	if len(scope.CreatorIDs) > 0 {
		var ph string
		ph, args, argIdx = buildPlaceholders(scope.CreatorIDs, args, argIdx)
		parts = append(parts, fmt.Sprintf("(%screator_id::TEXT = ANY(ARRAY[%s]) OR %sjoint_visit_user_id::TEXT = ANY(ARRAY[%s]))", prefix, ph, prefix, ph))
	}
	if len(scope.TeamIDs) > 0 {
		var ph string
		ph, args, argIdx = buildPlaceholders(scope.TeamIDs, args, argIdx)
		parts = append(parts, fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, ph))
	}
	if len(parts) > 0 {
		return "(" + strings.Join(parts, " OR ") + ")", args, argIdx
	}
	return "", args, argIdx
}

// targetScopeConditionsAliased builds the RBAC scope WHERE clause for targets with optional table alias.
func targetScopeConditionsAliased(alias string, scope rbac.TargetScope, args []any, argIdx int) (sql string, outArgs []any, nextIdx int) {
	if scope.DenyAll {
		return "FALSE", args, argIdx
	}
	if scope.AllTargets {
		return "", args, argIdx
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	var parts []string
	if len(scope.AssigneeIDs) > 0 {
		var ph string
		ph, args, argIdx = buildPlaceholders(scope.AssigneeIDs, args, argIdx)
		parts = append(parts, fmt.Sprintf("%sassignee_id::TEXT = ANY(ARRAY[%s])", prefix, ph))
	}
	if len(scope.TeamIDs) > 0 {
		var ph string
		ph, args, argIdx = buildPlaceholders(scope.TeamIDs, args, argIdx)
		parts = append(parts, fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, ph))
	}
	if len(parts) > 0 {
		return "(" + strings.Join(parts, " OR ") + ")", args, argIdx
	}
	return "", args, argIdx
}
