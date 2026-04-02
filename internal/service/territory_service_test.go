package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	errDBMsg          = "db error"
	msgExpectedErr    = "expected error, got nil"
	testTeam99ID      = "team-99"
	testTerritoryName = "New Territory"
)

// --- stub territory repo ---

type stubTerritoryRepo struct {
	territory *domain.Territory
	list      []*domain.Territory
	created   *domain.Territory
	updated   *domain.Territory
	deleted   bool
	getErr    error
	listErr   error
	createErr error
	updateErr error
	deleteErr error
}

func (r *stubTerritoryRepo) Get(_ context.Context, _ string) (*domain.Territory, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.territory, nil
}

func (r *stubTerritoryRepo) List(_ context.Context, _ store.TerritoryFilter) ([]*domain.Territory, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.list, nil
}

func (r *stubTerritoryRepo) Create(_ context.Context, t *domain.Territory) (*domain.Territory, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	t.ID = "ter-1"
	r.created = t
	return t, nil
}

func (r *stubTerritoryRepo) Update(_ context.Context, t *domain.Territory) (*domain.Territory, error) {
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	r.updated = t
	return t, nil
}

func (r *stubTerritoryRepo) Delete(_ context.Context, _ string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	r.deleted = true
	return nil
}

func sampleTerritory() *domain.Territory {
	return &domain.Territory{
		ID:     "ter-1",
		Name:   "Bucharest North",
		TeamID: testTeamID,
		Region: "Bucharest",
	}
}

// --- List tests ---

func TestTerritory_List_AdminSeesAll(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{
		list: []*domain.Territory{sampleTerritory()},
	}
	svc := service.NewTerritoryService(repo)

	result, err := svc.List(context.Background(), adminUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 territory, got %d", len(result))
	}
}

func TestTerritory_List_NilResultBecomesEmptySlice(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{list: nil}
	svc := service.NewTerritoryService(repo)

	result, err := svc.List(context.Background(), adminUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if result == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 territories, got %d", len(result))
	}
}

func TestTerritory_List_RepoError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{listErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	_, err := svc.List(context.Background(), adminUser())
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

func TestTerritory_List_ManagerScopedByTeam(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{
		list: []*domain.Territory{sampleTerritory()},
	}
	svc := service.NewTerritoryService(repo)

	result, err := svc.List(context.Background(), managerUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 territory, got %d", len(result))
	}
}

// --- Get tests ---

func TestTerritory_Get_AdminCanView(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	ter, err := svc.Get(context.Background(), adminUser(), "ter-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if ter.ID != "ter-1" {
		t.Errorf("expected ID ter-1, got %s", ter.ID)
	}
}

func TestTerritory_Get_RepSameTeam(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	ter, err := svc.Get(context.Background(), repUser(), "ter-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if ter.ID != "ter-1" {
		t.Errorf("expected ID ter-1, got %s", ter.ID)
	}
}

func TestTerritory_Get_RepOtherTeamForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	otherRep := &domain.User{ID: "rep-2", Role: domain.RoleRep, TeamIDs: []string{testTeam99ID}}
	_, err := svc.Get(context.Background(), otherRep, "ter-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Get_RepoError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{getErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	_, err := svc.Get(context.Background(), adminUser(), "ter-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Create tests ---

func TestTerritory_Create_AdminSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{Name: testTerritoryName, TeamID: testTeamID, Region: "Cluj"}
	created, err := svc.Create(context.Background(), adminUser(), ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if created.Name != testTerritoryName {
		t.Errorf("expected name %s, got %s", testTerritoryName, created.Name)
	}
}

func TestTerritory_Create_ManagerSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{Name: testTerritoryName, TeamID: testTeamID}
	_, err := svc.Create(context.Background(), managerUser(), ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestTerritory_Create_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{Name: testTerritoryName, TeamID: testTeamID}
	_, err := svc.Create(context.Background(), repUser(), ter)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Create_EmptyNameInvalid(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{Name: "", TeamID: testTeamID}
	_, err := svc.Create(context.Background(), adminUser(), ter)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Fatalf(fmtExpectedInvalidInput, err)
	}
}

func TestTerritory_Create_InvalidBoundary(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{
		Name:     "Bad Boundary",
		TeamID:   testTeamID,
		Boundary: map[string]any{"invalid": true},
	}
	_, err := svc.Create(context.Background(), adminUser(), ter)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Fatalf(fmtExpectedInvalidInput, err)
	}
}

func TestTerritory_Create_RepoError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{createErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{Name: "Test", TeamID: testTeamID}
	_, err := svc.Create(context.Background(), adminUser(), ter)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Update tests ---

func TestTerritory_Update_AdminSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{ID: "ter-1", Name: "Updated", TeamID: testTeamID}
	updated, err := svc.Update(context.Background(), adminUser(), ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected name Updated, got %s", updated.Name)
	}
}

func TestTerritory_Update_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{ID: "ter-1", Name: "Hacked", TeamID: testTeamID}
	_, err := svc.Update(context.Background(), repUser(), ter)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Update_EmptyNameInvalid(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{ID: "ter-1", Name: "", TeamID: testTeamID}
	_, err := svc.Update(context.Background(), adminUser(), ter)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Fatalf(fmtExpectedInvalidInput, err)
	}
}

func TestTerritory_Update_ManagerOtherTeamForbidden(t *testing.T) {
	t.Parallel()
	ter := sampleTerritory()
	ter.TeamID = "team-99"
	repo := &stubTerritoryRepo{territory: ter}
	svc := service.NewTerritoryService(repo)

	update := &domain.Territory{ID: "ter-1", Name: "Updated", TeamID: testTeam99ID}
	_, err := svc.Update(context.Background(), managerUser(), update)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Update_GetError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{getErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	ter := &domain.Territory{ID: "ter-1", Name: "Updated", TeamID: testTeamID}
	_, err := svc.Update(context.Background(), adminUser(), ter)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Delete tests ---

func TestTerritory_Delete_AdminSuccess(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	err := svc.Delete(context.Background(), adminUser(), "ter-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if !repo.deleted {
		t.Error("expected territory to be deleted")
	}
}

func TestTerritory_Delete_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory()}
	svc := service.NewTerritoryService(repo)

	err := svc.Delete(context.Background(), repUser(), "ter-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Delete_ManagerOtherTeamForbidden(t *testing.T) {
	t.Parallel()
	ter := sampleTerritory()
	ter.TeamID = "team-99"
	repo := &stubTerritoryRepo{territory: ter}
	svc := service.NewTerritoryService(repo)

	err := svc.Delete(context.Background(), managerUser(), "ter-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf(fmtExpectedForbidden, err)
	}
}

func TestTerritory_Delete_GetError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{getErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	err := svc.Delete(context.Background(), adminUser(), "ter-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

func TestTerritory_Delete_DeleteError(t *testing.T) {
	t.Parallel()
	repo := &stubTerritoryRepo{territory: sampleTerritory(), deleteErr: errors.New(errDBMsg)}
	svc := service.NewTerritoryService(repo)

	err := svc.Delete(context.Background(), adminUser(), "ter-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}
