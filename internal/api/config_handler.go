package api

import (
	"net/http"

	"github.com/pebblr/pebblr/internal/config"
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
	if actor := requireActor(w, r); actor == nil {
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	writeJSON(w, r, h.cfg)
}
