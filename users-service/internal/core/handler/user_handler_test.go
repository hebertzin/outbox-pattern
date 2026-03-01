package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"users-service/internal/core/domain/entity"
	"users-service/internal/core/handler"
	"users-service/internal/core/usecase"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubUserRepository struct {
	insertFn func(ctx context.Context, user *entity.User, outbox *entity.Outbox) error
}

func (s *stubUserRepository) Insert(_ context.Context, user *entity.User, outbox *entity.Outbox) error {
	if s.insertFn != nil {
		return s.insertFn(context.Background(), user, outbox)
	}
	return nil
}

func newTestHandler(repo *stubUserRepository) *handler.Handler {
	f := usecase.NewFactory(repo, testLogger())
	return handler.NewHandlerFactory(f)
}

func TestHandleCreate_Returns201(t *testing.T) {
	h := newTestHandler(&stubUserRepository{})

	body, _ := json.Marshal(map[string]any{
		"email":    "user@example.com",
		"password": "strongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestHandleCreate_InvalidBody_Returns400(t *testing.T) {
	h := newTestHandler(&stubUserRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreate_ValidationError_Returns400(t *testing.T) {
	h := newTestHandler(&stubUserRepository{})

	body, _ := json.Marshal(map[string]any{
		"email":    "not-an-email",
		"password": "strongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreate_RepositoryError_Returns500(t *testing.T) {
	repo := &stubUserRepository{
		insertFn: func(_ context.Context, _ *entity.User, _ *entity.Outbox) error {
			return errors.New("db error")
		},
	}
	h := newTestHandler(repo)

	body, _ := json.Marshal(map[string]any{
		"email":    "user@example.com",
		"password": "strongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleCreate_ShortPassword_Returns400(t *testing.T) {
	h := newTestHandler(&stubUserRepository{})

	body, _ := json.Marshal(map[string]any{
		"email":    "user@example.com",
		"password": "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
