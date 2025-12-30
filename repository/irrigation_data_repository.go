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
