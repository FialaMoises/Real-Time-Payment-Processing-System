package transaction

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID              uuid.UUID      `json:"id"`
	IdempotencyKey  string         `json:"idempotency_key"`
	Type            string         `json:"type"`
	FromAccountID   sql.NullString `json:"from_account_id,omitempty"`
	ToAccountID     sql.NullString `json:"to_account_id,omitempty"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency"`
	Status          string         `json:"status"`
	Description     string         `json:"description,omitempty"`
	Metadata        interface{}    `json:"metadata,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	ProcessedAt     sql.NullTime   `json:"processed_at,omitempty"`
}

type DepositRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	AccountID      uuid.UUID `json:"account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	Description    string    `json:"description"`
}

type WithdrawalRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	AccountID      uuid.UUID `json:"account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	Description    string    `json:"description"`
}

type TransferRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	FromAccountID  uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID    uuid.UUID `json:"to_account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	Description    string    `json:"description"`
}

type TransactionResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// Transaction types
const (
	TypeDeposit    = "DEPOSIT"
	TypeWithdrawal = "WITHDRAWAL"
	TypeTransfer   = "TRANSFER"
)

// Transaction statuses
const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
	StatusCancelled  = "CANCELLED"
)
