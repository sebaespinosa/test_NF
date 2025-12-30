package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/sebaespinosa/test_NF/repository"
	"go.uber.org/zap"
)

// IrrigationAnalyticsService handles business logic for irrigation analytics
type IrrigationAnalyticsService struct {
	repo   AnalyticsRepository
	logger *logging.Logger
}

// AnalyticsRepository defines the data access contract for analytics operations.
type AnalyticsRepository interface {
	GetAnalyticsForFarmByDateRange(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string, limit, offset int) ([]repository.AnalyticsAggregation, int64, error)
	GetYoYComparison(ctx context.Context, farmID uint, startTime, endTime time.Time, aggregation string) (map[int]repository.YoYAnalyticsData, error)
	GetSectorBreakdownForFarm(ctx context.Context, farmID uint, sectorID *uint, startTime, endTime time.Time) ([]repository.SectorAnalyticsData, error)
}

// NewIrrigationAnalyticsService creates a new IrrigationAnalyticsService instance
func NewIrrigationAnalyticsService(
	repo AnalyticsRepository,
	logger *logging.Logger,
) *IrrigationAnalyticsService {
	return &IrrigationAnalyticsService{
		repo:   repo,
		logger: logger,
	}
}

// GetAnalytics returns comprehensive irrigation analytics for a farm with year-over-year comparison
func (s *IrrigationAnalyticsService) GetAnalytics(
	ctx context.Context,
	farmID uint,
	startDate, endDate *time.Time,
	sectorID *uint,
	aggregation string,
	page, limit int,
) (*model.IrrigationAnalyticsResponse, error) {
	s.logger.WithContext(ctx).Info(
		"fetching irrigation analytics",
		zap.Uint("farm_id", farmID),
		zap.String("aggregation", aggregation),
	)

	// Calculate date range (default to last 90 days if not provided)
	now := time.Now().UTC()
	var start, end time.Time

	if startDate == nil || endDate == nil {
		// Default: last 90 days
		end = now
		start = now.AddDate(0, 0, -90)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		start = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		end = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)
	}

	// Fetch current period analytics
	timeSeries, totalCount, err := s.repo.GetAnalyticsForFarmByDateRange(ctx, farmID, start, end, aggregation, limit, (page-1)*limit)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to get analytics for farm", zap.Error(err))
		return nil, err
	}

	// Fetch YoY comparison data
	yoyData, err := s.repo.GetYoYComparison(ctx, farmID, start, end, aggregation)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to get YoY comparison", zap.Error(err))
		return nil, err
	}

	// Fetch sector breakdown
	sectorBreakdown, err := s.repo.GetSectorBreakdownForFarm(ctx, farmID, sectorID, start, end)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to get sector breakdown", zap.Error(err))
		return nil, err
	}

	// Convert time-series data to response format
	timeSeriesEntries := s.convertTimeSeriesData(timeSeries)
	sectorBreakdownEntries := s.convertSectorBreakdownData(sectorBreakdown)

	// Calculate metrics for current period
	currentMetrics := s.calculateMetrics(timeSeries)

	// Calculate YoY comparison metrics
	currentYear := time.Now().Year()
	yoY1 := s.getYoYMetrics(yoyData, currentYear-1, "previous year")
	yoY2 := s.getYoYMetrics(yoyData, currentYear-2, "two years ago")

	// Calculate period comparison percentages
	periodComparison := s.calculatePeriodComparison(currentMetrics, yoY1, yoY2)

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// Build response
	response := &model.IrrigationAnalyticsResponse{
		FarmID:       farmID,
		FarmName:     "", // Will be populated if needed
		Period:       model.IrrigationAnalyticsPeriod{Start: start, End: end},
		Aggregation:  aggregation,
		Metrics:      currentMetrics,
		SamePeriod1Y: yoY1,
		SamePeriod2Y: yoY2,
		PeriodComparison: &model.PeriodComparisonSet{
			VsPeriod1Y: periodComparison.VsPeriod1Y,
			VsPeriod2Y: periodComparison.VsPeriod2Y,
		},
		TimeSeries: model.TimeSeries{
			Data: timeSeriesEntries,
			Pagination: model.PaginationMetadata{
				Page:       page,
				Limit:      limit,
				TotalCount: int(totalCount),
				TotalPages: totalPages,
			},
		},
		SectorBreakdown: sectorBreakdownEntries,
	}

	return response, nil
}

// calculateMetrics calculates aggregated metrics from time-series data
func (s *IrrigationAnalyticsService) calculateMetrics(data []repository.AnalyticsAggregation) model.AnalyticsMetrics {
	if len(data) == 0 {
		return model.AnalyticsMetrics{
			TotalIrrigationVolumeMM: 0,
			TotalIrrigationEvents:   0,
			AverageEfficiency:       nil,
			EfficiencyRange:         nil,
		}
	}

	var totalVolume float64
	var totalEvents int
	var efficiencies []float64
	var minEfficiency, maxEfficiency *float64

	for _, entry := range data {
		totalVolume += entry.TotalRealAmount
		totalEvents += entry.EventCount

		if entry.AvgEfficiency != nil {
			efficiencies = append(efficiencies, *entry.AvgEfficiency)
		}
		if entry.MinEfficiency != nil {
			if minEfficiency == nil || *entry.MinEfficiency < *minEfficiency {
				minEfficiency = entry.MinEfficiency
			}
		}
		if entry.MaxEfficiency != nil {
			if maxEfficiency == nil || *entry.MaxEfficiency > *maxEfficiency {
				maxEfficiency = entry.MaxEfficiency
			}
		}
	}

	metrics := model.AnalyticsMetrics{
		TotalIrrigationVolumeMM: totalVolume,
		TotalIrrigationEvents:   totalEvents,
	}

	// Calculate average efficiency from valid values
	if len(efficiencies) > 0 {
		var sum float64
		for _, e := range efficiencies {
			sum += e
		}
		avgEff := sum / float64(len(efficiencies))
		metrics.AverageEfficiency = &avgEff

		if minEfficiency != nil && maxEfficiency != nil {
			metrics.EfficiencyRange = &model.EfficiencyRange{
				Min: *minEfficiency,
				Max: *maxEfficiency,
			}
		}
	}

	return metrics
}

