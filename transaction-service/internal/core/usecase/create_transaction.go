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
		FromUserID     string
		ToUserID       string
		Amount         int64
		Description    string
		IdempotencyKey string
	}

	CreateOutput struct {
		ID         string
		Status     string
		Idempotent bool
	}

	CreateTransactionUseCase struct {
		repo ports.TransactionRepository
	}
)

func NewCreateTransactionUseCase(repo ports.TransactionRepository) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	if input.IdempotencyKey != "" {
		existing, err := uc.repo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil {
			return nil, apperrors.Unexpected(apperrors.WithError(err))
		}
		if existing != nil {
			return &CreateOutput{
				ID:         existing.ID,
				Status:     string(existing.Status),
				Idempotent: true,
			}, nil
		}
	}

	tx, err := entity.NewTransaction(input.FromUserID, input.ToUserID, input.Amount, input.Description)
	if err != nil {
		return nil, apperrors.BadRequest(apperrors.WithMessage(err.Error()))
	}

	if input.IdempotencyKey != "" {
		tx.IdempotencyKey = &input.IdempotencyKey
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
