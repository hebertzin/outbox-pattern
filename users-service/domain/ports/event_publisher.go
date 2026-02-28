package ports

import (
	"context"
	"users-services/domain/entity"
)

type EventPublisher interface {
	Publish(ctx context.Context, event *entity.Outbox) error
}
