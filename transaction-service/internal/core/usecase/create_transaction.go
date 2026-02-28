package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"transaction-service/internal/core/domain/entity"
	"transaction-service/internal/core/domain/ports"
)

type CreateTransactionInput struct {
	FromUserID  string
	ToUserID    string
	Amount      int64
	Description string
}

type CreateTransactionOutput struct {
	ID     string
	Status string
}

type CreateTransactionUseCase struct {
	repo ports.TransactionRepository
}

func NewCreateTransactionUseCase(repo ports.TransactionRepository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInput) (*CreateTransactionOutput, error) {
	tx, err := entity.NewTransaction(input.FromUserID, input.ToUserID, input.Amount, input.Description)
	if err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"transactionId": tx.ID,
		"fromUserId":    tx.FromUserID,
		"toUserId":      tx.ToUserID,
		"amount":        tx.Amount,
		"description":   tx.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	outbox := entity.NewOutbox("TransactionCreated", string(payload))

	if err := uc.repo.Create(ctx, tx, outbox); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	return &CreateTransactionOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
