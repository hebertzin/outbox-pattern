//go:build e2e

package e2e_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"transaction-service/infra/repository"
	"transaction-service/internal/core/handler"
	"transaction-service/internal/core/usecase"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	db, err := connectDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: connect DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: run migrations: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := repository.NewTransactionRepository(db)
	f := usecase.NewFactory(repo, logger)
	h := handler.NewHandlerFactory(f)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	testServer = httptest.NewServer(mux)
	defer testServer.Close()

	os.Exit(m.Run())
}

// ── Setup helpers ──────────────────────────────────────────────────────────

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envOr("TEST_DB_HOST", "localhost"),
		envOr("TEST_DB_PORT", "5432"),
		envOr("TEST_DB_USER", "postgres"),
		envOr("TEST_DB_PASSWORD", "postgres"),
		envOr("TEST_DB_NAME", "transaction_db"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed (configure via TEST_DB_* env vars): %w", err)
	}
	return db, nil
}

// migrationOrder defines the correct execution order for SQL migrations.
// Explicit ordering avoids issues with alphabetical sort (add_* before create_*).
var migrationOrder = []string{
	"create_transaction_table.sql",
	"create_outbox_table.sql",
	"add_idempotency_key.sql",
	"add_outbox_retry.sql",
}

func runMigrations(db *sql.DB) error {
	dir := filepath.Join("..", "..", "migrations")
	for _, name := range migrationOrder {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", name, err)
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ── HTTP helpers ───────────────────────────────────────────────────────────

type apiResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func doPost(t *testing.T, path string, body any, headers map[string]string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func doGet(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(testServer.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response) apiResponse {
	t.Helper()
	defer resp.Body.Close()
	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out
}

// createTransaction is a helper that creates a transaction and returns its ID.
func createTransaction(t *testing.T, fromUser, toUser string, amount int64) string {
	t.Helper()
	resp := doPost(t, "/api/v1/transactions", map[string]any{
		"from_user_id": fromUser,
		"to_user_id":   toUser,
		"amount":       amount,
	}, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createTransaction: expected 201, got %d", resp.StatusCode)
	}
	body := decodeResponse(t, resp)
	id, _ := body.Data["id"].(string)
	if id == "" {
		t.Fatal("createTransaction: response missing id")
	}
	return id
}

// ── Test cases ─────────────────────────────────────────────────────────────

func TestE2E_CreateTransaction_Returns201(t *testing.T) {
	fromUser := uuid.NewString()
	toUser := uuid.NewString()

	resp := doPost(t, "/api/v1/transactions", map[string]any{
		"from_user_id": fromUser,
		"to_user_id":   toUser,
		"amount":       500,
		"description":  "e2e payment",
	}, nil)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	body := decodeResponse(t, resp)
	if body.Data["id"] == "" {
		t.Fatal("expected non-empty id in response data")
	}
	if body.Data["status"] != "PENDING" {
		t.Fatalf("expected status PENDING, got %v", body.Data["status"])
	}
}

func TestE2E_CreateTransaction_InvalidBody_Returns400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/transactions", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestE2E_CreateTransaction_SameUser_Returns400(t *testing.T) {
	sameUser := uuid.NewString()

	resp := doPost(t, "/api/v1/transactions", map[string]any{
		"from_user_id": sameUser,
		"to_user_id":   sameUser,
		"amount":       100,
	}, nil)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestE2E_CreateTransaction_IdempotentRequest_Returns200(t *testing.T) {
	fromUser := uuid.NewString()
	toUser := uuid.NewString()
	idempotencyKey := uuid.NewString()

	payload := map[string]any{
		"from_user_id": fromUser,
		"to_user_id":   toUser,
		"amount":       750,
	}
	headers := map[string]string{"Idempotency-Key": idempotencyKey}

	first := doPost(t, "/api/v1/transactions", payload, headers)
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first request: expected 201, got %d", first.StatusCode)
	}
	firstBody := decodeResponse(t, first)
	firstID := firstBody.Data["id"].(string)

	second := doPost(t, "/api/v1/transactions", payload, headers)
	if second.StatusCode != http.StatusOK {
		t.Fatalf("second request: expected 200 (idempotent), got %d", second.StatusCode)
	}
	secondBody := decodeResponse(t, second)
	secondID := secondBody.Data["id"].(string)

	if firstID != secondID {
		t.Fatalf("idempotent requests must return the same transaction ID: %s != %s", firstID, secondID)
	}
}

func TestE2E_GetTransactionStatus_Returns200(t *testing.T) {
	fromUser := uuid.NewString()
	toUser := uuid.NewString()
	id := createTransaction(t, fromUser, toUser, 300)

	resp := doGet(t, "/api/v1/transactions/"+id)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(t, resp)
	if body.Data["ID"] == "" && body.Data["id"] == "" {
		t.Fatal("expected non-empty id in response data")
	}
	if body.Data["Status"] != "PENDING" && body.Data["status"] != "PENDING" {
		t.Fatalf("expected status PENDING, got %v", body.Data["Status"])
	}
}

func TestE2E_GetTransactionStatus_NotFound_Returns404(t *testing.T) {
	unknownID := uuid.NewString()

	resp := doGet(t, "/api/v1/transactions/"+unknownID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestE2E_GetBalance_Returns200(t *testing.T) {
	userID := uuid.NewString()

	resp := doGet(t, "/api/v1/balance/"+userID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := decodeResponse(t, resp)
	if body.Data == nil {
		t.Fatal("expected data in response")
	}
}

func TestE2E_GetTransactionStatus_PersistsInDatabase(t *testing.T) {
	fromUser := uuid.NewString()
	toUser := uuid.NewString()
	id := createTransaction(t, fromUser, toUser, 1000)

	// Retrieve the same transaction twice to confirm persistence
	for i := range 2 {
		resp := doGet(t, "/api/v1/transactions/"+id)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("attempt %d: expected 200, got %d", i+1, resp.StatusCode)
		}
		resp.Body.Close()
	}
}
