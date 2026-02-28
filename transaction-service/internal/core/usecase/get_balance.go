package usecase

import (
	"context"

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
		repo ports.TransactionRepository
	}
)

func NewGetBalanceUseCase(repo ports.TransactionRepository) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, input BalanceInput) (*BalanceOutput, error) {
	if input.UserID == "" {
		return nil, apperrors.BadRequest(apperrors.WithMessage("user_id is required"))
	}

	balance, err := uc.repo.GetBalance(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	return &BalanceOutput{
		UserID:  input.UserID,
		Balance: balance,
	}, nil
}
