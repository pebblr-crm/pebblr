package postgres

import (
	"fmt"
	"strings"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// queryBuilder accumulates SQL conditions and positional arguments ($1, $2, ...)
// for building parameterised queries. It removes the need to manually track
// argIdx across long query-construction code.
type queryBuilder struct {
	conditions []string
	args       []any
	argIdx     int
}

func newQueryBuilder() *queryBuilder {
	return &queryBuilder{argIdx: 1}
}

// add appends a condition using a format string that contains exactly one %d
// placeholder for the positional argument index.
func (b *queryBuilder) add(sqlFmt string, val any) {
	b.conditions = append(b.conditions, fmt.Sprintf(sqlFmt, b.argIdx))
	b.args = append(b.args, val)
	b.argIdx++
}

// addRaw appends a condition that requires no arguments.
func (b *queryBuilder) addRaw(sql string) {
	b.conditions = append(b.conditions, sql)
}

// where returns "WHERE c1 AND c2 ..." or an empty string when there are no conditions.
func (b *queryBuilder) where() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(b.conditions, " AND ")
}

// nextPlaceholder returns "$N" and advances the counter, appending val to args.
func (b *queryBuilder) nextPlaceholder(val any) string {
	ph := fmt.Sprintf("$%d", b.argIdx)
	b.args = append(b.args, val)
	b.argIdx++
	return ph
}

// applyActivityScope adds RBAC activity scope conditions. Returns false if
// the scope excludes all rows (caller should short-circuit with empty results).
func (b *queryBuilder) applyActivityScope(prefix string, scope rbac.ActivityScope) bool {
	if scope.AllActivities {
		return true
	}

	var parts []string
	if len(scope.CreatorIDs) > 0 {
		phs := b.placeholders(scope.CreatorIDs)
		parts = append(parts,
			fmt.Sprintf("(%screator_id::TEXT = ANY(ARRAY[%s]) OR %sjoint_visit_user_id::TEXT = ANY(ARRAY[%s]))",
				prefix, phs, prefix, phs))
	}
	if len(scope.TeamIDs) > 0 {
		phs := b.placeholders(scope.TeamIDs)
		parts = append(parts,
			fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, phs))
	}
	if len(parts) == 0 {
		return false
	}
	b.conditions = append(b.conditions, "("+strings.Join(parts, " OR ")+")")
	return true
}

// applyTargetScope adds RBAC target scope conditions. Returns false if
// the scope excludes all rows.
func (b *queryBuilder) applyTargetScope(prefix string, scope rbac.TargetScope) bool {
	if scope.AllTargets {
		return true
	}

	var parts []string
	if len(scope.AssigneeIDs) > 0 {
		phs := b.placeholders(scope.AssigneeIDs)
		parts = append(parts,
			fmt.Sprintf("%sassignee_id::TEXT = ANY(ARRAY[%s])", prefix, phs))
	}
	if len(scope.TeamIDs) > 0 {
		phs := b.placeholders(scope.TeamIDs)
		parts = append(parts,
			fmt.Sprintf("%steam_id::TEXT = ANY(ARRAY[%s])", prefix, phs))
	}
	if len(parts) == 0 {
		return false
	}
	b.conditions = append(b.conditions, "("+strings.Join(parts, " OR ")+")")
	return true
}

// applyDashboardFilter adds date-range and user/team filter conditions.
func (b *queryBuilder) applyDashboardFilter(prefix string, filter store.DashboardFilter) {
	if !filter.DateFrom.IsZero() {
		b.add(prefix+"due_date >= $%d", filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		b.add(prefix+"due_date <= $%d", filter.DateTo)
	}
	if filter.UserID != nil {
		b.add(prefix+"creator_id::TEXT = $%d", *filter.UserID)
	}
	if filter.TeamID != nil {
		b.add(prefix+"team_id::TEXT = $%d", *filter.TeamID)
	}
}

// placeholders registers len(ids) values and returns their joined "$N,$M,..."
// placeholders string.
func (b *queryBuilder) placeholders(ids []string) string {
	phs := make([]string, len(ids))
	for i, id := range ids {
		phs[i] = fmt.Sprintf("$%d", b.argIdx)
		b.args = append(b.args, id)
		b.argIdx++
	}
	return strings.Join(phs, ",")
}
