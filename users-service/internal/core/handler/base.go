package handler

import (
	"encoding/json"
	"net/http"
)

type (
	// Response is the standard success envelope.
	Response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}

	// ErrResponse follows a Problem Details-like shape for errors.
	ErrResponse struct {
		Title    string `json:"title"`
		Status   int    `json:"status"`
		Detail   string `json:"detail,omitempty"`
		Instance string `json:"instance,omitempty"`
	}

	// Base embeds shared response helpers into every handler.
	Base struct{}
)

func (b *Base) RespondWithError(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrResponse{
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.String(),
	})
}

func (b *Base) RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
