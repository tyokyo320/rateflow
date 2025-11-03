package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tyokyo320/rateflow/internal/application/query"
	"github.com/tyokyo320/rateflow/internal/domain/currency"
	"github.com/tyokyo320/rateflow/internal/domain/rate"
	"github.com/tyokyo320/rateflow/internal/infrastructure/logger"
	"github.com/tyokyo320/rateflow/pkg/genericrepo"
)

type mockListRatesRepository struct {
	mockRateRepository
	findAllFunc func(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error)
	countFunc   func(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error)
}

func (m *mockListRatesRepository) FindAll(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockListRatesRepository) Count(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, opts...)
	}
	return 0, errors.New("not implemented")
}

func TestListRatesHandler_Success(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	now := time.Now()

	// Create test rates
	rate1, _ := rate.NewRate(pair, 20.0, now, rate.SourceUnionPay)
	rate2, _ := rate.NewRate(pair, 20.5, now.Add(-24*time.Hour), rate.SourceUnionPay)

	// Setup mock repository
	repo := &mockListRatesRepository{
		findAllFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
			// Verify options were passed
			if len(opts) == 0 {
				t.Error("expected query options to be passed")
			}
			return []*rate.Rate{rate1, rate2}, nil
		},
		countFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
			return 100, nil
		},
	}

	log := logger.NewNoop()
	handler := query.NewListRatesHandler(repo, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.ListRatesQuery{
		Pair:     pair,
		Page:     1,
		PageSize: 10,
	})

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
	if result.Pagination.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Pagination.Total)
	}
	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}
	if result.Pagination.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", result.Pagination.PageSize)
	}

	// Verify first item
	if result.Items[0].Rate != 20.0 {
		t.Errorf("expected first rate 20.0, got %f", result.Items[0].Rate)
	}
}

func TestListRatesHandler_FindAllError(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	expectedErr := errors.New("database error")

	// Setup mock repository with error
	repo := &mockListRatesRepository{
		findAllFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
			return nil, expectedErr
		},
	}

	log := logger.NewNoop()
	handler := query.NewListRatesHandler(repo, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.ListRatesQuery{
		Pair:     pair,
		Page:     1,
		PageSize: 10,
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

func TestListRatesHandler_CountError(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)
	now := time.Now()

	rate1, _ := rate.NewRate(pair, 20.0, now, rate.SourceUnionPay)
	expectedErr := errors.New("count error")

	// Setup mock repository
	repo := &mockListRatesRepository{
		findAllFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
			return []*rate.Rate{rate1}, nil
		},
		countFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
			return 0, expectedErr
		},
	}

	log := logger.NewNoop()
	handler := query.NewListRatesHandler(repo, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.ListRatesQuery{
		Pair:     pair,
		Page:     1,
		PageSize: 10,
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

func TestListRatesHandler_EmptyResult(t *testing.T) {
	pair := currency.MustNewPair(currency.CNY, currency.JPY)

	// Setup mock repository with empty results
	repo := &mockListRatesRepository{
		findAllFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) ([]*rate.Rate, error) {
			return []*rate.Rate{}, nil
		},
		countFunc: func(ctx context.Context, opts ...genericrepo.QueryOption) (int64, error) {
			return 0, nil
		},
	}

	log := logger.NewNoop()
	handler := query.NewListRatesHandler(repo, log)

	// Execute query
	result, err := handler.Handle(context.Background(), query.ListRatesQuery{
		Pair:     pair,
		Page:     1,
		PageSize: 10,
	})

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.Pagination.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Pagination.Total)
	}
}
