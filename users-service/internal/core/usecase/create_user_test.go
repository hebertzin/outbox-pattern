package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"users-service/internal/core/domain/entity"
	apperrors "users-service/internal/core/errors"
	"users-service/internal/core/usecase"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockUserRepository struct {
	insertFn func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error
}

func (m *mockUserRepository) Insert(_ context.Context, user *entity.User, outbox *entity.Outbox) error {
	if m.insertFn != nil {
		return m.insertFn(context.Background(), user, outbox)
	}
	return nil
}

func assertException(t *testing.T, err error, expectedCode int) *apperrors.Exception {
	t.Helper()
	exc, ok := err.(*apperrors.Exception)
	if !ok {
		t.Fatalf("expected *apperrors.Exception, got: %T — %v", err, err)
	}
	if exc.Code != expectedCode {
		t.Fatalf("expected code %d, got %d", expectedCode, exc.Code)
	}
	return exc
}

func TestCreateUserUseCase_Success(t *testing.T) {
	repo := &mockUserRepository{}
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	out, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.ID == "" {
		t.Fatal("expected non-empty user ID")
	}
}

func TestCreateUserUseCase_SavesOutboxEvent(t *testing.T) {
	var capturedOutbox *entity.Outbox

	repo := &mockUserRepository{
		insertFn: func(_ context.Context, _ *entity.User, outbox *entity.Outbox) error {
			capturedOutbox = outbox
			return nil
		},
	}
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	_, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
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
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	_, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "invalid-email",
		Password: "password123",
	})

	_ = assertException(t, err, http.StatusBadRequest)
}

func TestCreateUserUseCase_ShortPassword(t *testing.T) {
	repo := &mockUserRepository{}
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	_, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "abc",
	})

	_ = assertException(t, err, http.StatusBadRequest)
}

func TestCreateUserUseCase_RepositoryError(t *testing.T) {
	repo := &mockUserRepository{
		insertFn: func(_ context.Context, _ *entity.User, _ *entity.Outbox) error {
			return errors.New("database unavailable")
		},
	}
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	_, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	_ = assertException(t, err, http.StatusInternalServerError)
}

func TestCreateUserUseCase_DoesNotCallRepositoryOnValidationError(t *testing.T) {
	called := false
	repo := &mockUserRepository{
		insertFn: func(_ context.Context, _ *entity.User, _ *entity.Outbox) error {
			called = true
			return nil
		},
	}
	uc := usecase.NewCreateUserUseCase(repo, testLogger())

	_, _ = uc.Execute(context.Background(), usecase.CreateUserInput{
		Email:    "",
		Password: "password123",
	})

	if called {
		t.Fatal("repository should not be called when validation fails")
	}
}

func TestNewFactory_WiresAllUseCases(t *testing.T) {
	repo := &mockUserRepository{}
	f := usecase.NewFactory(repo, testLogger())

	if f.Create == nil {
		t.Fatal("expected Create use case to be non-nil")
	}
}
