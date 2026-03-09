package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/transaction"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type TransactionHandler struct {
	txService transaction.Service
}

func NewTransactionHandler(txService transaction.Service) *TransactionHandler {
	return &TransactionHandler{txService: txService}
}

// Deposit godoc
// @Summary Deposit money
// @Description Deposit money into an account
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.DepositRequest true "Deposit details"
// @Success 201 {object} transaction.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Router /transactions/deposit [post]
func (h *TransactionHandler) Deposit(c *gin.Context) {
	var req transaction.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.txService.Deposit(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		// Log the actual error for debugging
		fmt.Printf("ERROR - Deposit failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process deposit",
			"code":  "INTERNAL_SERVER_ERROR",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Withdrawal godoc
// @Summary Withdraw money
// @Description Withdraw money from an account
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.WithdrawalRequest true "Withdrawal details"
// @Success 201 {object} transaction.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Router /transactions/withdrawal [post]
func (h *TransactionHandler) Withdrawal(c *gin.Context) {
	var req transaction.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.txService.Withdrawal(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process withdrawal",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Transfer godoc
// @Summary Transfer money
// @Description Transfer money between accounts
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.TransferRequest true "Transfer details"
// @Success 201 {object} transaction.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Router /transactions/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	var req transaction.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.txService.Transfer(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process transfer",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetTransactionByID godoc
// @Summary Get transaction by ID
// @Description Get transaction details by ID
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} transaction.Transaction
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	idParam := c.Param("id")
	txID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction ID",
			"code":  "INVALID_TRANSACTION_ID",
		})
		return
	}

	tx, err := h.txService.GetByID(c.Request.Context(), txID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Transaction not found",
				"code":  "TRANSACTION_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get transaction",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetTransactions godoc
// @Summary Get transactions for account
// @Description Get list of transactions for a specific account
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param account_id query string true "Account ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /transactions [get]
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	accountIDParam := c.Query("account_id")
	if accountIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "account_id is required",
			"code":  "MISSING_ACCOUNT_ID",
		})
		return
	}

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

	transactions, err := h.txService.GetByAccountID(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get transactions",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
		"total":        len(transactions),
	})
}
