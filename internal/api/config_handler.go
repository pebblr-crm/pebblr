package api

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/rbac"
)

// ConfigHandler handles HTTP requests for tenant configuration.
type ConfigHandler struct {
	cfg *config.TenantConfig
}

// NewConfigHandler constructs a ConfigHandler with the given tenant config.
func NewConfigHandler(cfg *config.TenantConfig) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

// Get handles GET /api/v1/config
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, h.cfg)
}
