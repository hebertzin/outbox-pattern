package presentation

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *CreateTransactionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/transactions", h.Create).Methods(http.MethodPost)
}
