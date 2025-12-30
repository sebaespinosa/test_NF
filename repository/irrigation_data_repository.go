package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sebaespinosa/test_NF/model"
	"gorm.io/gorm"
)

// IrrigationDataRepository handles database operations for IrrigationData entities
type IrrigationDataRepository struct {
	db *gorm.DB
}

// NewIrrigationDataRepository creates a new IrrigationDataRepository instance
func NewIrrigationDataRepository(db *gorm.DB) *IrrigationDataRepository {
	return &IrrigationDataRepository{db: db}
}

// Create creates a new irrigation data record
func (r *IrrigationDataRepository) Create(ctx context.Context, data *model.IrrigationData) error {
	if err := r.db.WithContext(ctx).Create(data).Error; err != nil {
		return fmt.Errorf("failed to create irrigation data: %w", err)
	}
	return nil
}

// Save saves or updates an irrigation data record (upsert based on primary key)
func (r *IrrigationDataRepository) Save(ctx context.Context, data *model.IrrigationData) error {
	if err := r.db.WithContext(ctx).Save(data).Error; err != nil {
		return fmt.Errorf("failed to save irrigation data: %w", err)
	}
	return nil
}

// FindByID retrieves irrigation data by its ID
func (r *IrrigationDataRepository) FindByID(ctx context.Context, id uint) (*model.IrrigationData, error) {
	var data model.IrrigationData
	if err := r.db.WithContext(ctx).Preload("Farm").Preload("IrrigationSector").First(&data, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find irrigation data by ID: %w", err)
	}
	return &data, nil
}

// FindByFarmIDAndTimeRange retrieves irrigation data for a farm within a time range
// Uses composite index (farm_id, start_time) for optimal performance
func (r *IrrigationDataRepository) FindByFarmIDAndTimeRange(ctx context.Context, farmID uint, startTime, endTime time.Time) ([]model.IrrigationData, error) {
	var data []model.IrrigationData
	if err := r.db.WithContext(ctx).
		Where("farm_id = ? AND start_time >= ? AND start_time <= ?", farmID, startTime, endTime).
		Order("start_time ASC").
		Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to find irrigation data by farm and time range: %w", err)
	}
	return data, nil
}

// FindBySectorIDAndTimeRange retrieves irrigation data for a sector within a time range
// Uses composite index (irrigation_sector_id, start_time) for optimal performance
func (r *IrrigationDataRepository) FindBySectorIDAndTimeRange(ctx context.Context, sectorID uint, startTime, endTime time.Time) ([]model.IrrigationData, error) {
	var data []model.IrrigationData
	if err := r.db.WithContext(ctx).
		Where("irrigation_sector_id = ? AND start_time >= ? AND start_time <= ?", sectorID, startTime, endTime).
		Order("start_time ASC").
		Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to find irrigation data by sector and time range: %w", err)
	}
	return data, nil
}

// AggregateByFarm aggregates irrigation data by farm within a time range
// Performs SQL-level aggregation to avoid N+1 queries and reduce memory overhead
type FarmAggregation struct {
	FarmID             uint    `json:"farm_id"`
	FarmName           string  `json:"farm_name"`
	TotalEvents        int64   `json:"total_events"`
	TotalNominalAmount float64 `json:"total_nominal_amount"`
	TotalRealAmount    float64 `json:"total_real_amount"`
	AvgNominalAmount   float64 `json:"avg_nominal_amount"`
	AvgRealAmount      float64 `json:"avg_real_amount"`
}

func (r *IrrigationDataRepository) AggregateByFarm(ctx context.Context, startTime, endTime time.Time) ([]FarmAggregation, error) {
	var results []FarmAggregation
	if err := r.db.WithContext(ctx).
		Table("irrigation_data").
		Select(`
			irrigation_data.farm_id,
			farms.name as farm_name,
			COUNT(*) as total_events,
			SUM(irrigation_data.nominal_amount) as total_nominal_amount,
			SUM(irrigation_data.real_amount) as total_real_amount,
			AVG(irrigation_data.nominal_amount) as avg_nominal_amount,
			AVG(irrigation_data.real_amount) as avg_real_amount
		`).
		Joins("JOIN farms ON farms.id = irrigation_data.farm_id").
		Where("irrigation_data.start_time >= ? AND irrigation_data.start_time <= ?", startTime, endTime).
		Group("irrigation_data.farm_id, farms.name").
		Order("irrigation_data.farm_id ASC").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate irrigation data by farm: %w", err)
	}
	return results, nil
}

// AggregateBySector aggregates irrigation data by sector within a time range
// Performs SQL-level aggregation to avoid N+1 queries and reduce memory overhead
type SectorAggregation struct {
	FarmID             uint    `json:"farm_id"`
	FarmName           string  `json:"farm_name"`
	SectorID           uint    `json:"sector_id"`
	SectorName         string  `json:"sector_name"`
	TotalEvents        int64   `json:"total_events"`
	TotalNominalAmount float64 `json:"total_nominal_amount"`
	TotalRealAmount    float64 `json:"total_real_amount"`
	AvgNominalAmount   float64 `json:"avg_nominal_amount"`
	AvgRealAmount      float64 `json:"avg_real_amount"`
}

