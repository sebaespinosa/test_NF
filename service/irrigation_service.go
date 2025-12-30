package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/sebaespinosa/test_NF/repository"
	"go.uber.org/zap"
)

// IrrigationSectorService handles business logic for irrigation sector operations
type IrrigationSectorService struct {
	repo   *repository.IrrigationSectorRepository
	logger *logging.Logger
}

// NewIrrigationSectorService creates a new IrrigationSectorService instance
func NewIrrigationSectorService(repo *repository.IrrigationSectorRepository, logger *logging.Logger) *IrrigationSectorService {
	return &IrrigationSectorService{
		repo:   repo,
		logger: logger,
	}
}

// GetByID retrieves an irrigation sector by its ID
func (s *IrrigationSectorService) GetByID(ctx context.Context, id uint) (*model.IrrigationSector, error) {
	s.logger.WithContext(ctx).Info("fetching irrigation sector by ID", zap.Uint("sector_id", id))
	return s.repo.FindByID(ctx, id)
}

// GetByFarmID retrieves all irrigation sectors for a farm
func (s *IrrigationSectorService) GetByFarmID(ctx context.Context, farmID uint) ([]model.IrrigationSector, error) {
	s.logger.WithContext(ctx).Info("fetching irrigation sectors by farm ID", zap.Uint("farm_id", farmID))
	return s.repo.FindByFarmID(ctx, farmID)
}

// Create creates a new irrigation sector
func (s *IrrigationSectorService) Create(ctx context.Context, sector *model.IrrigationSector) error {
	s.logger.WithContext(ctx).Info("creating irrigation sector",
		zap.String("name", sector.Name),
		zap.Uint("farm_id", sector.FarmID),
	)
	return s.repo.Create(ctx, sector)
}

// Delete deletes an irrigation sector by ID
func (s *IrrigationSectorService) Delete(ctx context.Context, id uint) error {
	s.logger.WithContext(ctx).Info("deleting irrigation sector", zap.Uint("sector_id", id))
	return s.repo.Delete(ctx, id)
}

// SeedSectors inserts or updates irrigation sectors from seed data (idempotent)
func (s *IrrigationSectorService) SeedSectors(ctx context.Context, sectors []model.IrrigationSector) error {
	s.logger.WithContext(ctx).Info("seeding irrigation sectors", zap.Int("count", len(sectors)))

	for _, sector := range sectors {
		if err := s.repo.Save(ctx, &sector); err != nil {
			return fmt.Errorf("failed to seed sector %d: %w", sector.ID, err)
		}
		s.logger.WithContext(ctx).Debug("irrigation sector seeded",
			zap.Uint("sector_id", sector.ID),
			zap.String("name", sector.Name),
			zap.Uint("farm_id", sector.FarmID),
		)
	}

	s.logger.WithContext(ctx).Info("irrigation sectors seeded successfully", zap.Int("count", len(sectors)))
	return nil
}

// RemoveSeedSectors removes all irrigation sectors (cascade will remove related data)
func (s *IrrigationSectorService) RemoveSeedSectors(ctx context.Context) error {
	s.logger.WithContext(ctx).Info("removing all irrigation sectors")
	if err := s.repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("failed to remove irrigation sectors: %w", err)
	}
	s.logger.WithContext(ctx).Info("irrigation sectors removed successfully")
	return nil
}

// IrrigationDataService handles business logic for irrigation data operations
type IrrigationDataService struct {
	repo   *repository.IrrigationDataRepository
	logger *logging.Logger
}

// NewIrrigationDataService creates a new IrrigationDataService instance
func NewIrrigationDataService(repo *repository.IrrigationDataRepository, logger *logging.Logger) *IrrigationDataService {
	return &IrrigationDataService{
		repo:   repo,
		logger: logger,
	}
}

