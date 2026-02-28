package broker

import (
	"context"
	"fmt"
	"time"

	"transaction-service/internal/core/domain/entity"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitMQPublisher(ch *amqp.Channel, exchange, routingKey string) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		channel:    ch,
		exchange:   exchange,
		routingKey: routingKey,
	}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, event *entity.Outbox) error {
	err := p.channel.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(event.Payload),
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.NewString(),
			Timestamp:    time.Now().UTC(),
			Headers: amqp.Table{
				"event_type":   event.Type,
				"aggregate_id": event.ID,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("publish event %s: %w", event.ID, err)
	}

	return nil
}
