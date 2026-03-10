package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/webhook"
)

type WebhookHandler struct {
	webhookService webhook.Service
}

func NewWebhookHandler(webhookService webhook.Service) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// CreateSubscription creates a new webhook subscription
func (h *WebhookHandler) CreateSubscription(c *gin.Context) {
	var req webhook.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sub, err := h.webhookService.CreateSubscription(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscriptions lists all subscriptions for the authenticated user
func (h *WebhookHandler) GetSubscriptions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subs, err := h.webhookService.GetSubscriptionsByUser(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subs,
		"count":         len(subs),
	})
}

// UpdateSubscription updates a webhook subscription
func (h *WebhookHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	var req webhook.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.webhookService.UpdateSubscription(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated successfully"})
}

// DeleteSubscription deletes a webhook subscription
func (h *WebhookHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.webhookService.DeleteSubscription(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
}

// GetDeliveryHistory gets delivery history for a subscription
func (h *WebhookHandler) GetDeliveryHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	deliveries, err := h.webhookService.GetDeliveryHistory(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get delivery history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deliveries": deliveries,
		"count":      len(deliveries),
		"limit":      limit,
		"offset":     offset,
	})
}

// RetryDelivery manually retries a failed webhook delivery
func (h *WebhookHandler) RetryDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery ID"})
		return
	}

	if err := h.webhookService.RetryDelivery(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retry delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delivery retry initiated"})
}
