package repository

import (
	"context"
	"database/sql"
	"time"
	"users-services/domain/entity"
)

type PostgresOutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

func (r *PostgresOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	const query = `
		SELECT id, type, payload, status, created_at
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*entity.Outbox
	for rows.Next() {
		var e entity.Outbox
		if err := rows.Scan(&e.ID, &e.Type, &e.Payload, &e.Status, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, rows.Err()
}

func (r *PostgresOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	const query = `
		UPDATE outbox
		SET status = 'PROCESSED', processed_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}
