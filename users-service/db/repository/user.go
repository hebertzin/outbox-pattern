package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"users-services/entity"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entity.User) (string, error)
}

type DbUserRepository struct {
	Db *sql.DB
}

func NewUserRepository(db *sql.DB) *DbUserRepository {
	return &DbUserRepository{Db: db}
}

func (r *DbUserRepository) Insert(ctx context.Context, user *entity.User) (string, error) {
	tx, err := r.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	const insertUser = `
		INSERT INTO users (id, email, password, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	if _, err := tx.ExecContext(ctx, insertUser, user.ID, user.Email, user.Password); err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(map[string]any{
		"userId": user.ID,
		"email":  user.Email,
	})
	if err != nil {
		return "", fmt.Errorf("marshal outbox payload: %w", err)
	}

	const insertOutbox = `
		INSERT INTO outbox (id, type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	outboxID := user.ID
	if _, err := tx.ExecContext(ctx, insertOutbox, outboxID, "UserCreated", string(payloadBytes), "PENDING"); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return user.ID, nil
}
