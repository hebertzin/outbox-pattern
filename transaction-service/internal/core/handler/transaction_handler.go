package handler

import (
	"encoding/json"
	"net/http"
	apperrors "transaction-service/internal/core/errors"
	"transaction-service/internal/core/usecase"
)

type (
	// Handler serves the transaction HTTP endpoints.
	Handler struct {
		create  *usecase.CreateTransactionUseCase
		status  *usecase.GetTransactionStatusUseCase
		balance *usecase.GetBalanceUseCase
		Base
	}

	// CreateReq is the request body for creating a transaction.
	CreateReq struct {
		FromUserID  string `json:"from_user_id"  example:"user-abc"`
		ToUserID    string `json:"to_user_id"    example:"user-xyz"`
		Amount      int64  `json:"amount"        example:"1000"`
		Description string `json:"description"   example:"payment for services"`
	}
)

func NewHandler(
	create *usecase.CreateTransactionUseCase,
	status *usecase.GetTransactionStatusUseCase,
	balance *usecase.GetBalanceUseCase,
) *Handler {
	return &Handler{create: create, status: status, balance: balance}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/transactions", h.wrap(h.handleCreate))
	mux.HandleFunc("GET /api/v1/transactions/{id}", h.wrap(h.handleGetStatus))
	mux.HandleFunc("GET /api/v1/balance/{userId}", h.wrap(h.handleGetBalance))
}

// handleCreate godoc
// @Summary      Create transaction
// @Description  Creates a new transaction and saves an outbox event atomically.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        request  body      CreateReq   true  "Transaction payload"
// @Success      201      {object}  Response
// @Failure      400      {object}  ErrResponse
// @Failure      500      {object}  ErrResponse
// @Router       /api/v1/transactions [post]
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return nil
	}

	out, err := h.create.Execute(r.Context(), usecase.CreateInput{
		FromUserID:  req.FromUserID,
		ToUserID:    req.ToUserID,
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusCreated, "transaction created", map[string]string{
		"id":     out.ID,
		"status": out.Status,
	})
	return nil
}

// handleGetStatus godoc
// @Summary      Get transaction status
// @Description  Returns the current status of a transaction by ID.
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  Response
// @Failure      400  {object}  ErrResponse
// @Failure      404  {object}  ErrResponse
// @Failure      500  {object}  ErrResponse
// @Router       /api/v1/transactions/{id} [get]
func (h *Handler) handleGetStatus(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		h.RespondWithError(w, r, http.StatusBadRequest, "missing parameter", "transaction id is required")
		return nil
	}

	out, err := h.status.Execute(r.Context(), id)
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusOK, "ok", out)
	return nil
}

// handleGetBalance godoc
// @Summary      Get user balance
// @Description  Returns the net balance for a given user ID.
// @Tags         balance
// @Produce      json
// @Param        userId  path      string  true  "User ID"
// @Success      200     {object}  Response
// @Failure      400     {object}  ErrResponse
// @Failure      500     {object}  ErrResponse
// @Router       /api/v1/balance/{userId} [get]
func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userId")
	if userID == "" {
		h.RespondWithError(w, r, http.StatusBadRequest, "missing parameter", "user id is required")
		return nil
	}

	out, err := h.balance.Execute(r.Context(), usecase.BalanceInput{UserID: userID})
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusOK, "ok", out)
	return nil
}

func (h *Handler) wrap(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		if exc, ok := err.(*apperrors.Exception); ok {
			h.RespondWithError(w, r, exc.Code, exc.Message, exc.Err)
			return
		}

		h.RespondWithError(w, r, http.StatusInternalServerError, "internal server error", err.Error())
	}
}
