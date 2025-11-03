package postgres

import (
	"context"
	"errors"
	"iter"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
	"github.com/tyokyo320/rateflow/pkg/timeutil"
)

// RateRepository implements rate.Repository interface.
type RateRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewRateRepository creates a new PostgreSQL rate repository.
func NewRateRepository(db *gorm.DB, logger *slog.Logger) rate.Repository {
	return &RateRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new rate into the database.
// If a rate with the same (base, quote, date, source) exists, it updates the existing rate.
func (r *RateRepository) Create(ctx context.Context, entity *rate.Rate) error {
	model := r.domainToModel(entity)

	// Use ON CONFLICT DO UPDATE to handle duplicates gracefully
	// This ensures idempotent behavior when re-running fetch commands
	result := r.db.WithContext(ctx).
		Where(&RateModel{
			BaseCurrency:  model.BaseCurrency,
			QuoteCurrency: model.QuoteCurrency,
			EffectiveDate: model.EffectiveDate,
			Source:        model.Source,
		}).
		Assign(&RateModel{
			Value:     model.Value,
			UpdatedAt: time.Now(),
		}).
		FirstOrCreate(model)

	return result.Error
}

// FindByID retrieves a rate by its ID.
func (r *RateRepository) FindByID(ctx context.Context, id string) (*rate.Rate, error) {
	var model RateModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, rate.ErrRateNotFound{ID: id}
		}
		return nil, err
	}

	return r.modelToDomain(&model)
}

// Update modifies an existing rate.
func (r *RateRepository) Update(ctx context.Context, entity *rate.Rate) error {
	model := r.domainToModel(entity)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a rate by its ID.
func (r *RateRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&RateModel{}, "id = ?", id).Error
}

// FindAll retrieves rates with optional filtering.
func (r *RateRepository) FindAll(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
	cfg := genericrepo.BuildQueryConfig(opts...)

	query := r.db.WithContext(ctx).Model(&RateModel{})

	// Apply filters
	for key, value := range cfg.Filters {
		query = query.Where(key+" = ?", value)
	}

	// Apply ordering
	if cfg.OrderBy != "" {
		query = query.Order(cfg.OrderBy)
	}

	// Apply pagination
	if cfg.Limit > 0 {
		query = query.Limit(cfg.Limit)
	}
	if cfg.Offset > 0 {
		query = query.Offset(cfg.Offset)
	}

	var models []RateModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	rates := make([]*rate.Rate, 0, len(models))
	for i := range models {
		domainRate, err := r.modelToDomain(&models[i])
		if err != nil {
			r.logger.Error("failed to convert model to domain", "error", err)
			continue
		}
		rates = append(rates, domainRate)
	}

	return rates, nil
}

// Count returns the total number of rates matching the criteria.
func (r *RateRepository) Count(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
	cfg := genericrepo.BuildQueryConfig(opts...)

	query := r.db.WithContext(ctx).Model(&RateModel{})

	for key, value := range cfg.Filters {
		query = query.Where(key+" = ?", value)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// Stream returns an iterator for memory-efficient traversal.
// Uses Go 1.23+ range over function feature.
func (r *RateRepository) Stream(ctx context.Context, opts ...genericrepo.QueryOption) iter.Seq[*rate.Rate] {
	return func(yield func(*rate.Rate) bool) {
		const batchSize = 100
		offset := 0
		cfg := genericrepo.BuildQueryConfig(opts...)

		for {
			query := r.db.WithContext(ctx).Model(&RateModel{}).
				Limit(batchSize).
				Offset(offset)

			// Apply filters
			for key, value := range cfg.Filters {
				query = query.Where(key+" = ?", value)
			}

			// Apply ordering
			if cfg.OrderBy != "" {
				query = query.Order(cfg.OrderBy)
			}

			var models []RateModel
			if err := query.Find(&models).Error; err != nil {
				r.logger.Error("stream error", "error", err)
				return
			}

			if len(models) == 0 {
				return
			}

			for i := range models {
				domainRate, err := r.modelToDomain(&models[i])
				if err != nil {
					r.logger.Error("failed to convert model", "error", err)
					continue
				}

				if !yield(domainRate) {
					return // Early termination
				}
			}

			offset += batchSize
		}
	}
}

// StreamWithError returns an iterator that also yields errors.
func (r *RateRepository) StreamWithError(ctx context.Context, opts ...genericrepo.QueryOption) iter.Seq2[*rate.Rate, error] {
	return func(yield func(*rate.Rate, error) bool) {
		const batchSize = 100
		offset := 0
		cfg := genericrepo.BuildQueryConfig(opts...)

		for {
			query := r.db.WithContext(ctx).Model(&RateModel{}).
				Limit(batchSize).
				Offset(offset)

			for key, value := range cfg.Filters {
				query = query.Where(key+" = ?", value)
			}

			if cfg.OrderBy != "" {
				query = query.Order(cfg.OrderBy)
			}

			var models []RateModel
			if err := query.Find(&models).Error; err != nil {
				var zero *rate.Rate
				yield(zero, err)
				return
			}

			if len(models) == 0 {
				return
			}

			for i := range models {
				domainRate, err := r.modelToDomain(&models[i])
				if !yield(domainRate, err) {
					return
				}
			}

			offset += batchSize
		}
	}
}

// Exists checks if a rate with the given ID exists.
func (r *RateRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RateModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindByPairAndDate finds a rate for a specific currency pair and date.
func (r *RateRepository) FindByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (*rate.Rate, error) {
	var model RateModel

	dateStr := timeutil.FormatDate(date)

	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND quote_currency = ? AND effective_date = ?",
			pair.Base().String(),
			pair.Quote().String(),
			dateStr,
		).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, rate.ErrRateNotFound{}
		}
		return nil, err
	}

	return r.modelToDomain(&model)
}

