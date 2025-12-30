//go:build ignore

package main

import (
	"context"
	"log"
	"time"

	"github.com/sebaespinosa/test_NF/config"
	"github.com/sebaespinosa/test_NF/internal/database"
	"github.com/sebaespinosa/test_NF/internal/logging"
	"github.com/sebaespinosa/test_NF/repository"
	"github.com/sebaespinosa/test_NF/service"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logger, err := logging.New(cfg.Server.Env)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("starting database seeding", zap.String("service", cfg.Service.Name))

	db, err := database.Initialize(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	farmRepo := repository.NewFarmRepository(db)
	sectorRepo := repository.NewIrrigationSectorRepository(db)
	dataRepo := repository.NewIrrigationDataRepository(db)

	farmService := service.NewFarmService(farmRepo, logger)
	sectorService := service.NewIrrigationSectorService(sectorRepo, logger)
	dataService := service.NewIrrigationDataService(dataRepo, logger)

	seedFilePath := "./internal/seeds/irrigation_seed.json"
	seedData, err := farmService.LoadSeedData(seedFilePath)
	if err != nil {
		logger.Fatal("failed to load seed data", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := farmService.SeedFarms(ctx, seedData.Farms); err != nil {
		logger.Fatal("failed to seed farms", zap.Error(err))
	}

	if err := sectorService.SeedSectors(ctx, seedData.IrrigationSectors); err != nil {
		logger.Fatal("failed to seed irrigation sectors", zap.Error(err))
	}

	if err := dataService.SeedData(ctx, seedData.IrrigationData); err != nil {
		logger.Fatal("failed to seed irrigation data", zap.Error(err))
	}

	logger.Info("database seeding completed successfully",
		zap.Int("farms", len(seedData.Farms)),
		zap.Int("sectors", len(seedData.IrrigationSectors)),
		zap.Int("irrigation_data", len(seedData.IrrigationData)),
	)
}
