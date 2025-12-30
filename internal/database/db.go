package database

import (
	"fmt"

	"github.com/sebaespinosa/test_NF/config"
	"github.com/sebaespinosa/test_NF/model"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

// Initialize initializes the database connection with GORM and runs migrations
func Initialize(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Add OpenTelemetry instrumentation so DB calls appear in traces
	if err := db.Use(otelgorm.NewPlugin(
		otelgorm.WithTracerProvider(otel.GetTracerProvider()),
		otelgorm.WithDBSystem("postgresql"),
	)); err != nil {
		return nil, fmt.Errorf("failed to init otel gorm plugin: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Run AutoMigrate for schema creation
	if err := db.AutoMigrate(
		&model.Farm{},
		&model.IrrigationSector{},
		&model.IrrigationData{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
