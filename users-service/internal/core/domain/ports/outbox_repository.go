package ports

import (
	"context"
	"users-services/internal/core/domain/entity"
)

type OutboxRepository interface {
	FetchPending(ctx context.Context, limit int) ([]*entity.Outbox, error)
	MarkProcessed(ctx context.Context, id string) error
}
