package ports

import (
	"context"
	"transaction-service/internal/core/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
	GetBalance(ctx context.Context, userID string) (int64, error)
}
