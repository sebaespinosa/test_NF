package repository

import (
	"context"

	"gorm.io/gorm"
)

// HealthRepository handles health check queries
type HealthRepository struct {
	db *gorm.DB
}

// NewHealthRepository creates a new instance of HealthRepository
func NewHealthRepository(db *gorm.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

// CheckDatabaseHealth verifies the database connection is alive
func (r *HealthRepository) CheckDatabaseHealth(ctx context.Context) error {
	return r.db.WithContext(ctx).Raw("SELECT 1").Row().Scan(new(int))
}
