package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

type leadRepository struct {
	pool *pgxpool.Pool
}

const leadColumns = `
	id, title, description, status,
	COALESCE(assignee_id::TEXT, ''), team_id::TEXT, customer_id::TEXT, customer_type,
	company, industry, location, value_cents, initials,
	created_at, updated_at, deleted_at`

func scanLead(row pgx.Row) (*domain.Lead, error) {
	var l domain.Lead
	var deletedAt *time.Time
	err := row.Scan(
		&l.ID, &l.Title, &l.Description, &l.Status,
		&l.AssigneeID, &l.TeamID, &l.CustomerID, &l.CustomerType,
		&l.Company, &l.Industry, &l.Location, &l.ValueCents, &l.Initials,
		&l.CreatedAt, &l.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("scanning lead: %w", err)
	}
	l.DeletedAt = deletedAt
	return &l, nil
}

func (r *leadRepository) Get(ctx context.Context, id string) (*domain.Lead, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+leadColumns+` FROM leads WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	return scanLead(row)
}

func (r *leadRepository) List(ctx context.Context, scope rbac.LeadScope, filter store.LeadFilter, page, limit int) (*store.LeadPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{}
	argIdx := 1

	var conditions []string
	conditions = append(conditions, "deleted_at IS NULL")

	// RBAC scope
	if !scope.AllLeads {
		var scopeParts []string
		if len(scope.AssigneeIDs) > 0 {
			placeholders := make([]string, len(scope.AssigneeIDs))
			for i, id := range scope.AssigneeIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("assignee_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		}
		if len(scope.TeamIDs) > 0 {
			placeholders := make([]string, len(scope.TeamIDs))
			for i, id := range scope.TeamIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, id)
				argIdx++
			}
			scopeParts = append(scopeParts, fmt.Sprintf("team_id::TEXT = ANY(ARRAY[%s])", strings.Join(placeholders, ",")))
		}
		if len(scopeParts) > 0 {
			conditions = append(conditions, "("+strings.Join(scopeParts, " OR ")+")")
		} else {
			// Impossible scope — return empty
			return &store.LeadPage{Leads: []*domain.Lead{}, Total: 0, Page: page, Limit: limit}, nil
		}
	}

	// Filters
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.Assignee != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id::TEXT = $%d", argIdx))
		args = append(args, *filter.Assignee)
		argIdx++
	}
	if filter.Team != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.Team)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM leads ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting leads: %w", err)
	}

	listQuery := `SELECT ` + leadColumns + ` FROM leads ` + where +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying leads: %w", err)
	}
	defer rows.Close()

	var leads []*domain.Lead
	for rows.Next() {
		lead, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		leads = append(leads, lead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating leads: %w", err)
	}

	return &store.LeadPage{Leads: leads, Total: total, Page: page, Limit: limit}, nil
}

func (r *leadRepository) Create(ctx context.Context, lead *domain.Lead) (*domain.Lead, error) {
	var assigneeID *string
	if lead.AssigneeID != "" {
		assigneeID = &lead.AssigneeID
	}

	row := r.pool.QueryRow(ctx,
		`INSERT INTO leads (title, description, status, assignee_id, team_id, customer_id, customer_type,
		                    company, industry, location, value_cents, initials)
		 VALUES ($1, $2, $3, $4::UUID, $5::UUID, $6::UUID, $7, $8, $9, $10, $11, $12)
		 RETURNING `+leadColumns,
		lead.Title, lead.Description, string(lead.Status),
		assigneeID, lead.TeamID, lead.CustomerID, lead.CustomerType,
		lead.Company, lead.Industry, lead.Location, lead.ValueCents, lead.Initials,
	)
	return scanLead(row)
}

func (r *leadRepository) Update(ctx context.Context, lead *domain.Lead) (*domain.Lead, error) {
	var assigneeID *string
	if lead.AssigneeID != "" {
		assigneeID = &lead.AssigneeID
	}

	row := r.pool.QueryRow(ctx,
		`UPDATE leads
		 SET title = $1, description = $2, status = $3,
		     assignee_id = $4::UUID, team_id = $5::UUID,
		     customer_id = $6::UUID, customer_type = $7,
		     company = $8, industry = $9, location = $10, value_cents = $11, initials = $12,
		     updated_at = NOW()
		 WHERE id = $13::UUID AND deleted_at IS NULL
		 RETURNING `+leadColumns,
		lead.Title, lead.Description, string(lead.Status),
		assigneeID, lead.TeamID, lead.CustomerID, lead.CustomerType,
		lead.Company, lead.Industry, lead.Location, lead.ValueCents, lead.Initials,
		lead.ID,
	)
	return scanLead(row)
}

func (r *leadRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE leads SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1::UUID AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting lead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}
