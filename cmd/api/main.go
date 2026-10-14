package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zainrofa/fms-api/internal/config"
	"github.com/zainrofa/fms-api/internal/database"
	"github.com/zainrofa/fms-api/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn(
			".env could not be loaded; using process environment",
			"error", err,
		)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load application config: %w", err)
	}

	switch cfg.AppEnv {
	case "development":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "production":
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	connectCtx, cancelConnect := context.WithTimeout(
		ctx,
		10*time.Second,
	)
	client, err := database.Connect(connectCtx, cfg.MongoURI)
	cancelConnect()

	if err != nil {
		return fmt.Errorf("connect to MongoDB: %w", err)
	}

	defer func() {
		disconnectCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := client.Disconnect(disconnectCtx); err != nil {
			slog.Error("MongoDB disconnect failed", "error", err)
			return
		}

		slog.Info("MongoDB disconnected")
	}()

	// Initialize indexes before accepting HTTP requests.
	db := client.Database("fms")

	indexCtx, cancelIndexes := context.WithTimeout(
		ctx,
		30*time.Second,
	)
	err = database.EnsureIndexes(indexCtx, db)
	cancelIndexes()

	if err != nil {
		return fmt.Errorf("ensure MongoDB indexes: %w", err)
	}

	slog.Info(
		"MongoDB indexes ensured",
		"database", "fms",
	)

	router := server.NewRouter(client, cfg.JWTAccessSecret)

	httpServer := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info(
			"HTTP server starting",
			"port", cfg.AppPort,
			"environment", cfg.AppEnv,
		)
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful HTTP shutdown failed", "error", err)

			if closeErr := httpServer.Close(); closeErr != nil {
				slog.Error(
					"force closing HTTP server failed",
					"error", closeErr,
				)
			}

			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		slog.Info("HTTP server stopped gracefully")

	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
	}

	return nil
}
