package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Url        string
}

func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{
		Url: url,
	}
}

func (r *RabbitMQ) Connect() (*amqp.Connection, *amqp.Channel) {
	connection, err := amqp.Dial(r.Url)
	if err != nil {
		panic(err)
	}

	channel, err := r.Connection.Channel()

	return connection, channel
}
