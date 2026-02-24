package messagehandler

import (
	"encoding/json"
	"net/http"
	"transaction-service/internal/core/broker"
)

func NewTransactionMessageHandler(b *broker.RabbitMQ) *Handler {
	pub := broker.NewPublisher(b, "transactions", "transaction.created")

	return &Handler{
		pub: pub,
	}
}

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

func (handler *Handler) handle(w http.ResponseWriter, r *http.Request) error {
	var (
		req messageEnvelope
	)

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = handler.pub.Publish(r.Context(), b)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}
