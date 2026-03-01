package usecase

import (
	"log/slog"

	"users-service/internal/core/domain/ports"
)

type Factory struct {
	Create *CreateUserUseCase
}

func NewFactory(repo ports.UserRepository, logger *slog.Logger) *Factory {
	return &Factory{
		Create: NewCreateUserUseCase(repo, logger),
	}
}
