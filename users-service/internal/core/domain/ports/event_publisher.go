package ports

import (
	"context"

	"users-service/internal/core/domain/entity"
)

type EventPublisher interface {
	Publish(ctx context.Context, event *entity.Outbox) error
}
