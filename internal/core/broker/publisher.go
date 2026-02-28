package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	rabbitMq   *RabbitMQ
	exchange   string
	routingKey string
}

func NewPublisher(rabbitMq *RabbitMQ, exchange string, routingKey string) *Publisher {
	return &Publisher{
		rabbitMq,
		exchange,
		routingKey,
	}
}

func (p *Publisher) Publish(ctx context.Context, body []byte, aggregateID uuid.UUID) error {
	err := p.rabbitMq.Channel.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.NewString(),
			Timestamp:    time.Now().UTC(),
			Headers: amqp.Table{
				"aggregate_id": aggregateID,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
