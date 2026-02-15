package usecase

import (
	"context"
	"net/http"
	"users-services/domain/entity"
	"users-services/infra/db/repository"
	"users-services/infra/errors"
)

type CreateUserUseCase struct {
	repo repository.UserRepository
}

type UserUseCase interface {
	Execute(ctx context.Context, user *entity.User) (string, *errors.Exception)
}

func NewCreateUserUseCase(repo repository.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo}
}

func (u *CreateUserUseCase) Execute(ctx context.Context, user *entity.User) (string, *errors.Exception) {
	id, err := u.repo.Insert(ctx, user)
	if err != nil {
		return "", errors.BadRequest(errors.WithCode(http.StatusBadRequest), errors.WithMessage("some error has been ocurred"))
	}

	return id, nil
}