// FindLatest finds the most recent rate for a currency pair.
func (r *RateRepository) FindLatest(ctx context.Context, pair currency.Pair) (*rate.Rate, error) {
	var model RateModel

	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND quote_currency = ?",
			pair.Base().String(),
			pair.Quote().String(),
		).
		Order("effective_date DESC").
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, rate.ErrRateNotFound{}
		}
		return nil, err
	}

	return r.modelToDomain(&model)
}

// FindByDateRange finds rates for a currency pair within a date range.
func (r *RateRepository) FindByDateRange(ctx context.Context, pair currency.Pair, start, end time.Time) ([]*rate.Rate, error) {
	var models []RateModel

	startStr := timeutil.FormatDate(start)
	endStr := timeutil.FormatDate(end)

	err := r.db.WithContext(ctx).
		Where("base_currency = ? AND quote_currency = ? AND effective_date BETWEEN ? AND ?",
			pair.Base().String(),
			pair.Quote().String(),
			startStr,
			endStr,
		).
		Order("effective_date DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	rates := make([]*rate.Rate, 0, len(models))
	for i := range models {
		domainRate, err := r.modelToDomain(&models[i])
		if err != nil {
			r.logger.Error("failed to convert model", "error", err)
			continue
		}
		rates = append(rates, domainRate)
	}

	return rates, nil
}

// FindByPairs finds the latest rates for multiple currency pairs.
func (r *RateRepository) FindByPairs(ctx context.Context, pairs []currency.Pair) ([]*rate.Rate, error) {
	if len(pairs) == 0 {
		return []*rate.Rate{}, nil
	}

	var rates []*rate.Rate

	for _, pair := range pairs {
		latestRate, err := r.FindLatest(ctx, pair)
		if err != nil {
			// Log error but continue with other pairs
			r.logger.Warn("failed to find rate for pair",
				"pair", pair.String(),
				"error", err,
			)
			continue
		}
		rates = append(rates, latestRate)
	}

	return rates, nil
}

// ExistsByPairAndDate checks if a rate exists for a specific pair and date.
func (r *RateRepository) ExistsByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (bool, error) {
	var count int64

	dateStr := timeutil.FormatDate(date)

	err := r.db.WithContext(ctx).Model(&RateModel{}).
		Where("base_currency = ? AND quote_currency = ? AND effective_date = ?",
			pair.Base().String(),
			pair.Quote().String(),
			dateStr,
		).
		Count(&count).Error

	return count > 0, err
}

// DeleteOlderThan deletes rates older than the specified date.
func (r *RateRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	dateStr := timeutil.FormatDate(date)

	result := r.db.WithContext(ctx).
		Where("effective_date < ?", dateStr).
		Delete(&RateModel{})

	return result.RowsAffected, result.Error
}

// domainToModel converts a domain Rate entity to a database model.
func (r *RateRepository) domainToModel(entity *rate.Rate) *RateModel {
	return &RateModel{
		ID:            entity.ID(),
		BaseCurrency:  entity.Pair().Base().String(),
		QuoteCurrency: entity.Pair().Quote().String(),
		Value:         entity.Value(),
		EffectiveDate: entity.EffectiveDate(),
		Source:        string(entity.Source()),
		CreatedAt:     entity.CreatedAt(),
		UpdatedAt:     entity.UpdatedAt(),
	}
}

// modelToDomain converts a database model to a domain Rate entity.
func (r *RateRepository) modelToDomain(model *RateModel) (*rate.Rate, error) {
	baseCode, err := currency.NewCode(model.BaseCurrency)
	if err != nil {
		return nil, err
	}

	quoteCode, err := currency.NewCode(model.QuoteCurrency)
	if err != nil {
		return nil, err
	}

	pair, err := currency.NewPair(baseCode, quoteCode)
	if err != nil {
		return nil, err
	}

	return rate.Reconstitute(
		model.ID,
		pair,
		model.Value,
		model.EffectiveDate,
		rate.Source(model.Source),
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}
