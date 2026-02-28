package usecase

import (
	"context"
	"transaction-service/internal/core/domain/ports"
	apperrors "transaction-service/internal/core/errors"
)

type GetBalanceInput struct {
	UserID string
}

type GetBalanceOutput struct {
	UserID  string
	Balance int64
}

type GetBalanceUseCase struct {
	repo ports.TransactionRepository
}

func NewGetBalanceUseCase(repo ports.TransactionRepository) *GetBalanceUseCase {
	return &GetBalanceUseCase{repo: repo}
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, input GetBalanceInput) (*GetBalanceOutput, error) {
	if input.UserID == "" {
		return nil, apperrors.BadRequest(apperrors.WithMessage("user_id is required"))
	}

	balance, err := uc.repo.GetBalance(ctx, input.UserID)
	if err != nil {
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	return &GetBalanceOutput{
		UserID:  input.UserID,
		Balance: balance,
	}, nil
}
