package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/internal/account"
	"github.com/yourusername/real-time-payments/internal/ledger"
	apperrors "github.com/yourusername/real-time-payments/pkg/errors"
	"github.com/yourusername/real-time-payments/pkg/logger"
	"github.com/yourusername/real-time-payments/pkg/metrics"
)

type Service interface {
	Deposit(ctx context.Context, req *DepositRequest) (*TransactionResponse, error)
	Withdrawal(ctx context.Context, req *WithdrawalRequest) (*TransactionResponse, error)
	Transfer(ctx context.Context, req *TransferRequest) (*TransactionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]Transaction, error)
}

type service struct {
	db            *sql.DB
	txRepo        Repository
	accountRepo   account.Repository
	ledgerRepo    ledger.Repository
	accountLocks  sync.Map // Map of account ID to mutex
}

func NewService(
	db *sql.DB,
	txRepo Repository,
	accountRepo account.Repository,
	ledgerRepo ledger.Repository,
) Service {
	return &service{
		db:          db,
		txRepo:      txRepo,
		accountRepo: accountRepo,
		ledgerRepo:  ledgerRepo,
	}
}

// Deposit adds money to an account
func (s *service) Deposit(ctx context.Context, req *DepositRequest) (*TransactionResponse, error) {
	start := time.Now()

	// Check idempotency
	existing, err := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &TransactionResponse{
			TransactionID: existing.ID,
			Status:        existing.Status,
			CreatedAt:     existing.CreatedAt,
		}, nil
	}

	// Lock the account
	mu := s.getAccountLock(req.AccountID.String())
	mu.Lock()
	defer mu.Unlock()

	// Start database transaction
	sqlTx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	// Get current account
	acc, err := s.accountRepo.GetByID(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	// Create transaction record
	tx := &Transaction{
		IdempotencyKey: req.IdempotencyKey,
		Type:           TypeDeposit,
		ToAccountID:    sql.NullString{String: req.AccountID.String(), Valid: true},
		Amount:         req.Amount,
		Currency:       "BRL",
		Status:         StatusProcessing,
		Description:    req.Description,
	}

	if err := s.txRepo.CreateWithTx(ctx, sqlTx, tx); err != nil {
		return nil, err
	}

	// Update account balance
	newBalance := acc.Balance + req.Amount
	if err := s.updateBalanceInTx(ctx, sqlTx, req.AccountID, newBalance); err != nil {
		return nil, err
	}

	// Create ledger entry
	ledgerEntry := &ledger.LedgerEntry{
		TransactionID: tx.ID,
		AccountID:     req.AccountID,
		Amount:        req.Amount, // Positive = credit
		BalanceAfter:  newBalance,
		OperationType: ledger.OperationCredit,
	}

	if err := s.ledgerRepo.CreateWithTx(ctx, sqlTx, ledgerEntry); err != nil {
		return nil, err
	}

	// Update transaction status
	if err := s.txRepo.UpdateStatusWithTx(ctx, sqlTx, tx.ID, StatusCompleted); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := sqlTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info().
		Str("transaction_id", tx.ID.String()).
		Str("type", TypeDeposit).
		Float64("amount", req.Amount).
		Str("account_id", req.AccountID.String()).
		Msg("deposit completed")

	// Record metrics
	duration := time.Since(start).Seconds()
	metrics.RecordTransaction(TypeDeposit, "success", duration, req.Amount)

	return &TransactionResponse{
		TransactionID: tx.ID,
		Status:        StatusCompleted,
		CreatedAt:     tx.CreatedAt,
	}, nil
}

// Withdrawal removes money from an account
func (s *service) Withdrawal(ctx context.Context, req *WithdrawalRequest) (*TransactionResponse, error) {
	start := time.Now()

	// Check idempotency
	existing, err := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &TransactionResponse{
			TransactionID: existing.ID,
			Status:        existing.Status,
			CreatedAt:     existing.CreatedAt,
		}, nil
	}

	// Lock the account
	mu := s.getAccountLock(req.AccountID.String())
	mu.Lock()
	defer mu.Unlock()

	// Start database transaction
	sqlTx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	// Get current account
	acc, err := s.accountRepo.GetByID(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	// Check sufficient balance
	if acc.Balance < req.Amount {
		return nil, apperrors.ErrInsufficientBalance
	}

	// Create transaction record
	tx := &Transaction{
		IdempotencyKey: req.IdempotencyKey,
		Type:           TypeWithdrawal,
		FromAccountID:  sql.NullString{String: req.AccountID.String(), Valid: true},
		Amount:         req.Amount,
		Currency:       "BRL",
		Status:         StatusProcessing,
		Description:    req.Description,
	}

	if err := s.txRepo.CreateWithTx(ctx, sqlTx, tx); err != nil {
		return nil, err
	}

	// Update account balance
	newBalance := acc.Balance - req.Amount
	if err := s.updateBalanceInTx(ctx, sqlTx, req.AccountID, newBalance); err != nil {
		return nil, err
	}

	// Create ledger entry
	ledgerEntry := &ledger.LedgerEntry{
		TransactionID: tx.ID,
		AccountID:     req.AccountID,
		Amount:        -req.Amount, // Negative = debit
		BalanceAfter:  newBalance,
		OperationType: ledger.OperationDebit,
	}

	if err := s.ledgerRepo.CreateWithTx(ctx, sqlTx, ledgerEntry); err != nil {
		return nil, err
	}

	// Update transaction status
	if err := s.txRepo.UpdateStatusWithTx(ctx, sqlTx, tx.ID, StatusCompleted); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := sqlTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info().
		Str("transaction_id", tx.ID.String()).
		Str("type", TypeWithdrawal).
		Float64("amount", req.Amount).
		Str("account_id", req.AccountID.String()).
		Msg("withdrawal completed")

	// Record metrics
	duration := time.Since(start).Seconds()
	metrics.RecordTransaction(TypeWithdrawal, "success", duration, req.Amount)

	return &TransactionResponse{
		TransactionID: tx.ID,
		Status:        StatusCompleted,
		CreatedAt:     tx.CreatedAt,
	}, nil
}

