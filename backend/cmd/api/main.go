package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Jovial-Kanwadia/store-platform/backend/internal/api"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/config"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/domain"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/infrastructure/k8s"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/infrastructure/limiter"
	"github.com/Jovial-Kanwadia/store-platform/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("starting store platform backend",
		"environment", cfg.Environment,
		"listen_addr", cfg.ListenAddr,
	)

	k8sClient, err := k8s.NewClient(cfg.KubeConfig)
	if err != nil {
		slog.Error("failed to initialize kubernetes client", "error", err)
		os.Exit(1)
	}

	var limiterSvc domain.Limiter
	if cfg.RedisAddr != "" {
		limiterSvc = limiter.NewRedisLimiter(cfg.RedisAddr)
		slog.Info("using redis rate limiter", "addr", cfg.RedisAddr)
	} else {
		limiterSvc = limiter.NewMemoryLimiter(cfg.Rate, cfg.RateWindow)
		slog.Info("using memory rate limiter",
			"rate", cfg.Rate,
			"window", cfg.RateWindow,
		)
	}

	storeSvc := service.NewStoreService(k8sClient)

	router := api.SetupRouter(storeSvc, limiterSvc, cfg)

	srv := startHTTPServer(cfg.ListenAddr, router)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	slog.Info("received shutdown signal", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server gracefully stopped")
}

func startHTTPServer(addr string, router *gin.Engine) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("http server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	return srv
}