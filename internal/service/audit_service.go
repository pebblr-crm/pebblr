package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// AuditService handles audit log business logic. Only admins can list and review.
type AuditService struct {
	audit store.AuditRepository
}

// NewAuditService constructs an AuditService.
func NewAuditService(audit store.AuditRepository) *AuditService {
	return &AuditService{audit: audit}
}

// List returns paginated audit entries. Admin only.
func (s *AuditService) List(ctx context.Context, actor *domain.User, filter store.AuditFilter) ([]*domain.AuditEntry, int, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, 0, ErrForbidden
	}
	entries, total, err := s.audit.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit entries: %w", err)
	}
	if entries == nil {
		entries = []*domain.AuditEntry{}
	}
	return entries, total, nil
}

// UpdateStatus updates the review status of an audit entry. Admin only.
func (s *AuditService) UpdateStatus(ctx context.Context, actor *domain.User, id, status string) error {
	if actor.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	validStatuses := map[string]bool{"pending": true, "accepted": true, "false_positive": true}
	if !validStatuses[status] {
		return ErrInvalidInput
	}

	if err := s.audit.UpdateStatus(ctx, id, status, actor.ID); err != nil {
		return fmt.Errorf("updating audit status: %w", err)
	}
	return nil
}
