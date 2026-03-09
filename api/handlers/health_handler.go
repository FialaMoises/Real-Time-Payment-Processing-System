package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

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

	services := make(map[string]string)

	// Check database
	if err := h.db.Ping(); err != nil {
		services["database"] = "down"
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else {
		services["database"] = "up"
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"services":  services,
	})
}
