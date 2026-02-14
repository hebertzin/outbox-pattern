package usecase

import (
	"context"

	"users-services/db/repository"
	"users-services/entity"
)

type CreateUserUseCase struct {
	repo repository.UserRepository
}

type UserUseCase interface {
	Execute(ctx context.Context, user *entity.User) (string, error)
}

func NewCreateUserUseCase(repo repository.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo}
}

func (u *CreateUserUseCase) Execute(ctx context.Context, user *entity.User) (string, error) {
	id, err := u.repo.Insert(ctx, user)
	if err != nil {
		return "", err
	}

	// Outbox Pattern: o evento "UserCreated" já foi gravado no banco junto do usuário.
	// Um worker separado lê a outbox e publica no broker.
	return id, nil
}
