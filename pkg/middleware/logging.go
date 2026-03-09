package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log after request
		duration := time.Since(startTime)

		logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("client_ip", c.ClientIP()).
			Msg("request completed")
	}
}
