package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"users-service/internal/core/domain/entity"
	"users-service/internal/core/domain/ports"
	apperrors "users-service/internal/core/errors"
)

type (
	CreateUserInput struct {
		Email    string
		Password string
	}

	CreateUserOutput struct {
		ID string
	}

	CreateUserUseCase struct {
		userRepo ports.UserRepository
		logger   *slog.Logger
	}
)

func NewCreateUserUseCase(userRepo ports.UserRepository, logger *slog.Logger) *CreateUserUseCase {
	return &CreateUserUseCase{userRepo: userRepo, logger: logger}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	uc.logger.InfoContext(ctx, "create user request")

	user, err := entity.NewUser(input.Email, input.Password)
	if err != nil {
		uc.logger.WarnContext(ctx, "user validation failed", slog.String("reason", err.Error()))
		return nil, apperrors.BadRequest(apperrors.WithMessage(err.Error()))
	}

	outbox := entity.NewOutbox("UserCreated", fmt.Sprintf(`{"userId":%q}`, user.ID))

	if err := uc.userRepo.Insert(ctx, user, outbox); err != nil {
		uc.logger.ErrorContext(ctx, "persist user failed", slog.String("error", err.Error()))
		return nil, apperrors.Unexpected(apperrors.WithError(err))
	}

	uc.logger.InfoContext(ctx, "user created", slog.String("user_id", user.ID))

	return &CreateUserOutput{ID: user.ID}, nil
}
