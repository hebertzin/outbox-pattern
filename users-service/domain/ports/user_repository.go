package ports

import (
	"context"
	"users-services/domain/entity"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entity.User, outbox *entity.Outbox) error
}
