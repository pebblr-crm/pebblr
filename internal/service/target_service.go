package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/geo"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// geocodeConcurrency is the maximum number of concurrent geocoding requests
// during a target import. This prevents flooding the upstream API while still
// completing much faster than sequential geocoding.
const geocodeConcurrency = 5

const errGettingTarget = "getting target: %w"

// TargetService handles target business logic with RBAC enforcement.
type TargetService struct {
	targets  store.TargetRepository
	users    store.UserRepository
	audit    store.AuditRepository
	enforcer rbac.Enforcer
	cfg      *config.TenantConfig
	geocoder geo.Geocoder // optional; nil skips geocoding
}

// NewTargetService constructs a TargetService with the given dependencies.
func NewTargetService(targets store.TargetRepository, enforcer rbac.Enforcer, cfg *config.TenantConfig, opts ...TargetServiceOption) *TargetService {
	s := &TargetService{targets: targets, enforcer: enforcer, cfg: cfg}
	for _, o := range opts {
		o(s)
	}
	return s
}

// TargetServiceOption configures optional dependencies on TargetService.
type TargetServiceOption func(*TargetService)

// WithGeocoder sets the geocoder used to enrich targets during import.
func WithGeocoder(g geo.Geocoder) TargetServiceOption {
	return func(s *TargetService) { s.geocoder = g }
}

// WithUsers sets the user repository for assignment validation.
func WithUsers(u store.UserRepository) TargetServiceOption {
	return func(s *TargetService) { s.users = u }
}

// WithAudit sets the audit repository for recording assignment changes.
func WithAudit(a store.AuditRepository) TargetServiceOption {
	return func(s *TargetService) { s.audit = a }
}

// Create persists a new target. Only managers and admins may create targets.
func (s *TargetService) Create(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if err := s.validateTarget(target); err != nil {
		return nil, err
	}

	created, err := s.targets.Create(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("creating target: %w", err)
	}
	return created, nil
}

// Get retrieves a target by ID with RBAC enforcement.
func (s *TargetService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Target, error) {
	target, err := s.targets.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(errGettingTarget, err)
	}
	if !s.enforcer.CanViewTarget(ctx, actor, target) {
		return nil, ErrForbidden
	}
	return target, nil
}

// List returns a paginated list of targets scoped to the actor's permissions.
func (s *TargetService) List(ctx context.Context, actor *domain.User, filter store.TargetFilter, page, limit int) (*store.TargetPage, error) {
	scope := s.enforcer.ScopeTargetQuery(ctx, actor)
	result, err := s.targets.List(ctx, scope, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing targets: %w", err)
	}
	return result, nil
}

// Update persists changes to an existing target. Reps can only update editable fields
// on their own targets; managers/admins can update any field.
func (s *TargetService) Update(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	existing, err := s.targets.Get(ctx, target.ID)
	if err != nil {
		return nil, fmt.Errorf(errGettingTarget, err)
	}
	if !s.enforcer.CanUpdateTarget(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if err := s.validateTarget(target); err != nil {
		return nil, err
	}

	updated, err := s.targets.Update(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("updating target: %w", err)
	}
	return updated, nil
}

// Assign updates the assignee (and optionally team) of a target.
// Only managers (for their teams) and admins may assign targets.
func (s *TargetService) Assign(ctx context.Context, actor *domain.User, targetID, assigneeID, teamID string) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if assigneeID == "" {
		return nil, ErrInvalidInput
	}

	existing, err := s.targets.Get(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf(errGettingTarget, err)
	}
	if !s.enforcer.CanUpdateTarget(ctx, actor, existing) {
		return nil, ErrForbidden
	}

	// Validate that the assignee user exists.
	if s.users != nil {
		if _, err := s.users.GetByID(ctx, assigneeID); err != nil {
			return nil, fmt.Errorf("validating assignee: %w", err)
		}
	}

	oldAssignee := existing.AssigneeID
	existing.AssigneeID = assigneeID
	if teamID != "" {
		existing.TeamID = teamID
	}

	updated, err := s.targets.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("assigning target: %w", err)
	}

	if s.audit != nil {
		_ = s.audit.Record(ctx, &domain.AuditEntry{
			EntityType: "target",
			EntityID:   targetID,
			EventType:  "assigned",
			ActorID:    actor.ID,
			OldValue:   map[string]any{"assigneeId": oldAssignee},
			NewValue:   map[string]any{"assigneeId": assigneeID},
		})
	}

	return updated, nil
}

// Import bulk-upserts targets by external ID. Admin-only.
// When a geocoder is configured, targets without lat/lng are geocoded from their address fields.
func (s *TargetService) Import(ctx context.Context, actor *domain.User, targets []*domain.Target) (*store.ImportResult, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}
	for i, t := range targets {
		if t.ExternalID == "" {
			return nil, fmt.Errorf("target at index %d: %w: externalId is required", i, ErrInvalidInput)
		}
		if err := s.validateTarget(t); err != nil {
			return nil, fmt.Errorf("target at index %d: %w", i, err)
		}
	}

	// Geocode targets that have an address but no coordinates.
	if s.geocoder != nil {
		s.geocodeTargets(ctx, targets)
	}

	result, err := s.targets.Upsert(ctx, targets)
	if err != nil {
		return nil, fmt.Errorf("importing targets: %w", err)
	}
	return result, nil
}

