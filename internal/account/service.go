package account

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID) (*Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error)
	GetBalance(ctx context.Context, accountID uuid.UUID) (*GetBalanceResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID) (*Account, error) {
	// Check if account already exists for this user
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Generate account number
	accountNumber := generateAccountNumber()

	account := &Account{
		UserID:        userID,
		AccountNumber: accountNumber,
		Balance:       0.00,
		Currency:      "BRL",
		Status:        StatusActive,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *service) GetBalance(ctx context.Context, accountID uuid.UUID) (*GetBalanceResponse, error) {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceResponse{
		AccountID: account.ID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		AsOf:      time.Now(),
	}, nil
}

// Helper function to generate account number
func generateAccountNumber() string {
	// Simple implementation: 0001-XXXXXXXX (branch-number)
	// In production, you'd want a more sophisticated system
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("0001-%s", uuid)
}
