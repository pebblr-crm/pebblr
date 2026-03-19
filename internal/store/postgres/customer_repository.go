package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type customerRepository struct {
	pool *pgxpool.Pool
}

const customerColumns = `
	id::TEXT, name, customer_type,
	COALESCE(street, ''), COALESCE(city, ''), COALESCE(state, ''), COALESCE(country, ''), COALESCE(zip, ''),
	COALESCE(phone, ''), COALESCE(email, ''), COALESCE(notes, ''),
	created_at, updated_at`

func scanCustomer(row pgx.Row) (*domain.Customer, error) {
	var c domain.Customer
	err := row.Scan(
		&c.ID, &c.Name, &c.Type,
		&c.Address.Street, &c.Address.City, &c.Address.State, &c.Address.Country, &c.Address.Zip,
		&c.Phone, &c.Email, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("scanning customer: %w", err)
	}
	return &c, nil
}

func (r *customerRepository) Get(ctx context.Context, id string) (*domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+customerColumns+` FROM customers WHERE id = $1::UUID`,
		id,
	)
	return scanCustomer(row)
}

func (r *customerRepository) List(ctx context.Context, filter store.CustomerFilter, page, limit int) (*store.CustomerPage, error) {
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

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("customer_type = $%d", argIdx))
		args = append(args, string(*filter.Type))
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM customers ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting customers: %w", err)
	}

	listQuery := `SELECT ` + customerColumns + ` FROM customers ` + where +
		fmt.Sprintf(` ORDER BY name ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		c, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating customers: %w", err)
	}

	return &store.CustomerPage{Customers: customers, Total: total, Page: page, Limit: limit}, nil
}

func (r *customerRepository) Create(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO customers (name, customer_type, street, city, state, country, zip, phone, email, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING `+customerColumns,
		c.Name, string(c.Type),
		nullIfEmpty(c.Address.Street), nullIfEmpty(c.Address.City),
		nullIfEmpty(c.Address.State), nullIfEmpty(c.Address.Country),
		nullIfEmpty(c.Address.Zip), nullIfEmpty(c.Phone),
		nullIfEmpty(c.Email), nullIfEmpty(c.Notes),
	)
	return scanCustomer(row)
}

func (r *customerRepository) Update(ctx context.Context, c *domain.Customer) (*domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE customers
		 SET name = $1, customer_type = $2,
		     street = $3, city = $4, state = $5, country = $6, zip = $7,
		     phone = $8, email = $9, notes = $10,
		     updated_at = NOW()
		 WHERE id = $11::UUID
		 RETURNING `+customerColumns,
		c.Name, string(c.Type),
		nullIfEmpty(c.Address.Street), nullIfEmpty(c.Address.City),
		nullIfEmpty(c.Address.State), nullIfEmpty(c.Address.Country),
		nullIfEmpty(c.Address.Zip), nullIfEmpty(c.Phone),
		nullIfEmpty(c.Email), nullIfEmpty(c.Notes),
		c.ID,
	)
	return scanCustomer(row)
}

// ListLeads returns non-deleted leads associated with the given customer ID.
func (r *customerRepository) ListLeads(ctx context.Context, customerID string) ([]*domain.Lead, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+leadColumns+` FROM leads WHERE customer_id = $1::UUID AND deleted_at IS NULL ORDER BY created_at DESC`,
		customerID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying customer leads: %w", err)
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
		return nil, fmt.Errorf("iterating customer leads: %w", err)
	}
	return leads, nil
}

// nullIfEmpty converts an empty string to nil (for nullable TEXT columns).
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
