package repository

import (
	"context"
	"database/sql"
	"fmt"
	"transaction-service/internal/core/domain/entity"
)

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
	dbTx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = dbTx.Rollback() }()

	const insertTx = `
		INSERT INTO transactions (id, amount, description, from_user_id, to_user_id, transaction_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := dbTx.ExecContext(ctx, insertTx,
		tx.ID, tx.Amount, tx.Description, tx.FromUserID, tx.ToUserID, string(tx.Status), tx.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	const insertOutbox = `
		INSERT INTO outbox (id, type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := dbTx.ExecContext(ctx, insertOutbox,
		outbox.ID, outbox.Type, outbox.Payload, string(outbox.Status), outbox.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}

	return dbTx.Commit()
}

func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	const query = `
		SELECT id, amount, description, from_user_id, to_user_id, transaction_status, created_at
		FROM transactions
		WHERE id = $1
	`
	var tx entity.Transaction
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.Amount, &tx.Description,
		&tx.FromUserID, &tx.ToUserID, &tx.Status, &tx.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find transaction: %w", err)
	}
	return &tx, nil
}

func (r *PostgresTransactionRepository) GetBalance(ctx context.Context, userID string) (int64, error) {
	const query = `
		SELECT
			COALESCE(SUM(CASE WHEN to_user_id   = $1 THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN from_user_id = $1 THEN amount ELSE 0 END), 0) AS balance
		FROM transactions
		WHERE transaction_status = 'COMPLETED'
		  AND (from_user_id = $1 OR to_user_id = $1)
	`
	var balance int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}
