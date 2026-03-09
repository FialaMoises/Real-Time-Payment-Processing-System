package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/ledger"
)

type LedgerHandler struct {
	ledgerService ledger.Service
}

func NewLedgerHandler(ledgerService ledger.Service) *LedgerHandler {
	return &LedgerHandler{ledgerService: ledgerService}
}

// GetLedgerByAccountID godoc
// @Summary Get ledger history for account
// @Description Get immutable ledger history for a specific account
// @Tags ledger
// @Produce json
// @Security BearerAuth
// @Param account_id path string true "Account ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ledger.LedgerHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Router /ledger/{account_id} [get]
func (h *LedgerHandler) GetLedgerByAccountID(c *gin.Context) {
	accountIDParam := c.Param("account_id")
	accountID, err := uuid.Parse(accountIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
			"code":  "INVALID_ACCOUNT_ID",
		})
		return
	}

	limit := 50
	offset := 0

	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	response, err := h.ledgerService.GetByAccountID(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get ledger history",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
