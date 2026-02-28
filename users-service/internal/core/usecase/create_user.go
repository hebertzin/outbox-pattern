package usecase

import (
	"context"
	"encoding/json"
	corerrors "users-services/internal/core/errors"

	"users-services/internal/core/domain/entity"
	"users-services/internal/core/domain/ports"
)

type CreateUserInput struct {
	Email    string
	Password string
}

type CreateUserOutput struct {
	ID string
}

type CreateUserUseCase struct {
	userRepo ports.UserRepository
}

func NewCreateUserUseCase(userRepo ports.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{userRepo: userRepo}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, *corerrors.Exception) {
	user, err := entity.NewUser(input.Email, input.Password)
	if err != nil {
		return nil, corerrors.BadRequest(
			corerrors.WithMessage(err.Error()),
			corerrors.WithError(err),
		)
	}

	payload, err := json.Marshal(map[string]string{
		"userId": user.ID,
		"email":  user.Email,
	})
	if err != nil {
		return nil, corerrors.Unexpected(corerrors.WithError(err))
	}

	outbox := entity.NewOutbox("UserCreated", string(payload))

	if err := uc.userRepo.Insert(ctx, user, outbox); err != nil {
		return nil, corerrors.Unexpected(corerrors.WithError(err))
	}

	return &CreateUserOutput{ID: user.ID}, nil
}