// GetByID retrieves irrigation data by its ID
func (s *IrrigationDataService) GetByID(ctx context.Context, id uint) (*model.IrrigationData, error) {
	s.logger.WithContext(ctx).Info("fetching irrigation data by ID", zap.Uint("data_id", id))
	return s.repo.FindByID(ctx, id)
}

// GetByFarmAndTimeRange retrieves irrigation data for a farm within a time range
func (s *IrrigationDataService) GetByFarmAndTimeRange(ctx context.Context, farmID uint, startTime, endTime time.Time) ([]model.IrrigationData, error) {
	s.logger.WithContext(ctx).Info("fetching irrigation data by farm and time range",
		zap.Uint("farm_id", farmID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)
	return s.repo.FindByFarmIDAndTimeRange(ctx, farmID, startTime, endTime)
}

// GetBySectorAndTimeRange retrieves irrigation data for a sector within a time range
func (s *IrrigationDataService) GetBySectorAndTimeRange(ctx context.Context, sectorID uint, startTime, endTime time.Time) ([]model.IrrigationData, error) {
	s.logger.WithContext(ctx).Info("fetching irrigation data by sector and time range",
		zap.Uint("sector_id", sectorID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)
	return s.repo.FindBySectorIDAndTimeRange(ctx, sectorID, startTime, endTime)
}

// AggregateByFarm aggregates irrigation data by farm within a time range
func (s *IrrigationDataService) AggregateByFarm(ctx context.Context, startTime, endTime time.Time) ([]repository.FarmAggregation, error) {
	s.logger.WithContext(ctx).Info("aggregating irrigation data by farm",
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)
	return s.repo.AggregateByFarm(ctx, startTime, endTime)
}

// AggregateBySector aggregates irrigation data by sector within a time range
func (s *IrrigationDataService) AggregateBySector(ctx context.Context, startTime, endTime time.Time) ([]repository.SectorAggregation, error) {
	s.logger.WithContext(ctx).Info("aggregating irrigation data by sector",
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)
	return s.repo.AggregateBySector(ctx, startTime, endTime)
}

// Create creates a new irrigation data record
func (s *IrrigationDataService) Create(ctx context.Context, data *model.IrrigationData) error {
	s.logger.WithContext(ctx).Info("creating irrigation data",
		zap.Uint("farm_id", data.FarmID),
		zap.Uint("sector_id", data.IrrigationSectorID),
		zap.Time("start_time", data.StartTime),
	)
	return s.repo.Create(ctx, data)
}

// Delete deletes irrigation data by ID
func (s *IrrigationDataService) Delete(ctx context.Context, id uint) error {
	s.logger.WithContext(ctx).Info("deleting irrigation data", zap.Uint("data_id", id))
	return s.repo.Delete(ctx, id)
}

// SeedData inserts or updates irrigation data from seed data (idempotent)
func (s *IrrigationDataService) SeedData(ctx context.Context, data []model.IrrigationData) error {
	s.logger.WithContext(ctx).Info("seeding irrigation data", zap.Int("count", len(data)))

	for _, record := range data {
		if err := s.repo.Save(ctx, &record); err != nil {
			return fmt.Errorf("failed to seed irrigation data %d: %w", record.ID, err)
		}
		s.logger.WithContext(ctx).Debug("irrigation data seeded",
			zap.Uint("data_id", record.ID),
			zap.Uint("farm_id", record.FarmID),
			zap.Uint("sector_id", record.IrrigationSectorID),
		)
	}

	s.logger.WithContext(ctx).Info("irrigation data seeded successfully", zap.Int("count", len(data)))
	return nil
}

// RemoveSeedData removes all irrigation data
func (s *IrrigationDataService) RemoveSeedData(ctx context.Context) error {
	s.logger.WithContext(ctx).Info("removing all irrigation data")
	if err := s.repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("failed to remove irrigation data: %w", err)
	}
	s.logger.WithContext(ctx).Info("irrigation data removed successfully")
	return nil
}
