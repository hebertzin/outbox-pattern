package entity_test

import (
	"testing"

	"transaction-service/internal/core/domain/entity"
)

func TestNewTransaction_Success(t *testing.T) {
	tx, err := entity.NewTransaction("user-1", "user-2", 500, "payment")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tx.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if tx.FromUserID != "user-1" {
		t.Fatalf("expected from_user_id 'user-1', got '%s'", tx.FromUserID)
	}
	if tx.ToUserID != "user-2" {
		t.Fatalf("expected to_user_id 'user-2', got '%s'", tx.ToUserID)
	}
	if tx.Amount != 500 {
		t.Fatalf("expected amount 500, got %d", tx.Amount)
	}
	if tx.Status != entity.StatusPending {
		t.Fatalf("expected status PENDING, got '%s'", tx.Status)
	}
	if tx.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestNewTransaction_UniqueIDs(t *testing.T) {
	tx1, _ := entity.NewTransaction("user-1", "user-2", 100, "")
	tx2, _ := entity.NewTransaction("user-1", "user-2", 100, "")

	if tx1.ID == tx2.ID {
		t.Fatal("expected unique IDs for each transaction")
	}
}

func TestNewTransaction_MissingFromUserID(t *testing.T) {
	_, err := entity.NewTransaction("", "user-2", 100, "")

	if err != entity.ErrFromUserIDRequired {
		t.Fatalf("expected ErrFromUserIDRequired, got: %v", err)
	}
}

func TestNewTransaction_MissingToUserID(t *testing.T) {
	_, err := entity.NewTransaction("user-1", "", 100, "")

	if err != entity.ErrToUserIDRequired {
		t.Fatalf("expected ErrToUserIDRequired, got: %v", err)
	}
}

func TestNewTransaction_SameUser(t *testing.T) {
	_, err := entity.NewTransaction("user-1", "user-1", 100, "")

	if err != entity.ErrSameUser {
		t.Fatalf("expected ErrSameUser, got: %v", err)
	}
}

func TestNewTransaction_InvalidAmount(t *testing.T) {
	cases := []int64{0, -1, -100}
	for _, amount := range cases {
		_, err := entity.NewTransaction("user-1", "user-2", amount, "")
		if err != entity.ErrAmountMustBePositive {
			t.Fatalf("amount %d: expected ErrAmountMustBePositive, got: %v", amount, err)
		}
	}
}
