package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	whereDeletedAtIsNull = "deleted_at IS NULL"
	sqlWhere             = "WHERE "
)

type dashboardRepository struct {
	pool dbPool
}

// ActivityStats returns activity counts grouped by status and category for the given scope and filter.
func (r *dashboardRepository) ActivityStats(ctx context.Context, scope rbac.ActivityScope, filter store.DashboardFilter) (*store.ActivityStats, error) {
	args := []any{}
	argIdx := 1
	conditions := []string{whereDeletedAtIsNull}

	scopeSQL, args, argIdx := activityScopeConditions(scope, args, argIdx)
	if scopeSQL != "" {
		conditions = append(conditions, scopeSQL)
	} else if !scope.AllActivities {
		return &store.ActivityStats{ByStatus: map[string]int{}, ByCategory: map[string]int{}}, nil
	}

	conditions, args, _ = appendDashboardFilter(conditions, args, argIdx, filter)

	where := sqlWhere + strings.Join(conditions, " AND ")

	// By status
	byStatus := map[string]int{}
	statusQuery := `SELECT status, COUNT(*) FROM activities ` + where + ` GROUP BY status`
	rows, err := r.pool.Query(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by status: %w", err)
	}
	total := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scanning status row: %w", err)
		}
		byStatus[status] = count
		total += count
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating status rows: %w", err)
	}

	// By category (activity_type → category mapping done at service layer, here just group by activity_type)
	byCategory := map[string]int{}
	typeQuery := `SELECT activity_type, COUNT(*) FROM activities ` + where + ` GROUP BY activity_type`
	rows, err = r.pool.Query(ctx, typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by type: %w", err)
	}
	for rows.Next() {
		var activityType string
		var count int
		if err := rows.Scan(&activityType, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scanning type row: %w", err)
		}
		byCategory[activityType] = count
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating type rows: %w", err)
	}

	return &store.ActivityStats{ByStatus: byStatus, ByCategory: byCategory, Total: total}, nil
}

// CoverageStats returns the total vs visited target counts.
func (r *dashboardRepository) CoverageStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter store.DashboardFilter) (*store.CoverageStats, error) {
	// Count total targets in scope.
	tArgs := []any{}
	tIdx := 1
	tConds := []string{}

	tScopeSQL, tArgs, tIdx := targetScopeConditions(targetScope, tArgs, tIdx)
	if tScopeSQL != "" {
		tConds = append(tConds, tScopeSQL)
	} else if !targetScope.AllTargets {
		return &store.CoverageStats{}, nil
	}

	if filter.UserID != nil {
		tConds = append(tConds, fmt.Sprintf("assignee_id::TEXT = $%d", tIdx))
		tArgs = append(tArgs, *filter.UserID)
		tIdx++
	}
	if filter.TeamID != nil {
		tConds = append(tConds, fmt.Sprintf("team_id::TEXT = $%d", tIdx))
		tArgs = append(tArgs, *filter.TeamID)
		// tIdx++ not needed — last use
	}

	tWhere := ""
	if len(tConds) > 0 {
		tWhere = sqlWhere + strings.Join(tConds, " AND ")
	}

	var totalTargets int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM targets `+tWhere, tArgs...).Scan(&totalTargets); err != nil {
		return nil, fmt.Errorf("counting targets: %w", err)
	}

	// Count distinct targets visited (field activities with a target in the period).
	aArgs := []any{}
	aIdx := 1
	aConds := []string{"a." + whereDeletedAtIsNull, "a.target_id IS NOT NULL"}

	aScopeSQL, aArgs, aIdx := activityScopeConditions(scope, aArgs, aIdx)
	if aScopeSQL != "" {
		aConds = append(aConds, aScopeSQL)
	} else if !scope.AllActivities {
		return &store.CoverageStats{TotalTargets: totalTargets}, nil
	}

	aConds = append(aConds, fmt.Sprintf("a.due_date >= $%d", aIdx))
	aArgs = append(aArgs, filter.DateFrom)
	aIdx++
	aConds = append(aConds, fmt.Sprintf("a.due_date <= $%d", aIdx))
	aArgs = append(aArgs, filter.DateTo)
	aIdx++

	if filter.UserID != nil {
		aConds = append(aConds, fmt.Sprintf("a.creator_id::TEXT = $%d", aIdx))
		aArgs = append(aArgs, *filter.UserID)
		aIdx++
	}
	if filter.TeamID != nil {
		aConds = append(aConds, fmt.Sprintf("a.team_id::TEXT = $%d", aIdx))
		aArgs = append(aArgs, *filter.TeamID)
		// aIdx++ not needed — last use
	}

	aWhere := sqlWhere + strings.Join(aConds, " AND ")

	var visitedTargets int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT a.target_id) FROM activities a `+aWhere,
		aArgs...,
	).Scan(&visitedTargets); err != nil {
		return nil, fmt.Errorf("counting visited targets: %w", err)
	}

	return &store.CoverageStats{TotalTargets: totalTargets, VisitedTargets: visitedTargets}, nil
}

