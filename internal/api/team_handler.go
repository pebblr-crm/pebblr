package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// TeamServicer is the interface the TeamHandler depends on for business logic.
type TeamServicer interface {
	List(ctx context.Context) ([]*domain.Team, error)
	Get(ctx context.Context, id string) (*domain.Team, error)
	ListMembers(ctx context.Context, teamID string) ([]*domain.User, error)
}

// TeamHandler handles HTTP requests for team read operations.
type TeamHandler struct {
	svc TeamServicer
}

// NewTeamHandler constructs a TeamHandler backed by the given service.
func NewTeamHandler(svc TeamServicer) *TeamHandler {
	return &TeamHandler{svc: svc}
}

// NewTeamRouter returns an http.Handler with all team sub-routes mounted.
// Mount at /api/v1/teams in the main router.
func NewTeamRouter(h *TeamHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	return r
}

// teamDetailResponse is the JSON envelope for a team with its members.
type teamDetailResponse struct {
	Team    *domain.Team   `json:"team"`
	Members []*domain.User `json:"members"`
}

// teamListResponse is the JSON envelope for a team list.
type teamListResponse struct {
	Items []*domain.Team `json:"items"`
	Total int            `json:"total"`
}

func mapTeamServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
}

// List handles GET /api/v1/teams
func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	teams, err := h.svc.List(r.Context())
	if err != nil {
		mapTeamServiceError(w, err)
		return
	}

	if teams == nil {
		teams = []*domain.Team{}
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, teamListResponse{Items: teams, Total: len(teams)})
}

// Get handles GET /api/v1/teams/{id}
func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", errMissingUser)
		return
	}

	id := chi.URLParam(r, "id")
	team, err := h.svc.Get(r.Context(), id)
	if err != nil {
		mapTeamServiceError(w, err)
		return
	}

	members, err := h.svc.ListMembers(r.Context(), id)
	if err != nil {
		mapTeamServiceError(w, err)
		return
	}
	if members == nil {
		members = []*domain.User{}
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, teamDetailResponse{Team: team, Members: members})
}
