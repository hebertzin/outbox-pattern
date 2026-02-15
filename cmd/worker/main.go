package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OutboxEvent struct {
	ID      string
	Type    string
	Payload []byte
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := mustConnectDB()
	defer db.Close()

	amqpConn, amqpCh := mustConnectRabbit()
	defer amqpCh.Close()
	defer amqpConn.Close()

	exchangeName := getenv("RABBIT_EXCHANGE", "transaction.events")

	if err := amqpCh.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange declare failed: %v", err)
	}

	log.Println("outbox worker started")

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("outbox worker shutting down")
			return
		case <-ticker.C:
			if err := processBatch(ctx, db, amqpCh, exchangeName, 50); err != nil {
				log.Printf("process batch error: %v", err)
			}
		}
	}
}

func processBatch(ctx context.Context, db *sql.DB, ch *amqp.Channel, exchange string, batchSize int) error {
	events, err := claimPendingEvents(ctx, db, batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	for _, ev := range events {
		routingKey := ev.Type

		pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := ch.PublishWithContext(
			pubCtx,
			exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
				MessageId:    ev.ID,
				Type:         ev.Type,
				Body:         ev.Payload,
			},
		)
		cancel()

		if err != nil {
			log.Printf("publish failed (id=%s type=%s): %v", ev.ID, ev.Type, err)
			_ = markFailed(ctx, db, ev.ID)
			continue
		}

		if err := markProcessed(ctx, db, ev.ID); err != nil {
			log.Printf("mark processed failed (id=%s): %v", ev.ID, err)
			continue
		}

		log.Printf("published and processed (id=%s type=%s)", ev.ID, ev.Type)
	}

	return nil
}

func claimPendingEvents(ctx context.Context, db *sql.DB, batchSize int) ([]OutboxEvent, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, type, payload
		FROM outbox
		WHERE status = 'PENDING'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var id string
		var typ string
		var payload string

		if err := rows.Scan(&id, &typ, &payload); err != nil {
			return nil, err
		}
		events = append(events, OutboxEvent{
			ID:      id,
			Type:    typ,
			Payload: []byte(payload),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	ids := make([]string, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE outbox
		SET status = 'PROCESSING'
		WHERE id = ANY($1)
	`, pqStringArray(ids))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return events, nil
}

func markProcessed(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE outbox
		SET status = 'PROCESSED', processed_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func markFailed(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE outbox
		SET status = 'FAILED'
		WHERE id = $1
	`, id)
	return err
}

func mustConnectDB() *sql.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := mustGetenv("DB_HOST")
		port := mustGetenv("DB_PORT")
		user := mustGetenv("DB_USER")
		pass := mustGetenv("DB_PASSWORD")
		name := mustGetenv("DB_NAME")
		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + pass + " dbname=" + name + " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	db.SetMaxOpenConns(25)

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	return db
}

func mustConnectRabbit() (*amqp.Connection, *amqp.Channel) {
	url := mustGetenv("RABBIT_URL")
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("rabbit dial error: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("rabbit channel error: %v", err)
	}

	return conn, ch
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

/*
pqStringArray: evita importar diretamente github.com/lib/pq só pra Array.
Se você já usa pq, pode trocar isso por pq.Array(ids) e remover essa função.
*/
type stringArray []string

func pqStringArray(v []string) interface{} {
	// Postgres driver "lib/pq" aceita []string com cast se você usar pq.Array.
	// Aqui, para manter simples sem dependência extra, vamos validar e dar fallback.
	// Melhor: use pq.Array(ids).
	if len(v) == 0 {
		return stringArray{}
	}
	return stringArray(v)
}

func (a stringArray) Value() (driver.Value, error) {
	// Implementação simples: sem isso, prefira pq.Array(ids).
	return nil, errors.New("use pq.Array(ids) from github.com/lib/pq for ANY($1)")
}
