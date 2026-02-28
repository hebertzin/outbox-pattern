package handler

import "transaction-service/internal/core/usecase"

func NewHandlerFactory(f *usecase.Factory) *Handler {
	return NewHandler(f.Create, f.Status, f.Balance)
}
