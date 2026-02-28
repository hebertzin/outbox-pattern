package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email is invalid")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

type User struct {
	ID        string
	Email     string
	Password  string
	CreatedAt time.Time
}

func NewUser(email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return nil, ErrEmailRequired
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, ErrEmailInvalid
	}

	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	return &User{
		ID:        uuid.NewString(),
		Email:     email,
		Password:  password,
		CreatedAt: time.Now().UTC(),
	}, nil
}
