package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateWithTx(ctx context.Context, sqlTx *sql.Tx, tx *Transaction) error
	UpdateStatusWithTx(ctx context.Context, sqlTx *sql.Tx, id uuid.UUID, status string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO transactions (idempotency_key, type, from_account_id, to_account_id, amount, currency, status, description, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		tx.IdempotencyKey,
		tx.Type,
		nullStringFromUUID(tx.FromAccountID),
		nullStringFromUUID(tx.ToAccountID),
		tx.Amount,
		tx.Currency,
		tx.Status,
		tx.Description,
		nil, // metadata
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *repository) CreateWithTx(ctx context.Context, sqlTx *sql.Tx, tx *Transaction) error {
	query := `
		INSERT INTO transactions (idempotency_key, type, from_account_id, to_account_id, amount, currency, status, description, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err := sqlTx.QueryRowContext(
		ctx,
		query,
		tx.IdempotencyKey,
		tx.Type,
		nullStringFromUUID(tx.FromAccountID),
		nullStringFromUUID(tx.ToAccountID),
		tx.Amount,
		tx.Currency,
		tx.Status,
		tx.Description,
		nil,
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction with tx: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT id, idempotency_key, type, from_account_id, to_account_id, amount, currency, status, description, created_at, processed_at
		FROM transactions
		WHERE id = $1
	`

	tx := &Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.IdempotencyKey,
		&tx.Type,
		&tx.FromAccountID,
		&tx.ToAccountID,
		&tx.Amount,
		&tx.Currency,
		&tx.Status,
		&tx.Description,
		&tx.CreatedAt,
		&tx.ProcessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by ID: %w", err)
	}

	return tx, nil
}

func (r *repository) GetByIdempotencyKey(ctx context.Context, key string) (*Transaction, error) {
	query := `
		SELECT id, idempotency_key, type, from_account_id, to_account_id, amount, currency, status, description, created_at, processed_at
		FROM transactions
		WHERE idempotency_key = $1
	`

	tx := &Transaction{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&tx.ID,
		&tx.IdempotencyKey,
		&tx.Type,
		&tx.FromAccountID,
		&tx.ToAccountID,
		&tx.Amount,
		&tx.Currency,
		&tx.Status,
		&tx.Description,
		&tx.CreatedAt,
		&tx.ProcessedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not an error, just doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by idempotency key: %w", err)
	}

	return tx, nil
}

func (r *repository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]Transaction, error) {
	query := `
		SELECT id, idempotency_key, type, from_account_id, to_account_id, amount, currency, status, description, created_at, processed_at
		FROM transactions
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, accountID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by account ID: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		tx := Transaction{}
		err := rows.Scan(
			&tx.ID,
			&tx.IdempotencyKey,
			&tx.Type,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Currency,
			&tx.Status,
			&tx.Description,
			&tx.CreatedAt,
			&tx.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE transactions
		SET status = $1::VARCHAR, processed_at = CASE WHEN $1::VARCHAR IN ('COMPLETED', 'FAILED', 'CANCELLED') THEN NOW() ELSE processed_at END
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *repository) UpdateStatusWithTx(ctx context.Context, sqlTx *sql.Tx, id uuid.UUID, status string) error {
	query := `
		UPDATE transactions
		SET status = $1::VARCHAR, processed_at = CASE WHEN $1::VARCHAR IN ('COMPLETED', 'FAILED', 'CANCELLED') THEN NOW() ELSE processed_at END
		WHERE id = $2
	`

	result, err := sqlTx.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status with tx: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

// Helper function
func nullStringFromUUID(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}