// Transfer moves money between two accounts (WITH DEADLOCK PREVENTION)
func (s *service) Transfer(ctx context.Context, req *TransferRequest) (*TransactionResponse, error) {
	start := time.Now()

	// Validate
	if req.FromAccountID == req.ToAccountID {
		return nil, apperrors.New("INVALID_TRANSFER", "Cannot transfer to same account", 400)
	}

	// Check idempotency
	existing, err := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &TransactionResponse{
			TransactionID: existing.ID,
			Status:        existing.Status,
			CreatedAt:     existing.CreatedAt,
		}, nil
	}

	// Lock accounts in sorted order to prevent deadlock
	accountIDs := []string{req.FromAccountID.String(), req.ToAccountID.String()}
	sort.Strings(accountIDs)

	locks := make([]*sync.Mutex, 2)
	for i, id := range accountIDs {
		locks[i] = s.getAccountLock(id)
		locks[i].Lock()
		defer locks[i].Unlock()
	}

	// Start database transaction
	sqlTx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	// Get both accounts
	fromAcc, err := s.accountRepo.GetByID(ctx, req.FromAccountID)
	if err != nil {
		return nil, err
	}

	toAcc, err := s.accountRepo.GetByID(ctx, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	// Check sufficient balance
	if fromAcc.Balance < req.Amount {
		return nil, apperrors.ErrInsufficientBalance
	}

	// Create transaction record
	tx := &Transaction{
		IdempotencyKey: req.IdempotencyKey,
		Type:           TypeTransfer,
		FromAccountID:  sql.NullString{String: req.FromAccountID.String(), Valid: true},
		ToAccountID:    sql.NullString{String: req.ToAccountID.String(), Valid: true},
		Amount:         req.Amount,
		Currency:       "BRL",
		Status:         StatusProcessing,
		Description:    req.Description,
	}

	if err := s.txRepo.CreateWithTx(ctx, sqlTx, tx); err != nil {
		return nil, err
	}

	// Debit from source account
	newFromBalance := fromAcc.Balance - req.Amount
	if err := s.updateBalanceInTx(ctx, sqlTx, req.FromAccountID, newFromBalance); err != nil {
		return nil, err
	}

	// Create debit ledger entry
	debitEntry := &ledger.LedgerEntry{
		TransactionID: tx.ID,
		AccountID:     req.FromAccountID,
		Amount:        -req.Amount,
		BalanceAfter:  newFromBalance,
		OperationType: ledger.OperationDebit,
	}
	if err := s.ledgerRepo.CreateWithTx(ctx, sqlTx, debitEntry); err != nil {
		return nil, err
	}

	// Credit to destination account
	newToBalance := toAcc.Balance + req.Amount
	if err := s.updateBalanceInTx(ctx, sqlTx, req.ToAccountID, newToBalance); err != nil {
		return nil, err
	}

	// Create credit ledger entry
	creditEntry := &ledger.LedgerEntry{
		TransactionID: tx.ID,
		AccountID:     req.ToAccountID,
		Amount:        req.Amount,
		BalanceAfter:  newToBalance,
		OperationType: ledger.OperationCredit,
	}
	if err := s.ledgerRepo.CreateWithTx(ctx, sqlTx, creditEntry); err != nil {
		return nil, err
	}

	// Update transaction status
	if err := s.txRepo.UpdateStatusWithTx(ctx, sqlTx, tx.ID, StatusCompleted); err != nil {
		return nil, err
	}

	// Commit transaction
	if err := sqlTx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info().
		Str("transaction_id", tx.ID.String()).
		Str("type", TypeTransfer).
		Float64("amount", req.Amount).
		Str("from_account", req.FromAccountID.String()).
		Str("to_account", req.ToAccountID.String()).
		Msg("transfer completed")

	// Record metrics
	duration := time.Since(start).Seconds()
	metrics.RecordTransaction(TypeTransfer, "success", duration, req.Amount)

	return &TransactionResponse{
		TransactionID: tx.ID,
		Status:        StatusCompleted,
		CreatedAt:     tx.CreatedAt,
	}, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s *service) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]Transaction, error) {
	return s.txRepo.GetByAccountID(ctx, accountID, limit, offset)
}

// Helper: Get or create mutex for account
func (s *service) getAccountLock(accountID string) *sync.Mutex {
	val, _ := s.accountLocks.LoadOrStore(accountID, &sync.Mutex{})
	return val.(*sync.Mutex)
}

// Helper: Update balance within transaction
func (s *service) updateBalanceInTx(ctx context.Context, tx *sql.Tx, accountID uuid.UUID, newBalance float64) error {
	query := `
		UPDATE accounts
		SET balance = $1, version = version + 1
		WHERE id = $2
	`
	_, err := tx.ExecContext(ctx, query, newBalance, accountID)
	return err
}
