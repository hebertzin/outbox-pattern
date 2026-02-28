package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"users-services/domain/entity"
	"users-services/domain/ports"
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

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	user, err := entity.NewUser(input.Email, input.Password)
	if err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	payload, err := json.Marshal(map[string]string{
		"userId": user.ID,
		"email":  user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	outbox := entity.NewOutbox("UserCreated", string(payload))

	if err := uc.userRepo.Insert(ctx, user, outbox); err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &CreateUserOutput{ID: user.ID}, nil
}
