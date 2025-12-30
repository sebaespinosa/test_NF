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
func (c *HealthController) GetHealth(ctx *gin.Context) {
	health, err := c.service.GetHealth(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, health)
}
