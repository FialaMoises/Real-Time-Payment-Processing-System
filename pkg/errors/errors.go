package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common errors
var (
	ErrInvalidRequest      = &AppError{Code: "INVALID_REQUEST", Message: "Invalid request", Status: http.StatusBadRequest}
	ErrUnauthorized        = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden           = &AppError{Code: "FORBIDDEN", Message: "Forbidden", Status: http.StatusForbidden}
	ErrNotFound            = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: http.StatusNotFound}
	ErrInternalServer      = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error", Status: http.StatusInternalServerError}
	ErrDuplicateEntry      = &AppError{Code: "DUPLICATE_ENTRY", Message: "Resource already exists", Status: http.StatusConflict}
	ErrInsufficientBalance = &AppError{Code: "INSUFFICIENT_BALANCE", Message: "Insufficient balance", Status: http.StatusBadRequest}
	ErrInvalidCredentials  = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password", Status: http.StatusUnauthorized}
	ErrAccountNotFound     = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "Account not found", Status: http.StatusNotFound}
	ErrTransactionFailed   = &AppError{Code: "TRANSACTION_FAILED", Message: "Transaction processing failed", Status: http.StatusInternalServerError}
)

func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
