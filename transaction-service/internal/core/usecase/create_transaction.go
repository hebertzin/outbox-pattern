package usecase

import (
	"context"
	"encoding/json"

	"transaction-service/internal/core/domain/entity"
	"transaction-service/internal/core/domain/ports"
	apperrors "transaction-service/internal/core/errors"
)

type (
	CreateInput struct {
		FromUserID  string
		ToUserID    string
		Amount      int64
		Description string
	}

	CreateOutput struct {
		ID     string
		Status string
	}

	CreateTransactionUseCase struct {
		repo ports.TransactionRepository
	}
)

func NewCreateTransactionUseCase(repo ports.TransactionRepository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	tx, err := entity.NewTransaction(input.FromUserID, input.ToUserID, input.Amount, input.Description)
	if err != nil {
		return nil, apperrors.BadRequest(apperrors.WithMessage(err.Error()))
	}

	payload, err := json.Marshal(map[string]any{
		"transactionId": tx.ID,
		"fromUserId":    tx.FromUserID,
		"toUserId":      tx.ToUserID,
		"amount":        tx.Amount,
		"description":   tx.Description,
	})
	if err != nil {
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	outbox := entity.NewOutbox("TransactionCreated", string(payload))

	if err := uc.repo.Create(ctx, tx, outbox); err != nil {
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	return &CreateOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