func (r *IrrigationDataRepository) AggregateBySector(ctx context.Context, startTime, endTime time.Time) ([]SectorAggregation, error) {
	var results []SectorAggregation
	if err := r.db.WithContext(ctx).
		Table("irrigation_data").
		Select(`
			irrigation_data.farm_id,
			farms.name as farm_name,
			irrigation_data.irrigation_sector_id as sector_id,
			irrigation_sectors.name as sector_name,
			COUNT(*) as total_events,
			SUM(irrigation_data.nominal_amount) as total_nominal_amount,
			SUM(irrigation_data.real_amount) as total_real_amount,
			AVG(irrigation_data.nominal_amount) as avg_nominal_amount,
			AVG(irrigation_data.real_amount) as avg_real_amount
		`).
		Joins("JOIN farms ON farms.id = irrigation_data.farm_id").
		Joins("JOIN irrigation_sectors ON irrigation_sectors.id = irrigation_data.irrigation_sector_id").
		Where("irrigation_data.start_time >= ? AND irrigation_data.start_time <= ?", startTime, endTime).
		Group("irrigation_data.farm_id, farms.name, irrigation_data.irrigation_sector_id, irrigation_sectors.name").
		Order("irrigation_data.farm_id ASC, irrigation_data.irrigation_sector_id ASC").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate irrigation data by sector: %w", err)
	}
	return results, nil
}

// Delete deletes an irrigation data record by ID
func (r *IrrigationDataRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.IrrigationData{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete irrigation data: %w", err)
	}
	return nil
}

// DeleteAll deletes all irrigation data records
func (r *IrrigationDataRepository) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("DELETE FROM irrigation_data").Error; err != nil {
		return fmt.Errorf("failed to delete all irrigation data: %w", err)
	}
	return nil
}

// AnalyticsAggregation represents aggregated analytics data for a time period
type AnalyticsAggregation struct {
	Period             time.Time `gorm:"column:period"`
	Year               int       `gorm:"column:year"`
	TotalRealAmount    float64   `gorm:"column:total_real_amount"`
	TotalNominalAmount float64   `gorm:"column:total_nominal_amount"`
	EventCount         int       `gorm:"column:event_count"`
	AvgEfficiency      *float64  `gorm:"column:avg_efficiency"`
	MinEfficiency      *float64  `gorm:"column:min_efficiency"`
	MaxEfficiency      *float64  `gorm:"column:max_efficiency"`
}

// GetAnalyticsForFarmByDateRange retrieves aggregated analytics for a farm within a time range
// Uses SQL GROUP BY with DATE_TRUNC for efficient aggregation at database level
// Leverages composite index (farm_id, start_time) for optimal performance
func (r *IrrigationDataRepository) GetAnalyticsForFarmByDateRange(
	ctx context.Context,
	farmID uint,
	startTime, endTime time.Time,
	aggregation string,
	limit, offset int,
) ([]AnalyticsAggregation, int64, error) {
	var results []AnalyticsAggregation
	var totalCount int64

	// Determine DATE_TRUNC format based on aggregation type
	truncFormat := "'day'"
	if aggregation == "weekly" {
		truncFormat = "'week'"
	} else if aggregation == "monthly" {
		truncFormat = "'month'"
	}

	// Count total records for pagination
	countQuery := r.db.WithContext(ctx).
		Table("irrigation_data").
		Where("farm_id = ? AND start_time >= ? AND start_time <= ?", farmID, startTime, endTime)
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count irrigation data: %w", err)
	}

	// Fetch aggregated data using DATE_TRUNC
	if err := r.db.WithContext(ctx).
		Table("irrigation_data").
		Select(`
			DATE_TRUNC(`+truncFormat+`, start_time) as period,
			EXTRACT(YEAR FROM start_time)::int as year,
			SUM(real_amount) as total_real_amount,
			SUM(nominal_amount) as total_nominal_amount,
			COUNT(*) as event_count,
			AVG(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as avg_efficiency,
			MIN(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as min_efficiency,
			MAX(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as max_efficiency
		`).
		Where("farm_id = ? AND start_time >= ? AND start_time <= ?", farmID, startTime, endTime).
		Group("DATE_TRUNC(" + truncFormat + ", start_time), year").
		Order("period ASC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get analytics for farm: %w", err)
	}

	return results, totalCount, nil
}

// YoYAnalyticsData represents year-over-year aggregated data
type YoYAnalyticsData struct {
	Year               int      `gorm:"column:year"`
	TotalRealAmount    float64  `gorm:"column:total_real_amount"`
	TotalNominalAmount float64  `gorm:"column:total_nominal_amount"`
	EventCount         int      `gorm:"column:event_count"`
	AvgEfficiency      *float64 `gorm:"column:avg_efficiency"`
	MinEfficiency      *float64 `gorm:"column:min_efficiency"`
	MaxEfficiency      *float64 `gorm:"column:max_efficiency"`
}

