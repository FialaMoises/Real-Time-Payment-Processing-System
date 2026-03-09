package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/account"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type AccountHandler struct {
	accountService account.Service
}

func NewAccountHandler(accountService account.Service) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

// GetMyAccount godoc
// @Summary Get current user's account
// @Description Get account details for authenticated user
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} account.Account
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/me [get]
func (h *AccountHandler) GetMyAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	uid := userID.(uuid.UUID)
	acc, err := h.accountService.GetByUserID(c.Request.Context(), uid)
	if err != nil {
		if err == apperrors.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.(*apperrors.AppError).Message,
				"code":  err.(*apperrors.AppError).Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get account",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, acc)
}

// GetAccountByID godoc
// @Summary Get account by ID
// @Description Get account details by account ID
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} account.Account
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetAccountByID(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
			"code":  "INVALID_ACCOUNT_ID",
		})
		return
	}

	acc, err := h.accountService.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if err == apperrors.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.(*apperrors.AppError).Message,
				"code":  err.(*apperrors.AppError).Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get account",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, acc)
}

// GetBalance godoc
// @Summary Get account balance
// @Description Get current balance for an account
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} account.GetBalanceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/{id}/balance [get]
func (h *AccountHandler) GetBalance(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account ID",
			"code":  "INVALID_ACCOUNT_ID",
		})
		return
	}

	balance, err := h.accountService.GetBalance(c.Request.Context(), accountID)
	if err != nil {
		if err == apperrors.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.(*apperrors.AppError).Message,
				"code":  err.(*apperrors.AppError).Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get balance",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}
