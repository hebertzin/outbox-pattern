package handler

import (
	"encoding/json"
	"net/http"
)

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

type BaseHandler struct{}

func (b *BaseHandler) RespondWithError(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.String(),
	})
}

func (b *BaseHandler) RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(HttpResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
