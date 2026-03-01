package usecase

import (
	"context"
	"encoding/json"
	"log/slog"

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
		repo   ports.TransactionRepository
		logger *slog.Logger
	}
)

func NewCreateTransactionUseCase(repo ports.TransactionRepository, logger *slog.Logger) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{repo: repo, logger: logger}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	uc.logger.InfoContext(ctx, "create transaction request",
		slog.Bool("has_idempotency_key", input.IdempotencyKey != ""),
	)

	if input.IdempotencyKey != "" {
		existing, err := uc.repo.FindByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil {
			uc.logger.ErrorContext(ctx, "find by idempotency key failed", slog.String("error", err.Error()))
			return nil, apperrors.Unexpected(apperrors.WithError(err))
		}
		if existing != nil {
			uc.logger.InfoContext(ctx, "idempotent request — returning existing transaction",
				slog.String("transaction_id", existing.ID),
			)
			return &CreateOutput{
				ID:         existing.ID,
				Status:     string(existing.Status),
				Idempotent: true,
			}, nil
		}
	}

	tx, err := entity.NewTransaction(input.FromUserID, input.ToUserID, input.Amount, input.Description)
	if err != nil {
		uc.logger.WarnContext(ctx, "transaction validation failed", slog.String("reason", err.Error()))
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
		uc.logger.ErrorContext(ctx, "marshal outbox payload failed", slog.String("error", err.Error()))
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	outbox := entity.NewOutbox("TransactionCreated", string(payload))

	if err := uc.repo.Create(ctx, tx, outbox); err != nil {
		uc.logger.ErrorContext(ctx, "persist transaction failed", slog.String("error", err.Error()))
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	uc.logger.InfoContext(ctx, "transaction created",
		slog.String("transaction_id", tx.ID),
		slog.String("status", string(tx.Status)),
	)

	return &CreateOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
