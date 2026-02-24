package broker

import (
	"context"
	"encoding/json"
	"log"

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

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	for _, data := range body {
		bytes, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}

		err = p.rabbitMq.Channel.PublishWithContext(
			ctx,
			p.exchange,
			p.routingKey,
			true,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         bytes,
				DeliveryMode: amqp.Persistent,
			},
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
