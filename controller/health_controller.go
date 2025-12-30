package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sebaespinosa/test_NF/service"
)

// HealthController handles health check HTTP requests
type HealthController struct {
	service *service.HealthService
}

// NewHealthController creates a new instance of HealthController
func NewHealthController(service *service.HealthService) *HealthController {
	return &HealthController{service: service}
}

// GetHealth handles GET /health requests
// @Summary Health check
// @Description Returns the service health status and version
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 500 {object} map[string]string
// @Router /health [get]
func (c *HealthController) GetHealth(ctx *gin.Context) {
	health, err := c.service.GetHealth(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, health)
}

// FarmController handles farm-related HTTP requests
type FarmController struct {
	service *service.FarmService
}

// NewFarmController creates a new instance of FarmController
func NewFarmController(service *service.FarmService) *FarmController {
	return &FarmController{service: service}
}

// GetAllFarms handles GET /test_farms requests
// @Summary Get all farms
// @Description Returns all farms in the database
// @Tags farms
// @Produce json
// @Success 200 {array} model.Farm
// @Failure 500 {object} map[string]string
// @Router /test_farms [get]
func (c *FarmController) GetAllFarms(ctx *gin.Context) {
	farms, err := c.service.GetAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, farms)
}
