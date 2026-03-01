package handler

import "users-service/internal/core/usecase"

func NewHandlerFactory(f *usecase.Factory) *Handler {
	return NewHandler(f.Create)
}
