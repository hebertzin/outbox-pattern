package handler

import (
	"encoding/json"
	"net/http"

	apperrors "users-service/internal/core/errors"
	"users-service/internal/core/usecase"
)

type (
	// Handler serves the user HTTP endpoints.
	Handler struct {
		create *usecase.CreateUserUseCase
		Base
	}

	// CreateReq is the request body for creating a user.
	CreateReq struct {
		Email    string `json:"email"    example:"user@example.com"`
		Password string `json:"password" example:"s3cur3p@ss"`
	}
)

func NewHandler(create *usecase.CreateUserUseCase) *Handler {
	return &Handler{create: create}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users", h.wrap(h.handleCreate))
}

// handleCreate godoc
// @Summary      Create user
// @Description  Creates a new user and saves an outbox event atomically.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateReq  true  "User payload"
// @Success      201      {object}  Response
// @Failure      400      {object}  ErrResponse
// @Failure      500      {object}  ErrResponse
// @Router       /api/v1/users [post]
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	var req CreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return nil
	}

	out, err := h.create.Execute(r.Context(), usecase.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	h.RespondWithSuccess(w, http.StatusCreated, "user created", map[string]string{
		"id": out.ID,
	})
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
