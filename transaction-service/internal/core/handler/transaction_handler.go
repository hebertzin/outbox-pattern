package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"transaction-service/internal/core/domain/entity"
	"transaction-service/internal/core/usecase"
)

type TransactionHandler struct {
	createUC  *usecase.CreateTransactionUseCase
	statusUC  *usecase.GetTransactionStatusUseCase
	balanceUC *usecase.GetBalanceUseCase
	BaseHandler
}

type createTransactionRequest struct {
	FromUserID  string `json:"from_user_id"`
	ToUserID    string `json:"to_user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

func NewTransactionHandler(
	createUC *usecase.CreateTransactionUseCase,
	statusUC *usecase.GetTransactionStatusUseCase,
	balanceUC *usecase.GetBalanceUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		createUC:  createUC,
		statusUC:  statusUC,
		balanceUC: balanceUC,
	}
}

func (h *TransactionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/transactions", h.wrap(h.handleCreate))
	mux.HandleFunc("GET /api/v1/transactions/{id}", h.wrap(h.handleGetStatus))
	mux.HandleFunc("GET /api/v1/balance/{userId}", h.wrap(h.handleGetBalance))
}

func (h *TransactionHandler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return nil
	}

	out, err := h.createUC.Execute(r.Context(), usecase.CreateTransactionInput{
		FromUserID:  req.FromUserID,
		ToUserID:    req.ToUserID,
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, entity.ErrAmountMustBePositive) ||
			errors.Is(err, entity.ErrSameUser) ||
			errors.Is(err, entity.ErrFromUserIDRequired) ||
			errors.Is(err, entity.ErrToUserIDRequired) {
			h.RespondWithError(w, r, http.StatusBadRequest, "validation error", err.Error())
			return nil
		}
		return err
	}

	h.RespondWithSuccess(w, http.StatusCreated, "transaction created", map[string]string{
		"id":     out.ID,
		"status": out.Status,
	})
	return nil
}

func (h *TransactionHandler) handleGetStatus(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		h.RespondWithError(w, r, http.StatusBadRequest, "missing parameter", "transaction id is required")
		return nil
	}

	out, err := h.statusUC.Execute(r.Context(), id)
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusOK, "ok", out)
	return nil
}

func (h *TransactionHandler) handleGetBalance(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userId")
	if userID == "" {
		h.RespondWithError(w, r, http.StatusBadRequest, "missing parameter", "user id is required")
		return nil
	}

	out, err := h.balanceUC.Execute(r.Context(), usecase.GetBalanceInput{UserID: userID})
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusOK, "ok", out)
	return nil
}

func (h *TransactionHandler) wrap(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			h.RespondWithError(w, r, http.StatusInternalServerError, "internal server error", err.Error())
		}
	}
}
