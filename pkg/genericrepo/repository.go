// Package genericrepo provides a generic repository pattern implementation
// using Go 1.23+ generics and range over function for efficient data streaming.
package genericrepo

import (
	"context"
	"iter"
)

// Entity defines the interface that all domain entities must implement.
type Entity interface {
	comparable
	GetID() string
	Validate() error
}

// Repository defines the generic repository interface with common CRUD operations.
type Repository[T Entity] interface {
	// Create inserts a new entity into the repository.
	Create(ctx context.Context, entity T) error

	// FindByID retrieves an entity by its ID.
	FindByID(ctx context.Context, id string) (T, error)

	// Update modifies an existing entity.
	Update(ctx context.Context, entity T) error

	// Delete removes an entity by its ID.
	Delete(ctx context.Context, id string) error

	// FindAll retrieves entities with optional filtering and pagination.
	FindAll(ctx context.Context, opts ...QueryOption) ([]T, error)

	// Count returns the total number of entities matching the criteria.
	Count(ctx context.Context, opts ...QueryOption) (int64, error)

	// Stream returns an iterator for memory-efficient traversal of large datasets.
	// Uses Go 1.23+ range over function feature.
	Stream(ctx context.Context, opts ...QueryOption) iter.Seq[T]

	// StreamWithError returns an iterator that also yields errors.
	StreamWithError(ctx context.Context, opts ...QueryOption) iter.Seq2[T, error]

	// Exists checks if an entity with the given ID exists.
	Exists(ctx context.Context, id string) (bool, error)
}

// QueryConfig holds configuration for repository queries.
type QueryConfig struct {
	Filters  map[string]any
	OrderBy  string
	Limit    int
	Offset   int
	Preloads []string
}

// QueryOption is a functional option for configuring queries.
type QueryOption func(*QueryConfig)

// WithFilter adds a filter condition.
func WithFilter(key string, value any) QueryOption {
	return func(c *QueryConfig) {
		if c.Filters == nil {
			c.Filters = make(map[string]any)
		}
		c.Filters[key] = value
	}
}

// WithFilters adds multiple filter conditions.
func WithFilters(filters map[string]any) QueryOption {
	return func(c *QueryConfig) {
		if c.Filters == nil {
			c.Filters = make(map[string]any)
		}
		for k, v := range filters {
			c.Filters[k] = v
		}
	}
}

// WithOrderBy sets the ordering of results.
func WithOrderBy(orderBy string) QueryOption {
	return func(c *QueryConfig) {
		c.OrderBy = orderBy
	}
}

// WithLimit sets the maximum number of results.
func WithLimit(limit int) QueryOption {
	return func(c *QueryConfig) {
		c.Limit = limit
	}
}

// WithOffset sets the starting offset for results.
func WithOffset(offset int) QueryOption {
	return func(c *QueryConfig) {
		c.Offset = offset
	}
}

// WithPagination is a convenience function to set both limit and offset.
func WithPagination(page, pageSize int) QueryOption {
	return func(c *QueryConfig) {
		c.Limit = pageSize
		c.Offset = (page - 1) * pageSize
	}
}

// WithPreload specifies relations to preload.
func WithPreload(relations ...string) QueryOption {
	return func(c *QueryConfig) {
		c.Preloads = append(c.Preloads, relations...)
	}
}

// BuildQueryConfig creates a QueryConfig from options.
func BuildQueryConfig(opts ...QueryOption) *QueryConfig {
	cfg := &QueryConfig{
		Filters: make(map[string]any),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Pagination represents pagination metadata.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// CalculateTotalPages calculates the total number of pages.
func (p *Pagination) CalculateTotalPages() {
	if p.PageSize > 0 {
		p.TotalPages = int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
}

// PagedResult represents a paginated query result.
type PagedResult[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// NewPagedResult creates a new paginated result.
func NewPagedResult[T any](items []T, page, pageSize int, total int64) PagedResult[T] {
	pagination := Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	pagination.CalculateTotalPages()

	return PagedResult[T]{
		Items:      items,
		Pagination: pagination,
	}
}
