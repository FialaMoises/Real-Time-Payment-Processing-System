package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/fraud"
)

type FraudHandler struct {
	fraudService fraud.Service
}

func NewFraudHandler(fraudService fraud.Service) *FraudHandler {
	return &FraudHandler{
		fraudService: fraudService,
	}
}

// GetAlertByTransaction godoc
// @Summary Get fraud alert for a transaction
// @Description Get fraud detection alert for a specific transaction
// @Tags fraud
// @Accept json
// @Produce json
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} fraud.FraudAlert
// @Failure 404 {object} map[string]interface{}
// @Router /fraud/alerts/transaction/{transaction_id} [get]
// @Security BearerAuth
func (h *FraudHandler) GetAlertByTransaction(c *gin.Context) {
	txIDStr := c.Param("transaction_id")
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	alert, err := h.fraudService.GetAlertByTransactionID(c.Request.Context(), txID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// ListPendingAlerts godoc
// @Summary List pending fraud alerts
// @Description Get list of fraud alerts pending review
// @Tags fraud
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} fraud.FraudAlert
// @Failure 500 {object} map[string]interface{}
// @Router /fraud/alerts/pending [get]
// @Security BearerAuth
func (h *FraudHandler) ListPendingAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	alerts, err := h.fraudService.ListPendingAlerts(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"limit":  limit,
		"offset": offset,
	})
}

// ReviewAlert godoc
// @Summary Review a fraud alert
// @Description Mark a fraud alert as confirmed, false positive, or resolved
// @Tags fraud
// @Accept json
// @Produce json
// @Param alert_id path string true "Alert ID"
// @Param body body ReviewAlertRequest true "Review details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /fraud/alerts/{alert_id}/review [post]
// @Security BearerAuth
func (h *FraudHandler) ReviewAlert(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert ID"})
		return
	}

	var req ReviewAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context (set by auth middleware)
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.fraudService.ReviewAlert(c.Request.Context(), alertID, reviewerID.(uuid.UUID), fraud.AlertStatus(req.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to review alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "alert reviewed successfully",
		"alert_id": alertID,
		"status": req.Status,
	})
}

// GetAccountRiskHistory godoc
// @Summary Get account risk history
// @Description Get fraud risk scores history for an account
// @Tags fraud
// @Accept json
// @Produce json
// @Param account_id path string true "Account ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} fraud.FraudScore
// @Failure 400 {object} map[string]interface{}
// @Router /fraud/accounts/{account_id}/risk-history [get]
// @Security BearerAuth
func (h *FraudHandler) GetAccountRiskHistory(c *gin.Context) {
	accountIDStr := c.Param("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	scores, err := h.fraudService.GetAccountRiskHistory(c.Request.Context(), accountID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get risk history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
		"scores":     scores,
	})
}

// Request/Response types
type ReviewAlertRequest struct {
	Status string `json:"status" binding:"required,oneof=CONFIRMED FALSE_POSITIVE RESOLVED"`
}
