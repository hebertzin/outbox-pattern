package handler

import (
	"encoding/json"
	"net/http"
	corerrors "users-services/internal/core/errors"
	"users-services/internal/core/usecase"
)

type UserHandler struct {
	createUser *usecase.CreateUserUseCase
	BaseHandler
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUserHandler(createUser *usecase.CreateUserUseCase) *UserHandler {
	return &UserHandler{createUser: createUser}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithException(w, corerrors.BadRequest(
			corerrors.WithMessage("invalid request body"),
			corerrors.WithError(err),
		))
		return
	}

	out, exc := h.createUser.Execute(r.Context(), usecase.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if exc != nil {
		h.RespondWithException(w, exc)
		return
	}

	h.RespondWithSuccess(w, http.StatusCreated, "user created successfully", map[string]string{"id": out.ID})
}
