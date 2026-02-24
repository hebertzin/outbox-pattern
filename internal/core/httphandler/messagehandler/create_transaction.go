package messagehandler

import (
	"encoding/json"
	"net/http"
	"transaction-service/internal/core/broker"
)

type (
	Handler struct {
		pub *broker.Publisher
	}

	messageEnvelope struct {
		RequestID string `json:"requestId"`
		Message   string `json:"message"`
		Data      []byte `json:"data"`
	}
)

func NewTransactionMessageHandler(b *broker.RabbitMQ) *Handler {
	pub := broker.NewPublisher(b, "transactions", "transaction.created")

	return &Handler{
		pub: pub,
	}
}

func (handler *Handler) handle(w http.ResponseWriter, r *http.Request) error {
	var (
		req         messageEnvelope
		aggregateID string
	)

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = handler.pub.Publish(r.Context(), b, aggregateID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}
