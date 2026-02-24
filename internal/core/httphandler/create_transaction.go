package httphandler

import (
	"context"
	"encoding/json"
	"transaction-service/internal/core/broker"
)

func NewCreateTransaction(b *broker.RabbitMQ) *Handler {
	pub := broker.NewPublisher(b, "created.transaction", "")

	return &Handler{
		pub: pub,
	}
}

type Handler struct {
	pub *broker.Publisher
}

func (handler *Handler) handle() {
	data, err := json.Marshal(struct {
		Test string
	}{})
	if err != nil {
		panic(err)
	}

	handler.pub.Publish(context.Background(), data)
}
