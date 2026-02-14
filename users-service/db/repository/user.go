package repository

import (
	"database/sql"
	"users-services/entity"
)

type DbUserRepository struct {
	Db *sql.DB
}

func NewUserRepository(db *sql.DB) *DbUserRepository {
	return &DbUserRepository{Db: db}
}

func (u *DbUserRepository) Insert(user *entity.User) {

}
