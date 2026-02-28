package messaging

import (
	"context"
	"log/slog"
	"users-services/domain/entity"
)

// LogEventPublisher is a development publisher that logs events instead of
// sending them to a message broker. Replace with a Kafka or RabbitMQ
// implementation for production use.
type LogEventPublisher struct {
	logger *slog.Logger
}

func NewLogEventPublisher(logger *slog.Logger) *LogEventPublisher {
	return &LogEventPublisher{logger: logger}
}

func (p *LogEventPublisher) Publish(ctx context.Context, event *entity.Outbox) error {
	p.logger.InfoContext(ctx, "event published",
		slog.String("id", event.ID),
		slog.String("type", event.Type),
		slog.String("payload", event.Payload),
	)
	return nil
}
