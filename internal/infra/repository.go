package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"transaction-service/internal/domain"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, txEntity *domain.Transaction) error
}

type DBTransactionRepository struct {
	DB *sql.DB
}

func NewTransactionRepository(db *sql.DB) *DBTransactionRepository {
	return &DBTransactionRepository{DB: db}
}

func (r *DBTransactionRepository) CreateTransaction(ctx context.Context, txEntity *domain.Transaction) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
	INSERT INTO transactions (
		id,
		amount,
		description,
		from_user_id,
		to_user_id,
		transaction_status,
		created_at,
		processed_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,NOW(),NULL)
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		txEntity.Id,
		txEntity.Amount,
		txEntity.Description,
		txEntity.FromUserId,
		txEntity.ToUserId,
		txEntity.TransactionStatus,
	)

	if err != nil {
		return err
	}

	payload, err := json.Marshal(txEntity)
	if err != nil {
		return err
	}

	outboxID := txEntity.Id

	// Save the event in the outbox table with status "PENDING" so it can be processed later by a worker and published to a message queue.
	outboxQuery := `
	INSERT INTO outbox (
		id,
		type,
		payload,
		status,
		created_at
	)
	VALUES ($1,$2,$3,'PENDING',NOW())
	`

	_, err = tx.ExecContext(
		ctx,
		outboxQuery,
		outboxID,
		"TransactionCreated",
		payload,
	)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
