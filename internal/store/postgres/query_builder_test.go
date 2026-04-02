package postgres

import (
	"testing"

	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
	"time"
)

func TestQueryBuilder_Add(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	qb.add("name = $%d", "alice")
	qb.add("age > $%d", 30)

	where := qb.where()
	expected := "WHERE name = $1 AND age > $2"
	if where != expected {
		t.Errorf("expected %q, got %q", expected, where)
	}
	if len(qb.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(qb.args))
	}
}

func TestQueryBuilder_AddRaw(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	qb.addRaw("deleted_at IS NULL")
	qb.add("status = $%d", "active")

	where := qb.where()
	expected := "WHERE deleted_at IS NULL AND status = $1"
	if where != expected {
		t.Errorf("expected %q, got %q", expected, where)
	}
}

func TestQueryBuilder_EmptyWhere(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	if w := qb.where(); w != "" {
		t.Errorf("expected empty string, got %q", w)
	}
}

func TestQueryBuilder_ActivityScope_AllActivities(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	ok := qb.applyActivityScope("", rbac.ActivityScope{AllActivities: true})
	if !ok {
		t.Error("expected true for AllActivities")
	}
	if len(qb.conditions) != 0 {
		t.Errorf("expected no conditions, got %d", len(qb.conditions))
	}
}

func TestQueryBuilder_ActivityScope_CreatorIDs(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	ok := qb.applyActivityScope("a.", rbac.ActivityScope{CreatorIDs: []string{"u1", "u2"}})
	if !ok {
		t.Error("expected true")
	}
	if len(qb.conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(qb.conditions))
	}
	if len(qb.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(qb.args))
	}
}

func TestQueryBuilder_ActivityScope_Empty(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	ok := qb.applyActivityScope("", rbac.ActivityScope{})
	if ok {
		t.Error("expected false for empty scope")
	}
}

func TestQueryBuilder_TargetScope_AssigneeIDs(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	ok := qb.applyTargetScope("t.", rbac.TargetScope{AssigneeIDs: []string{"u1"}})
	if !ok {
		t.Error("expected true")
	}
	if len(qb.args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(qb.args))
	}
}

func TestQueryBuilder_DashboardFilter(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	userID := "user-1"
	qb.applyDashboardFilter("", store.DashboardFilter{
		DateFrom: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		UserID:   &userID,
	})
	if len(qb.conditions) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(qb.conditions))
	}
	if len(qb.args) != 3 {
		t.Errorf("expected 3 args, got %d", len(qb.args))
	}
}

func TestQueryBuilder_NextPlaceholder(t *testing.T) {
	t.Parallel()
	qb := newQueryBuilder()
	ph1 := qb.nextPlaceholder("val1")
	ph2 := qb.nextPlaceholder("val2")
	if ph1 != "$1" {
		t.Errorf("expected $1, got %s", ph1)
	}
	if ph2 != "$2" {
		t.Errorf("expected $2, got %s", ph2)
	}
	if len(qb.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(qb.args))
	}
}
