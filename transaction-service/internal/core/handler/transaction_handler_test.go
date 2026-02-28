package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"transaction-service/internal/core/domain/entity"
	apperrors "transaction-service/internal/core/errors"
	"transaction-service/internal/core/handler"
	"transaction-service/internal/core/usecase"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// stubTransactionRepository implements ports.TransactionRepository for handler tests.
type stubTransactionRepository struct {
	createFn               func(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error
	findByIDFn             func(ctx context.Context, id string) (*entity.Transaction, error)
	findByIdempotencyKeyFn func(ctx context.Context, key string) (*entity.Transaction, error)
	getBalanceFn           func(ctx context.Context, userID string) (int64, error)
}

func (s *stubTransactionRepository) Create(_ context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
	if s.createFn != nil {
		return s.createFn(context.Background(), tx, outbox)
	}
	return nil
}

func (s *stubTransactionRepository) FindByID(_ context.Context, id string) (*entity.Transaction, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(context.Background(), id)
	}
	return nil, nil
}

func (s *stubTransactionRepository) FindByIdempotencyKey(_ context.Context, key string) (*entity.Transaction, error) {
	if s.findByIdempotencyKeyFn != nil {
		return s.findByIdempotencyKeyFn(context.Background(), key)
	}
	return nil, nil
}

func (s *stubTransactionRepository) GetBalance(_ context.Context, userID string) (int64, error) {
	if s.getBalanceFn != nil {
		return s.getBalanceFn(context.Background(), userID)
	}
	return 0, nil
}

func newTestHandler(repo *stubTransactionRepository) *handler.Handler {
	f := usecase.NewFactory(repo, testLogger())
	return handler.NewHandlerFactory(f)
}

func TestHandleCreate_Returns201(t *testing.T) {
	h := newTestHandler(&stubTransactionRepository{})

	body, _ := json.Marshal(map[string]any{
		"from_user_id": "user-1",
		"to_user_id":   "user-2",
		"amount":       1000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestHandleCreate_IdempotentReturns200(t *testing.T) {
	existing := &entity.Transaction{ID: "tx-existing", Status: entity.StatusPending}
	repo := &stubTransactionRepository{
		findByIdempotencyKeyFn: func(_ context.Context, _ string) (*entity.Transaction, error) {
			return existing, nil
		},
	}
	h := newTestHandler(repo)

	body, _ := json.Marshal(map[string]any{
		"from_user_id": "user-1",
		"to_user_id":   "user-2",
		"amount":       1000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Idempotency-Key", "key-abc")
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for idempotent request, got %d", rec.Code)
	}
}

func TestHandleCreate_InvalidBody_Returns400(t *testing.T) {
	h := newTestHandler(&stubTransactionRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreate_ValidationError_Returns400(t *testing.T) {
	h := newTestHandler(&stubTransactionRepository{})

	body, _ := json.Marshal(map[string]any{
		"from_user_id": "user-1",
		"to_user_id":   "user-1",
		"amount":       100,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleGetStatus_Returns200(t *testing.T) {
	repo := &stubTransactionRepository{
		findByIDFn: func(_ context.Context, _ string) (*entity.Transaction, error) {
			return &entity.Transaction{ID: "tx-1", Status: entity.StatusPending}, nil
		},
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/tx-1", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleGetStatus_NotFound_Returns404(t *testing.T) {
	repo := &stubTransactionRepository{
		findByIDFn: func(_ context.Context, _ string) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/unknown", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleGetStatus_RepoError_Returns500(t *testing.T) {
	repo := &stubTransactionRepository{
		findByIDFn: func(_ context.Context, _ string) (*entity.Transaction, error) {
			return nil, apperrors.Unexpected()
		},
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/tx-1", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleGetBalance_Returns200(t *testing.T) {
	repo := &stubTransactionRepository{
		getBalanceFn: func(_ context.Context, _ string) (int64, error) {
			return 2500, nil
		},
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balance/user-1", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