// geocodeTargets enriches targets with lat/lng from their address fields.
// Geocoding runs concurrently (up to geocodeConcurrency goroutines) so that
// large imports are not blocked by sequential API calls. Individual failures
// are logged but do not block the import.
func (s *TargetService) geocodeTargets(ctx context.Context, targets []*domain.Target) {
	// Collect targets that need geocoding.
	type geocodeJob struct {
		target *domain.Target
		addr   string
	}
	var jobs []geocodeJob
	for _, t := range targets {
		if t.Fields == nil {
			continue
		}
		if _, hasLat := t.Fields["lat"]; hasLat {
			continue
		}
		addr := buildAddress(t.Fields)
		if addr == "" {
			continue
		}
		jobs = append(jobs, geocodeJob{target: t, addr: addr})
	}
	if len(jobs) == 0 {
		return
	}

	sem := make(chan struct{}, geocodeConcurrency)
	var wg sync.WaitGroup

	for i := range jobs {
		job := jobs[i]
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release

			result, err := s.geocoder.Geocode(ctx, job.addr)
			if err != nil {
				slog.Warn("geocoding failed, skipping", "target", job.target.Name, "address", job.addr, "err", err)
				return
			}
			// Each goroutine writes to its own target's Fields map — no shared
			// state across goroutines — so no mutex needed.
			job.target.Fields["lat"] = result.Lat
			job.target.Fields["lng"] = result.Lng
			job.target.Fields["formatted_address"] = result.FormattedAddress
		}()
	}
	wg.Wait()
}

// buildAddress assembles a geocodable address string from target fields.
func buildAddress(fields map[string]any) string {
	addr, _ := fields["address"].(string)
	city, _ := fields["city"].(string)
	county, _ := fields["county"].(string)
	if addr == "" && city == "" {
		return ""
	}
	s := addr
	if city != "" {
		if s != "" {
			s += ", "
		}
		s += city
	}
	if county != "" && county != city {
		s += ", " + county
	}
	return s
}

// VisitStatus returns the last visit date for each of the actor's targets.
func (s *TargetService) VisitStatus(ctx context.Context, actor *domain.User) ([]store.TargetVisitStatus, error) {
	scope := s.enforcer.ScopeTargetQuery(ctx, actor)
	fieldTypes := s.fieldActivityTypes()
	result, err := s.targets.VisitStatus(ctx, scope, fieldTypes)
	if err != nil {
		return nil, fmt.Errorf("querying visit status: %w", err)
	}
	return result, nil
}

// TargetFrequencyItem holds per-target frequency compliance for the API response.
type TargetFrequencyItem struct {
	TargetID       string  `json:"targetId"`
	Classification string  `json:"classification"`
	VisitCount     int     `json:"visitCount"`
	Required       int     `json:"required"`
	Compliance     float64 `json:"compliance"`
}

// FrequencyStatus returns per-target visit compliance for the given period.
func (s *TargetService) FrequencyStatus(ctx context.Context, actor *domain.User, dateFrom, dateTo time.Time) ([]TargetFrequencyItem, error) {
	scope := s.enforcer.ScopeTargetQuery(ctx, actor)
	fieldTypes := s.fieldActivityTypes()

	rows, err := s.targets.FrequencyStatus(ctx, scope, fieldTypes, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("querying frequency status: %w", err)
	}

	months := frequencyMonths(dateFrom, dateTo)
	items := make([]TargetFrequencyItem, 0, len(rows))
	for _, row := range rows {
		required := 0
		if s.cfg != nil {
			required = s.cfg.Rules.Frequency[row.Classification]
		}
		expected := required * months
		var compliance float64
		if expected > 0 {
			compliance = float64(row.VisitCount) / float64(expected) * 100
			if compliance > 100 {
				compliance = 100
			}
		}
		items = append(items, TargetFrequencyItem{
			TargetID:       row.TargetID,
			Classification: row.Classification,
			VisitCount:     row.VisitCount,
			Required:       required,
			Compliance:     compliance,
		})
	}
	return items, nil
}

// frequencyMonths returns the number of calendar months spanned by the date range (minimum 1).
func frequencyMonths(from, to time.Time) int {
	if to.Before(from) {
		return 1
	}
	months := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month()) + 1
	if months < 1 {
		return 1
	}
	return months
}

// fieldActivityTypes returns the keys of all field-category activity types from config.
func (s *TargetService) fieldActivityTypes() []string {
	if s.cfg == nil {
		return []string{"visit"}
	}
	var types []string
	for i := range s.cfg.Activities.Types {
		if s.cfg.Activities.Types[i].Category == "field" {
			types = append(types, s.cfg.Activities.Types[i].Key)
		}
	}
	if len(types) == 0 {
		types = []string{"visit"}
	}
	return types
}

// validateTarget checks that the target has a valid type and name.
func (s *TargetService) validateTarget(target *domain.Target) error {
	if target.Name == "" {
		return ErrInvalidInput
	}
	if s.cfg != nil {
		if s.cfg.AccountType(target.TargetType) == nil {
			return ErrInvalidInput
		}
	}
	return nil
}
