package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"users-services/internal/core/domain/entity"
	"users-services/internal/core/usecase"
)

type mockUserRepository struct {
	insertFn func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error
}

func (m *mockUserRepository) Insert(ctx context.Context, user *entity.User, outbox *entity.Outbox) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, user, outbox)
	}
	return nil
}

func TestCreateUserUseCase_Success(t *testing.T) {
	repo := &mockUserRepository{}
	uc := usecase.NewCreateUserUseCase(repo)

	out, exc := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if exc != nil {
		t.Fatalf("expected no error, got: %v", exc)
	}
	if out.ID == "" {
		t.Fatal("expected non-empty user ID")
	}
}

func TestCreateUserUseCase_SavesOutboxEvent(t *testing.T) {
	var capturedOutbox *entity.Outbox

	repo := &mockUserRepository{
		insertFn: func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error {
			capturedOutbox = outbox
			return nil
		},
	}
	uc := usecase.NewCreateUserUseCase(repo)

	_, exc := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if exc != nil {
		t.Fatalf("expected no error, got: %v", exc)
	}
	if capturedOutbox == nil {
		t.Fatal("expected outbox event to be saved")
	}
	if capturedOutbox.Type != "UserCreated" {
		t.Fatalf("expected outbox type 'UserCreated', got '%s'", capturedOutbox.Type)
	}
	if capturedOutbox.Status != entity.OutboxStatusPending {
		t.Fatalf("expected outbox status PENDING, got '%s'", capturedOutbox.Status)
	}
}

func TestCreateUserUseCase_InvalidEmail(t *testing.T) {
	repo := &mockUserRepository{}
	uc := usecase.NewCreateUserUseCase(repo)

	_, exc := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "invalid-email",
		Password: "password123",
	})

	if exc == nil {
		t.Fatal("expected validation error, got nil")
	}
	if exc.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %d", exc.Code)
	}
}

func TestCreateUserUseCase_ShortPassword(t *testing.T) {
	repo := &mockUserRepository{}
	uc := usecase.NewCreateUserUseCase(repo)

	_, exc := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "abc",
	})

	if exc == nil {
		t.Fatal("expected validation error, got nil")
	}
	if exc.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400, got %d", exc.Code)
	}
}

func TestCreateUserUseCase_RepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		insertFn: func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error {
			return errors.New("database unavailable")
		},
	}
	uc := usecase.NewCreateUserUseCase(repo)

	_, exc := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if exc == nil {
		t.Fatal("expected error from repository, got nil")
	}
	if exc.Code != http.StatusInternalServerError {
		t.Fatalf("expected HTTP 500, got %d", exc.Code)
	}
}

func TestCreateUserUseCase_DoesNotCallRepositoryOnValidationError(t *testing.T) {
	called := false
	repo := &mockUserRepository{
		insertFn: func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error {
			called = true
			return nil
		},
	}
	uc := usecase.NewCreateUserUseCase(repo)

	uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "",
		Password: "password123",
	})

	if called {
		t.Fatal("repository should not be called when validation fails")
	}
}
