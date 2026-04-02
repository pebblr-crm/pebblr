package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub audit repo for audit service tests ---

type stubAuditSvcRepo struct {
	entries     []*domain.AuditEntry
	total       int
	listErr     error
	updateErr   error
	updatedID   string
	updatedStat string
}

func (r *stubAuditSvcRepo) Record(_ context.Context, _ *domain.AuditEntry) error {
	return nil
}

func (r *stubAuditSvcRepo) ListByEntity(_ context.Context, _, _ string) ([]*domain.AuditEntry, error) {
	return r.entries, nil
}

func (r *stubAuditSvcRepo) List(_ context.Context, _ store.AuditFilter) ([]*domain.AuditEntry, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	return r.entries, r.total, nil
}

func (r *stubAuditSvcRepo) UpdateStatus(_ context.Context, id, status, _ string) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedID = id
	r.updatedStat = status
	return nil
}

// --- List tests ---

func TestAudit_List_AdminSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{
		entries: []*domain.AuditEntry{{ID: "a-1", Status: "pending"}},
		total:   1,
	}
	svc := service.NewAuditService(repo)

	entries, total, err := svc.List(context.Background(), adminUser(), store.AuditFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestAudit_List_NilBecomesEmptySlice(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{entries: nil, total: 0}
	svc := service.NewAuditService(repo)

	entries, _, err := svc.List(context.Background(), adminUser(), store.AuditFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries == nil {
		t.Error("expected empty slice, got nil")
	}
}

func TestAudit_List_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	_, _, err := svc.List(context.Background(), repUser(), store.AuditFilter{})
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAudit_List_ManagerForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	_, _, err := svc.List(context.Background(), managerUser(), store.AuditFilter{})
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAudit_List_RepoError(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{listErr: errors.New("db error")}
	svc := service.NewAuditService(repo)

	_, _, err := svc.List(context.Background(), adminUser(), store.AuditFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- UpdateStatus tests ---

func TestAudit_UpdateStatus_AdminSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), adminUser(), "a-1", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedID != "a-1" {
		t.Errorf("expected updatedID a-1, got %s", repo.updatedID)
	}
	if repo.updatedStat != "accepted" {
		t.Errorf("expected status accepted, got %s", repo.updatedStat)
	}
}

func TestAudit_UpdateStatus_FalsePositive(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), adminUser(), "a-1", "false_positive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAudit_UpdateStatus_Pending(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), adminUser(), "a-1", "pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAudit_UpdateStatus_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), repUser(), "a-1", "accepted")
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAudit_UpdateStatus_InvalidStatus(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), adminUser(), "a-1", "bogus")
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAudit_UpdateStatus_RepoError(t *testing.T) {
	t.Parallel()
	repo := &stubAuditSvcRepo{updateErr: errors.New("db error")}
	svc := service.NewAuditService(repo)

	err := svc.UpdateStatus(context.Background(), adminUser(), "a-1", "accepted")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
