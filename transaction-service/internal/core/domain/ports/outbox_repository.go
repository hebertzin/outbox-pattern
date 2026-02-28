package ports

import (
	"context"
	"transaction-service/internal/core/domain/entity"
)

type OutboxRepository interface {
	FetchPending(ctx context.Context, limit int) ([]*entity.Outbox, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
}
