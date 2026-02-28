package worker_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"
	"users-services/internal/core/domain/entity"
	"users-services/internal/core/worker"
)

type mockOutboxRepository struct {
	fetchPendingFn  func(ctx context.Context, limit int) ([]*entity.Outbox, error)
	markProcessedFn func(ctx context.Context, id string) error
}

func (m *mockOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	if m.fetchPendingFn != nil {
		return m.fetchPendingFn(ctx, limit)
	}
	return nil, nil
}

func (m *mockOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	if m.markProcessedFn != nil {
		return m.markProcessedFn(ctx, id)
	}
	return nil
}

type mockEventPublisher struct {
	publishFn func(ctx context.Context, event *entity.Outbox) error
}

func (m *mockEventPublisher) Publish(ctx context.Context, event *entity.Outbox) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, event)
	}
	return nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestOutboxWorker_PublishesAndMarksProcessed(t *testing.T) {
	events := []*entity.Outbox{
		{ID: "event-1", Type: "UserCreated", Payload: `{}`, Status: entity.OutboxStatusPending},
		{ID: "event-2", Type: "UserCreated", Payload: `{}`, Status: entity.OutboxStatusPending},
	}

	var publishedIDs []string
	var markedIDs []string

	repo := &mockOutboxRepository{
		fetchPendingFn: func(ctx context.Context, limit int) ([]*entity.Outbox, error) {
			return events, nil
		},
		markProcessedFn: func(ctx context.Context, id string) error {
			markedIDs = append(markedIDs, id)
			return nil
		},
	}
	pub := &mockEventPublisher{
		publishFn: func(ctx context.Context, event *entity.Outbox) error {
			publishedIDs = append(publishedIDs, event.ID)
			return nil
		},
	}

	w := worker.NewOutboxWorker(repo, pub, time.Millisecond, newTestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if len(publishedIDs) < 2 {
		t.Fatalf("expected at least 2 published events, got %d", len(publishedIDs))
	}
	if len(markedIDs) < 2 {
		t.Fatalf("expected at least 2 marked-processed events, got %d", len(markedIDs))
	}
}

func TestOutboxWorker_ContinuesAfterPublishError(t *testing.T) {
	events := []*entity.Outbox{
		{ID: "fail-event", Type: "UserCreated", Payload: `{}`, Status: entity.OutboxStatusPending},
		{ID: "ok-event", Type: "UserCreated", Payload: `{}`, Status: entity.OutboxStatusPending},
	}

	var markedIDs []string

	repo := &mockOutboxRepository{
		fetchPendingFn: func(ctx context.Context, limit int) ([]*entity.Outbox, error) {
			return events, nil
		},
		markProcessedFn: func(ctx context.Context, id string) error {
			markedIDs = append(markedIDs, id)
			return nil
		},
	}
	pub := &mockEventPublisher{
		publishFn: func(ctx context.Context, event *entity.Outbox) error {
			if event.ID == "fail-event" {
				return errors.New("broker unavailable")
			}
			return nil
		},
	}

	w := worker.NewOutboxWorker(repo, pub, time.Millisecond, newTestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	for _, id := range markedIDs {
		if id == "ok-event" {
			return
		}
	}
	t.Fatal("expected 'ok-event' to be marked as processed even when 'fail-event' fails")
}

func TestOutboxWorker_StopsOnContextCancel(t *testing.T) {
	fetchCount := 0

	repo := &mockOutboxRepository{
		fetchPendingFn: func(ctx context.Context, limit int) ([]*entity.Outbox, error) {
			fetchCount++
			return nil, nil
		},
	}
	pub := &mockEventPublisher{}

	w := worker.NewOutboxWorker(repo, pub, 5*time.Millisecond, newTestLogger())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		w.Run(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not stop after context cancellation")
	}

	if fetchCount == 0 {
		t.Fatal("expected worker to have run at least once")
	}
}

func TestOutboxWorker_SkipsMarkOnFetchError(t *testing.T) {
	markCalled := false

	repo := &mockOutboxRepository{
		fetchPendingFn: func(ctx context.Context, limit int) ([]*entity.Outbox, error) {
			return nil, errors.New("db error")
		},
		markProcessedFn: func(ctx context.Context, id string) error {
			markCalled = true
			return nil
		},
	}
	pub := &mockEventPublisher{}

	w := worker.NewOutboxWorker(repo, pub, time.Millisecond, newTestLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if markCalled {
		t.Fatal("expected MarkProcessed NOT to be called when FetchPending fails")
	}
}
