package presentation

import (
	"encoding/json"
	"errors"
	"net/http"
	"users-services/application/usecase"
	"users-services/domain/entity"
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

// Create godoc
// @Summary      Create user
// @Description  Creates a new user in the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body createUserRequest true "User data"
// @Success      201 {object} HttpResponse "User created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or validation error"
// @Router       /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondWithError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	out, err := h.createUser.Execute(r.Context(), usecase.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, entity.ErrEmailRequired) || errors.Is(err, entity.ErrEmailInvalid) || errors.Is(err, entity.ErrPasswordTooShort) {
			h.RespondWithError(w, r, http.StatusBadRequest, "validation error", err.Error())
			return
		}
		h.RespondWithError(w, r, http.StatusInternalServerError, "internal server error", err.Error())
		return
	}

	h.RespondWithSuccess(w, http.StatusCreated, "user created successfully", map[string]string{"id": out.ID})
}
