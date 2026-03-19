package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// CustomerServicer is the interface the CustomerHandler depends on for business logic.
// Defined here so handler tests can provide stubs without a real database.
type CustomerServicer interface {
	Create(ctx context.Context, actor *domain.User, customer *domain.Customer) (*domain.Customer, error)
	Get(ctx context.Context, id string) (*service.CustomerDetail, error)
	List(ctx context.Context, filter store.CustomerFilter, page, limit int) (*store.CustomerPage, error)
	Update(ctx context.Context, actor *domain.User, customer *domain.Customer) (*domain.Customer, error)
}

// CustomerHandler handles HTTP requests for customer CRUD operations.
type CustomerHandler struct {
	svc CustomerServicer
}

// NewCustomerHandler constructs a CustomerHandler backed by the given service.
func NewCustomerHandler(svc CustomerServicer) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// NewCustomerRouter returns an http.Handler with all customer sub-routes mounted.
// Mount at /api/v1/customers in the main router.
func NewCustomerRouter(h *CustomerHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	return r
}

// customerRequest is the body for create and update operations.
type customerRequest struct {
	Name    string              `json:"name"`
	Type    domain.CustomerType `json:"type"`
	Address domain.Address      `json:"address"`
	Phone   string              `json:"phone"`
	Email   string              `json:"email"`
	Notes   string              `json:"notes"`
}

// customerResponse is the JSON envelope for a single customer.
type customerResponse struct {
	Customer *domain.Customer `json:"customer"`
}

// customerDetailResponse is the JSON envelope for a customer with associated leads.
type customerDetailResponse struct {
	Customer *domain.Customer `json:"customer"`
	Leads    []*domain.Lead   `json:"leads"`
}

// customerListResponse is the JSON envelope for a paginated customer list.
type customerListResponse struct {
	Items []*domain.Customer `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

func mapCustomerServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "customer not found")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

// List handles GET /api/v1/customers
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	var filter store.CustomerFilter
	if t := r.URL.Query().Get("type"); t != "" {
		ct := domain.CustomerType(t)
		if !ct.Valid() {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid type filter")
			return
		}
		filter.Type = &ct
	}

	result, err := h.svc.List(r.Context(), filter, page, limit)
	if err != nil {
		mapCustomerServiceError(w, err)
		return
	}

	customers := result.Customers
	if customers == nil {
		customers = []*domain.Customer{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(customerListResponse{
		Items: customers,
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

// Create handles POST /api/v1/customers
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	var req customerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	customer := &domain.Customer{
		Name:    req.Name,
		Type:    req.Type,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		Notes:   req.Notes,
	}

	created, err := h.svc.Create(r.Context(), actor, customer)
	if err != nil {
		mapCustomerServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(customerResponse{Customer: created})
}

// Get handles GET /api/v1/customers/{id}
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")
	detail, err := h.svc.Get(r.Context(), id)
	if err != nil {
		mapCustomerServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(customerDetailResponse{
		Customer: detail.Customer,
		Leads:    detail.Leads,
	})
}

// Update handles PUT /api/v1/customers/{id}
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, err := rbac.UserFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated user")
		return
	}

	id := chi.URLParam(r, "id")

	var req customerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	customer := &domain.Customer{
		ID:      id,
		Name:    req.Name,
		Type:    req.Type,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		Notes:   req.Notes,
	}

	updated, err := h.svc.Update(r.Context(), actor, customer)
	if err != nil {
		mapCustomerServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(customerResponse{Customer: updated})
}
