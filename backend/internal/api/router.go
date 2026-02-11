package api

import (
	"github.com/gin-gonic/gin"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/api/handlers"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/api/middleware"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/config"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/service"
)

func SetupRouter(storeSvc *service.StoreService, limiter domain.Limiter, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())

	healthHandler := handlers.NewHealthHandler()
	r.GET("/healthz", healthHandler.Liveness)
	r.GET("/readyz", healthHandler.Readiness)

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimitMiddleware(limiter))
	api.Use(middleware.AuditLogger())

	storeHandler := handlers.NewStoreHandler(storeSvc)
	api.POST("/stores", storeHandler.Create)
	api.GET("/stores", storeHandler.List)
	api.GET("/stores/:name", storeHandler.Get)
	api.DELETE("/stores/:name", storeHandler.Delete)

	return r
}