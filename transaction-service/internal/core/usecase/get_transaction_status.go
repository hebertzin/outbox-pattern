package usecase

import (
	"context"
	"log/slog"

	"transaction-service/internal/core/domain/ports"
	apperrors "transaction-service/internal/core/errors"
)

type (
	StatusOutput struct {
		ID     string
		Status string
	}

	GetTransactionStatusUseCase struct {
		repo   ports.TransactionRepository
		logger *slog.Logger
	}
)

func NewGetTransactionStatusUseCase(repo ports.TransactionRepository, logger *slog.Logger) *GetTransactionStatusUseCase {
	return &GetTransactionStatusUseCase{repo: repo, logger: logger}
}

func (uc *GetTransactionStatusUseCase) Execute(ctx context.Context, id string) (*StatusOutput, error) {
	uc.logger.InfoContext(ctx, "get transaction status request", slog.String("transaction_id", id))

	tx, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.ErrorContext(ctx, "find transaction failed",
			slog.String("transaction_id", id),
			slog.String("error", err.Error()),
		)
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}
	if tx == nil {
		uc.logger.WarnContext(ctx, "transaction not found", slog.String("transaction_id", id))
		return nil, apperrors.NotFound(apperrors.WithMessage("transaction not found"))
	}

	uc.logger.InfoContext(ctx, "transaction found",
		slog.String("transaction_id", tx.ID),
		slog.String("status", string(tx.Status)),
	)

	return &StatusOutput{
		ID:     tx.ID,
		Status: string(tx.Status),
	}, nil
}
