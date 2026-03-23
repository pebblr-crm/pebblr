package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

type dashboardRepository struct {
	pool *pgxpool.Pool
}

// buildScopeConditions returns WHERE clause fragments and args for RBAC scoping.
// argIdx is the next available $N placeholder index.
func buildScopeConditions(scope rbac.ActivityScope, args []any, argIdx int) (conditions []string, outArgs []any, nextIdx int) {
	if scope.AllActivities {
		return nil, args, argIdx
	}
	var scopeParts []string
	if len(scope.CreatorIDs) > 0 {
		placeholders := make([]string, len(scope.CreatorIDs))
		for i, id := range scope.CreatorIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		joined := strings.Join(placeholders, ",")
		scopeParts = append(scopeParts, fmt.Sprintf("(creator_id::TEXT = ANY(ARRAY[%s]) OR joint_visit_user_id::TEXT = ANY(ARRAY[%s]))", joined, joined))
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
	if len(scopeParts) == 0 {
		// No scope => user can see nothing; add impossible condition.
		return []string{"FALSE"}, args, argIdx
	}
	return []string{"(" + strings.Join(scopeParts, " OR ") + ")"}, args, argIdx
}

func buildTargetScopeConditions(scope rbac.TargetScope, args []any, argIdx int) (conditions []string, outArgs []any, nextIdx int) {
	if scope.AllTargets {
		return nil, args, argIdx
	}
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
	if len(scopeParts) == 0 {
		return []string{"FALSE"}, args, argIdx
	}
	return []string{"(" + strings.Join(scopeParts, " OR ") + ")"}, args, argIdx
}

func (r *dashboardRepository) ActivityStats(ctx context.Context, scope rbac.ActivityScope, filter store.DashboardFilter) (*store.ActivityStats, error) {
	args := []any{filter.DateFrom, filter.DateTo}
	argIdx := 3
	conditions := []string{
		"deleted_at IS NULL",
		"due_date >= $1",
		"due_date <= $2",
	}

	scopeConds, args, argIdx := buildScopeConditions(scope, args, argIdx)
	conditions = append(conditions, scopeConds...)

	if filter.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Query 1: total and submitted count.
	var stats store.ActivityStats
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE submitted_at IS NOT NULL) FROM activities `+where,
		args...,
	).Scan(&stats.Total, &stats.SubmittedCount)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats totals: %w", err)
	}

	// Query 2: by status.
	rows, err := r.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM activities `+where+` GROUP BY status ORDER BY COUNT(*) DESC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sc store.StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, fmt.Errorf("scanning status count: %w", err)
		}
		stats.ByStatus = append(stats.ByStatus, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating status counts: %w", err)
	}
	if stats.ByStatus == nil {
		stats.ByStatus = []store.StatusCount{}
	}

	// Query 3: by type.
	rows2, err := r.pool.Query(ctx,
		`SELECT activity_type, COUNT(*) FROM activities `+where+` GROUP BY activity_type ORDER BY COUNT(*) DESC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("querying activity stats by type: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var tc store.TypeCount
		if err := rows2.Scan(&tc.ActivityType, &tc.Count); err != nil {
			return nil, fmt.Errorf("scanning type count: %w", err)
		}
		stats.ByType = append(stats.ByType, tc)
	}
	if err := rows2.Err(); err != nil {
		return nil, fmt.Errorf("iterating type counts: %w", err)
	}
	if stats.ByType == nil {
		stats.ByType = []store.TypeCount{}
	}

	return &stats, nil
}