// GetYoYComparison retrieves year-over-year data for the same date range across 3 years
// Uses single SQL UNION ALL query for efficiency (follows DatabaseOptimization.md best practices)
// Returns data for all 3 years; caller handles year-specific extraction
func (r *IrrigationDataRepository) GetYoYComparison(
	ctx context.Context,
	farmID uint,
	startTime, endTime time.Time,
	aggregation string,
) (map[int]YoYAnalyticsData, error) {
	var results []YoYAnalyticsData

	// Calculate date ranges for each year
	currentYear := time.Now().Year()
	year1Start := time.Date(currentYear, startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
	year1End := time.Date(currentYear, endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.UTC)
	year2Start := time.Date(currentYear-1, startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
	year2End := time.Date(currentYear-1, endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.UTC)
	year3Start := time.Date(currentYear-2, startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
	year3End := time.Date(currentYear-2, endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.UTC)

	// Build UNION ALL query using raw SQL for efficiency
	unionQuery := `
	SELECT
		EXTRACT(YEAR FROM start_time)::int as year,
		SUM(real_amount) as total_real_amount,
		SUM(nominal_amount) as total_nominal_amount,
		COUNT(*) as event_count,
		AVG(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as avg_efficiency,
		MIN(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as min_efficiency,
		MAX(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as max_efficiency
	FROM irrigation_data
	WHERE farm_id = ? AND start_time >= ? AND start_time <= ?
	GROUP BY EXTRACT(YEAR FROM start_time)
	
	UNION ALL
	
	SELECT
		EXTRACT(YEAR FROM start_time)::int as year,
		SUM(real_amount) as total_real_amount,
		SUM(nominal_amount) as total_nominal_amount,
		COUNT(*) as event_count,
		AVG(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as avg_efficiency,
		MIN(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as min_efficiency,
		MAX(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as max_efficiency
	FROM irrigation_data
	WHERE farm_id = ? AND start_time >= ? AND start_time <= ?
	GROUP BY EXTRACT(YEAR FROM start_time)
	
	UNION ALL
	
	SELECT
		EXTRACT(YEAR FROM start_time)::int as year,
		SUM(real_amount) as total_real_amount,
		SUM(nominal_amount) as total_nominal_amount,
		COUNT(*) as event_count,
		AVG(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as avg_efficiency,
		MIN(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as min_efficiency,
		MAX(CASE WHEN nominal_amount > 0 THEN real_amount::numeric / nominal_amount::numeric ELSE NULL END)::float as max_efficiency
	FROM irrigation_data
	WHERE farm_id = ? AND start_time >= ? AND start_time <= ?
	GROUP BY EXTRACT(YEAR FROM start_time)
	`

	if err := r.db.WithContext(ctx).Raw(unionQuery,
		farmID, year1Start, year1End,
		farmID, year2Start, year2End,
		farmID, year3Start, year3End,
	).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get YoY comparison: %w", err)
	}

	// Convert results to map indexed by year
	resultMap := make(map[int]YoYAnalyticsData)
	for _, result := range results {
		resultMap[result.Year] = result
	}

	return resultMap, nil
}

// SectorAnalyticsData represents aggregated data by sector
type SectorAnalyticsData struct {
	SectorID           uint     `gorm:"column:sector_id"`
	SectorName         string   `gorm:"column:sector_name"`
	TotalRealAmount    float64  `gorm:"column:total_real_amount"`
	TotalNominalAmount float64  `gorm:"column:total_nominal_amount"`
	AvgEfficiency      *float64 `gorm:"column:avg_efficiency"`
}

// GetSectorBreakdownForFarm retrieves aggregated metrics by irrigation sector
// Optionally filters by specific sector_id for better performance
func (r *IrrigationDataRepository) GetSectorBreakdownForFarm(
	ctx context.Context,
	farmID uint,
	sectorID *uint,
	startTime, endTime time.Time,
) ([]SectorAnalyticsData, error) {
	var results []SectorAnalyticsData

	query := r.db.WithContext(ctx).
		Table("irrigation_data").
		Select(`
			irrigation_data.irrigation_sector_id as sector_id,
			irrigation_sectors.name as sector_name,
			SUM(irrigation_data.real_amount) as total_real_amount,
			SUM(irrigation_data.nominal_amount) as total_nominal_amount,
			AVG(CASE WHEN irrigation_data.nominal_amount > 0 THEN irrigation_data.real_amount::numeric / irrigation_data.nominal_amount::numeric ELSE NULL END)::float as avg_efficiency
		`).
		Joins("JOIN irrigation_sectors ON irrigation_sectors.id = irrigation_data.irrigation_sector_id").
		Where("irrigation_data.farm_id = ? AND irrigation_data.start_time >= ? AND irrigation_data.start_time <= ?", farmID, startTime, endTime)

	// Filter by specific sector if provided
	if sectorID != nil {
		query = query.Where("irrigation_data.irrigation_sector_id = ?", *sectorID)
	}

	if err := query.
		Group("irrigation_data.irrigation_sector_id, irrigation_sectors.name").
		Order("irrigation_data.irrigation_sector_id ASC").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get sector breakdown: %w", err)
	}

	return results, nil
}
