package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error().
					Interface("error", err).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_SERVER_ERROR",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
