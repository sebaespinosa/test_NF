package service

import (
	"context"

	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/sebaespinosa/test_NF/repository"
	"go.uber.org/zap"
)

// HealthService handles business logic for health checks
type HealthService struct {
	repo    *repository.HealthRepository
	logger  *logging.Logger
	version string
}

// NewHealthService creates a new instance of HealthService
func NewHealthService(repo *repository.HealthRepository, logger *logging.Logger, version string) *HealthService {
	return &HealthService{
		repo:    repo,
		logger:  logger,
		version: version,
	}
}

// GetHealth returns the health status of the service
func (s *HealthService) GetHealth(ctx context.Context) (*model.HealthResponse, error) {
	s.logger.WithContext(ctx).Info("checking service health")

	// Check database health
	if err := s.repo.CheckDatabaseHealth(ctx); err != nil {
		s.logger.WithContext(ctx).Error("database health check failed", zap.Error(err))
		return &model.HealthResponse{
			Status:  "unhealthy",
			Message: "database connection failed",
			Version: s.version,
		}, nil
	}

	s.logger.WithContext(ctx).Info("health check passed")
	return &model.HealthResponse{
		Status:  "healthy",
		Message: "service is running",
		Version: s.version,
	}, nil
}
