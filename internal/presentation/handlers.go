package presentation

import (
	"encoding/json"
	"net/http"

	"transaction-service/internal/usecase"
)

type CreateTransactionHandler struct {
	uc usecase.CreateTransactionExecutor
	BaseHandler
}

func NewCreateTransactionHandler(uc usecase.CreateTransactionExecutor) *CreateTransactionHandler {
	return &CreateTransactionHandler{uc: uc}
}

type createTransactionRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	FromUserId  string `json:"fromUserId"`
	ToUserId    string `json:"toUserId"`
}

type createTransactionResponse struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
}

func (h *CreateTransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "Bad Request", "invalid JSON body")
		return
	}

	out, ex := h.uc.Execute(r.Context(), usecase.CreateTransactionInput{
		Amount:      req.Amount,
		Description: req.Description,
		FromUserId:  req.FromUserId,
		ToUserId:    req.ToUserId,
	})

	if ex != nil {
		h.RespondWithError(w, r, ex.Code, ex.Message, ex.Error())
		return
	}

	result := createTransactionResponse{
		TransactionID: out.TransactionID,
		Status:        out.Status,
	}

	h.RespondWithSuccess(w, http.StatusCreated, "transaction created", result)

}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
