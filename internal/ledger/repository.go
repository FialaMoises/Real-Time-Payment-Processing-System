package ledger

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, entry *LedgerEntry) error
	CreateWithTx(ctx context.Context, tx *sql.Tx, entry *LedgerEntry) error
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]LedgerEntry, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]LedgerEntry, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, entry *LedgerEntry) error {
	query := `
		INSERT INTO ledger (transaction_id, account_id, amount, balance_after, operation_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		entry.TransactionID,
		entry.AccountID,
		entry.Amount,
		entry.BalanceAfter,
		entry.OperationType,
	).Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}

	return nil
}

func (r *repository) CreateWithTx(ctx context.Context, tx *sql.Tx, entry *LedgerEntry) error {
	query := `
		INSERT INTO ledger (transaction_id, account_id, amount, balance_after, operation_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := tx.QueryRowContext(
		ctx,
		query,
		entry.TransactionID,
		entry.AccountID,
		entry.Amount,
		entry.BalanceAfter,
		entry.OperationType,
	).Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ledger entry with tx: %w", err)
	}

	return nil
}

func (r *repository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, balance_after, operation_type, created_at
		FROM ledger
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries by account ID: %w", err)
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		entry := LedgerEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.Amount,
			&entry.BalanceAfter,
			&entry.OperationType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *repository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, balance_after, operation_type, created_at
		FROM ledger
		WHERE transaction_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries by transaction ID: %w", err)
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		entry := LedgerEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.Amount,
			&entry.BalanceAfter,
			&entry.OperationType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
