package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/sebaespinosa/test_NF/repository"
	"go.uber.org/zap"
)

// FarmService handles business logic for farm operations
type FarmService struct {
	repo   *repository.FarmRepository
	logger *logging.Logger
}

// NewFarmService creates a new FarmService instance
func NewFarmService(repo *repository.FarmRepository, logger *logging.Logger) *FarmService {
	return &FarmService{
		repo:   repo,
		logger: logger,
	}
}

// GetByID retrieves a farm by its ID
func (s *FarmService) GetByID(ctx context.Context, id uint) (*model.Farm, error) {
	s.logger.WithContext(ctx).Info("fetching farm by ID", zap.Uint("farm_id", id))
	return s.repo.FindByID(ctx, id)
}

// GetAll retrieves all farms
func (s *FarmService) GetAll(ctx context.Context) ([]model.Farm, error) {
	s.logger.WithContext(ctx).Info("fetching all farms")
	return s.repo.FindAll(ctx)
}

// Create creates a new farm
func (s *FarmService) Create(ctx context.Context, farm *model.Farm) error {
	s.logger.WithContext(ctx).Info("creating farm", zap.String("name", farm.Name))
	return s.repo.Create(ctx, farm)
}

// Delete deletes a farm by ID
func (s *FarmService) Delete(ctx context.Context, id uint) error {
	s.logger.WithContext(ctx).Info("deleting farm", zap.Uint("farm_id", id))
	return s.repo.Delete(ctx, id)
}

// SeedData represents the structure of the seed data JSON file
type SeedData struct {
	Farms             []model.Farm             `json:"farms"`
	IrrigationSectors []model.IrrigationSector `json:"irrigation_sectors"`
	IrrigationData    []model.IrrigationData   `json:"irrigation_data"`
}

// LoadSeedData loads seed data from a JSON file
func (s *FarmService) LoadSeedData(filePath string) (*SeedData, error) {
	s.logger.Info("loading seed data", zap.String("file_path", filePath))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse seed data: %w", err)
	}

	s.logger.Info("seed data loaded",
		zap.Int("farms", len(seedData.Farms)),
		zap.Int("sectors", len(seedData.IrrigationSectors)),
		zap.Int("irrigation_data", len(seedData.IrrigationData)),
	)

	return &seedData, nil
}

// SeedFarms inserts or updates farms from seed data (idempotent)
func (s *FarmService) SeedFarms(ctx context.Context, farms []model.Farm) error {
	s.logger.WithContext(ctx).Info("seeding farms", zap.Int("count", len(farms)))

	for _, farm := range farms {
		if err := s.repo.Save(ctx, &farm); err != nil {
			return fmt.Errorf("failed to seed farm %d: %w", farm.ID, err)
		}
		s.logger.WithContext(ctx).Debug("farm seeded",
			zap.Uint("farm_id", farm.ID),
			zap.String("name", farm.Name),
		)
	}

	s.logger.WithContext(ctx).Info("farms seeded successfully", zap.Int("count", len(farms)))
	return nil
}

// RemoveSeedFarms removes all farms (cascade will remove related data)
func (s *FarmService) RemoveSeedFarms(ctx context.Context) error {
	s.logger.WithContext(ctx).Info("removing all farms")
	if err := s.repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("failed to remove farms: %w", err)
	}
	s.logger.WithContext(ctx).Info("farms removed successfully")
	return nil
}
