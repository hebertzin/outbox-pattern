package usecase

import (
	"context"

	"transaction-service/internal/core/domain/ports"
	apperrors "transaction-service/internal/core/errors"
)

type (
	StatusOutput struct {
		ID     string
		Status string
	}

	GetTransactionStatusUseCase struct {
		repo ports.TransactionRepository
	}
)

func NewGetTransactionStatusUseCase(repo ports.TransactionRepository) *GetTransactionStatusUseCase {
	return &GetTransactionStatusUseCase{repo: repo}
}

func (uc *GetTransactionStatusUseCase) Execute(ctx context.Context, id string) (*StatusOutput, error) {
	tx, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}
	if tx == nil {
		return nil, apperrors.NotFound(apperrors.WithMessage("transaction not found"))
	}

	return &StatusOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
