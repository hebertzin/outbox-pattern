package factory

import (
	"database/sql"
	"users-services/infra/db/repository"

	"users-services/usecase"
)

func UsersFactory(db *sql.DB) *usecase.CreateUserUseCase {
	repo := repository.NewUserRepository(db)
	return usecase.NewCreateUserUseCase(repo)
}
