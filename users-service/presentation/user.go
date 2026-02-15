package presentation

import (
	"encoding/json"
	"net/http"
	"users-services/domain/entity"
	"users-services/usecase"
)

type (
	UserHandler struct {
		UserService *usecase.UserUseCase
		BaseHandler
	}

	createUserRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
)

func NewUserHandler(userService *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// Create godoc
// @Summary      Create user
// @Description  Creates a new user in the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body createUserRequest true "User data"
// @Success      201 {object} HttpResponse "User created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      409 {object} ErrorResponse "User already exists"
// @Router       /users [post]
func (handler *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondWithError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	u := entity.User{
		Email:    req.Email,
		Password: req.Password,
	}

	_, err := handler.UserService.Execute(r.Context(), &u)
	if err != nil {
		handler.RespondWithError(w, r, err.Code, err.Error(), err.Message)
		return
	}

	handler.RespondWithSuccess(w, http.StatusCreated, "user created successfully with outbox", nil)
}
