package presentation

import (
	"encoding/json"
	"net/http"
)

type (
	HttpResponse struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}

	ErrorResponse struct {
		Title    string `json:"title"`
		Status   int    `json:"status"`
		Detail   string `json:"detail,omitempty"`
		Instance string `json:"instance,omitempty"`
	}

	BaseHandler struct{}
)

func (b *BaseHandler) RespondWithError(w http.ResponseWriter, r *http.Request, status int, title string, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.String(),
	}

	json.NewEncoder(w).Encode(resp)
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
