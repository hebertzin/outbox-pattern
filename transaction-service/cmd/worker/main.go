package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"transaction-service/config"
	infradb "transaction-service/infra/db"
	"transaction-service/infra/repository"
	"transaction-service/internal/core/broker"
	"transaction-service/internal/core/domain/entity"
	"transaction-service/internal/core/domain/ports"
)

const (
	pollInterval   = 500 * time.Millisecond
	batchSize      = 50
	publishTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := infradb.Connect(
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
	)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to database")

	rabbit := broker.NewRabbitMQ(cfg.RabbitMQ.URL)
	if err := rabbit.Connect(); err != nil {
		logger.Error("failed to connect to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer rabbit.Close()

	if err := rabbit.Channel.ExchangeDeclare(cfg.RabbitMQ.Exchange, "topic", true, false, false, false, nil); err != nil {
		logger.Error("failed to declare exchange", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("connected to rabbitmq", slog.String("exchange", cfg.RabbitMQ.Exchange))

	outboxRepo := repository.NewOutboxRepository(db)
	publisher := broker.NewRabbitMQPublisher(rabbit.Channel, cfg.RabbitMQ.Exchange, "transaction.created")

	logger.Info("outbox worker started", slog.Duration("interval", pollInterval))

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("outbox worker stopped")
			return
		case <-ticker.C:
			processBatch(ctx, logger, outboxRepo, publisher)
		}
	}
}

func processBatch(ctx context.Context, logger *slog.Logger, repo ports.OutboxRepository, pub ports.EventPublisher) {
	events, err := repo.FetchPending(ctx, batchSize)
	if err != nil {
		logger.ErrorContext(ctx, "fetch pending events failed", slog.String("error", err.Error()))
		return
	}

	if len(events) == 0 {
		return
	}

	logger.InfoContext(ctx, "processing batch", slog.Int("count", len(events)))

	for _, event := range events {
		processEvent(ctx, logger, repo, pub, event)
	}
}

func processEvent(ctx context.Context, logger *slog.Logger, repo ports.OutboxRepository, pub ports.EventPublisher, event *entity.Outbox) {
	pubCtx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	if err := pub.Publish(pubCtx, event); err != nil {
		logger.ErrorContext(ctx, "publish failed",
			slog.String("event_id", event.ID),
			slog.String("event_type", event.Type),
			slog.String("error", err.Error()),
		)
		_ = repo.MarkForRetry(ctx, event.ID)
		return
	}

	if err := repo.MarkProcessed(ctx, event.ID); err != nil {
		logger.ErrorContext(ctx, "mark processed failed",
			slog.String("event_id", event.ID),
			slog.String("error", err.Error()),
		)
		return
	}

	logger.InfoContext(ctx, "event processed",
		slog.String("event_id", event.ID),
		slog.String("event_type", event.Type),
	)
}
