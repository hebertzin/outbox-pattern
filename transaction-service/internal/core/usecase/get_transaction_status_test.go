package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"transaction-service/internal/core/domain/entity"
	"transaction-service/internal/core/usecase"
)

func TestGetTransactionStatusUseCase_Success(t *testing.T) {
	repo := &mockTransactionRepository{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transaction, error) {
			return &entity.Transaction{
				ID:     "tx-1",
				Status: entity.StatusPending,
			}, nil
		},
	}
	uc := usecase.NewGetTransactionStatusUseCase(repo)

	out, err := uc.Execute(context.Background(), "tx-1")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.ID != "tx-1" {
		t.Fatalf("expected ID 'tx-1', got '%s'", out.ID)
	}
	if out.Status != string(entity.StatusPending) {
		t.Fatalf("expected status PENDING, got '%s'", out.Status)
	}
}

func TestGetTransactionStatusUseCase_NotFound(t *testing.T) {
	repo := &mockTransactionRepository{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	uc := usecase.NewGetTransactionStatusUseCase(repo)

	_, err := uc.Execute(context.Background(), "tx-unknown")

	_ = assertException(t, err, http.StatusNotFound)
}

func TestGetTransactionStatusUseCase_RepositoryError(t *testing.T) {
	repo := &mockTransactionRepository{
		findByIDFn: func(ctx context.Context, id string) (*entity.Transaction, error) {
			return nil, errors.New("db error")
		},
	}
	uc := usecase.NewGetTransactionStatusUseCase(repo)

	_, err := uc.Execute(context.Background(), "tx-1")

	_ = assertException(t, err, http.StatusInternalServerError)
}
