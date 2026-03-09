package account

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error)
	UpdateBalance(ctx context.Context, accountID uuid.UUID, newBalance float64) error
	UpdateBalanceWithVersion(ctx context.Context, accountID uuid.UUID, newBalance float64, version int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, account *Account) error {
	query := `
		INSERT INTO accounts (user_id, account_number, balance, currency, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at, version
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		account.UserID,
		account.AccountNumber,
		account.Balance,
		account.Currency,
		account.Status,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt, &account.Version)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at, version
		FROM accounts
		WHERE id = $1
	`

	account := &Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.Version,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}

	return account, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at, version
		FROM accounts
		WHERE user_id = $1
	`

	account := &Account{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.Version,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by user ID: %w", err)
	}

	return account, nil
}

func (r *repository) GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error) {
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at, version
		FROM accounts
		WHERE account_number = $1
	`

	account := &Account{}
	err := r.db.QueryRowContext(ctx, query, accountNumber).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.Balance,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
		&account.Version,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by account number: %w", err)
	}

	return account, nil
}

func (r *repository) UpdateBalance(ctx context.Context, accountID uuid.UUID, newBalance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, version = version + 1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, newBalance, accountID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrAccountNotFound
	}

	return nil
}

func (r *repository) UpdateBalanceWithVersion(ctx context.Context, accountID uuid.UUID, newBalance float64, version int) error {
	query := `
		UPDATE accounts
		SET balance = $1, version = version + 1
		WHERE id = $2 AND version = $3
	`

	result, err := r.db.ExecContext(ctx, query, newBalance, accountID, version)
	if err != nil {
		return fmt.Errorf("failed to update balance with version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("optimistic lock failed or account not found")
	}

	return nil
}
