package handler

import (
	"encoding/json"
	"net/http"
	corerrors "users-services/internal/core/errors"
)

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type BaseHandler struct{}

func (b *BaseHandler) RespondWithException(w http.ResponseWriter, exc *corerrors.Exception) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(exc.Code)
	json.NewEncoder(w).Encode(exc)
}

func (b *BaseHandler) RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := HttpResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(resp)
}
