package messagehandler

import (
	"encoding/json"
	"net/http"
	"transaction-service/internal/core/broker"
	"transaction-service/internal/core/httphandler"

	"github.com/google/uuid"
)

type (
	Handler struct {
		pub *broker.Publisher
		httphandler.BaseHandler
	}

	transactionRequest struct {
		RequestID   string  `json:"request_id"`
		FromUserID  string  `json:"from_user_id"`
		ToUserID    string  `json:"to_user_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}
)

func NewTransactionMessageHandler(b *broker.RabbitMQ) *Handler {
	pub := broker.NewPublisher(b, "transactions", "transaction.created")

	return &Handler{
		pub: pub,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /transactions", h.handleError(h.handle))
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) error {
	var (
		req         transactionRequest
		aggregateID uuid.UUID
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "Invalid request payload")
		return nil
	}

	aggregateID, err := uuid.NewUUID()
	if err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "error generating aggregate ID")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if err := h.pub.Publish(r.Context(), payload, aggregateID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (h *Handler) handleError(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			h.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}
