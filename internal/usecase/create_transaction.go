package usecase

import (
	"context"
	"transaction-service/internal/domain"
)

type (
	createTransactionFeature interface {
		CreateTransaction(ctx context.Context, input domain.Transaction) (*domain.Transaction, error)
	}

	CreateTransactionUseCase struct {
		feature createTransactionFeature
	}
)

func NewCreateTransaction(feature createTransactionFeature) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{feature: feature}
}

func (uc *CreateTransactionUseCase) CreateTransaction(ctx context.Context, input domain.Transaction) (string, error) {
	// here have to be a producer
	res, err := uc.feature.CreateTransaction(ctx, input)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}
