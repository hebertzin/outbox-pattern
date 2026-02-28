package worker

import (
	"context"
	"log/slog"
	"time"
	"users-services/domain/ports"
)

const defaultBatchSize = 100

type OutboxWorker struct {
	outboxRepo ports.OutboxRepository
	publisher  ports.EventPublisher
	interval   time.Duration
	logger     *slog.Logger
}

func NewOutboxWorker(
	outboxRepo ports.OutboxRepository,
	publisher ports.EventPublisher,
	interval time.Duration,
	logger *slog.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		interval:   interval,
		logger:     logger,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	w.logger.InfoContext(ctx, "outbox worker started", slog.Duration("interval", w.interval))
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "outbox worker stopped")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	events, err := w.outboxRepo.FetchPending(ctx, defaultBatchSize)
	if err != nil {
		w.logger.ErrorContext(ctx, "failed to fetch pending events", slog.String("error", err.Error()))
		return
	}

	if len(events) == 0 {
		return
	}

	w.logger.InfoContext(ctx, "processing outbox events", slog.Int("count", len(events)))

	for _, event := range events {
		if err := w.publisher.Publish(ctx, event); err != nil {
			w.logger.ErrorContext(ctx, "failed to publish event",
				slog.String("event_id", event.ID),
				slog.String("event_type", event.Type),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := w.outboxRepo.MarkProcessed(ctx, event.ID); err != nil {
			w.logger.ErrorContext(ctx, "failed to mark event as processed",
				slog.String("event_id", event.ID),
				slog.String("error", err.Error()),
			)
		}
	}
}
