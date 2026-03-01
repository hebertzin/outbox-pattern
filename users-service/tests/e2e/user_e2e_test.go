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

	_ "github.com/lib/pq"

	"users-service/infra/repository"
	"users-service/internal/core/handler"
	"users-service/internal/core/usecase"
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
	repo := repository.NewUserRepository(db)
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
		envOr("TEST_DB_NAME", "users_db"),
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
	"create_user_table.sql",
	"create_outbox_table.sql",
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

func doPost(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
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

// ── Test cases ─────────────────────────────────────────────────────────────

func TestE2E_CreateUser_Returns201(t *testing.T) {
	resp := doPost(t, "/api/v1/users", map[string]any{
		"email":    "e2e-success@example.com",
		"password": "securepassword123",
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	body := decodeResponse(t, resp)
	if body.Data["id"] == "" {
		t.Fatal("expected non-empty id in response data")
	}
}

func TestE2E_CreateUser_InvalidBody_Returns400(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/users", bytes.NewReader([]byte("not-json")))
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

func TestE2E_CreateUser_InvalidEmail_Returns400(t *testing.T) {
	resp := doPost(t, "/api/v1/users", map[string]any{
		"email":    "not-an-email",
		"password": "securepassword123",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestE2E_CreateUser_ShortPassword_Returns400(t *testing.T) {
	resp := doPost(t, "/api/v1/users", map[string]any{
		"email":    "short-pass@example.com",
		"password": "abc",
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestE2E_CreateUser_PersistsInDatabase(t *testing.T) {
	resp := doPost(t, "/api/v1/users", map[string]any{
		"email":    "e2e-persist@example.com",
		"password": "persistpassword123",
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	body := decodeResponse(t, resp)
	id, _ := body.Data["id"].(string)
	if id == "" {
		t.Fatal("expected non-empty id in response")
	}

	// Create again with same email — should conflict (unique constraint)
	resp2 := doPost(t, "/api/v1/users", map[string]any{
		"email":    "e2e-persist@example.com",
		"password": "persistpassword123",
	})
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 on duplicate email, got %d", resp2.StatusCode)
	}
}
