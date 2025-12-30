package model

import "time"

// EfficiencyRange represents min/max efficiency values
type EfficiencyRange struct {
	Min float64 `json:"min" example:"0.72" description:"Minimum efficiency value"`
	Max float64 `json:"max" example:"0.98" description:"Maximum efficiency value"`
}

// AnalyticsMetrics represents aggregated irrigation metrics for a period
type AnalyticsMetrics struct {
	TotalIrrigationVolumeMM float64          `json:"total_irrigation_volume_mm" example:"450.5" description:"Sum of all real_amount values in mm"`
	TotalIrrigationEvents   int              `json:"total_irrigation_events" example:"120" description:"Count of irrigation events"`
	AverageEfficiency       *float64         `json:"average_efficiency" example:"0.85" description:"Average of (real_amount / nominal_amount); null if no valid data"`
	EfficiencyRange         *EfficiencyRange `json:"efficiency_range" description:"Min and max efficiency values; null if no valid data"`
}

// YoYComparison represents metrics for the same period in a previous year
// Null fields indicate no data was available for that period
type YoYComparison struct {
	TotalIrrigationVolumeMM *float64         `json:"total_irrigation_volume_mm" description:"Sum of all real_amount values in mm; null if no data for period"`
	TotalIrrigationEvents   *int             `json:"total_irrigation_events" description:"Count of irrigation events; null if no data for period"`
	AverageEfficiency       *float64         `json:"average_efficiency" description:"Average efficiency; null if no valid data or period missing"`
	EfficiencyRange         *EfficiencyRange `json:"efficiency_range" description:"Min and max efficiency; null if no valid data or period missing"`
	DataIncomplete          bool             `json:"data_incomplete" description:"True if no data exists for this period"`
	Note                    string           `json:"note,omitempty" description:"Explanation for null/missing data"`
}

// PeriodComparison represents year-over-year percentage changes
type PeriodComparison struct {
	VolumeChangePercent     *float64 `json:"volume_change_percent" example:"7.2" description:"((current - previous) / previous) * 100; null if previous period missing or zero"`
	EventsChangePercent     *float64 `json:"events_change_percent" example:"4.3" description:"((current - previous) / previous) * 100; null if previous period missing or zero"`
	EfficiencyChangePercent *float64 `json:"efficiency_change_percent" example:"3.7" description:"((current - previous) / previous) * 100; null if previous period missing or zero"`
}

// PeriodComparisonSet represents both year-over-year comparisons
type PeriodComparisonSet struct {
	VsPeriod1Y *PeriodComparison `json:"vs_same_period_-1" description:"Percentage changes vs last year; null if previous year missing"`
	VsPeriod2Y *PeriodComparison `json:"vs_same_period_-2" description:"Percentage changes vs two years ago; null if data missing"`
}

// TimeSeriesEntry represents aggregated data for a single time bucket (day/week/month)
type TimeSeriesEntry struct {
	Date            string   `json:"date" example:"2024-01-01" description:"Date or week/month identifier depending on aggregation"`
	NominalAmountMM float64  `json:"nominal_amount_mm" example:"12.5" description:"Sum of nominal amounts for the period"`
	RealAmountMM    float64  `json:"real_amount_mm" example:"10.8" description:"Sum of real amounts for the period"`
	Efficiency      *float64 `json:"efficiency" example:"0.864" description:"Average efficiency for the period: (sum real / sum nominal); null if no valid data"`
	EventCount      int      `json:"event_count" example:"3" description:"Number of irrigation events in this period"`
}

// SectorBreakdown represents aggregated metrics by irrigation sector
type SectorBreakdown struct {
	SectorID          uint     `json:"sector_id" example:"1" description:"Irrigation sector ID"`
	SectorName        string   `json:"sector_name" example:"North Field" description:"Irrigation sector name"`
	TotalVolumeMM     float64  `json:"total_volume_mm" example:"150.2" description:"Sum of real_amount values"`
	AverageEfficiency *float64 `json:"average_efficiency" example:"0.88" description:"Average efficiency for the sector; null if no valid data"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int `json:"page" example:"1" description:"Current page number (1-indexed)"`
	Limit      int `json:"limit" example:"50" description:"Results per page"`
	TotalCount int `json:"total_count" example:"250" description:"Total number of records available"`
	TotalPages int `json:"total_pages" example:"5" description:"Total number of pages: ceil(total_count / limit)"`
}

// IrrigationAnalyticsPeriod represents the date range analyzed
type IrrigationAnalyticsPeriod struct {
	Start time.Time `json:"start" example:"2024-01-01T00:00:00Z" description:"Start of analysis period (UTC)"`
	End   time.Time `json:"end" example:"2024-01-31T23:59:59Z" description:"End of analysis period (UTC)"`
}

// TimeSeries wraps paginated time-series results
type TimeSeries struct {
	Data       []TimeSeriesEntry  `json:"data" description:"Time-series entries for the period"`
	Pagination PaginationMetadata `json:"pagination" description:"Pagination metadata"`
}

// IrrigationAnalyticsResponse is the complete response for irrigation analytics endpoint
type IrrigationAnalyticsResponse struct {
	FarmID           uint                      `json:"farm_id" example:"1" description:"Farm identifier"`
	FarmName         string                    `json:"farm_name" example:"Green Valley Farm" description:"Farm name"`
	Period           IrrigationAnalyticsPeriod `json:"period" description:"Date range analyzed"`
	Aggregation      string                    `json:"aggregation" example:"daily" description:"Aggregation granularity: daily, weekly, monthly"`
	Metrics          AnalyticsMetrics          `json:"metrics" description:"Current period metrics"`
	SamePeriod1Y     *YoYComparison            `json:"same_period_-1" description:"Same period last year; null if no data"`
	SamePeriod2Y     *YoYComparison            `json:"same_period_-2" description:"Same period two years ago; null if no data"`
	PeriodComparison *PeriodComparisonSet      `json:"period_comparison" description:"Year-over-year percentage change analysis"`
	TimeSeries       TimeSeries                `json:"time_series" description:"Aggregated metrics by time bucket with pagination"`
	SectorBreakdown  []SectorBreakdown         `json:"sector_breakdown" description:"Aggregated metrics by sector"`
}
