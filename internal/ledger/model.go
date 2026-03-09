package ledger

import (
	"time"

	"github.com/google/uuid"
)

type LedgerEntry struct {
	ID            int64     `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AccountID     uuid.UUID `json:"account_id"`
	Amount        float64   `json:"amount"` // Positive = credit, Negative = debit
	BalanceAfter  float64   `json:"balance_after"`
	OperationType string    `json:"operation_type"`
	CreatedAt     time.Time `json:"created_at"`
}

type LedgerHistoryResponse struct {
	AccountID uuid.UUID      `json:"account_id"`
	Entries   []LedgerEntry  `json:"entries"`
	Total     int            `json:"total"`
}

const (
	OperationDebit  = "DEBIT"
	OperationCredit = "CREDIT"
)
