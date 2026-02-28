package usecase

import (
	"log/slog"

	"transaction-service/internal/core/domain/ports"
)

type Factory struct {
	Create  *CreateTransactionUseCase
	Status  *GetTransactionStatusUseCase
	Balance *GetBalanceUseCase
}

func NewFactory(repo ports.TransactionRepository, logger *slog.Logger) *Factory {
	return &Factory{
		Create:  NewCreateTransactionUseCase(repo, logger),
		Status:  NewGetTransactionStatusUseCase(repo, logger),
		Balance: NewGetBalanceUseCase(repo, logger),
	}
}
