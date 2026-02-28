package ports

import (
	"context"
	"users-services/internal/core/domain/entity"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entity.User, outbox *entity.Outbox) error
}
