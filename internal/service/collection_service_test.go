package service_test

import (
	"context"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub collection repo ---

type stubCollectionRepo struct {
	collection *domain.Collection
	list       []*domain.Collection
	created    *domain.Collection
	deleted    bool
	getErr     error
	saveErr    error
}

func (r *stubCollectionRepo) Create(_ context.Context, c *domain.Collection) (*domain.Collection, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	c.ID = "col-1"
	r.created = c
	return c, nil
}

func (r *stubCollectionRepo) List(_ context.Context, _ store.CollectionFilter) ([]*domain.Collection, error) {
	return r.list, nil
}

func (r *stubCollectionRepo) Get(_ context.Context, _ string) (*domain.Collection, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.collection, nil
}

func (r *stubCollectionRepo) Update(_ context.Context, c *domain.Collection) (*domain.Collection, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	return c, nil
}

func (r *stubCollectionRepo) Delete(_ context.Context, _ string) error {
	r.deleted = true
	return nil
}

func newCollectionSvc(repo *stubCollectionRepo) *service.CollectionService {
	return service.NewCollectionService(repo)
}

func sampleCollection() *domain.Collection {
	return &domain.Collection{
		ID:        "col-1",
		Name:      "Bucharest Mon",
		CreatorID: "rep-1",
		TeamID:    "team-1",
		TargetIDs: []string{"t1", "t2", "t3"},
	}
}

// --- Create tests ---

func TestCollection_Create_Success(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{}
	svc := newCollectionSvc(repo)

	c, err := svc.Create(context.Background(), repUser(), "Bucharest Mon", []string{"t1", "t2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "Bucharest Mon" {
		t.Errorf("name = %s, want Bucharest Mon", c.Name)
	}
	if c.CreatorID != "rep-1" {
		t.Errorf("creatorId = %s, want rep-1", c.CreatorID)
	}
	if len(c.TargetIDs) != 2 {
		t.Errorf("targetIds = %d, want 2", len(c.TargetIDs))
	}
}

func TestCollection_Create_EmptyName(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{}
	svc := newCollectionSvc(repo)

	_, err := svc.Create(context.Background(), repUser(), "", []string{"t1"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCollection_Create_NilTargetIDs(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{}
	svc := newCollectionSvc(repo)

	c, err := svc.Create(context.Background(), repUser(), "Empty", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.TargetIDs == nil {
		t.Error("targetIds should be empty slice, not nil")
	}
}

// --- List tests ---

func TestCollection_List_RepSeesOwn(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{
		list: []*domain.Collection{sampleCollection()},
	}
	svc := newCollectionSvc(repo)

	result, err := svc.List(context.Background(), repUser())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 collection, got %d", len(result))
	}
}

// --- Get tests ---

func TestCollection_Get_OwnerCanView(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	c, err := svc.Get(context.Background(), repUser(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "col-1" {
		t.Errorf("id = %s, want col-1", c.ID)
	}
}

func TestCollection_Get_OtherRepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	otherRep := &domain.User{ID: "rep-2", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
	_, err := svc.Get(context.Background(), otherRep, "col-1")
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestCollection_Get_AdminCanView(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	c, err := svc.Get(context.Background(), adminUser(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "col-1" {
		t.Errorf("id = %s, want col-1", c.ID)
	}
}

func TestCollection_Get_ManagerCanViewTeam(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	c, err := svc.Get(context.Background(), managerUser(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "col-1" {
		t.Errorf("id = %s, want col-1", c.ID)
	}
}

// --- Update tests ---

func TestCollection_Update_OwnerCanUpdate(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	c, err := svc.Update(context.Background(), repUser(), "col-1", "Renamed", []string{"t1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "Renamed" {
		t.Errorf("name = %s, want Renamed", c.Name)
	}
}

func TestCollection_Update_OtherRepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	otherRep := &domain.User{ID: "rep-2", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
	_, err := svc.Update(context.Background(), otherRep, "col-1", "Hijack", []string{})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestCollection_Update_EmptyName(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	_, err := svc.Update(context.Background(), repUser(), "col-1", "", []string{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

// --- Delete tests ---

func TestCollection_Delete_OwnerCanDelete(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	err := svc.Delete(context.Background(), repUser(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted {
		t.Error("expected collection to be deleted")
	}
}

func TestCollection_Delete_OtherRepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	otherRep := &domain.User{ID: "rep-2", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
	err := svc.Delete(context.Background(), otherRep, "col-1")
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestCollection_Delete_AdminCanDelete(t *testing.T) {
	t.Parallel()
	repo := &stubCollectionRepo{collection: sampleCollection()}
	svc := newCollectionSvc(repo)

	err := svc.Delete(context.Background(), adminUser(), "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted {
		t.Error("expected collection to be deleted")
	}
}
