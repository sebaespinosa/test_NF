package repository

import (
	"context"
	"testing"
	"time"

	"github.com/sebaespinosa/test_NF/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Farm{}, &model.IrrigationSector{}, &model.IrrigationData{})
	require.NoError(t, err)

	return db
}

func seedBasicData(t *testing.T, db *gorm.DB) {
	t.Helper()

	farm := model.Farm{ID: 1, Name: "Farm A"}
	sector := model.IrrigationSector{ID: 1, FarmID: 1, Name: "Sector A"}
	require.NoError(t, db.Create(&farm).Error)
	require.NoError(t, db.Create(&sector).Error)

	records := []model.IrrigationData{
		{
			FarmID:             1,
			IrrigationSectorID: 1,
			StartTime:          time.Date(2024, 3, 1, 6, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2024, 3, 1, 7, 0, 0, 0, time.UTC),
			NominalAmount:      20,
			RealAmount:         18,
		},
		{
			FarmID:             1,
			IrrigationSectorID: 1,
			StartTime:          time.Date(2024, 3, 1, 18, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2024, 3, 1, 19, 0, 0, 0, time.UTC),
			NominalAmount:      15,
			RealAmount:         12,
		},
		{
			FarmID:             1,
			IrrigationSectorID: 1,
			StartTime:          time.Date(2024, 3, 2, 6, 0, 0, 0, time.UTC),
			EndTime:            time.Date(2024, 3, 2, 7, 0, 0, 0, time.UTC),
			NominalAmount:      25,
			RealAmount:         20,
		},
	}

	require.NoError(t, db.Create(&records).Error)
}

// TestFindByFarmIDAndTimeRange tests basic time-range queries (SQLite-compatible)
func TestFindByFarmIDAndTimeRange(t *testing.T) {
	db := setupTestDB(t)
	seedBasicData(t, db)

	repo := NewIrrigationDataRepository(db)
	ctx := context.Background()

	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 2, 23, 59, 59, 0, time.UTC)

	results, err := repo.FindByFarmIDAndTimeRange(ctx, 1, start, end)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify ordering by start_time
	assert.True(t, results[0].StartTime.Before(results[1].StartTime))
	assert.True(t, results[1].StartTime.Before(results[2].StartTime))
}

// TestCreate tests creating irrigation records
func TestCreate(t *testing.T) {
	db := setupTestDB(t)

	farm := model.Farm{ID: 2, Name: "Farm B"}
	sector := model.IrrigationSector{ID: 2, FarmID: 2, Name: "Sector B"}
	require.NoError(t, db.Create(&farm).Error)
	require.NoError(t, db.Create(&sector).Error)

	repo := NewIrrigationDataRepository(db)
	ctx := context.Background()

	newData := &model.IrrigationData{
		FarmID:             2,
		IrrigationSectorID: 2,
		StartTime:          time.Date(2024, 4, 1, 6, 0, 0, 0, time.UTC),
		EndTime:            time.Date(2024, 4, 1, 7, 0, 0, 0, time.UTC),
		NominalAmount:      30,
		RealAmount:         27,
	}

	err := repo.Create(ctx, newData)
	require.NoError(t, err)
	assert.NotZero(t, newData.ID)

	// Verify it was created
	var count int64
	db.Model(&model.IrrigationData{}).Where("farm_id = ?", 2).Count(&count)
	assert.Equal(t, int64(1), count)
}
