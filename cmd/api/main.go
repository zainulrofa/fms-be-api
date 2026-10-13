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

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("APP_ENV") == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Cancel this context when the process receives Ctrl+C or SIGTERM.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Limit initial MongoDB connection and ping to 10 seconds.
	connectCtx, cancelConnect := context.WithTimeout(
		ctx,
		10*time.Second,
	)
	client, err := database.Connect(
		connectCtx,
		os.Getenv("MONGODB_URI"),
	)
	cancelConnect()

	if err != nil {
		return fmt.Errorf("connect to MongoDB: %w", err)
	}

	// Always disconnect MongoDB, including when the HTTP server fails.
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

	router := server.NewRouter(client)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("HTTP server starting", "port", port)
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
				slog.Error("force closing HTTP server failed", "error", closeErr)
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
