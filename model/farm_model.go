package model

import "time"

// Farm represents an agricultural farm entity
type Farm struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IrrigationSector represents a subdivision of a farm with irrigation capabilities
type IrrigationSector struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FarmID    uint      `gorm:"not null;index:idx_sector_farm" json:"farm_id"`
	Name      string    `gorm:"not null" json:"name"`
	Farm      Farm      `gorm:"foreignKey:FarmID;constraint:OnDelete:CASCADE" json:"farm,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IrrigationData represents irrigation event data with time-series metrics
// Optimized with composite indexes for common query patterns:
// - Time-range queries by farm
// - Time-range queries by sector
// - General time-based analytics
type IrrigationData struct {
	ID                 uint             `gorm:"primaryKey" json:"id"`
	FarmID             uint             `gorm:"not null;index:idx_irrigation_farm_time,priority:1;index:idx_irrigation_farm" json:"farm_id"`
	IrrigationSectorID uint             `gorm:"not null;index:idx_irrigation_sector_time,priority:1;index:idx_irrigation_sector" json:"irrigation_sector_id"`
	StartTime          time.Time        `gorm:"not null;index:idx_irrigation_farm_time,priority:2;index:idx_irrigation_sector_time,priority:2;index:idx_irrigation_time" json:"start_time"`
	EndTime            time.Time        `gorm:"not null" json:"end_time"`
	NominalAmount      float32          `gorm:"type:numeric(10,2)" json:"nominal_amount"` // in mm
	RealAmount         float32          `gorm:"type:numeric(10,2)" json:"real_amount"`    // in mm
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Farm               Farm             `gorm:"foreignKey:FarmID;constraint:OnDelete:CASCADE" json:"farm,omitempty"`
	IrrigationSector   IrrigationSector `gorm:"foreignKey:IrrigationSectorID;constraint:OnDelete:CASCADE" json:"irrigation_sector,omitempty"`
}
