package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAnalyticsRepo struct {
	getAnalyticsFn func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string, limit, offset int) ([]repository.AnalyticsAggregation, int64, error)
	getYoYFn       func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string) (map[int]repository.YoYAnalyticsData, error)
	getSectorFn    func(ctx context.Context, farmID uint, sectorID *uint, startTime, endTime time.Time) ([]repository.SectorAnalyticsData, error)
}

func (m *mockAnalyticsRepo) GetAnalyticsForFarmByDateRange(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string, limit, offset int) ([]repository.AnalyticsAggregation, int64, error) {
	return m.getAnalyticsFn(ctx, farmID, startTime, endTime, aggregation, limit, offset)
}

func (m *mockAnalyticsRepo) GetYoYComparison(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string) (map[int]repository.YoYAnalyticsData, error) {
	return m.getYoYFn(ctx, farmID, startTime, endTime, aggregation)
}

func (m *mockAnalyticsRepo) GetSectorBreakdownForFarm(ctx context.Context, farmID uint, sectorID *uint, startTime, endTime time.Time) ([]repository.SectorAnalyticsData, error) {
	return m.getSectorFn(ctx, farmID, sectorID, startTime, endTime)
}

func newTestLogger(t *testing.T) *logging.Logger {
	t.Helper()
	logger, err := logging.New("test")
	require.NoError(t, err)
	return logger
}

func TestGetAnalytics_Success(t *testing.T) {
	logger := newTestLogger(t)
	ctx := context.Background()

	currentYear := time.Now().Year()
	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

	repo := &mockAnalyticsRepo{
		getAnalyticsFn: func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string, limit, offset int) ([]repository.AnalyticsAggregation, int64, error) {
			return []repository.AnalyticsAggregation{
				{
					Period:             time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
					Year:               2024,
					TotalRealAmount:    30,
					TotalNominalAmount: 40,
					EventCount:         2,
					AvgEfficiency:      floatPtr(0.75),
					MinEfficiency:      floatPtr(0.7),
					MaxEfficiency:      floatPtr(0.8),
				},
			}, 1, nil
		},
		getYoYFn: func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string) (map[int]repository.YoYAnalyticsData, error) {
			return map[int]repository.YoYAnalyticsData{
				currentYear - 1: {
					Year:            currentYear - 1,
					TotalRealAmount: 25,
					EventCount:      2,
					AvgEfficiency:   floatPtr(0.6),
					MinEfficiency:   floatPtr(0.5),
					MaxEfficiency:   floatPtr(0.7),
				},
				currentYear - 2: {
					Year:            currentYear - 2,
					TotalRealAmount: 20,
					EventCount:      2,
					AvgEfficiency:   floatPtr(0.55),
					MinEfficiency:   floatPtr(0.5),
					MaxEfficiency:   floatPtr(0.6),
				},
			}, nil
		},
		getSectorFn: func(ctx context.Context, farmID uint, sectorID *uint, startTime, endTime time.Time) ([]repository.SectorAnalyticsData, error) {
			return []repository.SectorAnalyticsData{
				{SectorID: 1, SectorName: "S1", TotalRealAmount: 30, AvgEfficiency: floatPtr(0.75)},
			}, nil
		},
	}

	svc := NewIrrigationAnalyticsService(repo, logger)
	resp, err := svc.GetAnalytics(ctx, 1, &start, &end, nil, "daily", 1, 10)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.TimeSeries.Pagination.TotalPages)
	assert.Equal(t, 1, resp.TimeSeries.Pagination.TotalCount)
	require.NotNil(t, resp.PeriodComparison)
	require.NotNil(t, resp.PeriodComparison.VsPeriod1Y)
	assert.NotNil(t, resp.PeriodComparison.VsPeriod1Y.VolumeChangePercent)
	assert.Equal(t, 30.0, resp.Metrics.TotalIrrigationVolumeMM)
	assert.Len(t, resp.SectorBreakdown, 1)
}

func TestGetAnalytics_RepoError(t *testing.T) {
	logger := newTestLogger(t)
	ctx := context.Background()
	errExpected := errors.New("db error")

	repo := &mockAnalyticsRepo{
		getAnalyticsFn: func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string, limit, offset int) ([]repository.AnalyticsAggregation, int64, error) {
			return nil, 0, errExpected
		},
		getYoYFn: func(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string) (map[int]repository.YoYAnalyticsData, error) {
			return nil, nil
		},
		getSectorFn: func(ctx context.Context, farmID uint, sectorID *uint, startTime, endTime time.Time) ([]repository.SectorAnalyticsData, error) {
			return nil, nil
		},
	}

	svc := NewIrrigationAnalyticsService(repo, logger)
	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.GetAnalytics(ctx, 1, &start, &end, nil, "daily", 1, 10)
	require.ErrorIs(t, err, errExpected)
}

func floatPtr(v float64) *float64 { return &v }
