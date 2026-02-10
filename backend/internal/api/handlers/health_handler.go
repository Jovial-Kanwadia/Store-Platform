package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	// Dependencies for health checks can be added here
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	// Add actual readiness checks here (e.g., K8s connectivity, DB)
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}