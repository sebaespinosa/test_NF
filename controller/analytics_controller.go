package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebaespinosa/test_NF/model"
	"github.com/sebaespinosa/test_NF/service"
)

// AnalyticsService is the contract the controller depends on (facilitates mocking in tests).
type AnalyticsService interface {
	GetAnalytics(ctx context.Context, farmID uint, startDate, endDate *time.Time, sectorID *uint, aggregation string, page, limit int) (*model.IrrigationAnalyticsResponse, error)
}

// AnalyticsController handles HTTP requests for irrigation analytics
type AnalyticsController struct {
	service AnalyticsService
}

// NewAnalyticsController creates a new AnalyticsController instance
func NewAnalyticsController(service *service.IrrigationAnalyticsService) *AnalyticsController {
	return &AnalyticsController{service: service}
}

// GetAnalytics handles GET /v1/farms/:farm_id/irrigation/analytics requests
// @Summary Get irrigation analytics for a farm
// @Description Returns comprehensive irrigation analytics with year-over-year comparison, time-series data, and sector breakdown
// @Tags analytics
// @Produce json
// @Param farm_id path int true "Farm ID" example(1)
// @Param start_date query string false "Start date (YYYY-MM-DD format, defaults to 90 days ago)" example(2024-01-01)
// @Param end_date query string false "End date (YYYY-MM-DD format, defaults to today)" example(2024-01-31)
// @Param sector_id query int false "Filter by specific irrigation sector (optional)" example(5)
// @Param aggregation query string false "Aggregation granularity: daily, weekly, monthly (default: daily)" example(daily) enums(daily,weekly,monthly)
// @Param page query int false "Page number for time-series results (1-indexed, default: 1)" example(1)
// @Param limit query int false "Results per page (default: 50, max: 1000, use 'all' for all results)" example(50)
// @Success 200 {object} model.IrrigationAnalyticsResponse "Analytics data with complete year-over-year comparison"
// @Success 206 {object} model.IrrigationAnalyticsResponse "Partial content - previous year data incomplete or missing"
// @Failure 400 {object} map[string]string "Invalid request parameters or date format"
// @Failure 404 {object} map[string]string "Farm not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/farms/{farm_id}/irrigation/analytics [get]
func (c *AnalyticsController) GetAnalytics(ctx *gin.Context) {
	// Parse farm_id from path
	farmIDStr := ctx.Param("farm_id")
	farmID, err := strconv.ParseUint(farmIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid farm_id format"})
		return
	}

	// Parse optional query parameters
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	sectorIDStr := ctx.Query("sector_id")
	aggregation := ctx.DefaultQuery("aggregation", "daily")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "50")

	// Validate aggregation parameter
	if aggregation != "daily" && aggregation != "weekly" && aggregation != "monthly" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation type; must be daily, weekly, or monthly"})
		return
	}

	// Parse page and limit
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 50
	if limitStr == "all" {
		limit = 10000 // High limit for "all" results
	} else {
		limInt, err := strconv.Atoi(limitStr)
		if err == nil && limInt > 0 {
			if limInt > 1000 {
				limInt = 1000 // Cap at 1000
			}
			limit = limInt
		}
	}

	// Parse dates if provided (format: YYYY-MM-DD)
	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStart, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format; use YYYY-MM-DD"})
			return
		}
		startDate = &parsedStart
	}

	if endDateStr != "" {
		parsedEnd, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format; use YYYY-MM-DD"})
			return
		}
		endDate = &parsedEnd
	}

	// Parse optional sector_id filter
	var sectorID *uint
	if sectorIDStr != "" {
		sectorIDUint, err := strconv.ParseUint(sectorIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid sector_id format"})
			return
		}
		sectorID = (*uint)(&[]uint{uint(sectorIDUint)}[0])
	}

	// Call service with request context
	analytics, err := c.service.GetAnalytics(
		ctx.Request.Context(),
		uint(farmID),
		startDate,
		endDate,
		sectorID,
		aggregation,
		page,
		limit,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analytics: " + err.Error()})
		return
	}

	// Determine status code based on YoY data availability
	statusCode := http.StatusOK
	if (analytics.SamePeriod1Y != nil && analytics.SamePeriod1Y.DataIncomplete) ||
		(analytics.SamePeriod2Y != nil && analytics.SamePeriod2Y.DataIncomplete) {
		statusCode = http.StatusPartialContent // 206
	}

	ctx.JSON(statusCode, analytics)
}
