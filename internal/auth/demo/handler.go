package demo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pebblr/pebblr/internal/domain"
)

// Handler serves demo authentication endpoints.
type Handler struct {
	auth     *Authenticator
	personas []Persona
}

// NewHandler creates an HTTP handler for demo token issuance.
// The provided personas define the available demo identities.
func NewHandler(auth *Authenticator, personas []Persona) *Handler {
	return &Handler{auth: auth, personas: personas}
}

// tokenRequest is the JSON body for POST /demo/token.
type tokenRequest struct {
	PersonaID string `json:"persona_id"`
}

// tokenResponse is the JSON response for POST /demo/token.
type tokenResponse struct {
	Token   string  `json:"token"`
	Persona Persona `json:"persona"`
}

// ListPersonas returns the available demo personas.
// GET /demo/personas
func (h *Handler) ListPersonas(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.personas)
}

// IssueToken issues a JWT for the requested persona.
// POST /demo/token
func (h *Handler) IssueToken(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	persona, ok := h.findPersona(req.PersonaID)
	if !ok {
		writeError(w, http.StatusBadRequest, "INVALID_PERSONA", fmt.Sprintf("unknown persona %q", req.PersonaID))
		return
	}

	token, err := h.auth.IssueToken(persona)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to issue token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenResponse{
		Token:   token,
		Persona: persona,
	})
}

func (h *Handler) findPersona(id string) (Persona, bool) {
	for _, p := range h.personas {
		if p.ID == id {
			return p, true
		}
	}
	return Persona{}, false
}

// DefaultPersonas returns a standard set of demo personas.
func DefaultPersonas() []Persona {
	return []Persona{
		{
			ID:    "demo-rep",
			Email: "rep@demo.pebblr.com",
			Name:  "Riley Rep",
			Role:  domain.RoleRep,
		},
		{
			ID:    "demo-manager",
			Email: "manager@demo.pebblr.com",
			Name:  "Morgan Manager",
			Role:  domain.RoleManager,
		},
		{
			ID:    "demo-admin",
			Email: "admin@demo.pebblr.com",
			Name:  "Alex Admin",
			Role:  domain.RoleAdmin,
		},
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]map[string]string{
		"error": {"code": code, "message": message},
	})
}
