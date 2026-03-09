package account

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int       `json:"version"`
}

type CreateAccountRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	Currency string    `json:"currency"`
}

type GetBalanceResponse struct {
	AccountID uuid.UUID `json:"account_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	AsOf      time.Time `json:"as_of"`
}

const (
	StatusActive    = "ACTIVE"
	StatusSuspended = "SUSPENDED"
	StatusClosed    = "CLOSED"
)
