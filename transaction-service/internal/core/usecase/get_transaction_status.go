package usecase

import (
	"context"
	"fmt"
	"transaction-service/internal/core/domain/ports"
)

type GetTransactionStatusOutput struct {
	ID     string
	Status string
}

type GetTransactionStatusUseCase struct {
	repo ports.TransactionRepository
}

func NewGetTransactionStatusUseCase(repo ports.TransactionRepository) *GetTransactionStatusUseCase {
	return &GetTransactionStatusUseCase{repo: repo}
}

func (uc *GetTransactionStatusUseCase) Execute(ctx context.Context, id string) (*GetTransactionStatusOutput, error) {
	tx, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find transaction: %w", err)
	}

	return &GetTransactionStatusOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