// getYoYMetrics converts YoY data to response format with null handling
func (s *IrrigationAnalyticsService) getYoYMetrics(
	yoyData map[int]repository.YoYAnalyticsData,
	year int,
	yearLabel string,
) *model.YoYComparison {
	data, exists := yoyData[year]
	if !exists {
		return &model.YoYComparison{
			DataIncomplete: true,
			Note:           fmt.Sprintf("No data available for %s (%d)", yearLabel, year),
		}
	}

	comparison := &model.YoYComparison{
		DataIncomplete: false,
	}

	// Set total volume if data exists
	if data.TotalRealAmount > 0 {
		comparison.TotalIrrigationVolumeMM = &data.TotalRealAmount
	}

	// Set total events
	if data.EventCount > 0 {
		comparison.TotalIrrigationEvents = &data.EventCount
	} else {
		return &model.YoYComparison{
			DataIncomplete: true,
			Note:           fmt.Sprintf("No events found for %s (%d)", yearLabel, year),
		}
	}

	// Set efficiency metrics
	if data.AvgEfficiency != nil {
		comparison.AverageEfficiency = data.AvgEfficiency
	}

	if data.MinEfficiency != nil && data.MaxEfficiency != nil {
		comparison.EfficiencyRange = &model.EfficiencyRange{
			Min: *data.MinEfficiency,
			Max: *data.MaxEfficiency,
		}
	}

	return comparison
}

// calculatePeriodComparison calculates year-over-year percentage changes
func (s *IrrigationAnalyticsService) calculatePeriodComparison(
	current model.AnalyticsMetrics,
	yoY1, yoY2 *model.YoYComparison,
) *model.PeriodComparisonSet {
	result := &model.PeriodComparisonSet{}

	// Compare with previous year
	if yoY1 != nil && !yoY1.DataIncomplete && yoY1.TotalIrrigationVolumeMM != nil {
		result.VsPeriod1Y = s.calculatePercentageChanges(current, *yoY1.TotalIrrigationVolumeMM, *yoY1.TotalIrrigationEvents, yoY1.AverageEfficiency)
	}

	// Compare with two years ago
	if yoY2 != nil && !yoY2.DataIncomplete && yoY2.TotalIrrigationVolumeMM != nil {
		result.VsPeriod2Y = s.calculatePercentageChanges(current, *yoY2.TotalIrrigationVolumeMM, *yoY2.TotalIrrigationEvents, yoY2.AverageEfficiency)
	}

	return result
}

// Calculate percentage changes between two periods
func (s *IrrigationAnalyticsService) calculatePercentageChanges(
	current model.AnalyticsMetrics,
	prevVolume float64,
	prevEvents int,
	prevEfficiency *float64,
) *model.PeriodComparison {
	comparison := &model.PeriodComparison{}

	// Volume change
	if prevVolume > 0 {
		change := ((current.TotalIrrigationVolumeMM - prevVolume) / prevVolume) * 100
		comparison.VolumeChangePercent = &change
	}

	// Events change
	if prevEvents > 0 {
		change := ((float64(current.TotalIrrigationEvents) - float64(prevEvents)) / float64(prevEvents)) * 100
		comparison.EventsChangePercent = &change
	}

	// Efficiency change
	if prevEfficiency != nil && current.AverageEfficiency != nil && *prevEfficiency > 0 {
		change := ((*current.AverageEfficiency - *prevEfficiency) / *prevEfficiency) * 100
		comparison.EfficiencyChangePercent = &change
	}

	return comparison
}

// convertTimeSeriesData converts repository data to response format
func (s *IrrigationAnalyticsService) convertTimeSeriesData(data []repository.AnalyticsAggregation) []model.TimeSeriesEntry {
	entries := make([]model.TimeSeriesEntry, 0, len(data))

	for _, item := range data {
		entry := model.TimeSeriesEntry{
			Date:            item.Period.Format("2006-01-02"),
			NominalAmountMM: item.TotalNominalAmount,
			RealAmountMM:    item.TotalRealAmount,
			Efficiency:      item.AvgEfficiency,
			EventCount:      item.EventCount,
		}
		entries = append(entries, entry)
	}

	return entries
}

// convertSectorBreakdownData converts repository data to response format
func (s *IrrigationAnalyticsService) convertSectorBreakdownData(data []repository.SectorAnalyticsData) []model.SectorBreakdown {
	breakdown := make([]model.SectorBreakdown, 0, len(data))

	for _, item := range data {
		breakdown = append(breakdown, model.SectorBreakdown{
			SectorID:          item.SectorID,
			SectorName:        item.SectorName,
			TotalVolumeMM:     item.TotalRealAmount,
			AverageEfficiency: item.AvgEfficiency,
		})
	}

	return breakdown
}
