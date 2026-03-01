package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	URL        string
}

func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{URL: url}
}

func (r *RabbitMQ) Connect() error {
	conn, err := amqp.Dial(r.URL)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("rabbitmq channel: %w", err)
	}

	r.Connection = conn
	r.Channel = ch

	return nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Connection != nil {
		r.Connection.Close()
	}
}