func (r *dashboardRepository) CoverageStats(ctx context.Context, scope rbac.ActivityScope, targetScope rbac.TargetScope, filter store.DashboardFilter, fieldTypes []string, realizedStatus string) (*store.CoverageStats, error) {
	if len(fieldTypes) == 0 {
		return &store.CoverageStats{}, nil
	}

	// Count total targets visible to user.
	tArgs := []any{}
	tArgIdx := 1
	tConds := []string{"TRUE"} // targets have no deleted_at

	tScopeConds, tArgs, tArgIdx := buildTargetScopeConditions(targetScope, tArgs, tArgIdx)
	tConds = append(tConds, tScopeConds...)
	_ = tArgIdx

	tWhere := "WHERE " + strings.Join(tConds, " AND ")

	var totalTargets int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM targets `+tWhere, tArgs...).Scan(&totalTargets); err != nil {
		return nil, fmt.Errorf("counting targets: %w", err)
	}

	if totalTargets == 0 {
		return &store.CoverageStats{TotalTargets: 0, VisitedTargets: 0, CoveragePercent: 0}, nil
	}

	// Count distinct targets visited (realized field activities) in the period.
	aArgs := []any{filter.DateFrom, filter.DateTo, realizedStatus, fieldTypes}
	aArgIdx := 5
	aConds := []string{
		"deleted_at IS NULL",
		"due_date >= $1",
		"due_date <= $2",
		"status = $3",
		"activity_type = ANY($4)",
		"target_id IS NOT NULL",
	}

	aScopeConds, aArgs, aArgIdx := buildScopeConditions(scope, aArgs, aArgIdx)
	aConds = append(aConds, aScopeConds...)

	if filter.CreatorID != nil {
		aConds = append(aConds, fmt.Sprintf("creator_id::TEXT = $%d", aArgIdx))
		aArgs = append(aArgs, *filter.CreatorID)
		aArgIdx++
	}
	if filter.TeamID != nil {
		aConds = append(aConds, fmt.Sprintf("team_id::TEXT = $%d", aArgIdx))
		aArgs = append(aArgs, *filter.TeamID)
		aArgIdx++
	}
	_ = aArgIdx

	aWhere := "WHERE " + strings.Join(aConds, " AND ")

	var visitedTargets int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT target_id) FROM activities `+aWhere, aArgs...).Scan(&visitedTargets); err != nil {
		return nil, fmt.Errorf("counting visited targets: %w", err)
	}

	pct := 0.0
	if totalTargets > 0 {
		pct = float64(visitedTargets) / float64(totalTargets) * 100
	}

	return &store.CoverageStats{
		TotalTargets:    totalTargets,
		VisitedTargets:  visitedTargets,
		CoveragePercent: pct,
	}, nil
}

func (r *dashboardRepository) UserStats(ctx context.Context, scope rbac.ActivityScope, filter store.DashboardFilter) ([]store.UserActivityStats, error) {
	args := []any{filter.DateFrom, filter.DateTo}
	argIdx := 3
	conditions := []string{
		"a.deleted_at IS NULL",
		"a.due_date >= $1",
		"a.due_date <= $2",
	}

	// Adapt scope conditions for aliased table.
	if !scope.AllActivities {
		var scopeParts []string
		if len(scope.CreatorIDs) > 0 {
			placeholders := make([]string, len(scope.CreatorIDs))
			for i, id := range scope.CreatorIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			joined := strings.Join(placeholders, ",")
			scopeParts = append(scopeParts, fmt.Sprintf("(a.creator_id::TEXT = ANY(ARRAY[%s]) OR a.joint_visit_user_id::TEXT = ANY(ARRAY[%s]))", joined, joined))
		}
		if len(scope.TeamIDs) > 0 {
			placeholders := make([]string, len(scope.TeamIDs))
			for i, id := range scope.TeamIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("a.team_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		}
		if len(scopeParts) == 0 {
			return []store.UserActivityStats{}, nil
		}
		conditions = append(conditions, "("+strings.Join(scopeParts, " OR ")+")")
	}

	if filter.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("a.creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("a.team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Get per-user totals with user name from JOIN.
	query := `SELECT a.creator_id::TEXT, COALESCE(u.name, ''), COUNT(*)
		FROM activities a
		LEFT JOIN users u ON u.id = a.creator_id
		` + where + `
		GROUP BY a.creator_id, u.name
		ORDER BY COUNT(*) DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying user stats: %w", err)
	}
	defer rows.Close()

	type userRow struct {
		id    string
		name  string
		total int
	}
	var users []userRow
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.id, &u.name, &u.total); err != nil {
			return nil, fmt.Errorf("scanning user stats: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user stats: %w", err)
	}

	if len(users) == 0 {
		return []store.UserActivityStats{}, nil
	}

	// Get per-user status breakdown.
	statusQuery := `SELECT a.creator_id::TEXT, a.status, COUNT(*)
		FROM activities a
		` + where + `
		GROUP BY a.creator_id, a.status
		ORDER BY a.creator_id, COUNT(*) DESC`

	rows2, err := r.pool.Query(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying user status breakdown: %w", err)
	}
	defer rows2.Close()

	statusMap := make(map[string][]store.StatusCount)
	for rows2.Next() {
		var userID, status string
		var count int
		if err := rows2.Scan(&userID, &status, &count); err != nil {
			return nil, fmt.Errorf("scanning user status breakdown: %w", err)
		}
		statusMap[userID] = append(statusMap[userID], store.StatusCount{Status: status, Count: count})
	}
	if err := rows2.Err(); err != nil {
		return nil, fmt.Errorf("iterating user status breakdown: %w", err)
	}

	result := make([]store.UserActivityStats, len(users))
	for i, u := range users {
		byStatus := statusMap[u.id]
		if byStatus == nil {
			byStatus = []store.StatusCount{}
		}
		result[i] = store.UserActivityStats{
			UserID:   u.id,
			UserName: u.name,
			Total:    u.total,
			ByStatus: byStatus,
		}
	}

	return result, nil
}
