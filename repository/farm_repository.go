package repository

import (
	"context"
	"fmt"

	"github.com/sebaespinosa/test_NF/model"
	"gorm.io/gorm"
)

// FarmRepository handles database operations for Farm entities
type FarmRepository struct {
	db *gorm.DB
}

// NewFarmRepository creates a new FarmRepository instance
func NewFarmRepository(db *gorm.DB) *FarmRepository {
	return &FarmRepository{db: db}
}

// Create creates a new farm
func (r *FarmRepository) Create(ctx context.Context, farm *model.Farm) error {
	if err := r.db.WithContext(ctx).Create(farm).Error; err != nil {
		return fmt.Errorf("failed to create farm: %w", err)
	}
	return nil
}

// Save saves or updates a farm (upsert based on primary key)
func (r *FarmRepository) Save(ctx context.Context, farm *model.Farm) error {
	if err := r.db.WithContext(ctx).Save(farm).Error; err != nil {
		return fmt.Errorf("failed to save farm: %w", err)
	}
	return nil
}

// FindByID retrieves a farm by its ID
func (r *FarmRepository) FindByID(ctx context.Context, id uint) (*model.Farm, error) {
	var farm model.Farm
	if err := r.db.WithContext(ctx).First(&farm, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find farm by ID: %w", err)
	}
	return &farm, nil
}

// FindAll retrieves all farms
func (r *FarmRepository) FindAll(ctx context.Context) ([]model.Farm, error) {
	var farms []model.Farm
	if err := r.db.WithContext(ctx).Find(&farms).Error; err != nil {
		return nil, fmt.Errorf("failed to find all farms: %w", err)
	}
	return farms, nil
}

// Delete deletes a farm by ID
func (r *FarmRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Farm{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete farm: %w", err)
	}
	return nil
}

// DeleteAll deletes all farms
func (r *FarmRepository) DeleteAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("DELETE FROM farms").Error; err != nil {
		return fmt.Errorf("failed to delete all farms: %w", err)
	}
	return nil
}
