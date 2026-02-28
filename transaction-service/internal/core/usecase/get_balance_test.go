package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"transaction-service/internal/core/usecase"
)

func TestGetBalanceUseCase_Success(t *testing.T) {
	repo := &mockTransactionRepository{
		getBalanceFn: func(ctx context.Context, userID string) (int64, error) {
			return 1500, nil
		},
	}
	uc := usecase.NewGetBalanceUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.BalanceInput{UserID: "user-1"})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.UserID != "user-1" {
		t.Fatalf("expected userID 'user-1', got '%s'", out.UserID)
	}
	if out.Balance != 1500 {
		t.Fatalf("expected balance 1500, got %d", out.Balance)
	}
}

func TestGetBalanceUseCase_EmptyUserID(t *testing.T) {
	repo := &mockTransactionRepository{}
	uc := usecase.NewGetBalanceUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.BalanceInput{UserID: ""})

	_ = assertException(t, err, http.StatusBadRequest)
}

func TestGetBalanceUseCase_RepositoryError(t *testing.T) {
	repo := &mockTransactionRepository{
		getBalanceFn: func(ctx context.Context, userID string) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	uc := usecase.NewGetBalanceUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.BalanceInput{UserID: "user-1"})

	_ = assertException(t, err, http.StatusInternalServerError)
}
