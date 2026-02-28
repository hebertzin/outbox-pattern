package usecase

import (
	"context"
	"log/slog"

	"transaction-service/internal/core/domain/ports"
	apperrors "transaction-service/internal/core/errors"
)

type (
	BalanceInput struct {
		UserID string
	}

	BalanceOutput struct {
		UserID  string
		Balance int64
	}

	GetBalanceUseCase struct {
		repo   ports.TransactionRepository
		logger *slog.Logger
	}
)

func NewGetBalanceUseCase(repo ports.TransactionRepository, logger *slog.Logger) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo, logger: logger}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, input BalanceInput) (*BalanceOutput, error) {
	uc.logger.InfoContext(ctx, "get balance request", slog.String("user_id", input.UserID))

	if input.UserID == "" {
		uc.logger.WarnContext(ctx, "get balance validation failed", slog.String("reason", "user_id is required"))
		return nil, apperrors.BadRequest(apperrors.WithMessage("user_id is required"))
	}

	balance, err := uc.repo.GetBalance(ctx, input.UserID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "get balance failed",
			slog.String("user_id", input.UserID),
			slog.String("error", err.Error()),
		)
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	uc.logger.InfoContext(ctx, "balance retrieved",
		slog.String("user_id", input.UserID),
		slog.Int64("balance", balance),
	)

	return &BalanceOutput{
		UserID:  input.UserID,
		Balance: balance,
	}, nil
}
