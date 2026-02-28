package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCompleted  TransactionStatus = "COMPLETED"
	StatusFailed     TransactionStatus = "FAILED"
)

var (
	ErrFromUserIDRequired   = errors.New("from_user_id is required")
	ErrToUserIDRequired     = errors.New("to_user_id is required")
	ErrSameUser             = errors.New("from and to user cannot be the same")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
)

type Transaction struct {
	ID          string
	Amount      int64
	Description string
	FromUserID  string
	ToUserID    string
	Status      TransactionStatus
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

func NewTransaction(fromUserID, toUserID string, amount int64, description string) (*Transaction, error) {
	if fromUserID == "" {
		return nil, ErrFromUserIDRequired
	}
	if toUserID == "" {
		return nil, ErrToUserIDRequired
	}
	if fromUserID == toUserID {
		return nil, ErrSameUser
	}
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	return &Transaction{
		ID:          uuid.NewString(),
		Amount:      amount,
		Description: description,
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
