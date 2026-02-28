package repository

import (
	"context"
	"database/sql"
	"users-services/domain/entity"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Insert(ctx context.Context, user *entity.User, outbox *entity.Outbox) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertUser = `
		INSERT INTO users (id, email, password, created_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertUser, user.ID, user.Email, user.Password, user.CreatedAt); err != nil {
		return err
	}

	const insertOutbox = `
		INSERT INTO outbox (id, type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := tx.ExecContext(ctx, insertOutbox, outbox.ID, outbox.Type, outbox.Payload, string(outbox.Status), outbox.CreatedAt); err != nil {
		return err
	}

	return tx.Commit()
}
