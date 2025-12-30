package repository

import (
	"context"
	"fmt"

	"github.com/sebaespinosa/test_NF/model"
	"gorm.io/gorm"
)

// IrrigationSectorRepository handles database operations for IrrigationSector entities
type IrrigationSectorRepository struct {
	db *gorm.DB
}

// NewIrrigationSectorRepository creates a new IrrigationSectorRepository instance
func NewIrrigationSectorRepository(db *gorm.DB) *IrrigationSectorRepository {
	return &IrrigationSectorRepository{db: db}
}

// Create creates a new irrigation sector
func (r *IrrigationSectorRepository) Create(ctx context.Context, sector *model.IrrigationSector) error {
	if err := r.db.WithContext(ctx).Create(sector).Error; err != nil {
		return fmt.Errorf("failed to create irrigation sector: %w", err)
	}
	return nil
}

// Save saves or updates an irrigation sector (upsert based on primary key)
func (r *IrrigationSectorRepository) Save(ctx context.Context, sector *model.IrrigationSector) error {
	if err := r.db.WithContext(ctx).Save(sector).Error; err != nil {
		return fmt.Errorf("failed to save irrigation sector: %w", err)
	}
	return nil
}

// FindByID retrieves an irrigation sector by its ID
func (r *IrrigationSectorRepository) FindByID(ctx context.Context, id uint) (*model.IrrigationSector, error) {
	var sector model.IrrigationSector
	if err := r.db.WithContext(ctx).Preload("Farm").First(&sector, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find irrigation sector by ID: %w", err)
	}
	return &sector, nil
}

// FindByFarmID retrieves all irrigation sectors for a specific farm
func (r *IrrigationSectorRepository) FindByFarmID(ctx context.Context, farmID uint) ([]model.IrrigationSector, error) {
	var sectors []model.IrrigationSector
	if err := r.db.WithContext(ctx).Where("farm_id = ?", farmID).Find(&sectors).Error; err != nil {
		return nil, fmt.Errorf("failed to find irrigation sectors by farm ID: %w", err)
	}
	return sectors, nil
}

// FindAll retrieves all irrigation sectors
func (r *IrrigationSectorRepository) FindAll(ctx context.Context) ([]model.IrrigationSector, error) {
	var sectors []model.IrrigationSector
	if err := r.db.WithContext(ctx).Find(&sectors).Error; err != nil {
		return nil, fmt.Errorf("failed to find all irrigation sectors: %w", err)
	}
	return sectors, nil
}

// Delete deletes an irrigation sector by ID
func (r *IrrigationSectorRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.IrrigationSector{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete irrigation sector: %w", err)
	}
	return nil
}

// DeleteAll deletes all irrigation sectors
func (r *IrrigationSectorRepository) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("DELETE FROM irrigation_sectors").Error; err != nil {
		return fmt.Errorf("failed to delete all irrigation sectors: %w", err)
	}
	return nil
}
