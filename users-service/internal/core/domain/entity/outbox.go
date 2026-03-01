package entity

import (
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "PENDING"
	OutboxStatusProcessing OutboxStatus = "PROCESSING"
	OutboxStatusProcessed  OutboxStatus = "PROCESSED"
	OutboxStatusFailed     OutboxStatus = "FAILED"
)

type Outbox struct {
	ID          string
	Type        string
	Payload     string
	Status      OutboxStatus
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

func NewOutbox(eventType, payload string) *Outbox {
	return &Outbox{
		ID:        uuid.NewString(),
		Type:      eventType,
		Payload:   payload,
		Status:    OutboxStatusPending,
		CreatedAt: time.Now().UTC(),
	}
}
