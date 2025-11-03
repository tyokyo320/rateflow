package query_test

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/tyokyo320/rateflow/internal/application/dto"
	"github.com/tyokyo320/rateflow/internal/application/query"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
)

// Mock repository implements rate.Repository interface
type mockRateRepository struct {
	findLatestFunc func(ctx context.Context, pair currency.Pair) (*rate.Rate, error)
}

// Implement rate.Repository methods
func (m *mockRateRepository) FindLatest(ctx context.Context, pair currency.Pair) (*rate.Rate, error) {
	if m.findLatestFunc != nil {
		return m.findLatestFunc(ctx, pair)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) FindByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (*rate.Rate, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) FindByDateRange(ctx context.Context, pair currency.Pair, start, end time.Time) ([]*rate.Rate, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) FindByPairs(ctx context.Context, pairs []currency.Pair) ([]*rate.Rate, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) ExistsByPairAndDate(ctx context.Context, pair currency.Pair, date time.Time) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *mockRateRepository) DeleteOlderThan(ctx context.Context, date time.Time) (int64, error) {
	return 0, errors.New("not implemented")
}

// Implement genericrepo.Repository[*rate.Rate] methods
func (m *mockRateRepository) Create(ctx context.Context, entity *rate.Rate) error {
	return errors.New("not implemented")
}

func (m *mockRateRepository) FindByID(ctx context.Context, id string) (*rate.Rate, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) Update(ctx context.Context, entity *rate.Rate) error {
	return errors.New("not implemented")
}

func (m *mockRateRepository) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockRateRepository) FindAll(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRateRepository) Count(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
	return 0, errors.New("not implemented")
}

func (m *mockRateRepository) Stream(ctx context.Context, opts ...genericrepo.QueryOption) iter.Seq[*rate.Rate] {
	return func(yield func(*rate.Rate) bool) {}
}

func (m *mockRateRepository) StreamWithError(ctx context.Context, opts ...genericrepo.QueryOption) iter.Seq2[*rate.Rate, error] {
	return func(yield func(*rate.Rate, error) bool) {}
}

func (m *mockRateRepository) Exists(ctx context.Context, id string) (bool, error) {
	return false, errors.New("not implemented")
}

// Mock cache implements CacheInterface
type mockCache struct {
	getFunc func(ctx context.Context, key string, dest interface{}) error
	setFunc func(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if m.getFunc != nil {
		return m.getFunc(ctx, key, dest)
	}
	return errors.New("cache miss")
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (m *mockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 0, nil
}

func (m *mockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (m *mockCache) Ping(ctx context.Context) error {
	return nil
}

func (m *mockCache) Close() error {
	return nil
}

func TestGetLatestRateHandler_CacheHit(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	now := time.Now()

	// Setup mock cache with data
	cache := &mockCache{
		getFunc: func(ctx context.Context, key string, dest interface{}) error {
			// Simulate cache hit
			if resp, ok := dest.(*dto.RateResponse); ok {
				*resp = dto.RateResponse{
					ID:            "cached-123",
					Pair:          "CNY/JPY",
					BaseCurrency:  "CNY",
					QuoteCurrency: "JPY",
					Rate:          20.0,
					EffectiveDate: now,
					Source:        "unionpay",
					CreatedAt:     now,
					UpdatedAt:     now,
				}
			}
			return nil
		},
	}

	// Setup mock repository (should not be called)
	repo := &mockRateRepository{
		findLatestFunc: func(ctx context.Context, pair currency.Pair) (*rate.Rate, error) {
			t.Error("repository should not be called on cache hit")
			return nil, errors.New("should not be called")
		},
	}

	log := logger.NewNoop()
	handler := query.NewGetLatestRateHandler(repo, cache, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.GetLatestRateQuery{
		Pair: pair,
	})

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "cached-123" {
		t.Errorf("expected cached ID, got %s", result.ID)
	}
	if result.Rate != 20.0 {
		t.Errorf("expected rate 20.0, got %f", result.Rate)
	}
}

func TestGetLatestRateHandler_CacheMiss(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	now := time.Now()

	// Setup mock cache (miss)
	cache := &mockCache{
		getFunc: func(ctx context.Context, key string, dest interface{}) error {
			return errors.New("cache miss")
		},
		setFunc: func(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
			// Verify we're caching the result
			if _, ok := value.(*dto.RateResponse); !ok {
				t.Error("expected RateResponse to be cached")
			}
			if ttl != 5*time.Minute {
				t.Errorf("expected 5 minute TTL, got %v", ttl)
			}
			return nil
		},
	}

	// Create a rate for the repository to return
	testRate, _ := rate.NewRate(pair, 20.0, now, rate.SourceUnionPay)

	// Setup mock repository
	repo := &mockRateRepository{
		findLatestFunc: func(ctx context.Context, p currency.Pair) (*rate.Rate, error) {
			if !p.Equal(pair) {
				t.Errorf("expected pair %s, got %s", pair.String(), p.String())
			}
			return testRate, nil
		},
	}

	log := logger.NewNoop()
	handler := query.NewGetLatestRateHandler(repo, cache, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.GetLatestRateQuery{
		Pair: pair,
	})

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Pair != "CNY/JPY" {
		t.Errorf("expected pair CNY/JPY, got %s", result.Pair)
	}
	if result.Rate != 20.0 {
		t.Errorf("expected rate 20.0, got %f", result.Rate)
	}
}

func TestGetLatestRateHandler_RepositoryError(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)

	// Setup mock cache (miss)
	cache := &mockCache{
		getFunc: func(ctx context.Context, key string, dest interface{}) error {
			return errors.New("cache miss")
		},
	}

	// Setup mock repository with error
	expectedErr := errors.New("database error")
	repo := &mockRateRepository{
		findLatestFunc: func(ctx context.Context, p currency.Pair) (*rate.Rate, error) {
			return nil, expectedErr
		},
	}

	log := logger.NewNoop()
	handler := query.NewGetLatestRateHandler(repo, cache, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.GetLatestRateQuery{
		Pair: pair,
	})

	// Assert
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}
