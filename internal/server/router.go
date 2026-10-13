package server

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func NewRouter(client *mongo.Client) *gin.Engine {
	router := gin.New()

	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	router.Use(requestLogger())
	router.Use(recoveryMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/health/db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			c.Request.Context(),
			3*time.Second,
		)
		defer cancel()

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			slog.Error(
				"database health check failed",
				"error", err,
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "database_unavailable",
					"message": "Database is currently unavailable",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	return router
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		slog.Info(
			"http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error(
					"panic recovered",
					"panic", recovered,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"status": "error",
						"error": gin.H{
							"code":    "internal_server_error",
							"message": "An unexpected error occurred",
						},
					},
				)
			}
		}()

		c.Next()
	}
}
