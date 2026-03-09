package ledger

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) (*LedgerHistoryResponse, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]LedgerEntry, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) (*LedgerHistoryResponse, error) {
	entries, err := s.repo.GetByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &LedgerHistoryResponse{
		AccountID: accountID,
		Entries:   entries,
		Total:     len(entries),
	}, nil
}

func (s *service) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]LedgerEntry, error) {
	return s.repo.GetByTransactionID(ctx, transactionID)
}
