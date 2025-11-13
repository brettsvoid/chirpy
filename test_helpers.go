package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"

	_ "github.com/lib/pq"
)

const (
	testDBURL      = "postgres://postgres:postgres@localhost:5432/chirpy_test?sslmode=disable"
	postgresAdminURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	testJWTSecret  = "test-jwt-secret-key"
	testPolkaKey   = "test-polka-key"
)

// ensureTestDBExists creates the test database from scratch
func ensureTestDBExists(t *testing.T) {
	t.Helper()

	// Connect to default postgres database
	db, err := sql.Open("postgres", postgresAdminURL)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Drop the test database if it exists
	t.Log("Dropping existing test database if present...")
	_, err = db.Exec("DROP DATABASE IF EXISTS chirpy_test")
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// Create fresh test database
	t.Log("Creating fresh test database chirpy_test...")
	_, err = db.Exec("CREATE DATABASE chirpy_test")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
}

// setupTestDB creates a test database connection and runs migrations
func setupTestDB(t *testing.T) (*sql.DB, *database.Queries) {
	t.Helper()

	// Ensure test database exists
	ensureTestDBExists(t)

	// Connect to test database
	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Run migrations
	cmd := exec.Command("goose", "postgres", testDBURL, "-dir", "sql/schema/", "up")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Migrations might already be up, which is fine
		if !strings.Contains(string(output), "no migrations to run") && !strings.Contains(string(output), "OK") {
			t.Logf("Migration output: %s", output)
		}
	}

	queries := database.New(db)
	return db, queries
}

// cleanupTestDB truncates all tables to reset state
func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Delete all data (cascades will handle related tables)
	_, err := db.Exec("TRUNCATE users, chirps, refresh_tokens CASCADE")
	if err != nil {
		t.Logf("Cleanup warning: %v", err)
	}
}

// setupTestServer creates a test HTTP server with all routes configured
func setupTestServer(t *testing.T, cfg *apiConfig) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerDetailChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirp)

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerPolkaWebhooks)

	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	return httptest.NewServer(mux)
}

// getJSONField extracts a field from JSON using simple dot notation
func getJSONField(data map[string]interface{}, path string) (interface{}, bool) {
	// Handle array index notation like .[0].body
	if strings.HasPrefix(path, ".[") {
		// This is an array access on the root
		// Not implemented in this simple version
		return nil, false
	}

	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array index like [0]
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, false
			}
			var idx int
			fmt.Sscanf(part, "[%d]", &idx)
			if idx >= len(arr) {
				return nil, false
			}
			current = arr[idx]
			continue
		}

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

// checkJSONField validates a JSON field matches expected value
func checkJSONField(t *testing.T, body []byte, path string, expected interface{}) bool {
	t.Helper()

	// Handle array root
	if strings.HasPrefix(path, ".[") {
		var arr []interface{}
		if err := json.Unmarshal(body, &arr); err != nil {
			t.Errorf("Failed to parse JSON array: %v", err)
			return false
		}

		// Extract index and remaining path
		var idx int
		remainingPath := ""
		if dotIdx := strings.Index(path[1:], "."); dotIdx != -1 {
			fmt.Sscanf(path[1:dotIdx+1], "[%d]", &idx)
			remainingPath = path[dotIdx+2:]
		} else {
			fmt.Sscanf(path, ".[%d]", &idx)
		}

		if idx >= len(arr) {
			t.Errorf("Array index %d out of bounds (length %d)", idx, len(arr))
			return false
		}

		if remainingPath == "" {
			// Comparing the whole element
			if fmt.Sprintf("%v", arr[idx]) != fmt.Sprintf("%v", expected) {
				t.Errorf("Field %s = %v, expected %v", path, arr[idx], expected)
				return false
			}
			return true
		}

		// Navigate deeper
		m, ok := arr[idx].(map[string]interface{})
		if !ok {
			t.Errorf("Array element is not an object")
			return false
		}

		value, ok := getJSONField(m, remainingPath)
		if !ok {
			t.Errorf("Field %s not found in response", path)
			return false
		}

		if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", expected) {
			t.Errorf("Field %s = %v, expected %v", path, value, expected)
			return false
		}
		return true
	}

	// Handle object root
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
		return false
	}

	value, ok := getJSONField(data, path)
	if !ok {
		t.Errorf("Field %s not found in response", path)
		return false
	}

	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", expected) {
		t.Errorf("Field %s = %v, expected %v", path, value, expected)
		return false
	}

	return true
}

// createTestConfig creates a test apiConfig
func createTestConfig(t *testing.T, queries *database.Queries) *apiConfig {
	t.Helper()

	return &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             queries,
		platform:       "dev",
		jwtSecret:      testJWTSecret,
		polkaKey:       testPolkaKey,
	}
}
