package entity_test

import (
	"errors"
	"testing"
	"users-services/domain/entity"
)

func TestNewUser_Success(t *testing.T) {
	user, err := entity.NewUser("test@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestNewUser_NormalizesEmail(t *testing.T) {
	user, err := entity.NewUser("  TEST@EXAMPLE.COM  ", "password123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected normalized email, got '%s'", user.Email)
	}
}

func TestNewUser_GeneratesUniqueIDs(t *testing.T) {
	u1, _ := entity.NewUser("a@example.com", "password123")
	u2, _ := entity.NewUser("b@example.com", "password123")
	if u1.ID == u2.ID {
		t.Fatal("expected unique IDs for different users")
	}
}

func TestNewUser_EmptyEmail(t *testing.T) {
	_, err := entity.NewUser("", "password123")
	if !errors.Is(err, entity.ErrEmailRequired) {
		t.Fatalf("expected ErrEmailRequired, got: %v", err)
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	cases := []string{"not-an-email", "missing-at.com", "@nodomain"}
	for _, email := range cases {
		_, err := entity.NewUser(email, "password123")
		if !errors.Is(err, entity.ErrEmailInvalid) {
			t.Fatalf("email %q: expected ErrEmailInvalid, got: %v", email, err)
		}
	}
}

func TestNewUser_PasswordTooShort(t *testing.T) {
	_, err := entity.NewUser("test@example.com", "short")
	if !errors.Is(err, entity.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got: %v", err)
	}
}
