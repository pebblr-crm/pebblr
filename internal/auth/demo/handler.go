package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
)

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

// UserLister retrieves the set of users available for demo login.
type UserLister interface {
	List(ctx context.Context) ([]*domain.User, error)
}

// Handler serves demo authentication endpoints.
type Handler struct {
	auth  *Authenticator
	users UserLister
}

// NewHandler creates an HTTP handler for demo token issuance.
// The UserLister provides real user accounts from the database so that
// demo tokens map to actual user rows (and therefore real assignments/RBAC).
func NewHandler(auth *Authenticator, users UserLister) *Handler {
	return &Handler{auth: auth, users: users}
}

// Account is the JSON representation of a demo-selectable user.
type Account struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	Avatar string      `json:"avatar,omitempty"`
}

// tokenRequest is the JSON body for POST /demo/token.
type tokenRequest struct {
	UserID string `json:"user_id"`
}

// tokenResponse is the JSON response for POST /demo/token.
type tokenResponse struct {
	Token   string  `json:"token"`
	Account Account `json:"account"`
}

// ListAccounts returns the real user accounts available for demo login.
// GET /demo/accounts
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_ERROR", "failed to list accounts")
		return
	}

	accounts := make([]Account, len(users))
	for i, u := range users {
		accounts[i] = userToAccount(u)
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(accounts)
}

// IssueToken issues a JWT for the requested user account.
// POST /demo/token
func (h *Handler) IssueToken(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	users, err := h.users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_ERROR", "failed to list accounts")
		return
	}

	var user *domain.User
	for _, u := range users {
		if u.ID == req.UserID {
			user = u
			break
		}
	}
	if user == nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER", fmt.Sprintf("unknown user %q", req.UserID))
		return
	}

	token, err := h.auth.IssueToken(Persona{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to issue token")
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(tokenResponse{
		Token:   token,
		Account: userToAccount(user),
	})
}

func userToAccount(u *domain.User) Account {
	return Account{
		ID:     u.ID,
		Name:   u.Name,
		Email:  u.Email,
		Role:   u.Role,
		Avatar: u.Avatar,
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]map[string]string{
		"error": {"code": code, "message": message},
	})
}