// FrequencyStats returns per-classification visit counts.
func (r *dashboardRepository) FrequencyStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter store.DashboardFilter) ([]store.FrequencyRow, error) {
	// Join targets (for classification) with activity counts in the period.
	args := []any{}
	argIdx := 1

	// Target scope conditions (applied to t alias).
	tConds := []string{}
	tScopeSQL, args, argIdx := targetScopeConditionsAliased("t", targetScope, args, argIdx)
	if tScopeSQL != "" {
		tConds = append(tConds, tScopeSQL)
	} else if !targetScope.AllTargets {
		return []store.FrequencyRow{}, nil
	}

	// Only include targets that have a classification.
	tConds = append(tConds, "t.fields->>'potential' IS NOT NULL", "t.fields->>'potential' != ''")

	if filter.UserID != nil {
		tConds = append(tConds, fmt.Sprintf("t.assignee_id::TEXT = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.TeamID != nil {
		tConds = append(tConds, fmt.Sprintf("t.team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
		argIdx++
	}

	tWhere := ""
	if len(tConds) > 0 {
		tWhere = sqlWhere + strings.Join(tConds, " AND ")
	}

	// Activity conditions for the LEFT JOIN.
	aConds := []string{"a." + whereDeletedAtIsNull, "a.target_id = t.id"}

	aScopeSQL, args, argIdx := activityScopeConditionsAliased("a", scope, args, argIdx)
	if aScopeSQL != "" {
		aConds = append(aConds, aScopeSQL)
	}

	aConds = append(aConds, fmt.Sprintf("a.due_date >= $%d", argIdx))
	args = append(args, filter.DateFrom)
	argIdx++
	aConds = append(aConds, fmt.Sprintf("a.due_date <= $%d", argIdx))
	args = append(args, filter.DateTo)
	// argIdx++ not needed — last use

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

	rows, err := r.pool.Query(ctx, query, args...)
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

// WeekendFieldActivities returns dates of field-category activities on weekends.
func (r *dashboardRepository) WeekendFieldActivities(ctx context.Context, scope rbac.ActivityScope, fieldTypes []string, filter store.DashboardFilter) ([]store.WeekendActivity, error) {
	args := []any{}
	argIdx := 1
	conditions := []string{whereDeletedAtIsNull}

	// Weekend: extract DOW (0=Sun, 6=Sat)
	conditions = append(conditions, "EXTRACT(DOW FROM due_date) IN (0, 6)")

	// Field activity types
	args = append(args, fieldTypes)
	conditions = append(conditions, fmt.Sprintf("activity_type = ANY($%d)", argIdx))
	argIdx++

	scopeSQL, args, argIdx := activityScopeConditions(scope, args, argIdx)
	if scopeSQL != "" {
		conditions = append(conditions, scopeSQL)
	} else if !scope.AllActivities {
		return nil, nil
	}

	conditions, args, _ = appendDashboardFilter(conditions, args, argIdx, filter)

	where := sqlWhere + strings.Join(conditions, " AND ")
	query := `SELECT DISTINCT due_date FROM activities ` + where + ` ORDER BY due_date`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying weekend field activities: %w", err)
	}
	defer rows.Close()

	var result []store.WeekendActivity
	for rows.Next() {
		var wa store.WeekendActivity
		if err := rows.Scan(&wa.DueDate); err != nil {
			return nil, fmt.Errorf("scanning weekend activity: %w", err)
		}
		result = append(result, wa)
	}
	return result, rows.Err()
}

// RecoveryActivities returns dates of recovery-type activities taken.
func (r *dashboardRepository) RecoveryActivities(ctx context.Context, scope rbac.ActivityScope, recoveryType string, filter store.DashboardFilter) ([]time.Time, error) {
	args := []any{}
	argIdx := 1
	conditions := []string{whereDeletedAtIsNull}

	args = append(args, recoveryType)
	conditions = append(conditions, fmt.Sprintf("activity_type = $%d", argIdx))
	argIdx++

	scopeSQL, args, argIdx := activityScopeConditions(scope, args, argIdx)
	if scopeSQL != "" {
		conditions = append(conditions, scopeSQL)
	} else if !scope.AllActivities {
		return nil, nil
	}

	conditions, args, _ = appendDashboardFilter(conditions, args, argIdx, filter)

	where := sqlWhere + strings.Join(conditions, " AND ")
	query := `SELECT due_date FROM activities ` + where + ` ORDER BY due_date`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying recovery activities: %w", err)
	}
	defer rows.Close()

	var result []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scanning recovery activity: %w", err)
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

// appendDashboardFilter adds date range and user/team filter conditions.
func appendDashboardFilter(conditions []string, args []any, argIdx int, filter store.DashboardFilter) (conds []string, outArgs []any, nextIdx int) {
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("due_date >= $%d", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
		argIdx++
	}
	return conditions, args, argIdx
}

// activityScopeConditions builds the RBAC scope WHERE clause for activities (no alias).
func activityScopeConditions(scope rbac.ActivityScope, args []any, argIdx int) (sql string, outArgs []any, nextIdx int) {
	return activityScopeConditionsAliased("", scope, args, argIdx)
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
		placeholders := make([]string, len(scope.CreatorIDs))
		for i, id := range scope.CreatorIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		ph := strings.Join(placeholders, ",")
		parts = append(parts, fmt.Sprintf("(%screator_id::TEXT = ANY(ARRAY[%s]) OR %sjoint_visit_user_id::TEXT = ANY(ARRAY[%s]))", prefix, ph, prefix, ph))
	}
	if len(scope.TeamIDs) > 0 {
		placeholders := make([]string, len(scope.TeamIDs))
		for i, id := range scope.TeamIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		parts = append(parts, fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, strings.Join(placeholders, ",")))
	}
	if len(parts) > 0 {
		return "(" + strings.Join(parts, " OR ") + ")", args, argIdx
	}
	return "", args, argIdx
}

// targetScopeConditions builds the RBAC scope WHERE clause for targets (no alias).
func targetScopeConditions(scope rbac.TargetScope, args []any, argIdx int) (sql string, outArgs []any, nextIdx int) {
	return targetScopeConditionsAliased("", scope, args, argIdx)
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
		placeholders := make([]string, len(scope.AssigneeIDs))
		for i, id := range scope.AssigneeIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		parts = append(parts, fmt.Sprintf("%sassignee_id::TEXT = ANY(ARRAY[%s])", prefix, strings.Join(placeholders, ",")))
	}
	if len(scope.TeamIDs) > 0 {
		placeholders := make([]string, len(scope.TeamIDs))
		for i, id := range scope.TeamIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		parts = append(parts, fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, strings.Join(placeholders, ",")))
	}
	if len(parts) > 0 {
		return "(" + strings.Join(parts, " OR ") + ")", args, argIdx
	}
	return "", args, argIdx
}
