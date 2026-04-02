package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// UserServicer is the interface the UserHandler depends on for business logic.
type UserServicer interface {
	List(ctx context.Context, actor *domain.User) ([]*domain.User, error)
	Get(ctx context.Context, actor *domain.User, id string) (*domain.User, error)
}

// UserHandler handles HTTP requests for user read operations.
type UserHandler struct {
	svc UserServicer
}

// NewUserHandler constructs a UserHandler backed by the given service.
func NewUserHandler(svc UserServicer) *UserHandler {
	return &UserHandler{svc: svc}
}

// NewUserRouter returns an http.Handler with all user sub-routes mounted.
// Mount at /api/v1/users in the main router.
func NewUserRouter(h *UserHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	return r
}

// userResponse is the JSON envelope for a single user.
type userResponse struct {
	User *domain.User `json:"user"`
}

// userListResponse is the JSON envelope for a user list.
type userListResponse struct {
	Items []*domain.User `json:"items"`
	Total int            `json:"total"`
}

func mapUserServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", errUnexpected)
}

// List handles GET /api/v1/users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	users, err := h.svc.List(r.Context(), actor)
	if err != nil {
		mapUserServiceError(w, err)
		return
	}

	if users == nil {
		users = []*domain.User{}
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, userListResponse{Items: users, Total: len(users)})
}

// Get handles GET /api/v1/users/{id}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor := requireActor(w, r)
	if actor == nil {
		return
	}

	id := chi.URLParam(r, "id")
	user, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		mapUserServiceError(w, err)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, userResponse{User: user})
}
