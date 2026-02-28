package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"transaction-service/internal/core/domain/entity"
	apperrors "transaction-service/internal/core/errors"
	"transaction-service/internal/core/usecase"
)

type mockTransactionRepository struct {
	createFn     func(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error
	findByIDFn   func(ctx context.Context, id string) (*entity.Transaction, error)
	getBalanceFn func(ctx context.Context, userID string) (int64, error)
}

func (m *mockTransactionRepository) Create(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, outbox)
	}
	return nil
}

func (m *mockTransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTransactionRepository) GetBalance(ctx context.Context, userID string) (int64, error) {
	if m.getBalanceFn != nil {
		return m.getBalanceFn(ctx, userID)
	}
	return 0, nil
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

func TestCreateTransactionUseCase_Success(t *testing.T) {
	repo := &mockTransactionRepository{}
	uc := usecase.NewCreateTransactionUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.CreateInput{
		FromUserID:  "user-1",
		ToUserID:    "user-2",
		Amount:      1000,
		Description: "payment",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.ID == "" {
		t.Fatal("expected non-empty transaction ID")
	}
	if out.Status != string(entity.StatusPending) {
		t.Fatalf("expected status PENDING, got %s", out.Status)
	}
}

func TestCreateTransactionUseCase_SavesOutboxEvent(t *testing.T) {
	var capturedOutbox *entity.Outbox

	repo := &mockTransactionRepository{
		createFn: func(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
			capturedOutbox = outbox
			return nil
		},
	}
	uc := usecase.NewCreateTransactionUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreateInput{
		FromUserID: "user-1",
		ToUserID:   "user-2",
		Amount:     500,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedOutbox == nil {
		t.Fatal("expected outbox event to be saved")
	}
	if capturedOutbox.Type != "TransactionCreated" {
		t.Fatalf("expected outbox type 'TransactionCreated', got '%s'", capturedOutbox.Type)
	}
	if capturedOutbox.Status != entity.OutboxStatusPending {
		t.Fatalf("expected outbox status PENDING, got '%s'", capturedOutbox.Status)
	}
}

func TestCreateTransactionUseCase_SameUser(t *testing.T) {
	repo := &mockTransactionRepository{}
	uc := usecase.NewCreateTransactionUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreateInput{
		FromUserID: "user-1",
		ToUserID:   "user-1",
		Amount:     100,
	})

	exc := assertException(t, err, http.StatusBadRequest)
	if exc.Message != entity.ErrSameUser.Error() {
		t.Fatalf("expected message %q, got %q", entity.ErrSameUser.Error(), exc.Message)
	}
}

func TestCreateTransactionUseCase_InvalidAmount(t *testing.T) {
	cases := []int64{0, -1, -100}
	for _, amount := range cases {
		repo := &mockTransactionRepository{}
		uc := usecase.NewCreateTransactionUseCase(repo)

		_, err := uc.Execute(context.Background(), usecase.CreateInput{
			FromUserID: "user-1",
			ToUserID:   "user-2",
			Amount:     amount,
		})

		exc := assertException(t, err, http.StatusBadRequest)
		if exc.Message != entity.ErrAmountMustBePositive.Error() {
			t.Fatalf("amount %d: expected message %q, got %q", amount, entity.ErrAmountMustBePositive.Error(), exc.Message)
		}
	}
}

func TestCreateTransactionUseCase_RepositoryError(t *testing.T) {
	repo := &mockTransactionRepository{
		createFn: func(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
			return errors.New("db error")
		},
	}
	uc := usecase.NewCreateTransactionUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreateInput{
		FromUserID: "user-1",
		ToUserID:   "user-2",
		Amount:     100,
	})

	_ = assertException(t, err, http.StatusInternalServerError)
}

func TestCreateTransactionUseCase_DoesNotCallRepoOnValidationError(t *testing.T) {
	called := false
	repo := &mockTransactionRepository{
		createFn: func(ctx context.Context, tx *entity.Transaction, outbox *entity.Outbox) error {
			called = true
			return nil
		},
	}
	uc := usecase.NewCreateTransactionUseCase(repo)

	_, _ = uc.Execute(context.Background(), usecase.CreateInput{
		FromUserID: "",
		ToUserID:   "user-2",
		Amount:     100,
	})

	if called {
		t.Fatal("repository should not be called when validation fails")
	}
}
