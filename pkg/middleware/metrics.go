package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/pkg/metrics"
)

// MetricsMiddleware records metrics for all HTTP requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment concurrent requests
		metrics.ConcurrentRequests.Inc()
		defer metrics.ConcurrentRequests.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			status,
			duration,
		)
	}
}
