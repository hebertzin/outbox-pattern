package repository

import (
	"context"
	"database/sql"
	"time"
	"transaction-service/internal/core/domain/entity"

	"github.com/lib/pq"
)

type PostgresOutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

// FetchPending atomically claims PENDING events and marks them PROCESSING
// using FOR UPDATE SKIP LOCKED to avoid concurrent processing.
func (r *PostgresOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, type, payload
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}

	var events []*entity.Outbox
	var ids []string

	for rows.Next() {
		var e entity.Outbox
		if err := rows.Scan(&e.ID, &e.Type, &e.Payload); err != nil {
			rows.Close()
			return nil, err
		}
		events = append(events, &e)
		ids = append(ids, e.ID)
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE outbox SET status = 'PROCESSING'
		WHERE id = ANY($1)
	`, pq.Array(ids)); err != nil {
		return nil, err
	}

	return events, tx.Commit()
}

func (r *PostgresOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox SET status = 'PROCESSED', processed_at = $1 WHERE id = $2
	`, time.Now().UTC(), id)
	return err
}

func (r *PostgresOutboxRepository) MarkFailed(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox SET status = 'FAILED' WHERE id = $1
	`, id)
	return err
}
