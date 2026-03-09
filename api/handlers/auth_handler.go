package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/real-time-payments/internal/account"
	"github.com/yourusername/real-time-payments/internal/user"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type AuthHandler struct {
	userService    user.Service
	accountService account.Service
}

func NewAuthHandler(userService user.Service, accountService account.Service) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		accountService: accountService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.CreateUserRequest true "User registration data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Create user
	createdUser, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	// Create account for user
	acc, err := h.accountService.Create(c.Request.Context(), createdUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User created but failed to create account",
			"code":  "ACCOUNT_CREATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":        createdUser.ID,
		"account_id":     acc.ID,
		"account_number": acc.AccountNumber,
		"message":        "User registered successfully",
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.LoginRequest true "Login credentials"
// @Success 200 {object} user.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	response, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, gin.H{
				"error": appErr.Message,
				"code":  appErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to login",
			"code":  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
