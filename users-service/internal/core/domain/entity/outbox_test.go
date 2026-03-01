package entity_test

import (
	"testing"

	"users-service/internal/core/domain/entity"
)

func TestNewOutbox_DefaultFields(t *testing.T) {
	outbox := entity.NewOutbox("UserCreated", `{"userId":"u-1"}`)

	if outbox.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if outbox.Type != "UserCreated" {
		t.Fatalf("expected type 'UserCreated', got '%s'", outbox.Type)
	}
	if outbox.Payload != `{"userId":"u-1"}` {
		t.Fatalf("unexpected payload: %s", outbox.Payload)
	}
	if outbox.Status != entity.OutboxStatusPending {
		t.Fatalf("expected status PENDING, got '%s'", outbox.Status)
	}
	if outbox.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if outbox.ProcessedAt != nil {
		t.Fatal("expected nil ProcessedAt")
	}
}

func TestNewOutbox_UniqueIDs(t *testing.T) {
	o1 := entity.NewOutbox("UserCreated", "{}")
	o2 := entity.NewOutbox("UserCreated", "{}")

	if o1.ID == o2.ID {
		t.Fatal("expected unique IDs for each outbox event")
	}
}
