package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

var startTime = time.Now()

// Health godoc
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	statusCode := http.StatusOK

	services := make(map[string]interface{})

	// Check database with detailed info
	dbCheck := h.checkDatabase()
	services["database"] = dbCheck

	if dbCheck["status"] != "healthy" {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	// Add system info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(statusCode, gin.H{
		"status":    status,
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).Seconds(),
		"timestamp": time.Now().Format(time.RFC3339),
		"services":  services,
		"system": gin.H{
			"goroutines":     runtime.NumGoroutine(),
			"memory_alloc":   m.Alloc / 1024 / 1024, // MB
			"memory_total":   m.TotalAlloc / 1024 / 1024,
			"memory_sys":     m.Sys / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
	})
}

// Readiness godoc
// @Summary Readiness check
// @Description Check if service is ready to accept requests
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"reason": "database_unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness godoc
// @Summary Liveness check
// @Description Check if service is alive
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"uptime": time.Since(startTime).Seconds(),
	})
}

func (h *HealthHandler) checkDatabase() map[string]interface{} {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.db.PingContext(ctx)
	duration := time.Since(start).Milliseconds()

	result := map[string]interface{}{
		"response_time_ms": duration,
	}

	if err != nil {
		logger.Error().Err(err).Msg("database health check failed")
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		return result
	}

	// Get DB stats
	stats := h.db.Stats()
	result["status"] = "healthy"
	result["connections"] = map[string]interface{}{
		"open":        stats.OpenConnections,
		"in_use":      stats.InUse,
		"idle":        stats.Idle,
		"max_open":    stats.MaxOpenConnections,
		"wait_count":  stats.WaitCount,
		"wait_duration": stats.WaitDuration.Milliseconds(),
	}

	return result
}
